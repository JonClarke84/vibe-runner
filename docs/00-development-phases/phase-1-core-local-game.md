# Phase 1: Core Local Game (Client-Only)

**Version:** 1.5 **Date:** 2025-11-13

## Goal
Create a single-player, local-only version of the game.

## Tasks

* Set up the HTML page with a Pixi.js canvas.
* Create the `Player` class.
* Implement the `gameLoop` (`requestAnimationFrame`).
* Implement basic physics: gravity.
* Implement player input: "Jump" (e.g., Spacebar) that applies an upward velocity.
* Create a `Ground` object and basic collision detection (player vs. ground).
* Create a simple `Obstacle` class and add a few static obstacles to the screen.
* Implement player vs. obstacle collision detection (AABB).
* Implement a simple "death" state (e.g., game loop stops).

## Assets Required (Placeholders)

* **`player-static.svg`**: A static, non-animated sprite for the player.
* **`ground-tile.svg`**: A tileable graphic for the ground platform.
* **`obstacle-firewall.svg`**: A simple, static obstacle.
* **`font-debug.woff2`**: A basic monospaced font to display score/debug info.

*Note: All assets in this phase are functional placeholders. Refer to `Section 3. Art Style & Aesthetics` (docs/03-art-style-aesthetics.md) for visual guidance.*

## Success Criteria

- Player sprite renders on screen
- Player falls with gravity
- Player jumps when spacebar is pressed
- Player collides with ground and stops falling
- Static obstacles render on screen
- Player dies (game stops) when colliding with obstacles
- Basic score/debug info displays on screen
