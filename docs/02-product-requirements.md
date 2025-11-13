# Product Requirements Document (PRD)

**Version:** 1.5 **Date:** 2025-11-13

## 2.1. Game Overview

* **Product Name:** Vibe Runner
* **Concept:** An always-on, massively multiplayer 2D infinite runner. All players run the exact same procedurally generated level simultaneously.
* **Theme:** **80s Sci-Fi & Computer Aesthetic.** A blend of dark, synthwave-infused sci-fi (parallaxing neon cityscapes, distant nebulae) with 80s computer UI elements (glowing phosphor text, scan lines, `[ACCESSING...]` motifs). Player sprites are pixelated, but the overall presentation is modern and high-res.
* **Target Audience:** Players looking for a quick, skill-based challenge; fans of retro-games and the synthwave aesthetic; players who enjoy social, "shared-moment" gaming.
* **Core Loop:**
  1. **Splash Screen:** Player visits the URL and sees a main menu.
  2. **Join:** Player enters a name, clicks "RUN", and connects.
  3. **Spawn:** Player character instantly spawns into the live, in-progress game.
  4. **Survive:** Player character runs automatically. The player uses a single action (**Jump**) to dodge procedurally generated obstacles (themed as "glitches" or "firewalls").
  5. **Die:** Hitting an obstacle results in instant death.
  6. **Rank:** The player's "Time Survived" is their score. A real-time leaderboard shows the top survivors.
  7. **Respawn:** Player is returned to a "Game Over" screen (showing their score) with a "RUN AGAIN" button.

## 2.2. Core Features (MVP)

|  | Feature | Description |
| ----- | ----- | ----- |
|  | **Main Menu / Splash Screen** | A static entry screen with the game title. Includes a text input for "Player Name" and a "RUN" button to join the game. |
|  | **Player Controller** | Single-button **Jump** mechanic. Physics are tight, responsive, and deterministic. |
|  | **Infinite Level** | The level is **deterministically, procedurally generated** by the server. Every player receives the *exact same* sequence of obstacles at the *exact same time*. |
|  | **Massive Multiplayer** | Players see all other active players as "shadow" sprites with their names rendered above. Players pass through each other. |
|  | **Real-Time Leaderboard** | An on-screen UI element shows the Top 10 players ranked by "Time Survived" for the current session. |
|  | **Aesthetics & Audio** | 80s sci-fi theme with parallax scrolling backgrounds. UI elements have a CRT/phosphor-glow look. The provided `wav/mp3` soundtrack will be the core audio. |
|  | **Death & Respawn Cycle** | On death, the player is shown their final score and a "RUN AGAIN" button to instantly respawn. |

## 2.3. Developer & QA Tooling

| Feature | Description |
| ----- | ----- |
| **Debug HUD** | A client-side toggle (e.g., via a hotkey or URL query `?debug=true`) that overlays **collision boxes**, **network ping**, and server vs. client state. |
| **Admin Panel** | A simple, separate web-based dashboard for viewing server health, **current player count**, and server logs in real-time. |
| **Load Testing Script** | A script (e.g., in Go or Python) that can **simulate 500+ concurrent WebSocket clients** to stress-test the backend server and measure performance. |
