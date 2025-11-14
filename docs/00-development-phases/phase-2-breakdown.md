# Phase 2: Implementation Breakdown

**Version:** 1.0 **Date:** 2025-11-14

## Overview

This document breaks Phase 2 (Basic Server & Client Connection) into incremental, testable chunks. Each chunk should be implemented, tested, and verified before moving to the next.

**Phase 2 Goal**: Connect client to a Go server that controls the player's position server-side.

---

## Chunk 1: Go Server Setup & Basic WebSocket ⚡

### Tasks
- Initialize Go project with `go mod init vibe-runner-server`
- Create `server/main.go` with basic HTTP server on port 8080
- Add `gorilla/websocket` dependency
- Implement WebSocket upgrade handler at `/ws` endpoint
- Add basic connection logging

### Files Created
- `server/go.mod`
- `server/main.go`

### Test Criteria
- Server starts without errors
- Can connect from browser console:
  ```javascript
  const ws = new WebSocket('ws://localhost:8080/ws');
  ws.onopen = () => console.log('Connected!');
  ```
- Server logs show successful WebSocket upgrade

### Expected Output
```
Server starting on :8080
WebSocket upgrade: client connected
```

---

## Chunk 2: Message Protocol & Join Handshake 📨

### Tasks
- Define Go structs for message protocol:
  - `Message` (base with `e` and `d` fields)
  - `JoinMessage` (`n` = name)
  - `WelcomeMessage` (`id`, `seed`, `serverTime`)
- Implement JSON message parsing with error handling
- Handle `join` event
- Send `welcome` response with assigned player ID
- Implement simple player ID counter (thread-safe)

### Files Created
- `server/network/messages.go`
- `server/network/protocol.go`

### Test Criteria
- Client sends: `{"e": "join", "d": {"n": "TestPlayer"}}`
- Client receives: `{"e": "welcome", "d": {"id": 1, "seed": "...", "serverTime": ...}}`
- Each subsequent client gets incremented ID (2, 3, 4...)

### Expected Output
```javascript
// Browser console
ws.send('{"e": "join", "d": {"n": "TestPlayer"}}');
// Receives: {"e":"welcome","d":{"id":1,"seed":"vibe-runner-1","serverTime":1700000000000}}
```

---

## Chunk 3: Server Game State & Player Management 🎮

### Tasks
- Create `server/game/state.go` with `GameState` struct
- Create `server/game/player.go` with `Player` entity
  - Fields: `ID`, `Name`, `X`, `Y`, `VelocityY`, `IsGrounded`, `IsAlive`
  - Initial position: x=100, y=440 (on ground)
- Implement thread-safe player add/remove operations
- Store `GameState` as global or passed to handlers
- Add player to game state on successful join
- Remove player from game state on disconnect

### Files Created
- `server/game/state.go`
- `server/game/player.go`

### Test Criteria
- Multiple clients can join simultaneously
- Each client gets unique ID and player in game state
- Server logs show player count increasing/decreasing
- No race conditions (run with `go run -race`)

### Expected Output
```
Player TestPlayer (ID: 1) joined at position (100, 440)
Active players: 1
Player TestPlayer (ID: 1) disconnected
Active players: 0
```

---

## Chunk 4: Server Game Loop (20Hz Ticker) ⏱️

### Tasks
- Create `server/game/ticker.go` with `GameTicker` function
- Use `time.NewTicker(50 * time.Millisecond)` for 20Hz loop
- Apply physics to all players each tick:
  - Gravity: 1200 pixels/second² (convert to per-tick)
  - Update velocityY: `velocityY += gravity * deltaTime`
  - Update Y position: `y += velocityY * deltaTime`
- Implement ground collision detection (y >= 440)
  - Set `y = 440`, `velocityY = 0`, `isGrounded = true`
- Launch ticker goroutine on server start

### Files Modified
- `server/game/ticker.go` (new)
- `server/main.go` (launch ticker)

### Physics Constants
```go
const (
    Gravity       = 1200.0  // pixels/second²
    JumpVelocity  = -600.0  // pixels/second
    GroundY       = 440.0   // pixels
    PlayerWidth   = 40.0
    PlayerHeight  = 60.0
    TickRate      = 20      // Hz
    TickDuration  = time.Second / TickRate // 50ms
)
```

### Test Criteria
- Server logs player positions each tick (enable debug logging)
- Player with initial position (100, 0) falls to ground at y=440
- VelocityY increases due to gravity, stops at ground
- Ticker runs at consistent 20Hz (measure with timestamps)

### Expected Output
```
[Tick 0] Player 1 at (100.0, 0.0) velocityY=0.0
[Tick 1] Player 1 at (100.0, 3.0) velocityY=60.0
[Tick 2] Player 1 at (100.0, 9.0) velocityY=120.0
...
[Tick 15] Player 1 at (100.0, 440.0) velocityY=0.0 [grounded]
```

---

## Chunk 5: State Broadcasting 📡

### Tasks
- Create `StateMessage` struct with `t` (timestamp), `p` (player array)
- Implement broadcast function to send state to all clients
- Call broadcast function every tick (20Hz)
- Implement per-client write goroutine with buffered channel
- Handle slow client disconnection (write timeout)

