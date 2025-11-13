# Phase 5: The Full Game Loop

**Version:** 1.5 **Date:** 2025-11-13

## Goal
Implement the complete player experience from start to finish.

## Tasks

### Main Menu / Splash Screen
* Create the **Main Menu / Splash Screen** (HTML/CSS).
* Implement the "Join Game" flow (enter name, click "RUN", connect WebSocket).

### Death & Respawn Cycle
* Server detects player death and sends `{"e": "death", ...}`.
* Client receives death message, stops the game, and shows a "Game Over" UI.
* Client's "RUN AGAIN" button disconnects and reconnects (or sends a new "join" message).

### Leaderboard
* Connect Go server to **Redis**.
* On death, write `(score, name)` to the Redis Sorted Set.
* Create a new server endpoint (or message) for clients to fetch the Top 10 from Redis.
* Display the leaderboard in the client UI.

## Assets Required

### Logo
* **`logo-viberunner.svg`**: The main game title/logo for the splash screen.

### UI Assets (HTML/CSS/SVG)
* `ui-frame.svg` (A 9-slice "terminal" frame for UI boxes).
* `ui-button-run.svg` (The main "RUN" button).
* `ui-button-reboot.svg` (The "RUN AGAIN" / "REBOOT" button).
* `ui-text-input.css` (Styling for the "Enter Name" box to match the theme).
* `ui-fatal-error-text.svg` (Optional: The `FATAL ERROR` death message as a graphic).

*Note: This phase builds the complete UI shell. Refer to docs/03-art-style-aesthetics.md Section 3.6 (UI & HUD) for the "80s Computer Terminal" aesthetic.*

## Success Criteria

### Main Menu
- Splash screen displays game logo and title
- Player name input field present and functional
- "RUN" button triggers WebSocket connection
- Boot sequence UI shows `[SYSTEM BOOTING...]` and `[ACCESSING VIBE_GRID...]`
- Smooth transition from menu to gameplay

### Death & Respawn
- Server detects collision and marks player as dead
- Server sends targeted death message to player
- Client receives death message and stops local simulation
- Death screen displays with final score
- Death screen shows fake system crash: `FATAL ERROR: 0xDEADBEEF`, `PLAYER_PROCESS_TERMINATED`
- "RUN AGAIN" / "REBOOT" button returns player to game
- Respawn creates new player instance

### Leaderboard
- Redis connection established successfully
- Player score saved to Redis sorted set on death
- Player score saved to PostgreSQL for persistence
- Top 10 leaderboard fetches from Redis
- Leaderboard displays in game UI with player names and scores
- Leaderboard updates in real-time as players die
- Player name sanitization prevents XSS attacks

## Technical Notes

**Main Menu Flow:**
```
1. Page loads -> Show splash screen
2. User enters name -> Validate and sanitize
3. User clicks "RUN" -> Show boot sequence
4. Establish WebSocket connection
5. Send join message
6. Receive welcome message
7. Transition to gameplay
```

**Death Flow:**
```
Server:
1. Detect collision in game loop
2. Set player.isAlive = false
3. Calculate final score (time survived)
4. Save to Redis: ZADD leaderboard:current <score> <name>
5. Save to PostgreSQL: INSERT INTO scores (...)
6. Send: {"e": "death", "d": {"s": 120.5}}

Client:
1. Receive death message
2. Stop local simulation
3. Play death animation/effect
4. Show death screen with score
5. Display leaderboard
6. Wait for "RUN AGAIN" input
```

**Leaderboard Query:**
```redis
# Get top 10 players with scores
ZREVRANGE leaderboard:current 0 9 WITHSCORES
```

Refer to:
- docs/04-technical-architecture/database-schema.md for Redis/PostgreSQL details
- docs/04-technical-architecture/security.md for input sanitization
- docs/03-art-style-aesthetics.md Section 3.6 for UI styling
