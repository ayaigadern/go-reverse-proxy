package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ayaigadern/go-reverse-proxy/internal/admin"
	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
	"github.com/ayaigadern/go-reverse-proxy/internal/handler"
	"github.com/ayaigadern/go-reverse-proxy/internal/health"
	"github.com/ayaigadern/go-reverse-proxy/internal/loadbalancer"
	"github.com/ayaigadern/go-reverse-proxy/internal/proxyconfig"
	"github.com/ayaigadern/go-reverse-proxy/internal/serverpool"
)

func main() {
	// Load config.json
	cfg, err := proxyconfig.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize ServerPool
	pool := serverpool.NewServerpool()

	// Add backends dynamically
	for _, urlStr := range cfg.Backends {
		b, err := backend.NewBackend(urlStr)
		if err != nil {
			log.Println("Skipping invalid backend:", err)
			continue
		}
		pool.AddBackend(b)
	}
	// Start the health checker
	hc := health.NewHealthChecker(pool, cfg.HealthCheckFreq)
	hc.Start() //
	// Initialize LoadBalancer
	var lb loadbalancer.LoadBalancer
	switch cfg.Strategy {
	case "least-connections":
		lb = loadbalancer.NewLeastConn(pool)
	default: // round-robin
		lb = loadbalancer.NewRoundRobin(pool)
	}
	// Start admin api
	adminServer := admin.NewAdmin(pool)
	adminServer.Start(8081)

	// Setup HTTP handler
	http.HandleFunc("/", handler.Handler(lb))

	log.Printf("Proxy running on port %d using %s strategy\n", cfg.Port, cfg.Strategy)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
