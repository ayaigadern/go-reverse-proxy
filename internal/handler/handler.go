package handler

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/ayaigadern/go-reverse-proxy/internal/loadbalancer"
)

func Handler(lb loadbalancer.LoadBalancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend, err := lb.GetNextValidPeer()
		if err != nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}

		backend.IncrementConns()
		defer backend.DecrementConns()

		proxy := httputil.NewSingleHostReverseProxy(backend.URL)

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			log.Printf("Error proxying to backend %s: %v", backend.URL, err)
			//mark dead only if network connection refused
			http.Error(rw, "backend error", http.StatusServiceUnavailable)
		}

		proxy.ServeHTTP(w, r)
	}
}