### Files Created
- `server/network/broadcast.go`

### Message Format
```go
type PlayerState struct {
    I int     `json:"i"` // ID
    X float64 `json:"x"` // X position
    Y float64 `json:"y"` // Y position
}

type StateMessage struct {
    E string        `json:"e"` // "state"
    D StateData     `json:"d"`
}

type StateData struct {
    T int64         `json:"t"` // Server timestamp
    P []PlayerState `json:"p"` // Players
}
```

### Test Criteria
- Client receives state messages at ~20Hz
- Client console logs show player positions updating
- State includes all alive players
- Dead players excluded from broadcast
- Verify with browser console:
  ```javascript
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.e === 'state') {
      console.log('State:', msg.d.p);
    }
  };
  ```

### Expected Output
```javascript
// Browser console (20 times per second)
State: [{i: 1, x: 100, y: 440}]
State: [{i: 1, x: 100, y: 440}]
State: [{i: 1, x: 100, y: 440}, {i: 2, x: 100, y: 440}]
```

---

## Chunk 6: Client WebSocket Connection 🔌

### Tasks
- Create `client/public/src/network/WebSocketClient.js`
- Implement connection to `ws://localhost:8080/ws`
- Send `join` message with player name on connection open
- Receive and parse `welcome` message
- Store assigned player ID
- Log all incoming messages to console (debug mode)
- Handle connection errors and reconnection

### Files Created
- `client/public/src/network/WebSocketClient.js`

### Client Interface
```javascript
export class WebSocketClient {
    constructor(playerName) {
        this.playerName = playerName;
        this.playerId = null;
        this.ws = null;
        this.isConnected = false;
    }

    connect() {
        this.ws = new WebSocket('ws://localhost:8080/ws');
        this.ws.onopen = () => this.onOpen();
        this.ws.onmessage = (e) => this.onMessage(e);
        this.ws.onerror = (e) => this.onError(e);
        this.ws.onclose = () => this.onClose();
    }

    onOpen() {
        console.log('Connected to server');
        this.send('join', { n: this.playerName });
    }

    send(event, data) {
        const msg = { e: event, d: data };
        this.ws.send(JSON.stringify(msg));
    }
}
```

### Files Modified
- `client/public/src/main.js` (create WebSocketClient instance)

### Test Criteria
- Client connects to server on game start
- Client sends join message automatically
- Client receives welcome message with ID
- Client logs show connection status and messages
- Handle server offline gracefully (error message)

### Expected Console Output
```
Connecting to server...
Connected to server
Sent: {"e":"join","d":{"n":"Player1"}}
Received: {"e":"welcome","d":{"id":1,"seed":"vibe-runner-1","serverTime":1700000000000}}
Assigned player ID: 1
```

---

## Chunk 7: Client Receives Server State 📥

### Tasks
- Parse incoming `state` messages in `WebSocketClient`
- Extract player data for local player (match by ID)
- Update `Player` entity position from server state
- **Temporarily disable local physics**:
  - Comment out gravity application in `Player.update()`
  - Comment out jump velocity application
  - Keep `this.sprite.position.set()` for rendering
- Add method to set position from server: `Player.setServerPosition(x, y)`

### Files Modified
- `client/public/src/network/WebSocketClient.js`
- `client/public/src/game/Player.js`
- `client/public/src/main.js`

### Client Logic
```javascript
// WebSocketClient.js
onMessage(event) {
    const msg = JSON.parse(event.data);

    if (msg.e === 'welcome') {
        this.playerId = msg.d.id;
        console.log('Assigned player ID:', this.playerId);
    }

    if (msg.e === 'state') {
        const myPlayer = msg.d.p.find(p => p.i === this.playerId);
        if (myPlayer) {
            // Callback to update game player
            this.onStateUpdate(myPlayer.x, myPlayer.y);
        }
    }
}

// main.js
wsClient.onStateUpdate = (x, y) => {
    player.setServerPosition(x, y);
};
```

### Test Criteria
- Player sprite renders at server-provided position
- Player falls to ground due to server physics (not local)
- Client logs show state updates at ~20Hz
- Player position matches server state (verify in debug HUD)
- Multiple clients see their own players move independently

### Expected Behavior
- Player spawns at (100, 0)
- Player falls smoothly to ground at (100, 440)
- Client does NOT apply local gravity
- Position updates come from server state messages

---

## Chunk 8: Jump Input Flow 🦘

### Tasks
- Send `jump` message to server on spacebar press
- Include client timestamp in jump message
- Server receives jump message
- Server validates jump (player must be grounded and alive)
- Server applies jump velocity to player (-600 pixels/second)
- Client receives updated position in next state message
- Player sprite jumps on screen

### Files Modified
- `client/public/src/network/WebSocketClient.js`
- `client/public/src/main.js` (keyboard handler)
- `server/network/protocol.go` (handle jump message)
- `server/game/player.go` (add jump method)

