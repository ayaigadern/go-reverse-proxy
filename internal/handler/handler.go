package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/ayaigadern/go-reverse-proxy/internal/loadbalancer"
)

const BackendTimeout = 30 * time.Second

func Handler(lb loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend, err := lb.GetNextValidPeer()
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		backend.IncrementConns()
		defer backend.DecrementConns()

		ctx, cancel := context.WithTimeout(r.Context(), BackendTimeout)
		defer cancel()
		r = r.WithContext(ctx)

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		backendURL := backend.URL

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Printf("Error proxying to backend %s: %v", backendURL, err)

			if isConnectionError(err) {
				log.Printf("Marking backend %s as DOWN due to connection error", backendURL)
				lb.SetBackendStatus(backendURL, false)
			}

			http.Error(rw, "backend error", http.StatusServiceUnavailable)
		}

		proxy.ServeHTTP(w, r)
	}
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "EOF")
}
