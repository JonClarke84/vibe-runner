# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Vibe Runner** is a massively multiplayer 2D infinite runner where all players run the exact same deterministically procedurally generated level simultaneously. The game features an 80s synthwave aesthetic with server-authoritative gameplay, client-side prediction, and real-time multiplayer.

**Current Status:** Pre-implementation. All documentation is complete; no code has been written yet.

**Current Phase:** Phase 1 - Core Local Game (Client-Only)

## Documentation Structure

The project follows a comprehensive, phased development approach. **All architectural decisions, specifications, and requirements are documented in `docs/`**. Always reference these documents before implementing features.

### Critical Documentation

- **`docs/00-development-phases/`** - Six-phase roadmap from local prototype to production
  - Start with `phase-1-core-local-game.md` for current tasks and success criteria
  - Each phase document includes: Goal, Tasks, Required Assets, Success Criteria, Technical Notes

- **`docs/04-technical-architecture/`** - Complete implementation specifications
  - `frontend-pixijs.md` - Game loop, client-side prediction, entity interpolation
  - `backend-go.md` - Server concurrency model, game ticker, PRNG
  - `network-protocol.md` - All WebSocket message specifications
  - `database-schema.md` - Redis and PostgreSQL schemas
  - `security.md` - Anti-cheat, XSS prevention, rate limiting
  - `developer-tooling.md` - Debug HUD, load testing, admin panel

- **`docs/03-art-style-aesthetics.md`** - Complete visual identity (colors, fonts, effects)
- **`docs/02-product-requirements.md`** - Core features and MVP requirements

## Architecture at a Glance

### Three-Tier System

```
Browser (Pixi.js) ←→ Go Server (20Hz) ←→ Redis + PostgreSQL
  60 FPS client         WebSocket           Leaderboard + Storage
```

### Key Architectural Principles

1. **Server-Authoritative**: Server validates all actions and calculates all collision detection. Clients request, server executes.

2. **Client-Side Prediction**: Client applies jump immediately for instant feedback, then reconciles with server state. This eliminates perceived input lag.

3. **Deterministic Procedural Generation**: Level chunks generated server-side using seeded PRNG (hash(masterSeed + chunkID)). All players see identical obstacles.

4. **Entity Interpolation**: Ghost players (other players) interpolate between 20Hz server updates to appear smooth at 60 FPS client-side.

### Network Protocol

WebSocket communication using JSON with short keys (`e` = event, `d` = data):

**Client → Server:**
- `{"e": "join", "d": {"n": "PlayerName"}}`
- `{"e": "jump", "d": {"t": timestamp}}`

**Server → Client:**
- `{"e": "welcome", "d": {"id": 12345, "seed": "...", "serverTime": ...}}`
- `{"e": "state", "d": {"t": timestamp, "p": [{"i": id, "x": x, "y": y}, ...]}}` (20Hz)
- `{"e": "death", "d": {"s": score}}`
- `{"e": "chunk", "d": {"id": 10, "obs": [{"t": type, "x": x}, ...]}}`

## Tech Stack

### Frontend (Not Yet Implemented)
- **Pixi.js** - 2D WebGL rendering
- **JavaScript/TypeScript** - Game logic
- **requestAnimationFrame** - 60 FPS game loop
- Physics constants: `GRAVITY = 1200`, `JUMP_VELOCITY = -600`, `PLAYER_SPEED = 300`, `GROUND_Y = 500`

### Backend (Not Yet Implemented)
- **Go + gorilla/websocket** - Server with one goroutine per client
- **20Hz game ticker** (50ms per tick) - Central game loop
- **Redis** - Live leaderboard (sorted sets), player sessions (hashes)
- **PostgreSQL** - Persistent score history

### Future Structure (When Implemented)

```
vibe-runner/
├── client/                    # Frontend (Pixi.js)
│   ├── src/
│   │   ├── game/             # Game loop, physics, entities
│   │   ├── network/          # WebSocket client
│   │   ├── rendering/        # Pixi.js rendering
│   │   └── ui/               # Menus, HUD, leaderboard
│   ├── assets/               # SVG/PNG sprites, fonts, audio
│   └── public/               # index.html
│
├── server/                    # Backend (Go)
│   ├── main.go               # Entry point
│   ├── game/                 # Game state, physics, collision
│   ├── network/              # WebSocket handlers
│   ├── generation/           # Procedural level generation
│   └── database/             # Redis + PostgreSQL clients
│
├── tools/
│   └── load-test/            # Load testing script
│
└── docs/                      # All specifications (already exists)
```