### Client Implementation
```javascript
// main.js
window.addEventListener('keydown', (e) => {
    if (e.code === 'Space' && !e.repeat) {
        wsClient.sendJump();
    }
});

// WebSocketClient.js
sendJump() {
    this.send('jump', { t: Date.now() });
}
```

### Server Implementation
```go
// player.go
func (p *Player) Jump() {
    if p.IsGrounded && p.IsAlive {
        p.VelocityY = JumpVelocity // -600
        p.IsGrounded = false
    }
}

// protocol.go - handle jump message
case "jump":
    player := gameState.GetPlayer(playerID)
    if player != nil {
        player.Jump()
    }
```

### Test Criteria
- Press spacebar, player jumps on screen
- Expected 20Hz delay (50ms) between input and visual response
- Jump height matches Phase 1 (same physics constants)
- Cannot double-jump (server enforces grounded check)
- Server logs show jump events
- Multiple clients can jump independently

### Expected Behavior
1. Press spacebar
2. Client sends: `{"e":"jump","d":{"t":1700000000000}}`
3. Server receives jump, applies velocity
4. Next state message shows player rising: y=420, y=390, y=350...
5. Player arc reaches peak, falls back to ground
6. Player lands at y=440, can jump again

---

## Implementation Order

### Phase A: Server Implementation (Chunks 1-5)
Implement and test the complete server independently using browser console before touching client code.

**Order:**
1. Chunk 1 → Test WebSocket connection from console
2. Chunk 2 → Test join/welcome handshake from console
3. Chunk 3 → Test multiple clients joining
4. Chunk 4 → Verify physics with debug logging
5. Chunk 5 → Verify state broadcast with console logging

**Checkpoint**: Server fully functional, broadcasting state at 20Hz to console-connected clients.

---

### Phase B: Client Integration (Chunks 6-8)
Connect the existing Phase 1 client to the working server.

**Order:**
1. Chunk 6 → Client connects, logs messages
2. Chunk 7 → Disable local physics, render server position
3. Chunk 8 → Wire up jump input to server

**Checkpoint**: End-to-end playable game with server-authoritative movement.

---

## Success Criteria (Phase 2 Complete)

After all chunks are implemented:

- ✅ Client connects to Go server via WebSocket
- ✅ Server game loop runs at 20Hz
- ✅ Player falls with gravity (server-side physics)
- ✅ Player can jump (input sent to server, server applies)
- ✅ Client renders player at server-provided position
- ✅ Local physics disabled on client
- ✅ Multiple clients can connect simultaneously
- ✅ Each client controls their own player
- ✅ Connection remains stable during gameplay

---

## Testing Strategy

### Unit Testing
- **Server**: Test player physics calculations separately
- **Server**: Test message parsing with various inputs
- **Client**: Test WebSocket message handling

### Integration Testing
- **Single Client**: Connect one client, verify full flow
- **Multiple Clients**: Connect 2-3 clients, verify independent control
- **Network Conditions**: Test with simulated latency (later phase)

### Manual Testing Checklist
- [ ] Server starts without errors
- [ ] Client connects successfully
- [ ] Welcome message received with valid ID
- [ ] State messages received at ~20Hz
- [ ] Player falls to ground (server physics)
- [ ] Jump works (spacebar → server → visual response)
- [ ] Multiple clients can play simultaneously
- [ ] Server logs show correct player count
- [ ] Disconnecting client removes player from state

---

## Common Issues & Troubleshooting

### Issue: CORS errors
**Solution**: Add CORS headers to HTTP server:
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

### Issue: WebSocket connection refused
**Solution**: Verify server is running on port 8080, check firewall

### Issue: Client doesn't receive messages
**Solution**: Check WebSocket `onmessage` handler, verify JSON parsing

### Issue: Player jitters or teleports
**Solution**: Verify deltaTime calculations, check for race conditions on server

### Issue: Jump doesn't work
**Solution**: Check server logs for jump messages, verify isGrounded state

### Issue: Multiple players see same position
**Solution**: Verify player ID matching in client state parsing

---

## Performance Targets

### Server
- Tick rate: Solid 20Hz (50ms ± 2ms jitter)
- Memory: < 10 MB per 100 players
- CPU: < 5% on modern hardware with 100 players

### Client
- Frame rate: Maintain 60 FPS
- Network: ~4 KB/sec incoming, ~100 bytes/sec outgoing
- Latency: Jump response time ~50-100ms (network + tick)

---

## Next Phase Preview

**Phase 3: Client-Side Prediction & Ghost Players**

Once Phase 2 is complete and tested:
- Add client-side prediction (immediate jump response)
- Implement server reconciliation (rewind/replay)
- Add ghost players (other players' sprites)
- Implement entity interpolation (smooth 60 FPS ghosts from 20Hz data)

This will eliminate the perceived 50ms input lag and make the game feel responsive while maintaining server authority.

---

## Related Documentation

- **Phase 2 Overview**: `docs/00-development-phases/phase-2-server-connection.md`
- **Network Protocol**: `docs/04-technical-architecture/network-protocol.md`
- **Backend Architecture**: `docs/04-technical-architecture/backend-go.md`
- **Frontend Architecture**: `docs/04-technical-architecture/frontend-pixijs.md`
