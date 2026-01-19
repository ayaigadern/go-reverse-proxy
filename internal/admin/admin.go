package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ayaigadern/go-reverse-proxy/internal/backend"
	"github.com/ayaigadern/go-reverse-proxy/internal/serverpool"
)

type AdminServer struct {
	Pool *serverpool.ServerPool
}

func NewAdmin(pool *serverpool.ServerPool) *AdminServer {
	return &AdminServer{Pool: pool}
}

func (a *AdminServer) Start(port int) {
	muxServer := http.NewServeMux()
	muxServer.HandleFunc("/status", a.StatusHandler)
	muxServer.HandleFunc("/backends", a.BackendsRouter)

	addr := ":" + fmt.Sprint(port)
	log.Printf("Admin API running on port %d\n", port)
	go func() {
		if err := http.ListenAndServe(addr, muxServer); err != nil {
			log.Fatal("Admin API error:", err)
		}
	}()
}

// GET /status
func (a *AdminServer) StatusHandler(w http.ResponseWriter, r *http.Request) {
	a.Pool.Mu.RLock()
	defer a.Pool.Mu.RUnlock()

	type backendInfo struct {
		URL          string `json:"url"`
		Alive        bool   `json:"alive"`
		CurrentConns int64  `json:"current_connections"`
	}

	status := make([]backendInfo, len(a.Pool.Backends))
	for i, b := range a.Pool.Backends {
		status[i] = backendInfo{
			URL:          b.URL.String(),
			Alive:        b.IsAlive(),
			CurrentConns: b.GetConns(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Delete and Add
func (a *AdminServer) BackendsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.AddBackendHandler(w, r)
	case http.MethodDelete:
		a.RemoveBackendHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *AdminServer) AddBackendHandler(w http.ResponseWriter, r *http.Request) {
	var urlLoad struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&urlLoad); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	urlLoad.URL = strings.TrimSpace(urlLoad.URL)

	b, err := backend.NewBackend(urlLoad.URL)
	if err != nil {
		http.Error(w, "invalid backend URL", http.StatusBadRequest)
		return
	}

	a.Pool.AddBackend(b)
	w.WriteHeader(http.StatusCreated)
}

func (a *AdminServer) RemoveBackendHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	a.Pool.RemoveBackend(payload.URL)
	w.WriteHeader(http.StatusOK)
}
