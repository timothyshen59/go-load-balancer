package backend

import (
	"lb/internal/config"
	"lb/internal/health"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Backend with atomic fields
type Backend struct {
	URL           *url.URL
	alive         int32 // 0=dead, 1=alive (atomic)
	ReverseProxy  *httputil.ReverseProxy
	weight        int32
	currentWeight int64 // atomic
	failCount     int32

	//Stats
	totalRequests      int64
	failedRequests     int64
	successfulRequests int64
}

type ServerPool struct {
	backends []*Backend
	mux      sync.RWMutex // Only for slice modifications
	//Statistics

	totalRequests   int64
	healthyBackends int32
}

func NewServerPool() *ServerPool {
	return &ServerPool{}

}

func (p *ServerPool) AddBackend(b *Backend) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.backends = append(p.backends, b)
}

func (p *ServerPool) GetBackends() []*Backend {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.backends
}

func (b *Backend) GetBackendStats() (int64, int64) {
	success_count := atomic.LoadInt64(&b.successfulRequests)
	fail_count := atomic.LoadInt64(&b.failedRequests)

	return success_count, fail_count
}

func NewServerPoolFromConfig(cfg *config.Config) *ServerPool {
	pool := NewServerPool()

	for i, bcfg := range cfg.Backends {
		serverUrl, err := url.Parse(bcfg.URL)
		if err != nil {
			log.Fatalf("Invalid backend URL %s: %v", bcfg.URL, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		backend := &Backend{
			URL:           serverUrl,
			alive:         1,
			failCount:     0,
			ReverseProxy:  proxy,
			weight:        int32(bcfg.Weight),
			currentWeight: 0,
		}

		log.Printf("Backend %d: URL=%s, Weight=%d, Alive=%v",
			i, backend.URL, backend.weight, backend.alive)

		configureProxy(backend)
		pool.AddBackend(backend)
		log.Printf("Configured server[%d]: %s (weight=%d)", i, serverUrl, bcfg.Weight)
	}
	return pool

}

// SetAlive using atomic
func (b *Backend) SetAlive(alive bool) {
	if alive {
		atomic.StoreInt32(&b.alive, 1)
	} else {
		atomic.StoreInt32(&b.alive, 0)
	}
}

// IsAlive using atomic
func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, b := range s.backends {
		if b.URL == backendUrl {
			b.SetAlive(alive)
			break
		}
	}
}

// Lock-free GetNextPeer with atomic operations
func (s *ServerPool) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil
	}

	var best *Backend
	var bestWeight int64 = -1 << 62 // Very negative number
	var total int32

	// Lock-free reads!
	for _, backend := range s.backends {
		if atomic.LoadInt32(&backend.alive) == 0 {
			continue
		}

		w := atomic.LoadInt32(&backend.weight)
		newWeight := atomic.AddInt64(&backend.currentWeight, int64(w))
		total += w

		if newWeight > bestWeight {
			best = backend
			bestWeight = newWeight
		}
	}

	if best == nil {
		return nil
	}

	atomic.AddInt64(&s.totalRequests, 1)
	atomic.AddInt64(&best.totalRequests, 1)

	atomic.AddInt64(&best.currentWeight, -int64(total))
	return best
}

func (b *Backend) UpdateStats(statusCode int) {
	if statusCode >= 200 && statusCode < 400 {
		atomic.AddInt64(&b.successfulRequests, 1)

	} else {
		atomic.AddInt64(&b.failedRequests, 1)

	}
}

func (s *ServerPool) HealthCheck() {
	s.mux.RLock()
	defer s.mux.RUnlock()

	for _, b := range s.backends {
		status := "up"
		alive := health.IsBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func (s *ServerPool) PeriodicCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			log.Println("Starting health check...")
			s.HealthCheck()
			log.Println("Health check completed") // ← ADD THIS

		}
	}()
}

func configureProxy(backend *Backend) {
	backend.ReverseProxy.Transport = &http.Transport{
		MaxIdleConns:           5000,
		MaxIdleConnsPerHost:    2000,
		MaxConnsPerHost:        5000,
		IdleConnTimeout:        120 * time.Second,
		DisableCompression:     true,
		DisableKeepAlives:      false,
		MaxResponseHeaderBytes: 512 << 10,
	}

	// Move your retry logic here (needs context keys passed down)
	backend.ReverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[%s] %v", backend.URL.Host, err)

		fails := atomic.AddInt32(&backend.failCount, 1)
		if fails <= 2 {
			log.Printf("[%s] Transient retry %d/2", backend.URL.Host, fails)
			time.Sleep(50 * time.Millisecond)
			backend.ReverseProxy.ServeHTTP(w, r)
			return
		}

		log.Printf("[%s] Marked dead after %d fails", backend.URL.Host, fails)
		atomic.StoreInt32(&backend.alive, 0)
		atomic.StoreInt32(&backend.failCount, 0) // Reset for next life

		http.Error(w, "Backend unhealthy", http.StatusServiceUnavailable)
	}
}
