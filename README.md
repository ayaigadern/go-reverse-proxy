# go-reverse-proxy

A concurrent, load-balancing reverse proxy written in Go. It distributes incoming HTTP traffic across a pool of backend servers using configurable strategies, monitors backend health in the background, and exposes an admin API for dynamic management.

---

## Features

- **Load Balancing** — Round-Robin and Least-Connections strategies
- **Health Monitoring** — periodic background health checks; unhealthy backends are automatically removed from rotation
- **Immediate failure detection** — if a backend refuses a connection, it is marked DOWN right away (not just at the next health check)
- **Context propagation** — each proxied request carries a timeout context; client disconnections cancel backend requests automatically
- **Admin API** — add, remove, and inspect backends at runtime without restarting the proxy

---

## Project Structure

```
.
├── cmd/
│   ├── proxy/          # Main proxy entrypoint
│   └── testbackend/    # Simple test backend server
├── internal/
│   ├── admin/          # Admin HTTP API (port 8081)
│   ├── backend/        # Backend struct and connection tracking
│   ├── handler/        # Reverse proxy HTTP handler with context & error handling
│   ├── health/         # Periodic health checker
│   ├── loadbalancer/   # LoadBalancer interface, RoundRobin, LeastConn
│   ├── proxyconfig/    # Config file loader
│   └── serverpool/     # Thread-safe pool of backends
├── config.json         # Default configuration
└── go.mod
```

---

## Configuration

Edit `config.json` before starting:

```json
{
  "port": 8080,
  "strategy": "round-robin",
  "health_check_frequency": "15s",
  "backends": [
    "http://localhost:8082",
    "http://localhost:8083",
    "http://localhost:8084"
  ]
}
```

| Field | Description |
|---|---|
| `port` | Port the proxy listens on |
| `strategy` | `"round-robin"` or `"least-connections"` |
| `health_check_frequency` | Go duration string e.g. `"15s"`, `"1m"` |
| `backends` | Initial list of backend URLs |

---

## Running the Proxy

```bash
# Start the proxy (reads config.json from the current directory)
go run ./cmd/proxy

# Start test backends on ports 8082, 8083, 8084
PORT=8082 go run ./cmd/testbackend &
PORT=8083 go run ./cmd/testbackend &
PORT=8084 go run ./cmd/testbackend &

# Send a request through the proxy
curl http://localhost:8080/
```

---

## Admin API (port 8081)

### Check Status

```bash
curl http://localhost:8081/status
```

```json
{
  "active_backends": 2,
  "backends": [
    { "url": "http://localhost:8082", "alive": true,  "current_connections": 3 },
    { "url": "http://localhost:8083", "alive": false, "current_connections": 0 },
    { "url": "http://localhost:8084", "alive": true,  "current_connections": 1 }
  ],
  "total_backends": 3
}
```

### Add a Backend

```bash
curl -X POST http://localhost:8081/backends \
  -H "Content-Type: application/json" \
  -d '{"url": "http://localhost:8085"}'
```

### Remove a Backend

```bash
curl -X DELETE http://localhost:8081/backends \
  -H "Content-Type: application/json" \
  -d '{"url": "http://localhost:8082"}'
```

---

## Concurrency & Safety

- Backend `Alive` status is protected by `sync.RWMutex`
- Round-Robin counter uses `sync/atomic` for lock-free increments
- Connection counts use `atomic.AddInt64` for accurate per-backend tracking
- The server pool uses `sync.RWMutex` for safe concurrent reads and writes

---

## Context & Timeouts

Every proxied request is given a 30-second deadline via `context.WithTimeout`. If the **client disconnects** before the backend responds, the request context is cancelled automatically and the backend request is aborted — preventing wasted goroutines and connections.

Connection errors (refused, no such host, EOF) immediately mark the backend as DOWN so no further traffic is routed to it until the health checker confirms it is back up.