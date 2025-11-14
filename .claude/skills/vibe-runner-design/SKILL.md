---
name: vibe-runner-design
description: Frontend design guidance for implementing Vibe Runner's Hyper-Synthwave aesthetic. Use when creating UI components, styling elements, implementing visual effects, or making design decisions. Enforces 80s sci-fi terminal aesthetic with neon cyan/pink/purple colors, CRT effects, and over-the-top visual flair.
allowed-tools:
  - Read
  - Edit
  - Write
---

# Vibe Runner Design Skill

You are a specialist in implementing the Vibe Runner "Hyper-Synthwave" aesthetic - an over-the-top, modern interpretation of 80s sci-fi and computer terminal design.

## Core Design Philosophy

**"If it feels subtle, it's not enough."**

- More bloom, more glow, more neon
- Dark foundations make neon highlights explode
- Modern techniques (WebGL, shaders) to create an 80s dream
- Cyan and pink are your primaries - everything else supports them

**Key References:**
- *Far Cry 3: Blood Dragon* - Satirical 80s excess
- *Katana ZERO* - Hi-bit pixel art with neon lighting
- *Hotline Miami* - Bold color palette
- *Alien: Isolation* - CRT computer UI

---

## 1. Color Palette (EXACT VALUES REQUIRED)

### Primary Colors

**Electric Pink** - `#ff007f`
- **Use for:** Headers, titles, game logo, primary buttons, player visor, key highlights
- **Glow:** Strong bloom effect (20-40px spread)
- **Never:** Body text, backgrounds

**Hyper-Cyan** - `#00f0ff`
- **Use for:** Primary UI text, borders, ground platform glow, interactive elements, score displays
- **Glow:** Medium bloom effect (10-20px spread)
- **This is your main text color** (replaces traditional white)

**Magenta** - `#ff00ff`
- **Use for:** Emphasis text, important values, code snippet collectibles, special effects
- **Glow:** Medium bloom effect (10-20px spread)
- **Never:** Large text blocks

### Secondary Colors

**Violet Glow** - `#b388ff`
- **Use for:** Subtitles, descriptions, player names above ghosts, secondary information
- **Glow:** Subtle bloom effect (5-10px spread)

**Deep Purple** - `#301a4b`
- **Use for:** UI backgrounds, player suit, semi-transparent overlays, container fills
- **Glow:** None (this is a base color)

**Deep Indigo** - `#1a1a2e`
- **Use for:** Dark backgrounds, main canvas background, deep shadows
- **Glow:** None (this is a base color)

### Accent Colors (Use Sparingly)

**Solar Orange** - `#ff8c00`
- **Use for:** Collectibles only, rare highlights, warning states
- **Limited use:** Less than 5% of design elements

**Glitch Red** - `#ff003c`
- **Use for:** Obstacles ONLY, death effects, error messages
- **This is the danger color**

### NEVER USE

