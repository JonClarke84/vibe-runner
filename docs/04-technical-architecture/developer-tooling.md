# Developer Tooling Specifications

**Version:** 1.5 **Date:** 2025-11-13

## Overview

This document specifies the developer and QA tools required for debugging, testing, and monitoring Vibe Runner.

## Debug HUD (Client-Side)

A client-side debugging overlay for developers to visualize game state and network performance.

### Activation

Activated by URL query parameter:

```
http://localhost:3000/?debug=true
```

**Implementation:**

```javascript
const urlParams = new URLSearchParams(window.location.search);
const debugMode = urlParams.get('debug') === 'true';

if (debugMode) {
  initDebugHUD();
}
```

### Features

#### 1. Collision Boxes

Visualize AABB collision boxes for all entities.

**Implementation using Pixi.Graphics:**

```javascript
class DebugRenderer {
  constructor(stage) {
    this.graphics = new PIXI.Graphics();
    stage.addChild(this.graphics);
  }

  drawCollisionBox(entity, color = 0x00ff00) {
    this.graphics.lineStyle(2, color, 1);
    this.graphics.drawRect(
      entity.x,
      entity.y,
      entity.width,
      entity.height
    );
  }

  clear() {
    this.graphics.clear();
  }

  render(player, obstacles, ghosts) {
    this.clear();

    // Draw player (green)
    this.drawCollisionBox(player, 0x00ff00);

    // Draw obstacles (red)
    obstacles.forEach(obs => {
      this.drawCollisionBox(obs, 0xff0000);
    });

    // Draw ghosts (yellow)
    ghosts.forEach(ghost => {
      this.drawCollisionBox(ghost, 0xffff00);
    });
  }
}
```

**Visual Style:**
- Player: Green rectangle
- Obstacles: Red rectangle
- Ghosts: Yellow rectangle
- Line width: 2px

---

#### 2. Network Ping

Display real-time network latency (round-trip time).

**Implementation:**

```javascript
class PingMonitor {
  constructor(network) {
    this.network = network;
    this.pingHistory = [];
    this.currentPing = 0;
    this.averagePing = 0;
  }

  sendPing() {
    const timestamp = Date.now();
    this.network.send({
      e: "ping",
      d: { t: timestamp }
    });
  }

  onPong(timestamp) {
    const rtt = Date.now() - timestamp;
    this.currentPing = rtt;

    // Calculate rolling average (last 10 pings)
    this.pingHistory.push(rtt);
    if (this.pingHistory.length > 10) {
      this.pingHistory.shift();
    }
    this.averagePing = this.pingHistory.reduce((a, b) => a + b, 0) / this.pingHistory.length;
  }

  startMonitoring() {
    setInterval(() => this.sendPing(), 1000); // Every 1 second
  }
}
```

**Display:**
```
Ping: 45ms (avg: 48ms)
```

---

#### 3. State Display

Show server position vs client position for detecting desync.

**Implementation:**

```javascript
class StateMonitor {
  constructor(player) {
    this.player = player;
    this.serverX = 0;
    this.serverY = 0;
    this.clientX = 0;
    this.clientY = 0;
    this.desync = 0;
  }

  updateServerState(x, y) {
    this.serverX = x;
    this.serverY = y;
    this.calculateDesync();
  }

  updateClientState() {
    this.clientX = this.player.x;
    this.clientY = this.player.y;
    this.calculateDesync();
  }

  calculateDesync() {
    const dx = this.serverX - this.clientX;
    const dy = this.serverY - this.clientY;
    this.desync = Math.sqrt(dx * dx + dy * dy);
  }

  getDebugInfo() {
    return {
      serverPos: `(${this.serverX.toFixed(1)}, ${this.serverY.toFixed(1)})`,
      clientPos: `(${this.clientX.toFixed(1)}, ${this.clientY.toFixed(1)})`,
      desync: `${this.desync.toFixed(1)}px`
    };
  }
}
```

**Display:**
```
Server: (1024.0, 450.0)
Client: (1024.5, 450.2)
Desync: 0.5px
```

---

#### 4. Performance Metrics

Display FPS and other performance metrics.

```javascript
class PerformanceMonitor {
  constructor() {
    this.fps = 0;
    this.frameCount = 0;
    this.lastTime = performance.now();
  }

  update() {
    this.frameCount++;
    const currentTime = performance.now();
    const elapsed = currentTime - this.lastTime;

    if (elapsed >= 1000) {
      this.fps = Math.round((this.frameCount * 1000) / elapsed);
      this.frameCount = 0;
      this.lastTime = currentTime;
    }
  }
}
```

**Display:**
```
FPS: 60
Draw Calls: 145
Sprites: 523
```

---

### Debug HUD Layout

**Position:** Top-left corner overlay

