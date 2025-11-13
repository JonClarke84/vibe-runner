---
name: vibe-runner
description: Expert guidance for building Vibe Runner - a massively multiplayer 2D infinite runner game with 80s synthwave aesthetic. Use this when working on game development tasks, implementing any of the 6 phases, referencing technical architecture for Pixi.js frontend, Go backend with WebSocket networking, Redis/PostgreSQL databases, or when discussing procedural generation, client-side prediction, or multiplayer features.
allowed-tools:
  - Read
  - Glob
  - Grep
  - Edit
  - Write
  - Bash
  - Task
---

# Vibe Runner Development Skill

You are an expert Vibe Runner developer with deep knowledge of the project architecture, implementation details, and the complete technical specification.

## Project Overview

**Vibe Runner** is an always-on, massively multiplayer 2D infinite runner where all players run the exact same deterministically procedurally generated level simultaneously. The game features an over-the-top "Hyper-Synthwave" 80s sci-fi aesthetic with neon colors, parallax backgrounds, and CRT effects.

**Key Concept:** All players see the same obstacles at the same time, making it a fair competition. The server is the authoritative source of truth for all game logic.

## Essential Documentation Structure

All project documentation is in `docs/`:

### Phase Documentation (`docs/00-development-phases/`)
- **index.md** - Phase overview and progression strategy
- **phase-1-core-local-game.md** - Local single-player prototype
- **phase-2-server-connection.md** - WebSocket connection and server control
- **phase-3-prediction-ghosts.md** - Client prediction and multiplayer
- **phase-4-procedural-generation.md** - Deterministic level generation
- **phase-5-full-game-loop.md** - Complete UX (menu, death, leaderboard)
- **phase-6-polish-tooling.md** - Visual polish, audio, dev tools

### Technical Architecture (`docs/04-technical-architecture/`)
- **index.md** - Architecture overview with system diagrams
- **frontend-pixijs.md** - Client implementation (game loop, prediction, interpolation)
- **backend-go.md** - Server implementation (concurrency, game ticker, PRNG)
- **network-protocol.md** - All WebSocket message specifications
- **database-schema.md** - Redis and PostgreSQL schemas
- **developer-tooling.md** - Debug HUD, load testing, admin panel
- **security.md** - Anti-cheat, XSS prevention, rate limiting

### Other Key Documents
- **docs/01-user-stories.md** - User stories by epic
- **docs/02-product-requirements.md** - Core features and requirements
- **docs/03-art-style-aesthetics.md** - Complete visual identity guide

## Current Development Status

**Current Phase:** Phase 1 - Core Local Game (Client-Only)

**Next Steps:** Refer to `docs/00-development-phases/phase-1-core-local-game.md`

## Technical Stack

### Frontend
- **Pixi.js** - 2D WebGL rendering
- **JavaScript/TypeScript** - Game logic
- **WebSocket client** - Real-time communication
- **requestAnimationFrame** - Game loop (60 FPS target)

### Backend
- **Go (Golang)** - Server implementation
- **gorilla/websocket** - WebSocket handling
- **Redis** - In-memory cache (live leaderboard, sessions)
- **PostgreSQL** - Persistent storage (all-time scores)

### Network
- **WebSocket (JSON)** - Communication protocol
- **20Hz tick rate** - Server game loop
- **Client-side prediction** - Instant input response
- **Entity interpolation** - Smooth 60 FPS from 20Hz updates

## Core Architectural Principles

### 1. Server-Authoritative Design
**The server is the absolute source of truth.** Clients request actions, the server validates and executes them.

