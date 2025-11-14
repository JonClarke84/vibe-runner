# Vibe Runner - Client (Phase 1)

**Phase 1: Core Local Game (Client-Only)**

This is a local single-player prototype with basic physics, jump mechanics, and collision detection.

## Features Implemented

✅ Pixi.js canvas with game loop (requestAnimationFrame)
✅ Player entity with gravity physics
✅ Jump mechanic (Spacebar)
✅ Ground platform with collision detection
✅ Static obstacles with AABB collision detection
✅ Death state (game stops on collision)
✅ Debug display (FPS, player position, state)
✅ Score tracking (time survived)
✅ Restart functionality (Press R after death)

## Controls

- **SPACEBAR** - Jump
- **R** - Restart (after death)

## Running the Game

### Option 1: Using npm (Recommended)

```bash
cd client
npm install
npm run dev
```

The game will open automatically at `http://localhost:3000`

### Option 2: Using any HTTP server

```bash
cd client/public
# Python 3
python -m http.server 3000

# Python 2
python -m SimpleHTTPServer 3000

# Node.js
npx http-server -p 3000
```

Then open `http://localhost:3000` in your browser.

### Option 3: Direct file access (may have CORS issues)

Open `client/public/index.html` directly in your browser.

⚠️ **Note:** Some browsers may block ES6 modules when opening files directly. Use an HTTP server instead.

## Project Structure

```
client/
├── public/
│   └── index.html          # Main HTML page
├── src/
│   ├── main.js             # Game initialization and loop
│   └── game/
│       ├── Player.js       # Player entity
│       ├── Ground.js       # Ground platform
│       ├── Obstacle.js     # Obstacle entities
│       └── collision.js    # AABB collision detection
├── package.json            # npm configuration
└── README.md              # This file
```

## Placeholder Graphics

Phase 1 uses colored rectangles as placeholder graphics:

- **Player**: Dark purple rectangle with pink visor and cyan outline
- **Ground**: Dark indigo platform with glowing cyan top edge
- **Obstacles**: Glitch red rectangles with pink outline

These will be replaced with proper SVG/PNG assets in later phases.

## Physics Constants

- **Gravity**: 1200 pixels/second²
- **Jump Velocity**: -600 pixels/second
- **Player Size**: 40×60 pixels
- **Ground Y**: 500 pixels

## Success Criteria (Phase 1)

- [x] Player sprite renders on screen
- [x] Player falls with gravity
- [x] Player jumps when spacebar is pressed
- [x] Player collides with ground and stops falling
- [x] Static obstacles render on screen
- [x] Player dies (game stops) when colliding with obstacles
- [x] Basic score/debug info displays on screen

## Next Steps

Once Phase 1 is complete and tested, move to **Phase 2: Basic Server & Client Connection**.

See `docs/00-development-phases/phase-2-server-connection.md` for details.

## Troubleshooting

### Game doesn't load
- Check browser console for errors
- Ensure you're using an HTTP server (not file://)
- Verify Pixi.js CDN is accessible

### Controls not working
- Click on the game canvas to ensure it has focus
- Check that JavaScript is enabled
- Try refreshing the page

### Performance issues
- Phase 1 should run smoothly at 60 FPS
- Check debug display for actual FPS
- Try closing other browser tabs
