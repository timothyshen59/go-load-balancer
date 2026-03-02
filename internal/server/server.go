package server

import (
	// 👈 ADD THIS for r.Context()
	"context"
	"encoding/json"
	"lb/internal/backend"
	"lb/internal/balancer"
	"lb/internal/stats"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ctxKey struct{}

var (
	attemptKey = ctxKey{}
	retryKey   = ctxKey{}
)

type Server struct {
	balancer  balancer.Balancer
	port      string
	pool      *backend.ServerPool
	collector *stats.LBCollector
}

func InitializeServer(port string, bal balancer.Balancer, pool *backend.ServerPool) *Server {
	return &Server{
		balancer:  bal,
		port:      port,
		pool:      pool,
		collector: stats.NewLBCollector(pool),
	}
}

// Helper Context Methods
func (s *Server) GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(attemptKey).(int); ok {
		return attempts
	}
	return 1
}

func (s *Server) GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(retryKey).(int); ok {
		return retry
	}
	return 0
}

// Calrify how this key works and const.

type statusCapturingWriter struct {
	http.ResponseWriter
	statusCode int
}

func (scw *statusCapturingWriter) WriteHeader(code int) {
	scw.statusCode = code
	scw.ResponseWriter.WriteHeader(code)

}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	result := stats.GetStats(s.pool)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) recordLatency(backendHost string, statusCode int, totalDuration, backendDuration float64) {
	statusStr := strconv.Itoa(statusCode)
	s.collector.RequestDuration.WithLabelValues(statusStr).Observe(totalDuration)

	if backendHost != "" && statusCode < 500 {
		s.collector.BackendDuration.WithLabelValues(backendHost, statusStr).Observe(backendDuration)
	}
}

func (s *Server) handleLBRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	statsWriter := &statusCapturingWriter{
		ResponseWriter: w,
		statusCode:     http.StatusServiceUnavailable,
	}

	attempts := s.GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		s.recordLatency("", statsWriter.statusCode, time.Since(start).Seconds(), 0)
		return
	}

	peer := s.balancer.Select()

	if peer != nil {
		backendStart := time.Now()
		peer.ReverseProxy.ServeHTTP(statsWriter, r)

		backendTime := time.Since(backendStart).Seconds()
		totalTime := time.Since(start).Seconds()

		s.recordLatency(peer.URL.Host, statsWriter.statusCode, totalTime, backendTime)
		peer.UpdateStats(statsWriter.statusCode)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
	s.recordLatency("", statsWriter.statusCode, time.Since(start).Seconds(), 0)

}

func (s *Server) Start() {

	//Prometheus Metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(s.collector)

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/stats", s.statsHandler)
	mux.HandleFunc("/", s.handleLBRequest)
	server := &http.Server{
		Addr:         s.port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,

		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	log.Printf("Server starting on %s", s.port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

}
