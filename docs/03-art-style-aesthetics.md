# Art Style & Aesthetics

**Version:** 1.5 **Date:** 2025-11-13

This section defines the complete visual identity. The goal is a **"Hyper-Synthwave"** look: a self-aware, over-the-top, and technically modern interpretation of the 80s' vision of the future. It should be fun, absurdly vibrant, and drenched in neon.

**Key References:** *Far Cry 3: Blood Dragon* (for its satirical 80s tone), *Katana ZERO* / *Hyper Light Drifter* (for modern "hi-bit" pixel art and lighting), *Hotline Miami* (for its color palette and grit), *Alien: Isolation* (for its CRT computer UI).

## 3.1. Core Principles

* **Go Over The Top:** More bloom, more glow, more parallax, more neon. If it feels subtle, it's not enough.
* **Modern Techniques:** This is not a "retro" game. We use high-resolution rendering, smooth animations, WebGL shaders, and particle effects to create a *modern* game that *looks* like an 80s dream.
* **Contrast:** The world is built on a foundation of dark, deep colors (midnight blue, dark purple) to make the explosive neon highlights (cyan, pink, orange) "pop."

## 3.2. Color Palette

* **Base:** Deep Indigo (`#1a1a2e`), Dark Purple (`#301a4b`), Near-Black.
* **Primary Neon:** Electric Pink (`#ff007f`), Hyper-Cyan (`#00f0ff`).
* **Secondary Neon:** Magenta (`#ff00ff`), Violet Glow (`#b388ff`), Solar Orange (`#ff8c00` \- collectibles only).
* **Obstacle Color:** Glitch Red (`#ff003c`).
* **UI Text:** Hyper-Cyan (`#00f0ff`) for primary text, Electric Pink (`#ff007f`) for headers, Magenta (`#ff00ff`) for emphasis, Violet Glow (`#b388ff`) for secondary text.

## 3.3. Characters

* **Player Sprite:**
  * **Style:** Detailed "hi-bit" pixel art, designed for fluid animation (Phase 6).
  * **Visuals:** A "data-runner" with a dark suit (`#301a4b`) and glowing `Hyper-Cyan` trace lines.
  * **Key Feature:** An absurdly oversized, glowing **Electric Pink** visor.
  * **Fun Detail:** A comically large 80s hairstyle (e.g., mullet, big perm) that flows dramatically behind them as they run.
* **Ghost Sprites (Other Players):**
  * **Visuals:** A pure black silhouette of the player sprite.
  * **Effect:** The silhouette has a bright, 1px neon outline (player's chosen color, e.g., orange).
  * **Glitch Effect:** The ghost should have a subtle "VHS tracking" shader applied, making it flicker and occasionally have horizontal static lines pass through it, as if it's a "poor recording" of another player.

## 3.4. Environment & Background

The background is high-resolution and uses aggressive **parallax scrolling** to create a sense of depth and speed.

* **Layer 1 (Farthest):** A deep purple sky with a *massive*, low-hanging **OutRun-style sun** (a magenta circle with horizontal scan lines) setting.
* **Layer 2 (Skyline):** A dark silhouette of a futuristic, 80s-style metropolis. Thousands of tiny windows glow in pink and cyan.
* **Layer 3 (Midground):** Giant, chrome-reflective geometric shapes (pyramids, spheres) that slowly float by, reflecting the scene's neon lights. Think 80s CGI art.
* **Layer 4 (Foreground):** Large, detailed pixel-art structures that scroll by, such as the side of a massive, dark server rack with thousands of randomly blinking LEDs.
* **Layer 5 (Ground):** The `ground-tile` platform. A dark, near-black surface with a sharp, glowing `Hyper-Cyan` top edge. This edge should cast a bloom/glow effect onto the player's feet.

## 3.5. Obstacles & Collectibles

* **Obstacles ("Firewalls"):**
  * **Style:** "Digital Glitch" aesthetic.
  * **Visuals:** Not just a red block. It's a `Glitch Red` rectangle, but it constantly flickers, has chromatic aberration at its edges, and has "pixel-sorting" artifacts that drip down from it. Reference the glitch effects in *Katana ZERO*.
* **Collectibles ("Code Snippets"):**
  * **Style:** Self-aware and fun.
  * **Visuals:** Floating, glowing, monospaced text.
  * **Examples:** `<blink>`, `goto 10;`, `// TODO: fix this`, `[object Object]`.

## 3.6. UI & HUD

* **Style:** "80s Computer Terminal."
* **Reference:** The UI from *Alien: Isolation* or *Fallout*'s terminals.
* **Font:** A glowing, monospaced pixel font (like "VT323" or "Press Start 2P" from Google Fonts).
* **Text Colors:**
  * Primary UI text: `Hyper-Cyan` (`#00f0ff`)
  * Headers/Titles: `Electric Pink` (`#ff007f`)
  * Emphasis/Values: `Magenta` (`#ff00ff`)
  * Secondary text: `Violet Glow` (`#b388ff`)
* **Containers:** All UI elements (leaderboard, score) are in dark, 70% transparent black boxes with sharp, 1px `Hyper-Cyan` borders.
* **Effects:** The entire UI (and optionally the whole screen) has a very subtle, gently flickering **scan line** overlay shader.
* **Fun Detail:**
  * **Main Menu:** Shows `[SYSTEM BOOTING...]` and `[ACCESSING VIBE_GRID...]` in Hyper-Cyan.
  * **Death Screen:** Instead of "Game Over," it displays a fake system crash in Glitch Red: `FATAL ERROR: 0xDEADBEEF` `PLAYER_PROCESS_TERMINATED` with `[Press 'R' to REBOOT]` in Hyper-Cyan.
