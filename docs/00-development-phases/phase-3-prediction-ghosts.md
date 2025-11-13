# Phase 3: Client-Side Prediction & Multiplayer Ghosts

**Version:** 1.5 **Date:** 2025-11-13

## Goal
Re-enable client-side responsiveness and add "ghost" players.

## Tasks

### Client-Side Prediction
* Client jumps *immediately* on input.
* Client sends `{"e": "jump", ...}` to the server.
* Server validates the jump and includes it in the *next* state broadcast.
* Client implements basic state correction/snapping if its predicted position desyncs from the server's broadcasted position.

### Multiplayer Ghosts
* Allow multiple clients to connect.
* The server's `state` broadcast now includes *all* players.
* The client renders all other players as "ghost" sprites.
* Implement **Entity Interpolation (lerp)** on ghosts so their 20Hz movement looks smooth.

## Assets Required

* **`player-ghost.svg` / Shader**: A "ghost" version of the player sprite, or a GLSL shader to create the silhouette/glitch effect described in `3.3`.
* **`font-player-name.woff2`**: The chosen 80s-style font (e.g., "VT323") for rendering player names above ghosts.

*Note: Refer to docs/03-art-style-aesthetics.md Section 3.3 (Characters) for the specific "ghost" visual description.*

## Success Criteria

- Player jump feels instant and responsive
- Local player simulation runs independently
- Jump input messages sent to server with timestamps
- Server validates jump requests
- Client reconciles position differences with server state
- Multiple clients can connect simultaneously
- All connected players visible on each client
- Ghost players render with distinct visual style
- Ghost movement appears smooth (interpolated from 20Hz updates)
- Player names display above ghost sprites

## Technical Notes

**Client-Side Prediction Flow:**
```
1. User presses Spacebar
2. Client immediately applies velocityY to local player
3. Client sends: {"e": "jump", "d": {"t": timestamp}}
4. Client continues local simulation
5. Server validates and broadcasts state
6. Client reconciles if position differs
```

**Entity Interpolation:**
```javascript
function lerp(start, end, amount) {
  return (1 - amount) * start + amount * end;
}

// In render loop:
ghost.visual_x = lerp(ghost.visual_x, ghost.targetPosition.x, 0.3);
ghost.visual_y = lerp(ghost.visual_y, ghost.targetPosition.y, 0.3);
```

Refer to docs/04-technical-architecture/frontend-pixijs.md for detailed implementation guidance.
