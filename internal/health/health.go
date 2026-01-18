package health

import (
	"log"
	"net/http"
	"time"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
	"github.com/ayaigadern/go-reverse-proxy/internal/serverpool"
)

// Health checker runs periodic health checks for all backends
type HealthChecker struct {
	Pool     *serverpool.ServerPool
	Interval time.Duration
	Client   *http.Client
}

// Create a new HealthChecker
func NewHealthChecker(pool *serverpool.ServerPool, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		Pool:     pool,
		Interval: interval,
		Client: &http.Client{
			Timeout: 2 * time.Second, // timeout for each ping
		},
	}
}

// Start launches the health check loop as a goroutine
func (h *HealthChecker) Start() {
	ticker := time.NewTicker(h.Interval)

	go func() {
		for range ticker.C {
			h.checkBackends()
		}
	}()
}

// checkBackends pings each backend and updates its Alive status
func (h *HealthChecker) checkBackends() {
	for _, b := range h.Pool.Backends {
		alive := h.pingBackend(b)
		if b.IsAlive() != alive {
			b.SetAlive(alive)
			status := "UP"
			if !alive {
				status = "DOWN"
			}
			log.Printf("Backend %s is %s\n", b.URL, status)
		}
	}
}

// pingBackend performs a simple GET request to check if backend is alive
func (h *HealthChecker) pingBackend(b *backend.Backend) bool {
	resp, err := h.Client.Get(b.URL.String())
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider backend alive if response status is 2xx or 3xx
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