- ✅ Server calculates all collision detection
- ✅ Server validates jump requests (can't jump mid-air)
- ✅ Server controls player movement speed
- ❌ Clients cannot bypass obstacles by manipulating local code

### 2. Client-Side Prediction
Players experience instant input response despite network latency.

**Flow:**
1. Player presses spacebar → client immediately applies jump
2. Client sends `{"e": "jump", "d": {"t": timestamp}}` to server
3. Server validates and includes in next state broadcast
4. Client reconciles if position differs from server

### 3. Deterministic Procedural Generation
All players see identical obstacles generated from a seeded PRNG.

**Algorithm:**
- Master seed generated on server start
- Each chunk N uses `seed_N = hash(masterSeed + N)`
- PRNG generates obstacles for that chunk
- Server broadcasts chunk data to clients

### 4. Entity Interpolation (Ghosts)
Other players ("ghosts") move smoothly despite 20Hz updates.

**Implementation:**
```javascript
ghost.visual_x = lerp(ghost.visual_x, ghost.targetPosition.x, 0.3);
```

## When to Use This Skill

Activate this skill when you need to:

- Implement any game feature (physics, collision, rendering)
- Set up client or server architecture
- Work on networking or WebSocket communication
- Implement procedural generation or PRNG logic
- Add database operations (Redis/PostgreSQL)
- Reference art style or color palettes
- Understand message protocol specifications
- Implement security measures (anti-cheat, XSS prevention)
- Create developer tools (debug HUD, load testing)
- Check phase requirements and success criteria
- Review user stories or product requirements

## Development Workflow

### Starting a New Feature

1. **Check Phase Requirements**
   - Read current phase doc in `docs/00-development-phases/`
   - Review tasks, assets needed, and success criteria

2. **Reference Technical Specs**
   - Check `docs/04-technical-architecture/` for implementation details
   - Review relevant sections (frontend, backend, network, etc.)

3. **Follow Visual Guidelines**
   - Refer to `docs/03-art-style-aesthetics.md` for colors, fonts, effects
   - Use specified color palette (Electric Pink #ff007f, Hyper-Cyan #00f0ff)

4. **Implement with Best Practices**
   - Server validates all client inputs
   - Sanitize user inputs to prevent XSS
   - Use client-side prediction for responsive gameplay
   - Interpolate ghost positions for smooth movement

5. **Test Against Success Criteria**
   - Verify all tasks completed
   - Check success criteria from phase document
   - Test with multiple clients (Phase 3+)

## Key Constants and Values

### Physics
```javascript
const GRAVITY = 1200;          // pixels/second^2
const JUMP_VELOCITY = -600;    // pixels/second
const PLAYER_SPEED = 300;      // pixels/second
const GROUND_Y = 500;          // Y position of ground
```

### Network
- **Server Tick Rate:** 20Hz (50ms per tick)
- **Message Keys:** `e` (event), `d` (data)
- **Bandwidth per Client:** ~4-5 KB/sec

### Color Palette
- **Base:** Deep Indigo `#1a1a2e`, Dark Purple `#301a4b`
- **Primary Neon:** Electric Pink `#ff007f`, Hyper-Cyan `#00f0ff`
- **Secondary Neon:** Solar Orange `#ff8c00`, Phosphor Green `#33ff00`
- **Obstacle:** Glitch Red `#ff003c`

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

### Input Sanitization
```go
func sanitizePlayerName(name string) string {
    name = strings.TrimSpace(name)
    if len(name) > 30 {
        name = name[:30]
    }
    name = stripHTMLTags(name)
    name = html.EscapeString(name)
    if name == "" {
        name = "Player"
    }
    return name
}
```

### State Broadcast Message
```json
{
  "e": "state",
  "d": {
    "t": 1678886400550,
    "p": [
      {"i": 12345, "x": 1024, "y": 50},
      {"i": 12346, "x": 1022, "y": 80}
    ]
  }
}
```

## Security Checklist

Always consider these security measures:

- [ ] Server validates all client inputs
- [ ] Player names are sanitized (XSS prevention)
- [ ] Jump requests validated (is player grounded?)
- [ ] Collision detection runs only on server
- [ ] Rate limiting on connections and messages
- [ ] Use WSS (WebSocket Secure) in production
- [ ] Admin panel protected with authentication

## Quick Reference Commands

### Documentation
```bash
# View main index
cat docs/index.md

# Check current phase
cat docs/00-development-phases/phase-1-core-local-game.md

# Review architecture
cat docs/04-technical-architecture/index.md
```

### Development
```bash
# Start Redis
redis-server

# Start PostgreSQL
pg_ctl start

# Run Go server
cd server && go run main.go

# Run frontend dev server
cd client && npm run dev
```

## Troubleshooting Common Issues

### "Client and server positions desync"
- Check client reconciliation logic in frontend-pixijs.md
- Verify server is broadcasting state at 20Hz
- Ensure client snaps to server position when diff > threshold

### "Obstacles appear different for different players"
- Verify PRNG is seeded with deterministic value
- Check chunk generation uses hash(masterSeed + chunkID)
- Ensure server broadcasts chunk data before players reach it

### "Jumps feel laggy"
- Implement client-side prediction (Phase 3)
- Apply velocityY immediately on client
- Send jump message to server for validation

### "Ghost players are jittery"
- Implement entity interpolation with lerp
- Use 0.3 as lerp amount for smooth motion
- Update visual position in render loop, not update loop

## Tips for Success

1. **Read the phase document first** before implementing features
2. **Follow the phased approach** - don't skip ahead
3. **Refer to technical architecture** for detailed implementation patterns
4. **Use the color palette** exactly as specified in art style guide
5. **Test each phase thoroughly** before moving to the next
6. **Keep security in mind** - server authority, input sanitization, rate limiting
7. **Use the TodoWrite tool** to track multi-step tasks
8. **Commit working states** at the end of each phase

## Additional Resources

For deeper understanding of specific topics, always reference:

- **Game Loop & Physics:** frontend-pixijs.md
- **Server Concurrency:** backend-go.md
- **Message Formats:** network-protocol.md
- **Database Queries:** database-schema.md
- **Visual Style:** art-style-aesthetics.md
- **Security Measures:** security.md
- **Debug Tools:** developer-tooling.md

Remember: Documentation is your source of truth. When in doubt, reference the docs before making architectural decisions.
