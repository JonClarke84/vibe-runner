# 🌊 Vibe Runner

> An always-on, massively multiplayer 2D infinite runner with an explosive 80s synthwave aesthetic.

**Status:** 📋 Planning & Documentation Phase
**Current Development Phase:** Phase 1 (Local Prototype)

---

## 🎮 What is Vibe Runner?

Vibe Runner is a browser-based multiplayer infinite runner where **all players run the exact same level simultaneously**. It's not just about your personal best—it's about competing against hundreds of other players in real-time on a shared, deterministically generated track.

### Core Concept

- 🏃 **Single-Button Gameplay:** Jump to avoid obstacles
- 🌐 **Massive Multiplayer:** See other players as "ghost" silhouettes
- 🎲 **Fair Competition:** Everyone gets the same procedurally generated level
- ⚡ **Instant Respawn:** Die and jump back in immediately
- 🏆 **Live Leaderboard:** Real-time rankings of top survivors

### The Aesthetic

**"Hyper-Synthwave"** — An over-the-top, modern interpretation of the 80s' vision of the future:

- Neon pinks (#ff007f), electric cyans (#00f0ff), and phosphor green (#33ff00)
- Parallax scrolling backgrounds with OutRun-style sunset
- Glitched-out "firewall" obstacles with chromatic aberration
- 80s computer terminal UI with scan lines and CRT effects
- Retro pixel art meets modern high-res rendering

---

## ✨ Features

### MVP (Minimum Viable Product)

- **Splash Screen & Main Menu** — Terminal-style UI with `[SYSTEM BOOTING...]` aesthetic
- **Infinite Runner Gameplay** — Tight, responsive physics with single-button jump mechanic
- **Multiplayer Ghosts** — See all other active players with their names above them
- **Deterministic Level Generation** — Server-generated obstacles ensure fairness
- **Real-Time Leaderboard** — Top 10 players ranked by survival time
- **Death & Respawn** — Instant "FATAL ERROR" screen with score and "REBOOT" button

### Future Enhancements (Phase 6)

- Animated player sprites with run cycles
- Parallax background layers (cityscape, floating geometry, server racks)
- Particle effects and neon trails
- Soundtrack integration
- Developer debug HUD (`?debug=true`)
- Load testing tools for 500+ concurrent players

---

## 🏗️ Architecture

### High-Level Design

```
┌─────────────────┐
│  Browser        │  Pixi.js (WebGL)
│  60 FPS Client  │  Client-side prediction
└────────┬────────┘  Entity interpolation
         │
         │ WebSocket (JSON)
         │ 20Hz updates
         │
┌────────▼────────┐
│  Go Server      │  Server-authoritative
│  20Hz Ticker    │  Collision detection
└────────┬────────┘  Procedural generation
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼──────┐
│ Redis │ │ Postgres│
│ Live  │ │ History │
└───────┘ └─────────┘
```

### Key Principles

1. **Server-Authoritative** — Server validates all actions and calculates collisions. Clients request, server executes.
2. **Client-Side Prediction** — Jumps happen instantly on the client for responsive gameplay, then reconcile with server state.
3. **Deterministic Procedural Generation** — Seeded PRNG ensures all players see identical obstacles.
4. **Entity Interpolation** — Other players' movements smoothly interpolated from 20Hz server updates to 60 FPS.

---

## 🛠️ Tech Stack

### Frontend
- **Pixi.js** — 2D WebGL rendering engine
- **JavaScript/TypeScript** — Game logic
- **WebSocket** — Real-time client-server communication

### Backend
- **Go** — High-performance server
- **gorilla/websocket** — WebSocket implementation
- **Redis** — In-memory cache for live leaderboard and sessions
- **PostgreSQL** — Persistent storage for all-time scores

### Infrastructure
- **WebSocket Protocol** — JSON messages with short keys (`e`, `d`)
- **20Hz Server Tick** — Game loop updates 20 times per second
- **WSS (WebSocket Secure)** — Encrypted communication in production

---

## 📚 Documentation

Comprehensive documentation is available in the [`docs/`](./docs/) directory:

### Quick Links

- **[Documentation Index](./docs/index.md)** — Overview of all documentation
- **[Development Phases](./docs/00-development-phases/)** — Six-phase roadmap from prototype to production
- **[Technical Architecture](./docs/04-technical-architecture/)** — Deep dive into frontend, backend, networking, database, security
- **[Art Style Guide](./docs/03-art-style-aesthetics.md)** — Complete visual identity specifications
- **[Product Requirements](./docs/02-product-requirements.md)** — Core features and MVP definition

---

## 🚀 Development Phases

Vibe Runner is being built in six incremental phases to ensure stability and focus:

### Phase 1: Core Local Game *(Current)*
Build a single-player prototype with basic physics, jump mechanics, and collision detection.

### Phase 2: Server & WebSocket
Connect client to Go server with real-time communication and server-controlled gameplay.

### Phase 3: Multiplayer & Prediction
Add client-side prediction for responsive input and multiplayer "ghost" players.

### Phase 4: Procedural Generation
Implement deterministic level generation using seeded PRNG on the server.

### Phase 5: Full Game Loop
Complete the UX with main menu, death/respawn cycle, and leaderboard system.

### Phase 6: Polish & Tooling
Add visual effects, parallax backgrounds, audio, and developer tools.

📖 **See [Development Phases](./docs/00-development-phases/)** for detailed task breakdowns.

---

## 🏁 Getting Started

> **Note:** This project is currently in the planning phase. No implementation exists yet.

### Prerequisites (Future)

- **Node.js** (v18+) — For frontend development
- **Go** (v1.21+) — For backend server
- **Redis** — In-memory cache
- **PostgreSQL** — Persistent storage

### Development Setup (Future)

Once implementation begins:

```bash
# Clone the repository
git clone https://github.com/yourusername/vibe-runner.git
cd vibe-runner

# Start Redis
redis-server

# Start PostgreSQL
pg_ctl start

# Backend (Go server)
cd server
go run main.go

# Frontend (Pixi.js client)
cd client
npm install
npm run dev
```

Open your browser to `http://localhost:3000`

For debug mode: `http://localhost:3000/?debug=true`

---

## 🎯 Current Status

**Project Status:** Documentation Complete, Implementation Not Started

### Completed
- ✅ Complete documentation (19 files, 4000+ lines)
- ✅ Technical architecture specifications
- ✅ Six-phase development roadmap
- ✅ Art style and visual identity guide
- ✅ Network protocol specification
- ✅ Database schema design
- ✅ Security specifications
- ✅ Developer tooling requirements

### Next Steps
- [ ] Set up project structure (client/ and server/ directories)
- [ ] Initialize frontend with Pixi.js
- [ ] Implement Phase 1: Core local game prototype
- [ ] Create placeholder assets (player sprite, ground tile, obstacles)

---

## 🤝 Contributing

This is currently a personal project in active development. Contributions, suggestions, and feedback are welcome!

### Development Guidelines

1. **Follow the phased approach** — Don't skip ahead. Each phase must be complete and stable.
2. **Reference documentation** — All architectural decisions are documented in `docs/`.
3. **Server authority** — The server validates everything. Clients request, never command.
4. **Security first** — Sanitize inputs, rate limit connections, validate all client messages.

See [CLAUDE.md](./CLAUDE.md) for detailed development context.

---

## 📄 License

*License to be determined*

---

## 🙏 Acknowledgments

Inspired by:
- **Far Cry 3: Blood Dragon** — Satirical 80s tone
- **Katana ZERO** & **Hyper Light Drifter** — Modern "hi-bit" pixel art
- **Hotline Miami** — Color palette and grit
- **Alien: Isolation** — CRT computer UI

---

## 📞 Contact

*Contact information to be added*

---

<div align="center">

**[Documentation](./docs/index.md)** • **[Architecture](./docs/04-technical-architecture/)** • **[Phases](./docs/00-development-phases/)**

Made with 💖 and excessive amounts of neon

</div>
