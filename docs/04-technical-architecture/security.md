# Security Specifications

**Version:** 1.5 **Date:** 2025-11-13

## Overview

This document details measures to protect the game's integrity, player data, and server health from common attack vectors.

## 4.1. Anti-Cheat (Server-Side Authority)

### Core Principle

**The server is the absolute source of truth.** The client *requests* actions; the server *validates* and *executes* them.

### State Validation

The client's `{"e": "jump"}` message is a *request*, not a command.

**Server Validation:**

```go
func handleJumpRequest(player *Player, timestamp int64) {
    // Validate player is alive
    if !player.IsAlive {
        return // Ignore jumps from dead players
    }

    // Validate player is on ground (can't jump mid-air)
    if !player.IsGrounded {
        log.Printf("Player %d attempted to jump while airborne", player.ID)
        return
    }

    // Validate jump cooldown (prevent spam)
    timeSinceLastJump := time.Now().UnixMilli() - player.LastJumpTime
    if timeSinceLastJump < 100 { // Min 100ms between jumps
        log.Printf("Player %d jump spam detected", player.ID)
        return
    }

    // Validate timestamp (prevent replay attacks)
    if timestamp < player.LastInputTimestamp {
        log.Printf("Player %d sent outdated timestamp", player.ID)
        return
    }

    // ALL CHECKS PASSED - Execute jump
    player.VelocityY = -600.0
    player.IsGrounded = false
    player.LastJumpTime = time.Now().UnixMilli()
    player.LastInputTimestamp = timestamp
}
```

**What This Prevents:**
- Players jumping while mid-air (double-jump exploit)
- Players jumping after death (ghost exploit)
- Players spamming jump (input flooding)
- Replay attacks (reusing old jump messages)

---

### Collision Authority

All collision detection (player vs. obstacle) is calculated *only* on the server. A client cannot "lie" and say it missed an obstacle.

**Server-Only Collision:**

```go
func checkCollisions(state *GameState) {
    for _, player := range state.Players {
        if !player.IsAlive {
            continue
        }

        for _, obstacle := range state.Obstacles {
            if aabbCollision(player, obstacle) {
                // Server detected collision - player dies
                player.IsAlive = false
                player.DeathTime = time.Now()

                // Send death message
                sendDeathMessage(player)

                // Save score
                saveScore(player.Name, calculateScore(player))

                log.Printf("Player %d died at score %.2f", player.ID, calculateScore(player))
            }
        }
    }
}
```

**What This Prevents:**
- Clients modifying collision detection to "pass through" obstacles
- Clients reporting false collision data
- Position manipulation exploits

---

### Position Correction

If a client's local simulation desynchronizes (e.g., due to lag or tampering), the server's next `{"e": "state"}` broadcast will contain the *correct* position. The client *must* snap to the server's authoritative state.

**Client Reconciliation:**

```javascript
function handleStateUpdate(serverState) {
    const serverPlayer = serverState.p.find(p => p.i === myPlayerId);

    if (serverPlayer) {
        const positionDiff = Math.abs(player.x - serverPlayer.x) + Math.abs(player.y - serverPlayer.y);

        // If desync is too large, snap to server position
        if (positionDiff > DESYNC_THRESHOLD) {
            console.warn(`Position desync detected: ${positionDiff}px - snapping to server`);
            player.x = serverPlayer.x;
            player.y = serverPlayer.y;
        }
    }
}
```

**Death Authority:**

If the server detects a collision, it sends a *targeted* `{"e": "death"}` message, which the client must obey:

```javascript
function handleDeath(data) {
    // Server says we died - no arguing
    player.isAlive = false;
    showDeathScreen(data.s); // Show score
    stopGameLoop();
}
```

**What This Prevents:**
- Position hacking (teleportation, speed hacks)
- "Ignoring" death messages from server
- Clients staying alive after server detected collision

---

### Speed Hacking Prevention

Player movement is calculated based on the server's 20Hz tick rate, not client-provided `deltaTime`. This prevents speed hacks.

**Server-Controlled Movement:**

```go
const PLAYER_SPEED = 300.0 // pixels per second
const TICK_RATE = 0.05     // 50ms = 0.05 seconds

func updatePhysics(state *GameState) {
    for _, player := range state.Players {
        if !player.IsAlive {
            continue
        }

        // Movement is controlled by server's fixed timestep
        player.X += PLAYER_SPEED * TICK_RATE

        // Client CANNOT influence horizontal speed
    }
}
```

**What This Prevents:**
- Clients speeding up by manipulating deltaTime
- Clients slowing down to "dodge" obstacles in slow motion
- Any time manipulation exploits

---

## 4.2. Input Sanitization (XSS Prevention)

### Threat

