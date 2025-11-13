# Network Protocol: WebSockets (JSON)

**Version:** 1.5 **Date:** 2025-11-13

## Protocol Overview

All client-server communication happens over **WebSocket** connections using **JSON-encoded messages**. Messages use short keys (e.g., `e` = event, `d` = data) to minimize bandwidth.

### Message Structure

All messages follow this structure:

```json
{
  "e": "<event_type>",
  "d": { /* event data */ }
}
```

- **`e` (event):** String identifying the message type
- **`d` (data):** Object containing event-specific data

## Client-to-Server Messages (C->S)

### Join Game

Sent once after the WebSocket connects to register the player.

**Event:** `join`

```json
{
  "e": "join",
  "d": {
    "n": "VibeKing"
  }
}
```

**Fields:**
- `n` (name): Player's chosen nickname (max 30 characters)

**Server Response:** `welcome` message

---

### Player Input (Jump)

Sent every time the player presses the jump button.

**Event:** `jump`

```json
{
  "e": "jump",
  "d": {
    "t": 1678886400123
  }
}
```

**Fields:**
- `t` (timestamp): Client timestamp in milliseconds (from `Date.now()`)

**Purpose:** The timestamp helps the server validate input timing and detect potential cheating.

---

### Ping (Optional)

Sent to measure network latency.

**Event:** `ping`

```json
{
  "e": "ping",
  "d": {
    "t": 1678886400123
  }
}
```

**Fields:**
- `t` (timestamp): Client timestamp in milliseconds

**Server Response:** `pong` message with same timestamp

---

## Server-to-Client Messages (S->C)

### Welcome / Handshake

Sent once to a new player immediately after they connect and send a `join` message.

**Event:** `welcome`

```json
{
  "e": "welcome",
  "d": {
    "id": 12345,
    "seed": "vibe-runner-12345",
    "serverTime": 1678886400500
  }
}
```

**Fields:**
- `id`: The player's unique ID for this session (integer)
- `seed`: The level's master seed (string)
- `serverTime`: Current server timestamp in milliseconds (for clock synchronization)

---

### Game State (Broadcast)

Sent to *all* clients at the server's tick rate (20Hz / every 50ms). This is the most frequent message.

**Event:** `state`

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

**Fields:**
- `t` (timestamp): Server timestamp for this state snapshot (milliseconds)
- `p` (players): Array of player objects
  - `i` (id): Player ID (integer)
  - `x`: Player X position (float)
  - `y`: Player Y position (float)

**Notes:**
- Only *alive* players are included in the broadcast
- Clients use this to update ghost player positions
- Clients reconcile their local player position with the server's position

---

### Player Death (Targeted)

Sent *only* to the player who died (unicast, not broadcast).

**Event:** `death`

```json
{
  "e": "death",
  "d": {
    "s": 120.5
  }
}
```

**Fields:**
- `s` (score): Final score (time survived in seconds, float)

**Client Action:** Display death screen with score and "RUN AGAIN" button

---

### Level Data (Chunk)

Sent to clients as they approach a new, un-generated part of the level.

**Event:** `chunk`

```json
{
  "e": "chunk",
  "d": {
    "id": 10,
    "obs": [
      {"t": 1, "x": 15000},
      {"t": 2, "x": 15100},
      {"t": 3, "x": 15250}
    ]
  }
}
```

**Fields:**
- `id`: Chunk ID (integer)
- `obs` (obstacles): Array of obstacle objects
  - `t` (type): Obstacle type ID (1-3)
    - 1: Tall, thin firewall
    - 2: Low, wide data-block
    - 3: Small spike
  - `x`: Obstacle X position (float)

**Notes:**
- Y position is implicit (ground level: 500)
- Client looks up obstacle dimensions based on type
- Chunks are sent proactively (before player reaches them)

---

### Pong (Optional)

Sent in response to a `ping` message.

**Event:** `pong`

```json
{
  "e": "pong",
  "d": {
    "t": 1678886400123
  }
}
```

**Fields:**
- `t` (timestamp): Echoed timestamp from the `ping` message

**Purpose:** Client calculates RTT (round-trip time) as: `Date.now() - t`

---

### Leaderboard Update (Optional)

Sent periodically or on request to update the leaderboard.

**Event:** `leaderboard`

```json
{
  "e": "leaderboard",
  "d": {
    "top": [
      {"n": "VibeKing", "s": 180.5},
      {"n": "CodeRunner", "s": 156.2},
      {"n": "Glitch", "s": 142.8}
    ]
  }
}
```

**Fields:**
- `top`: Array of top 10 players
  - `n` (name): Player name (string)
  - `s` (score): Player score (float)

---

## Connection Flow

### Initial Connection

```
1. Client opens WebSocket: ws://server/ws
2. WebSocket handshake completes
3. Client sends: {"e": "join", "d": {"n": "PlayerName"}}
4. Server sends: {"e": "welcome", "d": {...}}
5. Server starts sending: {"e": "state", ...} at 20Hz
6. Server sends initial chunks: {"e": "chunk", ...}
```

### Gameplay Loop

```
Client -> Server: {"e": "jump", ...}  (when player presses jump)
Server -> Client: {"e": "state", ...} (20Hz continuous)
Server -> Client: {"e": "chunk", ...} (as player progresses)
```

### Death & Disconnect

```
Server -> Client: {"e": "death", "d": {"s": 120.5}}
Client displays death screen
User clicks "RUN AGAIN"
Client closes WebSocket
Client opens new WebSocket (reconnect)
Flow repeats from "Initial Connection"
```

## Bandwidth Optimization

### Message Frequency (per second)

**From Server (per client):**
- `state`: 20 messages/sec (~200 bytes each) = ~4 KB/sec
- `chunk`: ~0.1 messages/sec (occasional) = ~100 bytes/sec
- **Total:** ~4.1 KB/sec per client

**From Client:**
- `jump`: ~2 messages/sec (varies) = ~100 bytes/sec

### Optimization Strategies

1. **Short Keys:** Use single-letter keys (`e`, `d`, `i`, `x`, `y`) instead of full words
2. **Efficient Encoding:** Consider binary protocols (MessagePack, Protocol Buffers) for production
3. **Delta Compression:** Send only changed positions (future optimization)
4. **Client Prediction:** Reduces need for high-frequency server updates
5. **Entity Interpolation:** Smooths 20Hz updates to appear 60Hz on client

## Error Handling

### Connection Lost

- Client detects WebSocket `onclose` event
- Display "Connection Lost" message
- Attempt automatic reconnection (3 retries)
- If all retries fail, return to main menu

### Invalid Messages

- Server validates all incoming messages
- Invalid messages are logged and ignored
- Malformed JSON closes the connection
- Rate limiting prevents spam

## Security Considerations

- All messages must be validated server-side
- Player names sanitized to prevent XSS
- Jump timing validated to prevent speed hacks
- Rate limiting on all message types
- Use WSS (WebSocket Secure) in production

See docs/04-technical-architecture/security.md for detailed security specifications.

## Related Documentation

- Frontend Implementation: docs/04-technical-architecture/frontend-pixijs.md
- Backend Implementation: docs/04-technical-architecture/backend-go.md
- Security: docs/04-technical-architecture/security.md