**Visual Style:**
- Semi-transparent black background (rgba(0, 0, 0, 0.7))
- Phosphor green text (#33ff00)
- Monospace font (Courier New or VT323)
- 12px font size
- 10px padding

**Example Implementation:**

```javascript
class DebugHUD {
  constructor(stage) {
    this.container = new PIXI.Container();
    this.container.x = 10;
    this.container.y = 10;

    // Background
    this.bg = new PIXI.Graphics();
    this.bg.beginFill(0x000000, 0.7);
    this.bg.drawRect(0, 0, 300, 200);
    this.bg.endFill();
    this.container.addChild(this.bg);

    // Text
    this.text = new PIXI.Text('', {
      fontFamily: 'Courier New',
      fontSize: 12,
      fill: 0x33ff00,
      align: 'left'
    });
    this.text.x = 10;
    this.text.y = 10;
    this.container.addChild(this.text);

    stage.addChild(this.container);
  }

  update(debugInfo) {
    this.text.text = `
FPS: ${debugInfo.fps}
Ping: ${debugInfo.ping}ms (avg: ${debugInfo.avgPing}ms)
Server: ${debugInfo.serverPos}
Client: ${debugInfo.clientPos}
Desync: ${debugInfo.desync}
Players: ${debugInfo.playerCount}
Obstacles: ${debugInfo.obstacleCount}
    `.trim();
  }
}
```

---

## Load Testing Script

A command-line tool to simulate hundreds of concurrent players for stress testing.

### Requirements

- Simulate 500+ concurrent WebSocket connections
- Send realistic player behavior (join, jump periodically)
- Log connection failures and message latency
- Configurable via command-line flags

### Implementation (Go)

**File:** `tools/load-test/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "math/rand"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type Bot struct {
    ID   int
    Conn *websocket.Conn
    Name string
}

var (
    serverURL = flag.String("url", "ws://localhost:8080/ws", "WebSocket server URL")
    numBots   = flag.Int("c", 100, "Number of concurrent bots")
    duration  = flag.Int("d", 60, "Test duration in seconds")
)

func main() {
    flag.Parse()

    fmt.Printf("Starting load test:\n")
    fmt.Printf("  Server: %s\n", *serverURL)
    fmt.Printf("  Bots: %d\n", *numBots)
    fmt.Printf("  Duration: %d seconds\n", *duration)

    var wg sync.WaitGroup
    startTime := time.Now()

    for i := 0; i < *numBots; i++ {
        wg.Add(1)
        go func(botID int) {
            defer wg.Done()
            runBot(botID, *serverURL, *duration)
        }(i)

        // Stagger connection attempts
        time.Sleep(10 * time.Millisecond)
    }

    wg.Wait()
    elapsed := time.Since(startTime)

    fmt.Printf("\nLoad test completed in %v\n", elapsed)
}

func runBot(id int, url string, duration int) {
    // Connect
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Printf("Bot %d: Connection failed: %v", id, err)
        return
    }
    defer conn.Close()

    bot := &Bot{
        ID:   id,
        Conn: conn,
        Name: fmt.Sprintf("Bot%d", id),
    }

    // Send join message
    joinMsg := map[string]interface{}{
        "e": "join",
        "d": map[string]interface{}{
            "n": bot.Name,
        },
    }
    if err := conn.WriteJSON(joinMsg); err != nil {
        log.Printf("Bot %d: Join failed: %v", id, err)
        return
    }

    log.Printf("Bot %d: Connected as %s", id, bot.Name)

    // Start reading messages
    go readMessages(bot)

    // Simulate gameplay
    endTime := time.Now().Add(time.Duration(duration) * time.Second)
    for time.Now().Before(endTime) {
        // Random jump interval (1-5 seconds)
        sleepDuration := time.Duration(rand.Intn(4)+1) * time.Second
        time.Sleep(sleepDuration)

        // Send jump
        jumpMsg := map[string]interface{}{
            "e": "jump",
            "d": map[string]interface{}{
                "t": time.Now().UnixMilli(),
            },
        }
        if err := conn.WriteJSON(jumpMsg); err != nil {
            log.Printf("Bot %d: Jump failed: %v", id, err)
            return
        }
    }

    log.Printf("Bot %d: Disconnecting", id)
}

func readMessages(bot *Bot) {
    for {
        var msg map[string]interface{}
        err := bot.Conn.ReadJSON(&msg)
        if err != nil {
            return // Connection closed
        }

        // Log important events
        if event, ok := msg["e"].(string); ok {
            switch event {
            case "welcome":
                log.Printf("Bot %d: Received welcome", bot.ID)
            case "death":
                log.Printf("Bot %d: Died", bot.ID)
            }
        }
    }
}
```

### Usage

```bash
# Build
go build -o load-test tools/load-test/main.go

# Run with 500 bots for 60 seconds
./load-test -c=500 -url=ws://localhost:8080/ws -d=60

# Run with 1000 bots
./load-test -c=1000 -url=ws://production.server.com/ws -d=120
```

### Metrics to Log

- **Connection Success Rate:** % of bots that successfully connected
- **Average Connection Time:** Time to establish WebSocket connection
- **Message Latency:** Time from send to receive (for ping/pong)
- **Disconnection Count:** Number of unexpected disconnects
- **Server Response Time:** Time for server to respond to messages

---

## Admin Panel

A web-based dashboard for viewing server statistics and monitoring health.

### Features

1. **Current Player Count**
2. **Peak Player Count** (today/all-time)
3. **Server Uptime**
4. **Average Tick Duration** (game loop performance)
5. **Messages Per Second** (throughput)
6. **Recent Logs** (last 100 events)
7. **Top Players** (current session)

### Implementation

**Backend Endpoint (Go):**

```go
// File: server/admin.go

type AdminStats struct {
    CurrentPlayers int       `json:"currentPlayers"`
    PeakPlayers    int       `json:"peakPlayers"`
    Uptime         int64     `json:"uptime"` // seconds
    AvgTickTime    float64   `json:"avgTickTime"` // milliseconds
    MessagesPerSec float64   `json:"messagesPerSec"`
    RecentLogs     []string  `json:"recentLogs"`
    TopPlayers     []Player  `json:"topPlayers"`
}

func handleAdminStats(w http.ResponseWriter, r *http.Request) {
    stats := AdminStats{
        CurrentPlayers: len(gameState.Players),
        PeakPlayers:    gameState.PeakPlayers,
        Uptime:         int64(time.Since(serverStartTime).Seconds()),
        AvgTickTime:    gameState.AvgTickDuration.Milliseconds(),
        MessagesPerSec: gameState.MessageRate,
        RecentLogs:     getLast100Logs(),
        TopPlayers:     getTopPlayers(10),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

// Mount endpoint
http.HandleFunc("/admin/stats", handleAdminStats)
```

**Frontend (HTML + JavaScript):**

```html
<!-- File: admin/index.html -->
<!DOCTYPE html>
<html>
<head>
  <title>Vibe Runner - Admin Panel</title>
  <style>
    body {
      background: #1a1a2e;
      color: #33ff00;
      font-family: 'Courier New', monospace;
      padding: 20px;
    }
    .stat-box {
      background: rgba(0, 0, 0, 0.7);
      border: 2px solid #00f0ff;
      padding: 20px;
      margin: 10px 0;
    }
    h1 { color: #ff007f; }
    .value { font-size: 2em; color: #00f0ff; }
  </style>
</head>
<body>
  <h1>VIBE RUNNER - ADMIN PANEL</h1>

  <div class="stat-box">
    <div>Current Players</div>
    <div class="value" id="currentPlayers">-</div>
  </div>

  <div class="stat-box">
    <div>Peak Players</div>
    <div class="value" id="peakPlayers">-</div>
  </div>

  <div class="stat-box">
    <div>Server Uptime</div>
    <div class="value" id="uptime">-</div>
  </div>

  <div class="stat-box">
    <div>Average Tick Time</div>
    <div class="value" id="avgTickTime">-</div>
  </div>

  <script>
    async function updateStats() {
      const res = await fetch('/admin/stats');
      const stats = await res.json();

      document.getElementById('currentPlayers').textContent = stats.currentPlayers;
      document.getElementById('peakPlayers').textContent = stats.peakPlayers;
      document.getElementById('uptime').textContent = formatUptime(stats.uptime);
      document.getElementById('avgTickTime').textContent = stats.avgTickTime.toFixed(2) + 'ms';
    }

    function formatUptime(seconds) {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      return `${hours}h ${minutes}m`;
    }

    // Update every 2 seconds
    setInterval(updateStats, 2000);
    updateStats();
  </script>
</body>
</html>
```

### Access Control

**Basic Authentication:**

```go
func adminAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth()

        if !ok || username != "admin" || password != os.Getenv("ADMIN_PASSWORD") {
            w.Header().Set("WWW-Authenticate", `Basic realm="Admin Panel"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Mount with auth
http.Handle("/admin/", adminAuthMiddleware(http.FileServer(http.Dir("./admin"))))
```

**Environment Variable:**
```bash
export ADMIN_PASSWORD="secure-admin-password-here"
```

---

## Related Documentation

- Frontend Implementation: docs/04-technical-architecture/frontend-pixijs.md
- Backend Implementation: docs/04-technical-architecture/backend-go.md
- Phase 6 (Polish & Tooling): docs/00-development-phases/phase-6-polish-tooling.md