## Development Workflow

### Phase-Based Development

This project **must** be built in phases. Do not skip ahead. Each phase should be fully functional and tested before proceeding.

**Phase 1 (Current):** Local single-player game (no server)
- HTML + Pixi.js canvas
- Player class with jump physics
- Ground collision detection
- Static obstacles with AABB collision
- Simple death state

**Subsequent Phases:**
- Phase 2: Add Go server with WebSocket (server-controlled movement)
- Phase 3: Client-side prediction + multiplayer ghosts
- Phase 4: Procedural generation with seeded PRNG
- Phase 5: Full game loop (menu, death/respawn, leaderboard)
- Phase 6: Polish (parallax, shaders, audio, debug tools)

### Before Implementing Any Feature

1. Read the current phase document in `docs/00-development-phases/`
2. Check technical specs in `docs/04-technical-architecture/`
3. Verify visual requirements in `docs/03-art-style-aesthetics.md`
4. Implement following the documented patterns
5. Test against phase success criteria

### Important Constants

**Color Palette** (from `docs/03-art-style-aesthetics.md`):
- Electric Pink: `#ff007f` (primary neon)
- Hyper-Cyan: `#00f0ff` (primary neon)
- Phosphor Green: `#33ff00` (UI only)
- Glitch Red: `#ff003c` (obstacles)
- Deep Indigo: `#1a1a2e` (base)
- Dark Purple: `#301a4b` (base)

**Network**:
- Server tick rate: 20Hz (50ms)
- Target client FPS: 60
- Bandwidth per client: ~4-5 KB/sec

## Security Requirements

All implementations must follow these security principles (detailed in `docs/04-technical-architecture/security.md`):

1. **Server validates all inputs** - Jump requests, player names, all client messages
2. **Sanitize player names** - Strip HTML tags, escape entities, max 30 chars
3. **Rate limiting** - Connection attempts (10/min per IP) and messages (100/sec per client)
4. **Server-only collision** - Clients cannot manipulate collision detection
5. **WSS in production** - WebSocket Secure with TLS

## Common Patterns

### AABB Collision Detection
```javascript
function checkCollision(rect1, rect2) {
  return (
    rect1.x < rect2.x + rect2.width &&
    rect1.x + rect1.width > rect2.x &&
    rect1.y < rect2.y + rect2.height &&
    rect1.y + rect1.height > rect2.y
  );
}
```

### Input Sanitization (Go)
```go
func sanitizePlayerName(name string) string {
    name = strings.TrimSpace(name)
    if len(name) > 30 { name = name[:30] }
    name = stripHTMLTags(name)
    name = html.EscapeString(name)
    if name == "" { name = "Player" }
    return name
}
```

### PRNG Chunk Generation (Go)
```go
func generateChunk(masterSeed string, chunkID int) *Chunk {
    seedSource := fmt.Sprintf("%s-%d", masterSeed, chunkID)
    hash := sha256.Sum256([]byte(seedSource))
    seed := int64(binary.BigEndian.Uint64(hash[:8]))
    rng := rand.New(rand.NewSource(seed))
    // Generate obstacles using rng...
}
```

## Skills

This project has a custom skill at `.claude/skills/vibe-runner/SKILL.md` that provides quick reference to all documentation, constants, and patterns. The skill activates automatically when working on Vibe Runner features.

## Testing Strategy

- **Phase 1:** Manual browser testing
- **Phase 2:** Single client connection test
- **Phase 3:** 2-5 concurrent clients
- **Phase 4:** Verify deterministic generation across clients
- **Phase 5:** Full user flow (join → play → die → respawn)
- **Phase 6:** Load test with 500+ simulated clients

## Troubleshooting References

Common issues and solutions are documented in:
- `docs/04-technical-architecture/frontend-pixijs.md` - Client desync, jittery ghosts, laggy jumps
- `docs/04-technical-architecture/backend-go.md` - Server performance, goroutine management
- `docs/04-technical-architecture/security.md` - Anti-cheat, XSS, rate limiting issues

## Philosophy

**Documentation is the source of truth.** Before making any architectural decision, reference the existing specifications. The docs are comprehensive, battle-tested designs—don't deviate without good reason and explicit discussion.

**Build incrementally.** Resist the urge to implement multiple phases at once. Each phase should be a stable, committable state.

**Server authority is sacred.** The server validates everything. Clients request, never command.
