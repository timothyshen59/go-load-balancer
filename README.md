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

### YAML
```yaml
port: ":3030"
health_check_interval: 1
backends:
  - url: "http://backend1:8081"
    weight: 3
    health_path: "/health"
  - url: "http://backend2:8082"
    weight: 2
    health_path: "/health"
```
## Getting Started 
Clone the repo and spin up the load balancer + backends with Docker 
```bash
git clone https://github.com/timothyshen59/go-load-balancer.git
cd go-load-balancer
docker compose up -d
```

### Access Points
* Load Balancer: ```http://localhost:8080```

* Metrics: ```http://localhost:3030/metrics```

* Grafana: ```http://localhost:3000``` (admin/admin)

* Health: ```curl http://localhost:8080/health```

### Customize Config
Edit ```config.yaml```
```yaml
backends:
  - url: "http://backend1:8081"
    weight: 3
    health_check: "/health" 
  - url: "http://backend2:8082"
    weight: 1
```
To Restart: ```docker compose down && docker compose up -d```

### Benchmark
You can install Hey to benchmark the load balancer 
```bash 
brew install hey 
hey -n 10000 -c 100 http://localhost:3030
```







