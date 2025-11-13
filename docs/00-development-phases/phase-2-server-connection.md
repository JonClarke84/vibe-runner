# Phase 2: Basic Server & Client Connection

**Version:** 1.5 **Date:** 2025-11-13

## Goal
Connect a single client to the Go server and control the player's *existence* from the server.

## Tasks

* Set up the Go server with `gorilla/websocket`.
* Implement the WebSocket connection handshake.
* On client connect (`{"e": "join", ...}`), the server will spawn a "player" in its internal state.
* Implement the server's `gameTicker` (e.g., 20Hz).
* The server's game state (containing the single player's position) is broadcast to the client (`{"e": "state", ...}`).
* The client's player sprite is *driven* by the server's state broadcast (i.e., client-side prediction is *off* for now).

## Assets Required

*No new visual assets are required for this phase.*

## Success Criteria

- Go WebSocket server starts successfully
- Client establishes WebSocket connection to server
- Server receives and processes `join` message
- Server spawns player in internal game state
- Server game loop runs at 20Hz
- Server broadcasts state updates to client
- Client player position is controlled by server state (not local input)
- Connection remains stable during gameplay

## Technical Notes

**Server Architecture:**
- Use gorilla/websocket for WebSocket handling
- Implement readPump goroutine per client
- Implement central gameTicker goroutine for game loop
- Server tick rate: 50ms (20Hz)

**Message Flow:**
```
Client -> Server: {"e": "join", "d": {"n": "PlayerName"}}
Server -> Client: {"e": "welcome", "d": {"id": 12345, ...}}
Server -> Client: {"e": "state", "d": {"p": [...]}} (20Hz)
```

Refer to docs/04-technical-architecture/network-protocol.md for detailed message specifications.
