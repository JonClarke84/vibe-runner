// ChunkManager.js - Client-side chunk management and rendering

export class ChunkManager {
    constructor(app) {
        this.app = app;
        this.chunks = new Map(); // Map<chunkID, {obstacles: [], sprites: []}>
        this.obstacleSprites = []; // All obstacle sprites for easy iteration
    }

    /**
     * Receives a chunk from the server and renders its obstacles.
     * @param {Object} chunkData - Chunk data from server: {id, obs: [{t, x, y}]}
     */
    receiveChunk(chunkData) {
        const chunkID = chunkData.id;

        // Don't re-render if we already have this chunk
        if (this.chunks.has(chunkID)) {
            console.log(`[ChunkManager] Chunk ${chunkID} already exists, skipping`);
            return;
        }

        console.log(`[ChunkManager] Received chunk ${chunkID} with ${chunkData.obs.length} obstacles`);

        const obstacles = [];
        const sprites = [];

        // Render each obstacle
        for (const obsData of chunkData.obs) {
            const sprite = this.createObstacleSprite(obsData);
            sprites.push(sprite);
            obstacles.push({
                type: obsData.t,
                x: obsData.x,
                y: obsData.y,
                sprite: sprite
            });
            this.app.stage.addChild(sprite);
            this.obstacleSprites.push(sprite);
        }

        // Store chunk
        this.chunks.set(chunkID, {
            obstacles: obstacles,
            sprites: sprites
        });
    }

    /**
     * Creates a Pixi sprite for an obstacle based on type.
     * @param {Object} obsData - Obstacle data: {t: type, x, y}
     * @returns {PIXI.Graphics} The obstacle sprite
     */
    createObstacleSprite(obsData) {
        const sprite = new PIXI.Graphics();

        // Different appearance based on obstacle type
        switch (obsData.t) {
            case 1: // Tall obstacle
                sprite.beginFill(0xff003c); // Glitch Red
                sprite.drawRect(0, 0, 40, 100);
                sprite.endFill();
                break;

            case 2: // Low obstacle
                sprite.beginFill(0xff003c); // Glitch Red
                sprite.drawRect(0, 0, 60, 60);
                sprite.endFill();
                break;

            case 3: // Spike obstacle
                sprite.beginFill(0xff003c); // Glitch Red
                sprite.drawRect(0, 0, 30, 80);
                sprite.endFill();
                break;

            default:
                // Unknown type - render as red box
                sprite.beginFill(0xff003c);
                sprite.drawRect(0, 0, 40, 80);
                sprite.endFill();
        }

        // Add cyan outline for neon effect
        sprite.lineStyle(2, 0x00f0ff); // Hyper-Cyan
        sprite.drawRect(0, 0, sprite.width, sprite.height);

        // Position obstacle at absolute world coordinates
        // Y is measured from top, but obstacle needs to sit on ground (y=500)
        sprite.position.set(obsData.x, obsData.y === 0 ? 500 - sprite.height : obsData.y);

        return sprite;
    }

    /**
     * Returns all obstacles for collision detection.
     * @returns {Array} Array of obstacle objects with bounds
     */
    getAllObstacles() {
        const allObstacles = [];

        for (const chunk of this.chunks.values()) {
            for (const obs of chunk.obstacles) {
                allObstacles.push({
                    x: obs.x,
                    y: obs.sprite.y,
                    width: obs.sprite.width,
                    height: obs.sprite.height
                });
            }
        }

        return allObstacles;
    }

    /**
     * Cleans up chunks that are far behind the camera.
     * @param {number} minX - Minimum X position to keep (usually camera X - buffer)
     */
    cleanup(minX) {
        const chunkSize = 5000;
        const minChunkID = Math.floor(minX / chunkSize) - 2; // Keep 2 chunks behind

        for (const [chunkID, chunk] of this.chunks.entries()) {
            if (chunkID < minChunkID) {
                // Remove sprites from stage
                for (const sprite of chunk.sprites) {
                    this.app.stage.removeChild(sprite);

                    // Remove from obstacleSprites array
                    const index = this.obstacleSprites.indexOf(sprite);
                    if (index > -1) {
                        this.obstacleSprites.splice(index, 1);
                    }
                }

                // Remove chunk from map
                this.chunks.delete(chunkID);
                console.log(`[ChunkManager] Cleaned up chunk ${chunkID}`);
            }
        }
    }

    /**
     * Returns the number of currently loaded chunks.
     * @returns {number}
     */
    getChunkCount() {
        return this.chunks.size;
    }
}