A user provides a malicious `player_name` in the `{"e": "join"}` message, such as:

```json
{"e": "join", "d": {"n": "<script>document.location='http://evil.com'</script>"}}
```

If this name is rendered in the client without sanitization, it could execute arbitrary JavaScript.

### Defense-in-Depth Strategy

#### Layer 1: Server-Side Sanitization (On Ingest)

The Go server *must* sanitize all user-provided strings upon receipt.

**Implementation:**

```go
import (
    "html"
    "regexp"
    "strings"
)

func sanitizePlayerName(name string) string {
    // 1. Trim whitespace
    name = strings.TrimSpace(name)

    // 2. Length limit (max 30 characters)
    if len(name) > 30 {
        name = name[:30]
    }

    // 3. Strip all HTML tags
    name = stripHTMLTags(name)

    // 4. HTML entity encoding (defense in depth)
    name = html.EscapeString(name)

    // 5. Remove control characters
    name = removeControlCharacters(name)

    // 6. Default name if empty
    if name == "" {
        name = "Player"
    }

    return name
}

func stripHTMLTags(s string) string {
    // Remove anything between < and >
    re := regexp.MustCompile(`<[^>]*>`)
    return re.ReplaceAllString(s, "")
}

func removeControlCharacters(s string) string {
    // Remove characters < 0x20 except space
    re := regexp.MustCompile(`[\x00-\x1F]`)
    return re.ReplaceAllString(s, "")
}
```

**Usage:**

```go
func handleJoinMessage(player *Player, data map[string]interface{}) {
    rawName := data["n"].(string)
    player.Name = sanitizePlayerName(rawName)

    log.Printf("Player joined: %s (sanitized from: %s)", player.Name, rawName)
}
```

---

#### Layer 2: Storage Sanitization

Only the sanitized, plain-text name is stored in Redis and PostgreSQL.

```go
// Redis
rdb.HSet(ctx, fmt.Sprintf("player:%d", player.ID), "name", player.Name)

// PostgreSQL
db.Exec("INSERT INTO scores (player_name, score) VALUES ($1, $2)", player.Name, score)
```

---

#### Layer 3: Broadcast Sanitization

The server broadcasts only the sanitized name in state packets.

```go
playerData := map[string]interface{}{
    "i": player.ID,
    "x": player.X,
    "y": player.Y,
    "n": player.Name, // Already sanitized
}
```

---

#### Layer 4: Client-Side Rendering (Pixi.Text)

The client's rendering logic (Pixi.js) should render names as `Pixi.Text` objects, which are **not parsed as HTML**. This provides a second layer of defense.

```javascript
// Safe: Pixi.Text does not execute scripts
const nameText = new PIXI.Text(playerName, {
    fontFamily: 'VT323',
    fontSize: 14,
    fill: 0x33ff00
});

// NEVER do this (vulnerable):
// element.innerHTML = playerName; ❌
```

---

### Additional Input Validation

**Email validation (if added):**

```go
func isValidEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}
```

**Username validation:**

```go
func isValidUsername(username string) bool {
    // Only alphanumeric and underscore, 3-30 characters
    re := regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
    return re.MatchString(username)
}
```

---

## 4.3. Network-Layer Security (DDoS & Botting)

### Encryption (WSS)

Use **WSS (WebSocket Secure)** instead of WS. This encrypts all traffic (like TLS for WebSockets) and prevents man-in-the-middle attacks.

**Server Configuration:**

```go
// Development (ws://)
http.ListenAndServe(":8080", nil)

// Production (wss://)
certFile := "/path/to/cert.pem"
keyFile := "/path/to/key.pem"
http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
```

**Client Connection:**

```javascript
// Automatically use wss:// in production
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsURL = `${protocol}//${window.location.host}/ws`;
const socket = new WebSocket(wsURL);
```

---

### Rate Limiting

Implement rate limiting at the Go server (or at a load-balancer level).

#### Connection Rate Limiting

Limit the number of WebSocket connection attempts from a single IP address.

**Implementation:**

```go
import (
    "golang.org/x/time/rate"
    "sync"
)

type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  sync.RWMutex
    r   rate.Limit
    b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        r:   r,
        b:   b,
    }
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(i.r, i.b)
        i.ips[ip] = limiter
    }

    return limiter
}

// Usage
var limiter = NewIPRateLimiter(rate.Limit(10), 20) // 10 req/sec, burst 20

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getIP(r)
        limiter := limiter.GetLimiter(ip)

        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            log.Printf("Rate limit exceeded for IP: %s", ip)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**Configuration:**
- **Connection Limiting:** 10 connection attempts per minute per IP
- **Burst:** Allow burst of 20 connections

---

#### Message Rate Limiting

