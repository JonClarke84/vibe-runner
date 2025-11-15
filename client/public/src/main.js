// main.js - Main game entry point with game loop

import { Player } from './game/Player.js';
import { Ground } from './game/Ground.js';
import { Obstacle } from './game/Obstacle.js';
import { checkCollision } from './game/collision.js';
import { WebSocketClient } from './network/WebSocketClient.js';
import { GhostPlayer } from './game/GhostPlayer.js';
import { ChunkManager } from './game/ChunkManager.js';

// Game constants
const GAME_WIDTH = 1280;
const GAME_HEIGHT = 720;
const GROUND_Y = 500;

// Game state
let player;
let ground;
let obstacles = [];
let isGameRunning = true;
let score = 0;
let lastTime = 0;

// Network
let wsClient = null;
let myPlayerId = null; // Our player ID from server

// Ghost players (other connected players)
let ghostPlayers = new Map(); // Map<playerId, GhostPlayer>

// Chunk Manager (Phase 4)
let chunkManager = null;

// Debug display
let debugText;
let scoreText;

// Initialize Pixi.js Application
const app = new PIXI.Application({
    width: GAME_WIDTH,
    height: GAME_HEIGHT,
    backgroundColor: 0x1a1a2e, // Deep Indigo background
    antialias: false // Pixel-perfect rendering
});

// Add canvas to DOM
document.getElementById('game-container').appendChild(app.view);

// Initialize game objects
function init() {
    // Create player
    player = new Player(100, 100);
    app.stage.addChild(player.sprite);

    // Create ground
    ground = new Ground(GAME_WIDTH, GROUND_Y);
    app.stage.addChild(ground.sprite);

    // PHASE 4: Initialize Chunk Manager (procedural obstacles)
    chunkManager = new ChunkManager(app);
    console.log('[Game] ChunkManager initialized - obstacles will come from server');

    // Create debug text
    const debugStyle = new PIXI.TextStyle({
        fontFamily: 'Courier New',
        fontSize: 16,
        fill: 0x00f0ff, // Hyper-Cyan
        align: 'left'
    });

    debugText = new PIXI.Text('', debugStyle);
    debugText.position.set(10, 10);
    app.stage.addChild(debugText);

    // Create score text
    const scoreStyle = new PIXI.TextStyle({
        fontFamily: 'Courier New',
        fontSize: 24,
        fill: 0xff00ff, // Magenta
        align: 'left'
    });

    scoreText = new PIXI.Text('Score: 0.0s', scoreStyle);
    scoreText.position.set(10, 50);
    app.stage.addChild(scoreText);

    // Set up input
    setupInput();

    // Initialize WebSocket connection
    initializeNetwork();

    // Start game loop
    requestAnimationFrame(gameLoop);
}

// Initialize network connection
function initializeNetwork() {
    // Create WebSocket client
    wsClient = new WebSocketClient('Player1'); // Default name for now

    // Set up callbacks
    wsClient.onWelcome = (playerId, seed, serverTime) => {
        console.log(`[Game] Welcome! Player ID: ${playerId}, Seed: ${seed}`);
        myPlayerId = playerId; // Store our player ID
    };

    wsClient.onStateUpdate = (x, y, allPlayers) => {
        // Update local player position from server state
        player.setServerPosition(x, y);

        // PHASE 3: Render ghost players (other connected players)
        if (allPlayers && allPlayers.length > 0) {
            updateGhostPlayers(allPlayers);
        }
    };

    wsClient.onDisconnect = () => {
        console.log('[Game] Disconnected from server');
    };

    // PHASE 4: Handle incoming chunks
    wsClient.onChunkReceived = (chunkData) => {
        if (chunkManager) {
            chunkManager.receiveChunk(chunkData);
        }
    };

    // Connect to server
    wsClient.connect();
    console.log('[Game] Connecting to server...');
}

// Update ghost players from server state
function updateGhostPlayers(allPlayers) {
    // Track which player IDs are in the current state
    const currentPlayerIds = new Set();

    // Update or create ghost players
    for (const playerData of allPlayers) {
        const playerId = playerData.i; // Player ID
        const x = playerData.x;
        const y = playerData.y;

        // Skip our own player
        if (playerId === myPlayerId) {
            continue;
        }

        currentPlayerIds.add(playerId);

        // Create new ghost if it doesn't exist
        if (!ghostPlayers.has(playerId)) {
            const ghost = new GhostPlayer(playerId, `Player${playerId}`);
            ghost.setTargetPosition(x, y);
            ghost.x = x; // Initialize position immediately
            ghost.y = y;
            ghostPlayers.set(playerId, ghost);
            app.stage.addChild(ghost.getContainer());
            console.log(`[Game] New ghost player joined: ${playerId}`);
        } else {
            // Update existing ghost's target position
            const ghost = ghostPlayers.get(playerId);
            ghost.setTargetPosition(x, y);
        }
    }

    // Remove ghosts for disconnected players
    for (const [playerId, ghost] of ghostPlayers.entries()) {
        if (!currentPlayerIds.has(playerId)) {
            console.log(`[Game] Ghost player left: ${playerId}`);
            ghost.destroy();
            ghostPlayers.delete(playerId);
        }
    }
}

