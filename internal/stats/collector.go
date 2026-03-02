package stats

import (
	"lb/internal/backend"

	"github.com/prometheus/client_golang/prometheus"
)

type LBCollector struct {
	pool *backend.ServerPool

	globalTotal       *prometheus.Desc
	globalSuccess     *prometheus.Desc
	globalFailed      *prometheus.Desc
	globalSuccessRate *prometheus.Desc

	backendTotal       *prometheus.Desc
	backendSuccess     *prometheus.Desc
	backendFailed      *prometheus.Desc
	backendSuccessRate *prometheus.Desc

	//Latency Metrics
	RequestDuration *prometheus.HistogramVec
	BackendDuration *prometheus.HistogramVec
}

func NewLBCollector(pool *backend.ServerPool) *LBCollector {
	c := &LBCollector{
		pool: pool,
		globalTotal: prometheus.NewDesc(
			"lb_requests_total",
			"Total number of requests", nil, nil,
		),
		globalSuccess: prometheus.NewDesc(
			"lb_successful_requests_total",
			"Total successful requests", nil, nil,
		),
		globalFailed: prometheus.NewDesc(
			"lb_failed_requests_total",
			"Total failed requests", nil, nil,
		),
		globalSuccessRate: prometheus.NewDesc(
			"lb_success_rate",
			"Global success rate as a percentage", nil, nil,
		),
		backendTotal: prometheus.NewDesc(
			"backend_total_requests",
			"Total Requests Sent To Specific Backend",
			[]string{"backend"}, nil,
		),
		backendSuccess: prometheus.NewDesc(
			"backend_successful_requests_total",
			"Successful requests per backend",
			[]string{"backend"}, nil,
		),
		backendFailed: prometheus.NewDesc(
			"backend_failed_requests_total",
			"Failed requests per backend",
			[]string{"backend"}, nil,
		),
		backendSuccessRate: prometheus.NewDesc(
			"backend_success_rate",
			"Success rate per backend as a percentage",
			[]string{"backend"}, nil,
		),

		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "lb_request_duration_seconds",
			Help:    "Duration of LB Requests",
			Buckets: prometheus.DefBuckets,
		}, []string{"status_code"}),

		BackendDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "backend_proxy_duration_seconds",
			Help:    "Duration of proxying to backend",
			Buckets: prometheus.DefBuckets,
		}, []string{"backend", "status_code"}),
	}

	return c
}

func (c *LBCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.globalTotal
	ch <- c.globalFailed
	ch <- c.globalSuccess
	ch <- c.globalSuccessRate

	ch <- c.backendTotal
	ch <- c.backendFailed
	ch <- c.backendSuccess
	ch <- c.backendSuccessRate

	c.RequestDuration.Describe(ch)
	c.BackendDuration.Describe(ch)

}

func (c *LBCollector) Collect(ch chan<- prometheus.Metric) {
	s := GetStats(c.pool)

	ch <- prometheus.MustNewConstMetric(c.globalTotal, prometheus.CounterValue, float64(s.Global.TotalRequests))
	ch <- prometheus.MustNewConstMetric(c.globalSuccess, prometheus.CounterValue, float64(s.Global.SuccessfulRequests))
	ch <- prometheus.MustNewConstMetric(c.globalFailed, prometheus.CounterValue, float64(s.Global.FailedRequests))
	ch <- prometheus.MustNewConstMetric(c.globalSuccessRate, prometheus.GaugeValue, s.Global.SuccessRate)

	for _, b := range s.Backends {
		ch <- prometheus.MustNewConstMetric(c.backendSuccess, prometheus.CounterValue, float64(b.SuccessfulRequests), b.URL)
		ch <- prometheus.MustNewConstMetric(c.backendFailed, prometheus.CounterValue, float64(b.FailedRequests), b.URL)
		ch <- prometheus.MustNewConstMetric(c.backendSuccessRate, prometheus.GaugeValue, b.SuccessRate, b.URL)

	}
	c.RequestDuration.Collect(ch)
	c.BackendDuration.Collect(ch)

}