❌ Any shade of green (no #00ff00, #33ff00, #39ff14, etc.)
❌ Pastel or muted tones
❌ Pure white (#ffffff) - use Hyper-Cyan (#00f0ff) instead
❌ Pure black (#000000) - use Deep Indigo (#1a1a2e) instead
❌ Yellow or lime shades
❌ Brown, beige, or earth tones

---

## 2. Typography

### Font Families

**VT323** (Google Fonts)
- **Use for:** Primary UI text, labels, descriptions, body copy
- **Color:** Hyper-Cyan (#00f0ff)
- **Size:** 12-16px for body, 18-24px for labels
- **Effect:** Subtle cyan glow

**Press Start 2P** (Google Fonts)
- **Use for:** Headers, screen titles, game logo, major labels
- **Color:** Electric Pink (#ff007f)
- **Size:** 24-48px
- **Effect:** Strong pink glow/bloom

**Courier New** (System fallback)
- **Use for:** Fallback monospace, debug info
- **Color:** Hyper-Cyan (#00f0ff)

**JetBrains Mono** (Google Fonts)
- **Use for:** Code snippet collectibles, technical displays
- **Color:** Magenta (#ff00ff)
- **Effect:** Monospace with slight glow

### Typography Hierarchy

```css
/* Main Title */
.title-main {
  font-family: 'Press Start 2P', monospace;
  color: #ff007f; /* Electric Pink */
  font-size: 48px;
  text-shadow: 0 0 20px #ff007f, 0 0 40px #ff007f, 0 0 60px #ff007f;
}

/* Screen Headers */
.header {
  font-family: 'Press Start 2P', monospace;
  color: #ff007f; /* Electric Pink */
  font-size: 24px;
  text-shadow: 0 0 15px #ff007f, 0 0 30px #ff007f;
}

/* Primary UI Text */
.text-primary {
  font-family: 'VT323', monospace;
  color: #00f0ff; /* Hyper-Cyan */
  font-size: 16px;
  text-shadow: 0 0 10px #00f0ff, 0 0 20px #00f0ff;
}

/* Secondary Info */
.text-secondary {
  font-family: 'VT323', monospace;
  color: #b388ff; /* Violet Glow */
  font-size: 14px;
  text-shadow: 0 0 8px #b388ff;
}

/* Emphasis/Values */
.text-emphasis {
  font-family: 'VT323', monospace;
  color: #ff00ff; /* Magenta */
  font-size: 16px;
  text-shadow: 0 0 12px #ff00ff, 0 0 24px #ff00ff;
  font-weight: bold;
}

/* Error/Death Messages */
.text-error {
  font-family: 'Press Start 2P', monospace;
  color: #ff003c; /* Glitch Red */
  font-size: 18px;
  text-shadow: 0 0 15px #ff003c, 0 0 30px #ff003c;
}
```

### Text Effects

All text should have:
- Scanline overlay (subtle, 1px horizontal lines)
- Appropriate glow/bloom for its color
- Sharp, pixel-perfect rendering (no anti-aliasing for retro fonts)

---

## 3. Motion & Animation

### Core Animations

**Scanline Flicker** (continuous, subtle)
```css
@keyframes scanlines {
  0% { opacity: 0.1; }
  50% { opacity: 0.15; }
  100% { opacity: 0.1; }
}

.scanline-overlay {
  background: repeating-linear-gradient(
    0deg,
    rgba(0, 0, 0, 0.15),
    rgba(0, 0, 0, 0.15) 1px,
    transparent 1px,
    transparent 2px
  );
  animation: scanlines 0.1s infinite;
}
```

**Glow Pulse** (interactive elements)
```css
@keyframes glowPulse {
  0%, 100% {
    filter: drop-shadow(0 0 5px #00f0ff) drop-shadow(0 0 10px #00f0ff);
  }
  50% {
    filter: drop-shadow(0 0 15px #00f0ff) drop-shadow(0 0 30px #00f0ff);
  }
}

.interactive-element:hover {
  animation: glowPulse 1.5s ease-in-out infinite;
}
```

**Text Cursor Blink**
```css
@keyframes cursorBlink {
  0%, 49% { opacity: 1; }
  50%, 100% { opacity: 0; }
}

.cursor {
  animation: cursorBlink 1s step-end infinite;
  color: #00f0ff;
}
```

**Chromatic Aberration** (on hit/death)
```css
@keyframes chromaticAberration {
  0%, 100% { filter: none; }
  25% { filter: drop-shadow(2px 0 0 #ff007f) drop-shadow(-2px 0 0 #00f0ff); }
  50% { filter: drop-shadow(-2px 0 0 #ff007f) drop-shadow(2px 0 0 #00f0ff); }
  75% { filter: drop-shadow(2px 0 0 #ff007f) drop-shadow(-2px 0 0 #00f0ff); }
}

.hit-effect {
  animation: chromaticAberration 0.2s ease-in-out;
}
```

**VHS Tracking Glitch** (ghost players)
```css
@keyframes vhsGlitch {
  0%, 90% { transform: translateX(0); }
  92% { transform: translateX(2px); }
  94% { transform: translateX(-2px); }
  96% { transform: translateX(1px); }
  100% { transform: translateX(0); }
}

.ghost-player {
  animation: vhsGlitch 3s infinite;
}
```

### Animation Principles

**DO:**
- Quick, snappy transitions (0.1-0.3s)
- Digital, glitchy effects
- Continuous subtle movement (scanlines, glow pulse)
- Chromatic aberration on impacts
- VHS tracking effects

**DON'T:**
- Smooth, Apple-style animations
- Long transitions (>0.5s for UI)
- Bouncy/spring easing (too organic)
- Slow fades
- Realistic physics

---

## 4. Backgrounds & Parallax

### Parallax Layer Structure (back to front)

**Layer 1: Sky + Sunset** (slowest scroll, 0.2x speed)
- Deep purple (#301a4b) to indigo (#1a1a2e) gradient
- Massive OutRun-style sun: magenta (#ff00ff) circle with horizontal scan lines
- Low on horizon

**Layer 2: Cityscape Silhouette** (0.4x speed)
- Pure black (#000000) silhouette
- Futuristic 80s metropolis skyline
- Tiny window lights in pink (#ff007f) and cyan (#00f0ff)

**Layer 3: Floating Geometry** (0.6x speed)
- Chrome-reflective geometric shapes (pyramids, spheres)
- 80s CGI aesthetic
- Reflect scene's neon lights
- Subtle rotation/float

**Layer 4: Foreground Structures** (0.8x speed)
- Massive server racks with randomly blinking LEDs
- Detailed pixel art
- LEDs in cyan, pink, violet

**Layer 5: Ground Platform** (1.0x speed, matches game speed)
- Dark, near-black surface (#1a1a2e)
- Sharp glowing cyan (#00f0ff) top edge
- Strong bloom/glow effect casting light on player

### Background Implementation Pattern

```javascript
// Pixi.js parallax example
const layers = [
  { sprite: skyLayer, speed: 0.2 },
  { sprite: cityLayer, speed: 0.4 },
  { sprite: geometryLayer, speed: 0.6 },
  { sprite: rackLayer, speed: 0.8 },
  { sprite: groundLayer, speed: 1.0 }
];

function updateParallax(scrollSpeed) {
  layers.forEach(layer => {
    layer.sprite.x -= scrollSpeed * layer.speed;
    // Wrap around when off-screen
    if (layer.sprite.x <= -layer.sprite.width) {
      layer.sprite.x = 0;
    }
  });
}
```

### Visual Effects Requirements

- **Heavy bloom** on all light sources
- **Scanline overlay** on entire screen (1px, subtle)
- **Chromatic aberration** on fast-moving objects
- **Particle trails** on player (neon cyan glow)

---

## 5. UI Components

### Containers

```css
.ui-container {
  background: rgba(26, 26, 46, 0.7); /* Deep Indigo with transparency */
  border: 1px solid #00f0ff; /* Hyper-Cyan */
  box-shadow: 0 0 20px rgba(0, 240, 255, 0.5); /* Cyan glow */
  border-radius: 0; /* Sharp corners only */
}

/* Optional: ASCII border */
.ui-container-fancy::before {
  content: "╔═══════════╗";
  color: #00f0ff;
  font-family: 'Courier New', monospace;
}
```

### Buttons

```css
.button-primary {
  font-family: 'Press Start 2P', monospace;
  color: #00f0ff; /* Hyper-Cyan */
  background: transparent;
  border: 2px solid #00f0ff;
  padding: 10px 20px;
  text-shadow: 0 0 10px #00f0ff;
  cursor: pointer;
  position: relative;
}

.button-primary::before {
  content: "[> ";
}

.button-primary::after {
  content: " ]";
}

.button-primary:hover {
  color: #ff007f; /* Electric Pink on hover */
  border-color: #ff007f;
  text-shadow: 0 0 15px #ff007f, 0 0 30px #ff007f;
  animation: glowPulse 1s ease-in-out infinite;
}

.button-primary:active {
  animation: chromaticAberration 0.1s ease-in-out;
}
```

### Input Fields

```css
.text-input {
  font-family: 'VT323', monospace;
  font-size: 18px;
  color: #00f0ff; /* Hyper-Cyan */
  background: rgba(26, 26, 46, 0.9); /* Deep Indigo */
  border: 1px solid #00f0ff;
  padding: 10px;
  text-shadow: 0 0 10px #00f0ff;
}

.text-input::placeholder {
  color: #b388ff; /* Violet Glow */
  opacity: 0.6;
}

.text-input:focus {
  outline: none;
  border-color: #ff007f; /* Pink when focused */
  box-shadow: 0 0 20px rgba(255, 0, 127, 0.5);
}

/* Blinking cursor */
.text-input::after {
  content: "█";
  animation: cursorBlink 1s step-end infinite;
  color: #00f0ff;
}
```

### Leaderboard

```css
.leaderboard {
  background: rgba(26, 26, 46, 0.8);
  border: 1px solid #00f0ff;
}

.leaderboard-header {
  font-family: 'Press Start 2P', monospace;
  color: #ff007f; /* Electric Pink */
  font-size: 18px;
  text-shadow: 0 0 15px #ff007f;
}

.leaderboard-row {
  font-family: 'VT323', monospace;
  font-size: 16px;
}

.leaderboard-rank {
  color: #b388ff; /* Violet Glow */
}

.leaderboard-name {
  color: #00f0ff; /* Hyper-Cyan */
  text-shadow: 0 0 10px #00f0ff;
}

.leaderboard-score {
  color: #ff00ff; /* Magenta */
  text-shadow: 0 0 12px #ff00ff;
  font-weight: bold;
}
```

---

## 6. Specific Screen Implementations

### Main Menu / Splash Screen

```
╔═══════════════════════════════════╗
║                                   ║
║     VIBE RUNNER                   ║  <- Electric Pink (#ff007f), 48px
║                                   ║
║  [SYSTEM BOOTING...]              ║  <- Hyper-Cyan (#00f0ff), blinking
║  [ACCESSING VIBE_GRID...]         ║  <- Hyper-Cyan (#00f0ff)
║                                   ║
║  Enter Name: ████                 ║  <- Hyper-Cyan (#00f0ff), blinking cursor
║                                   ║
║     [> RUN ]                      ║  <- Hyper-Cyan (#00f0ff) button
║                                   ║
╚═══════════════════════════════════╝
```

### Death Screen

```
╔═══════════════════════════════════╗
║                                   ║
║  FATAL ERROR: 0xDEADBEEF          ║  <- Glitch Red (#ff003c)
║  PLAYER_PROCESS_TERMINATED        ║  <- Glitch Red (#ff003c)
║                                   ║
║  Time Survived: 120.5s            ║  <- Magenta (#ff00ff), large
║                                   ║
║  [Press 'R' to REBOOT]            ║  <- Hyper-Cyan (#00f0ff)
║                                   ║
╚═══════════════════════════════════╝

[Chromatic aberration effect on entire screen]
[Red flash overlay, fading out]
```

### In-Game HUD

```
Score: 045.2s            <- Magenta (#ff00ff), top-left
Rank: #7                 <- Violet Glow (#b388ff), below score

┌─ TOP 10 ─────────┐    <- Hyper-Cyan (#00f0ff) border, top-right
│ 1. Player  182.5 │    <- Rank: Violet, Name: Cyan, Score: Magenta
│ 2. Vibe    156.2 │
│ 3. Runner  142.8 │
└──────────────────┘
```

---

## 7. Pixi.js-Specific Implementation

### Bloom/Glow Filters

```javascript
// Apply bloom to neon elements
const bloomFilter = new PIXI.filters.BloomFilter({
  strength: 2,
  quality: 10,
  resolution: window.devicePixelRatio,
  kernelSize: 5
});

player.filters = [bloomFilter];
uiText.filters = [bloomFilter];
```

### Scanline Overlay

```javascript
// Create scanline texture
const scanlineGraphics = new PIXI.Graphics();
for (let i = 0; i < app.screen.height; i += 2) {
  scanlineGraphics.lineStyle(1, 0x000000, 0.15);
  scanlineGraphics.moveTo(0, i);
  scanlineGraphics.lineTo(app.screen.width, i);
}

const scanlineTexture = app.renderer.generateTexture(scanlineGraphics);
const scanlines = new PIXI.Sprite(scanlineTexture);
scanlines.alpha = 0.1;
app.stage.addChild(scanlines);
```

### Chromatic Aberration Shader

```glsl
// Fragment shader for chromatic aberration
precision mediump float;

varying vec2 vTextureCoord;
uniform sampler2D uSampler;
uniform float aberration;

void main() {
  vec2 offset = vec2(aberration, 0.0);

  float r = texture2D(uSampler, vTextureCoord + offset).r;
  float g = texture2D(uSampler, vTextureCoord).g;
  float b = texture2D(uSampler, vTextureCoord - offset).b;

  gl_FragColor = vec4(r, g, b, 1.0);
}
```

---

## 8. Implementation Checklist

When implementing any UI element, verify:

- [ ] Uses exact color codes from palette (no approximations)
- [ ] Text uses appropriate font family (VT323 or Press Start 2P)
- [ ] Text has appropriate glow/shadow for its color
- [ ] No green anywhere in the design
- [ ] Containers have sharp corners (no border-radius)
- [ ] Interactive elements have glow pulse on hover
- [ ] Scanline overlay is applied
- [ ] Animations are snappy (<0.3s)
- [ ] Bloom/glow effects are applied to neon elements
- [ ] Backgrounds use parallax (if applicable)
- [ ] Over-the-top visual flair (if it feels subtle, add more)

---

## 9. Common Mistakes to Avoid

❌ **Using green** - This palette has no green!
❌ **Rounded corners** - Sharp angles only for terminal aesthetic
❌ **Smooth animations** - Keep it digital and glitchy
❌ **Subtle effects** - Go bold or go home
❌ **White text** - Use Hyper-Cyan (#00f0ff) instead
❌ **Black backgrounds** - Use Deep Indigo (#1a1a2e) instead
❌ **Generic fonts** - No Inter, Roboto, Arial, Helvetica
❌ **Pastel colors** - Full saturation neon only
❌ **Minimal shadows** - Aggressive glows on everything neon
❌ **Professional polish** - We want over-the-top, not corporate

---

## When NOT to Use This Skill

This skill is for **frontend visual design only**. Do not use for:
- Backend/server code
- Game logic/physics
- Network protocol implementation
- Database queries
- Security implementations

For those topics, use the main `vibe-runner` skill instead.

---

## Reference Documentation

For complete specifications, always refer to:
- `docs/03-art-style-aesthetics.md` - Full visual identity guide
- `docs/04-technical-architecture/frontend-pixijs.md` - Pixi.js implementation patterns

**Remember:** Documentation is the source of truth. This skill provides practical implementation guidance, but defer to docs for authoritative specifications.
