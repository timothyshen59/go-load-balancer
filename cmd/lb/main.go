package main

import (
	"flag"
	"lb/internal/backend"
	"lb/internal/balancer"
	"lb/internal/config"
	"lb/internal/server"
	"log"
	"time"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Config File (YAML/JSON)")
	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Config error (%s): %v", *configFile, err)
	}

	pool := backend.NewServerPoolFromConfig(cfg)
	pool.PeriodicCheck(time.Duration(cfg.HealthCheckInterval) * time.Minute)
	bal := balancer.NewSmoothWRR(pool)

	server.InitializeServer(cfg.Port, bal, pool).Start()
}
