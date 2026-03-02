package stats

import (
	"lb/internal/backend"
)

type Stats struct {
	Global   GlobalStats    `json:"global"`
	Backends []BackendStats `json:"backends"`
}

type GlobalStats struct {
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
}

type BackendStats struct {
	URL                string  `json:"url"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"sucess_rate"`
}

func GetStats(pool *backend.ServerPool) Stats {
	var totalSuccess, totalFailed int64
	lb_backends := pool.GetBackends()
	backends := make([]BackendStats, len(lb_backends))

	for i, b := range lb_backends {
		success, failed := b.GetBackendStats()

		backends[i] = BackendStats{
			URL:                b.URL.String(),
			SuccessfulRequests: success,
			FailedRequests:     failed,
			SuccessRate:        successRate(success, failed),
		}
		totalSuccess += success
		totalFailed += failed
	}

	return Stats{
		Global: GlobalStats{
			TotalRequests:      totalSuccess + totalFailed,
			SuccessfulRequests: totalSuccess,
			FailedRequests:     totalFailed,
			SuccessRate:        successRate(totalSuccess, totalFailed),
		},
		Backends: backends,
	}
}

func successRate(success, failed int64) float64 {
	total := success + failed
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total) * 100
}
