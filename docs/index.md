# Vibe Runner Documentation Index

**Version:** 1.5 **Date:** 2025-11-13

This index provides quick access to all project documentation. Each document is self-contained and focuses on a specific aspect of the Vibe Runner project.

## Documentation Structure

### [00-development-phases/](./00-development-phases/)
Six-phase development plan from local single-player prototype to full multiplayer game with polish and tooling.
- [Phase 1: Core Local Game](./00-development-phases/phase-1-core-local-game.md)
- [Phase 2: Basic Server & Client Connection](./00-development-phases/phase-2-server-connection.md)
- [Phase 3: Client-Side Prediction & Multiplayer Ghosts](./00-development-phases/phase-3-prediction-ghosts.md)
- [Phase 4: Deterministic Procedural Generation](./00-development-phases/phase-4-procedural-generation.md)
- [Phase 5: The Full Game Loop](./00-development-phases/phase-5-full-game-loop.md)
- [Phase 6: Polish & Tooling](./00-development-phases/phase-6-polish-tooling.md)

### [01-user-stories.md](./01-user-stories.md)
User stories organized by epic covering onboarding, core gameplay, multiplayer features, and developer tools.

### [02-product-requirements.md](./02-product-requirements.md)
Complete product requirements including game overview, core features, and developer tooling specifications.

### [03-art-style-aesthetics.md](./03-art-style-aesthetics.md)
Comprehensive visual identity guide defining the "Hyper-Synthwave" aesthetic with color palettes, character designs, and UI specifications.

### [04-technical-architecture/](./04-technical-architecture/)
Deep technical specifications covering all system components and implementation details.
- [Frontend (Client): Pixi.js](./04-technical-architecture/frontend-pixijs.md)
- [Backend (Server): Go](./04-technical-architecture/backend-go.md)
- [Network Protocol: WebSockets (JSON)](./04-technical-architecture/network-protocol.md)
- [Database Schema](./04-technical-architecture/database-schema.md)
- [Developer Tooling](./04-technical-architecture/developer-tooling.md)
- [Security](./04-technical-architecture/security.md)

---

## Quick Reference

**Current Phase:** Phase 1 (Core Local Game)

**Tech Stack:**
- Frontend: Pixi.js (WebGL), HTML/CSS
- Backend: Go with gorilla/websocket
- Database: Redis (leaderboard), PostgreSQL (persistence)

**Key Concepts:**
- Client-side prediction for responsive gameplay
- Server-authoritative collision detection
- Deterministic procedural generation with seeded PRNG
- 20Hz server tick rate with entity interpolation
