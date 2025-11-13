# Technical Architecture Index

**Version:** 1.5 **Date:** 2025-11-13

This section provides deep technical specifications for all components of Vibe Runner. Each document focuses on a specific architectural layer or concern.

## Architecture Documents

### [Frontend (Client): Pixi.js](./frontend-pixijs.md)
Complete client-side implementation guide covering game loop, client-side prediction, entity interpolation, physics, and performance optimization.

### [Backend (Server): Go](./backend-go.md)
Server architecture using Go and gorilla/websocket, including concurrency model, game tick implementation, deterministic procedural generation, and WebSocket handling.

### [Network Protocol: WebSockets (JSON)](./network-protocol.md)
Detailed specification of all client-server messages, connection flow, bandwidth optimization, and error handling for WebSocket communication.

### [Database Schema](./database-schema.md)
Complete database design for Redis (in-memory cache) and PostgreSQL (persistent storage) including leaderboard management, player sessions, and queries.

### [Developer Tooling](./developer-tooling.md)
Specifications for debug HUD, load testing script, and admin panel for development, testing, and monitoring.

### [Security](./security.md)
Comprehensive security measures including anti-cheat (server-side authority), input sanitization (XSS prevention), rate limiting, encryption, and deployment security checklist.

---

## Quick Reference

### Core Technologies

**Frontend:**
- Pixi.js (WebGL 2D renderer)
- JavaScript/TypeScript
- WebSocket client
- RequestAnimationFrame game loop

**Backend:**
- Go (Golang)
- gorilla/websocket
- Redis (in-memory cache)
- PostgreSQL (persistent storage)

**Protocol:**
- WebSocket (JSON messages)
- 20Hz server tick rate
- Client-side prediction
- Entity interpolation

---

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     CLIENTS (Browser)                    │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Pixi.js Game Loop (60 FPS)                     │   │
│  │  - Client-side prediction                        │   │
│  │  - Entity interpolation                          │   │
│  │  - WebSocket client                              │   │
│  └──────────────────┬──────────────────────────────┘   │
└─────────────────────┼──────────────────────────────────┘
                      │
                      │ WebSocket (WSS)
                      │ JSON Messages
                      │
┌─────────────────────▼──────────────────────────────────┐
│                   GO SERVER                             │
│  ┌─────────────────────────────────────────────────┐  │
│  │  Game Ticker (20Hz)                             │  │
│  │  - Process inputs                                │  │
│  │  - Run simulation                                │  │
│  │  - Collision detection                           │  │
│  │  - Broadcast state                               │  │
│  └──────────────────┬──────────────────────────────┘  │
│                     │                                   │
│  ┌─────────────────▼──────────────┐                   │
│  │  WebSocket Handler              │                   │
│  │  (gorilla/websocket)            │                   │
│  │  - One goroutine per client     │                   │
│  └─────────────────┬───────────────┘                   │
└────────────────────┼───────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
┌───────▼──────┐         ┌────────▼─────────┐
│    REDIS     │         │   POSTGRESQL     │
│              │         │                  │
│ - Leaderboard│         │ - Scores (all)   │
│ - Sessions   │         │ - Long-term data │
│ - Live data  │         │                  │
└──────────────┘         └──────────────────┘
```

---

## Key Design Decisions

### 1. Server-Authoritative Architecture

**Decision:** All game logic runs on the server. Client predictions are corrected by server state.

**Rationale:**
- Prevents cheating (collision detection can't be bypassed)
- Ensures fair gameplay for all players
- Simplifies client code (less complex state management)

**Trade-off:** Requires robust network handling and prediction/reconciliation logic.

---

### 2. 20Hz Server Tick Rate

**Decision:** Server updates game state 20 times per second (every 50ms).

**Rationale:**
- Balance between responsiveness and server load
- Adequate for 2D side-scroller physics
- Reduces bandwidth (vs 60Hz)
- Smoothed on client with interpolation

**Bandwidth Impact:** ~4 KB/sec per client for state updates.

---

### 3. Client-Side Prediction

**Decision:** Client simulates jump immediately, then reconciles with server state.

**Rationale:**
- Eliminates perceived input lag
- Makes game feel responsive despite network latency
- Standard approach in multiplayer action games

**Complexity:** Requires prediction, reconciliation, and correction logic.

---

### 4. Deterministic Procedural Generation

**Decision:** Level generated by server using seeded PRNG, sent to all clients.

**Rationale:**
- All players see identical obstacles (fair competition)
- Server controls obstacle placement (prevents cheating)
- Reduces bandwidth (vs sending raw data)
- Infinite level without pre-designed content

**Alternative Considered:** Client-side generation (rejected due to cheat potential).

---

### 5. Redis + PostgreSQL

**Decision:** Use Redis for real-time data, PostgreSQL for persistence.

**Rationale:**
- Redis: Fast reads/writes for live leaderboard (sub-millisecond)
- PostgreSQL: Reliable long-term storage for analytics
- Best of both worlds: speed + durability

**Trade-off:** Additional infrastructure complexity (two databases).

---

## Performance Targets

### Server Performance

- **Tick Duration:** < 10ms average (target: 5ms)
- **Concurrent Players:** 500+ simultaneous connections
- **Messages Per Second:** 10,000+ (20 state broadcasts × 500 players)
- **Memory Usage:** < 512MB for 500 players
- **CPU Usage:** < 50% on single core for 500 players

### Client Performance

- **Frame Rate:** 60 FPS stable
- **Input Lag:** < 50ms (perceived, with prediction)
- **Memory Usage:** < 256MB
- **Network Usage:** ~4-5 KB/sec downstream, ~0.2 KB/sec upstream

### Network Performance

- **Latency Tolerance:** Playable up to 200ms RTT
- **Packet Loss Tolerance:** Playable up to 5% loss
- **Bandwidth Per Client:** ~5 KB/sec total

---

## Scaling Strategy

### Vertical Scaling (Phase 1)
- Single server instance
- Target: 500 concurrent players
- Upgrade: CPU (4-8 cores), RAM (8-16GB)

### Horizontal Scaling (Future)
- Multiple game server instances
- Load balancer distributes players
- Shared Redis cluster for global leaderboard
- Target: 5,000+ concurrent players

### Regional Deployment (Future)
- Deploy servers in multiple regions (US, EU, Asia)
- Route players to nearest server
- Minimize latency for global audience

---

## Development Workflow

### Local Development

```bash
# Start Redis
redis-server

# Start PostgreSQL
pg_ctl start

# Start Go server
cd server
go run main.go

# Start frontend dev server
cd client
npm run dev
```

### Testing

```bash
# Run unit tests
go test ./...
npm test

# Run load test
./tools/load-test -c=100 -url=ws://localhost:8080/ws

# Manual testing with ?debug=true
http://localhost:3000/?debug=true
```

### Deployment

```bash
# Build frontend
cd client
npm run build

# Build backend
cd server
go build -o viberunner-server

# Deploy to server (example)
scp viberunner-server user@server:/opt/viberunner/
ssh user@server 'systemctl restart viberunner'
```

---

## Related Documentation

- Development Phases: docs/00-development-phases/index.md
- User Stories: docs/01-user-stories.md
- Product Requirements: docs/02-product-requirements.md
- Art Style: docs/03-art-style-aesthetics.md
