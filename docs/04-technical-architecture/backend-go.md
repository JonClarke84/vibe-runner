# Backend (Server): Go (Golang)

**Version:** 1.5 **Date:** 2025-11-13

## Technology Stack

* **Language:** Go (Golang)
* **WebSocket Library:** `gorilla/websocket`
* **Database:** Redis (in-memory cache), PostgreSQL (persistent storage)

## Concurrency Model

One **Goroutine** will be spawned for *each* connected player to read their incoming messages (`readPump`). A single, central **Goroutine** (`gameTicker`) will run the main game loop, collect all inputs, and broadcast the game state.

### Goroutine Architecture

```
┌─────────────────┐
│  Client 1       │
│  (readPump)     │◄─── Goroutine 1
└────────┬────────┘
         │
┌────────▼────────┐
│  Client 2       │
│  (readPump)     │◄─── Goroutine 2
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│   Input Queue           │
│   (channel)             │
└──────────┬──────────────┘
           │
           ▼
┌──────────────────────────┐
│   Game Ticker            │◄─── Central Goroutine
│   (20Hz Game Loop)       │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│   Broadcast to All       │
│   Clients                │
└──────────────────────────┘
```

## Server Game Loop (The "Tick")

This is the heartbeat of the entire game.

### Tick Rate

```go
// Run the main game loop at 20Hz (50ms per tick)
ticker := time.NewTicker(50 * time.Millisecond)
```

### Tick Sequence

In each tick, the server does this, *in order*:

1. **Process Inputs:** Collect all buffered `jump` messages from all players.
2. **Run Simulation:** Update the game state.
   * Move every *alive* player forward.
   * Apply physics (gravity, jump velocity from inputs).
   * Run **collision detection** for *all* players against the level.
   * If `collision == true`, set `player.isAlive = false` and send a *targeted* death message:
     ```go
     S->C: {"e": "death", "d": {"s": 5000}}
     ```
3. **Broadcast State:** Bundle all *living* player positions into one large message and broadcast it to *all* connected clients:
   ```go
   S->C: {"e": "state", "d": {...}}
   ```

### Implementation Example

```go
package main

import (
    "time"
    "github.com/gorilla/websocket"
)

type Player struct {
    ID       int
    Name     string
    X        float64
    Y        float64
    VelocityY float64
    IsAlive  bool
    IsGrounded bool
    Conn     *websocket.Conn
}

type GameState struct {
    Players  map[int]*Player
    Obstacles []*Obstacle
    Seed     string
}

func gameTicker(state *GameState, inputChan chan PlayerInput) {
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // 1. Process Inputs
            processInputs(state, inputChan)

            // 2. Run Simulation
            updatePhysics(state, 0.05) // 50ms = 0.05 seconds
            checkCollisions(state)

            // 3. Broadcast State
            broadcastState(state)
        }
    }
}

func processInputs(state *GameState, inputChan chan PlayerInput) {
    // Drain all pending inputs
    for {
        select {
        case input := <-inputChan:
            player := state.Players[input.PlayerID]
            if player != nil && player.IsAlive {
                if input.Action == "jump" && player.IsGrounded {
                    player.VelocityY = -600.0
                    player.IsGrounded = false
                }
            }
        default:
            return // No more inputs
        }
    }
}

func updatePhysics(state *GameState, deltaTime float64) {
    const GRAVITY = 1200.0
    const GROUND_Y = 500.0
    const PLAYER_SPEED = 300.0

    for _, player := range state.Players {
        if !player.IsAlive {
            continue
        }

        // Apply gravity
        player.VelocityY += GRAVITY * deltaTime

        // Update position
        player.X += PLAYER_SPEED * deltaTime
        player.Y += player.VelocityY * deltaTime

        // Ground collision
        if player.Y >= GROUND_Y {
            player.Y = GROUND_Y
            player.VelocityY = 0
            player.IsGrounded = true
        }
    }
}

func checkCollisions(state *GameState) {
    for _, player := range state.Players {
        if !player.IsAlive {
            continue
        }

        // Check against obstacles
        for _, obstacle := range state.Obstacles {
            if aabbCollision(player, obstacle) {
                player.IsAlive = false

                // Send death message
                deathMsg := Message{
                    E: "death",
                    D: map[string]interface{}{
                        "s": calculateScore(player),
                    },
                }
                player.Conn.WriteJSON(deathMsg)

                // Save to leaderboard
                saveScore(player.Name, calculateScore(player))
            }
        }
    }
}

func broadcastState(state *GameState) {
    // Build state message
    playerData := []map[string]interface{}{}

    for _, player := range state.Players {
        if player.IsAlive {
            playerData = append(playerData, map[string]interface{}{
                "i": player.ID,
                "x": player.X,
                "y": player.Y,
            })
        }
    }

    stateMsg := Message{
        E: "state",
        D: map[string]interface{}{
            "t": time.Now().UnixMilli(),
            "p": playerData,
        },
    }

    // Broadcast to all clients
    for _, player := range state.Players {
        if player.Conn != nil {
            player.Conn.WriteJSON(stateMsg)
        }
    }
}
```

## Deterministic Procedural Generation

This ensures every player gets the same level.

### Generation Algorithm