// Input handling
function setupInput() {
    window.addEventListener('keydown', (event) => {
        if (event.code === 'Space') {
            // PHASE 3: Client-side prediction - jump immediately
            player.jump(); // Apply jump locally for instant feedback

            // Also notify server for validation
            if (wsClient && wsClient.isConnected) {
                wsClient.sendJump();
            }
        }
    });
}

// Main game loop
function gameLoop(currentTime) {
    // Calculate delta time in seconds
    const deltaTime = (currentTime - lastTime) / 1000;
    lastTime = currentTime;

    // Cap delta time to prevent large jumps
    const cappedDelta = Math.min(deltaTime, 0.1);

    if (isGameRunning) {
        update(cappedDelta);
    }

    render(cappedDelta);

    // Continue loop
    requestAnimationFrame(gameLoop);
}

// Update game logic
function update(deltaTime) {
    // Update player physics
    player.update(deltaTime);

    // Check ground collision
    player.checkGroundCollision(ground.getY());

    // PHASE 3: Update ghost players (entity interpolation)
    for (const ghost of ghostPlayers.values()) {
        ghost.update(deltaTime);
    }

    // PHASE 5: Camera follows player horizontally
    // Center camera on player (player at 1/3 from left edge for better forward visibility)
    const cameraX = -(player.x - GAME_WIDTH / 3);
    app.stage.x = cameraX;

    // TODO PHASE 5: Re-enable collision detection with server validation
    // const playerBounds = player.getBounds();
    // if (chunkManager) {
    //     const allObstacles = chunkManager.getAllObstacles();
    //     for (let obstacleBounds of allObstacles) {
    //         if (checkCollision(playerBounds, obstacleBounds)) {
    //             player.die();
    //             isGameRunning = false;
    //             break;
    //         }
    //     }
    // }

    // Update score (time survived)
    if (player.isAlive) {
        score += deltaTime;
    }
}

// Render/draw
function render(deltaTime) {
    // Update debug info
    const fps = Math.round(1 / deltaTime);
    const chunkCount = chunkManager ? chunkManager.getChunkCount() : 0;
    debugText.text = `FPS: ${fps}\nPlayer: (${Math.round(player.x)}, ${Math.round(player.y)})\nGrounded: ${player.isGrounded}\nAlive: ${player.isAlive}\nGhosts: ${ghostPlayers.size}\nChunks: ${chunkCount}`;

    // Update score display
    scoreText.text = `Score: ${score.toFixed(1)}s`;
    scoreText.style.fill = player.isAlive ? 0xff00ff : 0xff003c; // Magenta or Glitch Red

    // Add death message if dead
    if (!player.isAlive && !document.getElementById('death-message')) {
        showDeathScreen();
    }
}

// Show death screen
function showDeathScreen() {
    const deathStyle = new PIXI.TextStyle({
        fontFamily: 'Courier New',
        fontSize: 48,
        fill: 0xff003c, // Glitch Red
        align: 'center',
        stroke: 0xff003c,
        strokeThickness: 2
    });

    const deathText = new PIXI.Text('GAME OVER', deathStyle);
    deathText.anchor.set(0.5);
    deathText.position.set(GAME_WIDTH / 2, GAME_HEIGHT / 2);
    deathText.id = 'death-message';
    app.stage.addChild(deathText);

    const restartStyle = new PIXI.TextStyle({
        fontFamily: 'Courier New',
        fontSize: 20,
        fill: 0x00f0ff, // Hyper-Cyan
        align: 'center'
    });

    const restartText = new PIXI.Text('[Press R to Restart]', restartStyle);
    restartText.anchor.set(0.5);
    restartText.position.set(GAME_WIDTH / 2, GAME_HEIGHT / 2 + 60);
    app.stage.addChild(restartText);

    // Add restart functionality
    window.addEventListener('keydown', function restartHandler(event) {
        if (event.code === 'KeyR') {
            window.removeEventListener('keydown', restartHandler);
            restartGame();
        }
    });
}

// Restart game
function restartGame() {
    console.log('[Game] Restarting game...');

    // CRITICAL: Disconnect old WebSocket to prevent connection storm
    if (wsClient && wsClient.isConnected) {
        console.log('[Game] Disconnecting old WebSocket...');
        wsClient.disconnect();
        wsClient = null;
    }

    // Clear all ghost players
    for (const ghost of ghostPlayers.values()) {
        ghost.destroy();
    }
    ghostPlayers.clear();

    // Clear stage
    app.stage.removeChildren();

    // Reset game state
    obstacles = [];
    isGameRunning = true;
    score = 0;
    myPlayerId = null;

    // Reinitialize (this will create a new WebSocket connection)
    init();

    console.log('[Game] Restart complete');
}

// Start the game
init();

console.log('Vibe Runner - Phase 1 initialized!');
console.log('Controls: SPACEBAR to jump');
