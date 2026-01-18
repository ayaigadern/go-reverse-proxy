package loadbalancer

import (
	"errors"
	"net/url"
	"sync/atomic"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
	"github.com/ayaigadern/go-reverse-proxy/internal/serverpool"
)

type RoundRobin struct {
	Pool *serverpool.ServerPool
}

func NewRoundRobin(pool *serverpool.ServerPool) *RoundRobin {
	return &RoundRobin{Pool: pool}
}

func (r *RoundRobin) AddBackend(b *backend.Backend) {
	r.Pool.AddBackend(b)
}

func (r *RoundRobin) SetBackendStatus(uri *url.URL, alive bool) {
	r.Pool.SetBackendStatus(uri, alive)
}

func (r *RoundRobin) GetNextValidPeer() (*backend.Backend, error) {
	n := len(r.Pool.Backends)
	if n == 0 {
		return nil, errors.New("no backends available")
	}

	// Try up to n backends to find one alive
	for i := 0; i < n; i++ {
		idx := int(atomic.AddUint64(&r.Pool.Current, 1)-1) % n
		b := r.Pool.Backends[idx]

		if b.IsAlive() {
			return b, nil
		}
	}

	return nil, errors.New("no alive backends")
}
