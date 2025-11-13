# Phase 6: Polish & Tooling

**Version:** 1.5 **Date:** 2025-11-13

## Goal
Add final aesthetics, audio, and developer tools.

## Tasks

### Visual Polish
* Implement the 80s Sci-Fi theme: parallax backgrounds, phosphor UI, etc.
* Add player animations (run cycle, jump, death).
* Add particle effects and shaders (scanlines, bloom, trails).

### Audio Integration
* Integrate the provided `wav/mp3` soundtrack.
* Add sound effects for jump, death, and collectibles.

### Developer Tools
* Implement the **Developer Debug HUD** (`?debug=true`) to show hitboxes and ping.
* Create the **Load Testing Script** (Go/Python) to simulate 500+ users.
* Create the basic **Admin Panel** to view server stats.

## Assets Required

### Player Animation (Spritesheets)
* `player-run-spritesheet.png` (Full multi-frame run cycle).
* `player-jump-spritesheet.png` (Sprites for jump, peak, and fall).
* `player-die-effect.png` (Spritesheet for the "glitch-out" death animation).

### Parallax Backgrounds (High-res PNGs)
* `bg-layer-01-sky.png` (Farthest layer: Sky, OutRun sun).
* `bg-layer-02-skyline.png` (Silhouette city).
* `bg-layer-03-midground.png` (Floating chrome shapes).
* `bg-layer-04-foreground.png` (Scrolling server racks).

### Audio (User-provided)
* `soundtrack-loop.mp3`
* `sfx-jump.wav`
* `sfx-die.wav`
* `sfx-collect.wav`

### Effects & Shaders (GLSL / PNG)
* `shader-scanlines.glsl` (GLSL code for the UI/screen overlay).
* `shader-bloom.glsl` (GLSL code for the neon glow).
* `particle-trail.png` (Texture for the player's neon foot trail).

*Note: This is the "make it pretty" phase. Refer to docs/03-art-style-aesthetics.md for all visual and atmospheric details.*

## Success Criteria

### Visual Polish
- Parallax background layers scroll at different speeds
- Player sprite uses animated spritesheet (not static)
- Player run cycle plays smoothly during gameplay
- Jump animation transitions (ground -> jump -> peak -> fall -> ground)
- Death animation plays on collision (glitch-out effect)
- Scanline shader overlays entire screen or UI
- Bloom shader creates neon glow on key elements
- Particle trail follows player's feet
- Ghost players have VHS tracking glitch effect
- Ground platform has glowing cyan edge with bloom
- Obstacles have chromatic aberration and pixel-sorting effects
- UI uses phosphor green text with terminal styling

### Audio Integration
- Background music loops seamlessly
- Music starts on game start (or after user interaction for autoplay policy)
- Jump sound plays on player input
- Death sound plays on collision
- Collectible sound plays when collecting code snippets
- Audio volume is reasonable and balanced
- Audio can be muted/controlled by player

### Developer Tools
- Debug HUD activates with `?debug=true` URL parameter
- Debug HUD shows collision boxes around player and obstacles
- Debug HUD displays network ping (time to receive pong)
- Debug HUD shows server position vs client position
- Load testing script accepts command-line arguments (-c for count, -url for server)
- Load test creates specified number of WebSocket connections
- Load test bots send realistic join and jump messages
- Load test logs connection failures and latency
- Admin panel accessible via separate route/port
- Admin panel shows current player count
- Admin panel shows server health metrics
- Admin panel displays recent logs

## Technical Notes

**Parallax Implementation:**
```javascript
// Different scroll speeds for depth
layer1.x -= scrollSpeed * 0.2; // Farthest = slowest
layer2.x -= scrollSpeed * 0.4;
layer3.x -= scrollSpeed * 0.6;
layer4.x -= scrollSpeed * 0.8;
ground.x -= scrollSpeed * 1.0;  // Closest = fastest
```

**Debug HUD Activation:**
```javascript
const urlParams = new URLSearchParams(window.location.search);
const debugMode = urlParams.get('debug') === 'true';
```

**Load Test Script:**
```bash
./load-test -c=500 -url=ws://localhost:8080/ws
```

**Shader Application:**
```javascript
// Scanline shader on container
const scanlineFilter = new PIXI.Filter(null, scanlineShaderGLSL);
app.stage.filters = [scanlineFilter];

// Bloom on specific sprites
const bloomFilter = new PIXI.filters.BloomFilter();
player.filters = [bloomFilter];
```

Refer to:
- docs/03-art-style-aesthetics.md for complete visual specifications
- docs/04-technical-architecture/developer-tooling.md for debug tools
- docs/04-technical-architecture/frontend-pixijs.md for shader implementation
