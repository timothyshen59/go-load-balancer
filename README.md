# go-load-balancer

Go-Load-Balancer is a toy load balancer written in Go, implementing weighted round-robin with connection pooling for efficient traffic distribution. It includes atomic stats tracking (success/failure rates per backend), Prometheus/Grafana observability, and health checks, all configurable via JSON/YAML

## Features

### Backend Health Checks
Continuous HTTP health monitoring removes unhealthy backends from rotation and automatically re-adds recovered instances, ensuring zero traffic to failed servers.

### Atomic Stats Tracking
Per-backend success/failure counters (200-399 = success) with global aggregation. Real-time success rates via Prometheus histograms at /metrics.

### Observability Integration
Production-ready Prometheus + Grafana metrics export: request latency, RPS, error rates, backend health status. Dockerized demo with live dashboards.

### Simple Configuration
JSON/YAML configs for backends, weights, health check intervals. Zero boilerplate—single docker compose up for full stack.
## Configuration









