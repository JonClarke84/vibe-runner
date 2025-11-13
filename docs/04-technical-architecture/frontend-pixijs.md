# Frontend (Client): Pixi.js

**Version:** 1.5 **Date:** 2025-11-13

## Engine Overview

**Pixi.js** for 2D WebGL rendering. We will write our own simple, deterministic physics (Gravity, Jump Force) for maximum control.

## Game Loop

The core loop will be driven by `requestAnimationFrame`.

```javascript
let lastTime = 0;

function gameLoop(currentTime) {
  const deltaTime = (currentTime - lastTime) / 1000; // in seconds

  update(deltaTime); // Run game logic (physics, input)
  render(deltaTime); // Draw everything

  lastTime = currentTime;
  requestAnimationFrame(gameLoop);
}
```

### Loop Responsibilities

**`update(deltaTime)`:**
- Process player input
- Update local player physics (position, velocity)
- Apply client-side prediction
- Process network messages
- Update ghost player target positions
- Update animations and timers

**`render(deltaTime)`:**
- Interpolate ghost positions
- Update sprite positions
- Update parallax background layers
- Render particles and effects
- Update debug visualizations (if enabled)

## Client-Side Prediction

To feel "instant," the player's jump happens *immediately* on input, *before* the server confirms it.

### Prediction Flow

1. **User Input:** User presses "Spacebar".
2. **Immediate Response:** The client's `Player` object *immediately* gets a `velocityY` applied via `onInput()`.
3. **Server Notification:** A "jump" event is sent to the server with a timestamp:
   ```javascript
   network.send({"e": "jump", "d": {"t": Date.now()}})
   ```
4. **Local Simulation:** The client continues to simulate its own movement.
5. **Server Reconciliation:** If a "correction" packet arrives from the server:
   ```javascript
   {"e": "correction", "d": {"x": 100, "y": 50}}
   ```
   The client will *snap* the player's position to match the server's authoritative state.

### Implementation Details

```javascript
class Player {
  constructor() {
    this.x = 0;
    this.y = 0;
    this.velocityX = 0;
    this.velocityY = 0;
    this.isGrounded = false;
  }

  jump() {
    if (this.isGrounded) {
      this.velocityY = -600; // Upward velocity (negative Y)
      this.isGrounded = false;

      // Notify server
      network.send({
        e: "jump",
        d: { t: Date.now() }
      });
    }
  }

  update(deltaTime) {
    // Apply gravity
    this.velocityY += 1200 * deltaTime; // Gravity acceleration

    // Update position
    this.x += this.velocityX * deltaTime;
    this.y += this.velocityY * deltaTime;

    // Ground collision (simplified)
    if (this.y >= GROUND_Y) {
      this.y = GROUND_Y;
      this.velocityY = 0;
      this.isGrounded = true;
    }
  }

  reconcile(serverX, serverY) {
    // Snap to server position if too far off
    const distanceSquared =
      (this.x - serverX) ** 2 +
      (this.y - serverY) ** 2;

    if (distanceSquared > 100) { // Threshold
      this.x = serverX;
      this.y = serverY;
    }
  }
}
```

## Entity Interpolation (For "Ghosts")

To prevent jittery movement, we *smooth* other players' movements.

### Interpolation Concept

- Each "ghost" sprite has a `lastPosition` and a `targetPosition`.
- When a state packet arrives, we update the `targetPosition`.
- In our `render(deltaTime)` loop, we **linearly interpolate (lerp)** the ghost's *visual* position.

### LERP Function

```javascript
// Example LERP function
function lerp(start, end, amount) {
  return (1 - amount) * start + amount * end;
}

// In the render loop:
ghost.visual_x = lerp(ghost.visual_x, ghost.targetPosition.x, 0.3);
ghost.visual_y = lerp(ghost.visual_y, ghost.targetPosition.y, 0.3);
```

### Ghost Player Implementation