Limit the number of messages a single client can send.

```go
type ClientRateLimiter struct {
    limiter *rate.Limiter
}

func NewClientRateLimiter() *ClientRateLimiter {
    // 100 messages per second with burst of 200
    return &ClientRateLimiter{
        limiter: rate.NewLimiter(rate.Limit(100), 200),
    }
}

func (c *ClientRateLimiter) Allow() bool {
    return c.limiter.Allow()
}

// In readPump
func readPump(player *Player) {
    rateLimiter := NewClientRateLimiter()

    for {
        var msg Message
        err := player.Conn.ReadJSON(&msg)
        if err != nil {
            break
        }

        // Check rate limit
        if !rateLimiter.Allow() {
            log.Printf("Player %d exceeded message rate limit", player.ID)
            player.Conn.Close()
            break
        }

        // Process message
        handleMessage(player, msg)
    }
}
```

**Configuration:**
- **Message Limiting:** 100 messages per second per client
- **Burst:** Allow burst of 200 messages

---

### Bot Prevention (Future)

If automated bots become a problem (e.g., creating thousands of players to lag the server), a simple, non-intrusive CAPTCHA can be added to the Main Menu "RUN" button in a future phase.

**Options:**
- **hCaptcha:** Privacy-focused, GDPR compliant
- **Cloudflare Turnstile:** Invisible CAPTCHA
- **reCAPTCHA v3:** Invisible, score-based

**Implementation (Phase 5+):**

```html
<!-- Main Menu -->
<button id="runButton" disabled>RUN</button>

<script src="https://js.hcaptcha.com/1/api.js" async defer></script>
<div class="h-captcha" data-sitekey="YOUR_SITE_KEY" data-callback="onCaptchaSuccess"></div>

<script>
function onCaptchaSuccess(token) {
    document.getElementById('runButton').disabled = false;
    captchaToken = token;
}
</script>
```

**Server Verification:**

```go
func verifyCaptcha(token string) bool {
    resp, err := http.PostForm("https://hcaptcha.com/siteverify", url.Values{
        "secret":   {os.Getenv("HCAPTCHA_SECRET")},
        "response": {token},
    })
    // ... check response
}
```

---

## 4.4. Data Privacy & Compliance

### Data Collection

**What We Collect:**
- Player nickname (self-provided, no PII)
- Game scores and timestamps
- IP address (for rate limiting only, not stored long-term)

**What We DON'T Collect:**
- Email addresses (unless registration added)
- Payment information (game is free)
- Personal identifiable information

### GDPR Compliance (if applicable)

**Right to Erasure:**

```sql
-- Delete all scores for a player
DELETE FROM scores WHERE player_name = $1;

-- Delete from Redis
DEL player:12345
ZREM leaderboard:current "PlayerName"
```

**Data Retention:**
- Redis: 24 hours (automatic expiration)
- PostgreSQL: Indefinite (for leaderboards)

---

## 4.5. Logging & Monitoring

### Security Logging

Log all security-relevant events:

```go
// Failed login attempts
log.Printf("SECURITY: Failed admin login from IP: %s", ip)

// Rate limit violations
log.Printf("SECURITY: Rate limit exceeded for IP: %s", ip)

// Input validation failures
log.Printf("SECURITY: XSS attempt detected from player %d: %s", player.ID, rawInput)

// Suspicious behavior
log.Printf("SECURITY: Player %d attempted impossible jump (airborne)", player.ID)
```

### Monitoring Alerts

Set up alerts for:
- Sudden spike in connections (DDoS indicator)
- High rate of validation failures (attack indicator)
- Server tick time > 100ms (performance issue)
- Database connection failures

---

## 4.6. Deployment Security Checklist

### Before Production Deployment

- [ ] Enable WSS (WebSocket Secure) with valid TLS certificate
- [ ] Set strong admin panel password (environment variable)
- [ ] Enable rate limiting (connection + message)
- [ ] Configure Redis password authentication
- [ ] Configure PostgreSQL SSL mode (require)
- [ ] Review and sanitize all user inputs
- [ ] Set secure CORS policy (not allow-all)
- [ ] Disable debug endpoints in production
- [ ] Enable security logging
- [ ] Set up monitoring and alerts
- [ ] Review environment variables (no secrets in code)
- [ ] Implement firewall rules (only ports 80, 443, 8443 open)
- [ ] Regular security updates (Go, libraries, OS)

---

## Related Documentation

- Frontend Implementation: docs/04-technical-architecture/frontend-pixijs.md
- Backend Implementation: docs/04-technical-architecture/backend-go.md
- Network Protocol: docs/04-technical-architecture/network-protocol.md
- Database Schema: docs/04-technical-architecture/database-schema.md
