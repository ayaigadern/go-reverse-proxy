package backend

import (
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL `json:"url"`
	Alive        bool     `json:"alive"`
	CurrentConns int64    `json:"current_connections"`

	mux sync.RWMutex
}

// Constructor
func NewBackend(rawURL string) (*Backend, error) {
	pasedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse backend URL %q: %w", rawURL, err)
	}
	return &Backend{
		URL:          pasedURL,
		Alive:        true,
		CurrentConns: 0,
	}, nil
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive

}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func (b *Backend) IncrementConns() {
	atomic.AddInt64(&b.CurrentConns, 1)
}

func (b *Backend) DecrementConns() {
	atomic.AddInt64(&b.CurrentConns, -1)
}

func (b *Backend) GetConns() int64 {
	return int64(atomic.LoadInt64(&b.CurrentConns))
}