```javascript
class GhostPlayer {
  constructor(id, name) {
    this.id = id;
    this.name = name;

    // Visual position (what's rendered)
    this.visual_x = 0;
    this.visual_y = 0;

    // Target position (from server)
    this.targetPosition = { x: 0, y: 0 };

    // Sprite
    this.sprite = new PIXI.Sprite(ghostTexture);
    this.nameText = new PIXI.Text(name, nameStyle);
  }

  updateTarget(x, y) {
    this.targetPosition.x = x;
    this.targetPosition.y = y;
  }

  render(deltaTime) {
    // Smoothly interpolate toward target
    this.visual_x = lerp(this.visual_x, this.targetPosition.x, 0.3);
    this.visual_y = lerp(this.visual_y, this.targetPosition.y, 0.3);

    // Update sprite position
    this.sprite.x = this.visual_x;
    this.sprite.y = this.visual_y;

    // Position name above sprite
    this.nameText.x = this.visual_x;
    this.nameText.y = this.visual_y - 40;
  }
}
```

## Physics System

### Constants

```javascript
const GRAVITY = 1200;          // pixels/second^2
const JUMP_VELOCITY = -600;    // pixels/second (negative = up)
const PLAYER_SPEED = 300;      // pixels/second (horizontal)
const GROUND_Y = 500;          // Y position of ground
```

### Collision Detection (AABB)

```javascript
function checkCollision(rect1, rect2) {
  return (
    rect1.x < rect2.x + rect2.width &&
    rect1.x + rect1.width > rect2.x &&
    rect1.y < rect2.y + rect2.height &&
    rect1.y + rect1.height > rect2.y
  );
}
```

## Network Message Handling

```javascript
class NetworkManager {
  constructor(url) {
    this.ws = new WebSocket(url);
    this.setupHandlers();
  }

  setupHandlers() {
    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      switch(msg.e) {
        case "welcome":
          this.handleWelcome(msg.d);
          break;
        case "state":
          this.handleState(msg.d);
          break;
        case "death":
          this.handleDeath(msg.d);
          break;
        case "chunk":
          this.handleChunk(msg.d);
          break;
      }
    };
  }

  send(message) {
    this.ws.send(JSON.stringify(message));
  }

  handleState(data) {
    // Update all player positions
    data.p.forEach(playerData => {
      if (playerData.i === myPlayerId) {
        // Reconcile local player
        player.reconcile(playerData.x, playerData.y);
      } else {
        // Update ghost target position
        const ghost = ghosts.get(playerData.i);
        if (ghost) {
          ghost.updateTarget(playerData.x, playerData.y);
        }
      }
    });
  }
}
```

## Performance Considerations

### Object Pooling

Reuse sprite objects instead of creating/destroying them:

```javascript
class SpritePool {
  constructor(textureId, poolSize = 100) {
    this.pool = [];
    this.active = [];

    for (let i = 0; i < poolSize; i++) {
      const sprite = new PIXI.Sprite(PIXI.Texture.from(textureId));
      sprite.visible = false;
      this.pool.push(sprite);
    }
  }

  get() {
    const sprite = this.pool.pop() || new PIXI.Sprite(PIXI.Texture.from(textureId));
    sprite.visible = true;
    this.active.push(sprite);
    return sprite;
  }

  release(sprite) {
    sprite.visible = false;
    const index = this.active.indexOf(sprite);
    if (index > -1) {
      this.active.splice(index, 1);
      this.pool.push(sprite);
    }
  }
}
```

### Culling

Don't render objects outside the visible area:

```javascript
function cullSprites(sprites, camera) {
  sprites.forEach(sprite => {
    const inView = (
      sprite.x + sprite.width >= camera.x &&
      sprite.x <= camera.x + camera.width &&
      sprite.y + sprite.height >= camera.y &&
      sprite.y <= camera.y + camera.height
    );

    sprite.renderable = inView;
  });
}
```

## Related Documentation

- Network Protocol: See docs/04-technical-architecture/network-protocol.md
- Server Architecture: See docs/04-technical-architecture/backend-go.md
- Visual Styling: See docs/03-art-style-aesthetics.md
