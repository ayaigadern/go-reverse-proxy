package loadbalancer

import (
	"errors"
	"math"
	"net/url"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
	"github.com/ayaigadern/go-reverse-proxy/internal/serverpool"
)

type LeastConn struct {
	Pool *serverpool.ServerPool
}

func NewLeastConn(Pool *serverpool.ServerPool) *LeastConn {
	return &LeastConn{Pool: Pool}
}

func (l *LeastConn) AddBackend(b *backend.Backend) {
	l.Pool.AddBackend(b)
}

func (l *LeastConn) SetBackendStatus(uri *url.URL, alive bool) {
	l.Pool.SetBackendStatus(uri, alive)
}

func (l *LeastConn) GetNextValidPeer() (*backend.Backend, error) {
	backends := l.Pool.GetBackends()

	var selected *backend.Backend
	min := int64(math.MaxInt64)

	for _, b := range backends {
		if !b.IsAlive() {
			continue
		}

		if c := b.GetConns(); c < min {
			min = c
			selected = b
		}
	}

	if selected == nil {
		return nil, errors.New("no alive backends")
	}

	return selected, nil
}
