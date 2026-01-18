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
	return &RoundRobin{
		Pool: pool,
	}
}

func (r *RoundRobin) AddBackend(b *backend.Backend) {
	r.Pool.AddBackend(b)
}

func (r *RoundRobin) SetBackendStatus(uri *url.URL, alive bool) {
	r.Pool.SetBackendStatus(uri, alive)
}

func (r *RoundRobin) GetNextValidPeer() (*backend.Backend, error) {
	r.Pool.Mu.RLock()
	defer r.Pool.Mu.RUnlock()

	n := len(r.Pool.Backends)
	if n == 0 {
		return nil, errors.New("no backends available")
	}

	start := atomic.AddUint64(&r.Pool.Current, 1)

	for i := 0; i < n; i++ {
		idx := int((start + uint64(i)) % uint64(n))
		backend := r.Pool.Backends[idx]

		if backend.IsAlive() {
			return backend, nil
		}
	}

	return nil, errors.New("no alive backends")
}