1. **Master Seed:** On server start, generate a `masterSeed = "vibe-runner-12345"`.
2. **Level Chunks:** The level is generated in "chunks" (e.g., 5 screens wide).
3. **Seeded PRNG:** To generate Chunk `N`, the server initializes a **Pseudo-Random Number Generator (PRNG)** with a *deterministic* seed:
   ```go
   seed_N = hash(masterSeed + N)
   ```
4. The server uses this `prng_N` to generate all obstacles for that chunk.
5. The server broadcasts the obstacle data (`{"e": "chunk", ...}`) to clients *before* they reach it.

### Implementation Example

```go
import (
    "crypto/sha256"
    "encoding/binary"
    "math/rand"
)

const CHUNK_WIDTH = 5000.0 // 5 screen widths

type Obstacle struct {
    Type int
    X    float64
    Y    float64
    Width float64
    Height float64
}

type Chunk struct {
    ID       int
    Obstacles []*Obstacle
}

func generateChunk(masterSeed string, chunkID int) *Chunk {
    // Create deterministic seed for this chunk
    seedSource := fmt.Sprintf("%s-%d", masterSeed, chunkID)
    hash := sha256.Sum256([]byte(seedSource))
    seed := int64(binary.BigEndian.Uint64(hash[:8]))

    // Initialize PRNG with deterministic seed
    rng := rand.New(rand.NewSource(seed))

    chunk := &Chunk{
        ID:       chunkID,
        Obstacles: []*Obstacle{},
    }

    // Generate obstacles
    chunkStartX := float64(chunkID) * CHUNK_WIDTH
    numObstacles := rng.Intn(10) + 5 // 5-15 obstacles per chunk

    for i := 0; i < numObstacles; i++ {
        obstacleType := rng.Intn(3) + 1 // Types 1-3

        obstacle := &Obstacle{
            Type:   obstacleType,
            X:      chunkStartX + rng.Float64()*CHUNK_WIDTH,
            Y:      500.0, // Ground level
            Width:  getObstacleWidth(obstacleType),
            Height: getObstacleHeight(obstacleType),
        }

        chunk.Obstacles = append(chunk.Obstacles, obstacle)
    }

    return chunk
}

func getObstacleWidth(obstacleType int) float64 {
    switch obstacleType {
    case 1: return 40.0  // Tall, thin
    case 2: return 100.0 // Low, wide
    case 3: return 30.0  // Small spike
    default: return 50.0
    }
}

func getObstacleHeight(obstacleType int) float64 {
    switch obstacleType {
    case 1: return 150.0 // Tall
    case 2: return 60.0  // Low
    case 3: return 80.0  // Medium
    default: return 100.0
    }
}
```

### Chunk Broadcasting

```go
func broadcastChunk(chunk *Chunk, players map[int]*Player) {
    obstacleData := []map[string]interface{}{}

    for _, obs := range chunk.Obstacles {
        obstacleData = append(obstacleData, map[string]interface{}{
            "t": obs.Type,
            "x": obs.X,
        })
    }

    chunkMsg := Message{
        E: "chunk",
        D: map[string]interface{}{
            "id":  chunk.ID,
            "obs": obstacleData,
        },
    }

    // Broadcast to all clients
    for _, player := range players {
        if player.Conn != nil {
            player.Conn.WriteJSON(chunkMsg)
        }
    }
}
```

## WebSocket Connection Handling

### Connection Setup

```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Configure properly in production
    },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }

    player := &Player{
        ID:   generatePlayerID(),
        Conn: conn,
        IsAlive: true,
    }

    // Spawn readPump goroutine
    go readPump(player, inputChan)

    // Send welcome message
    welcomeMsg := Message{
        E: "welcome",
        D: map[string]interface{}{
            "id":         player.ID,
            "seed":       gameSeed,
            "serverTime": time.Now().UnixMilli(),
        },
    }
    conn.WriteJSON(welcomeMsg)

    // Add player to game state
    gameState.Players[player.ID] = player
}
```

### Read Pump

```go
func readPump(player *Player, inputChan chan PlayerInput) {
    defer func() {
        player.Conn.Close()
        delete(gameState.Players, player.ID)
    }()

    for {
        var msg Message
        err := player.Conn.ReadJSON(&msg)
        if err != nil {
            break // Connection closed
        }

        switch msg.E {
        case "join":
            handleJoin(player, msg.D)
        case "jump":
            inputChan <- PlayerInput{
                PlayerID: player.ID,
                Action:   "jump",
                Timestamp: msg.D["t"].(float64),
            }
        }
    }
}
```

## Performance Optimization

### Message Batching

Combine multiple small messages into one:

```go
type BatchMessage struct {
    E string        `json:"e"` // "batch"
    D []interface{} `json:"d"`
}
```

### Binary Protocol (Future)

For production, consider using binary protocol (MessagePack, Protocol Buffers) instead of JSON for reduced bandwidth.

## Related Documentation

- Frontend: See docs/04-technical-architecture/frontend-pixijs.md
- Network Protocol: See docs/04-technical-architecture/network-protocol.md
- Database: See docs/04-technical-architecture/database-schema.md
- Security: See docs/04-technical-architecture/security.md
