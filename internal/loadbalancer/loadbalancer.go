package loadbalancer

import (
	"net/url"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
)

type LoadBalancer interface {
	GetNextValidPeer() (*backend.Backend, error)
	AddBackend(backend *backend.Backend)
	SetBackendStatus(uri *url.URL, alive bool)
}
