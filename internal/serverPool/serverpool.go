package serverpool

import (
	"net/url"
	"sync"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
)

type ServerPool struct {
	Backends []*backend.Backend `json:"backends"`
	Current  uint64             `json:"current"`

	Mu sync.RWMutex
}

func NewServerpool() *ServerPool {
	return &ServerPool{
		Backends: make([]*backend.Backend, 0),
	}
}

func (s *ServerPool) AddBackend(b *backend.Backend) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Backends = append(s.Backends, b)

}

func (s *ServerPool) RemoveBackend(rawURL string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for idx, backend := range s.Backends {
		if backend.URL.String() == rawURL {
			s.Backends = append(s.Backends[:idx], s.Backends[idx+1:]...)
			return
		}

	}
}

func (s *ServerPool) SetBackendStatus(uri *url.URL, alive bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for _, backend := range s.Backends {
		if backend.URL.String() == uri.String() {
			backend.SetAlive(alive)
			return
		}
	}

}

func (s *ServerPool) GetBackends() []*backend.Backend {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	backends := make([]*backend.Backend, len(s.Backends))
	copy(backends, s.Backends)
	return backends
}
