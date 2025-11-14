// GhostPlayer.js - Ghost player entity for rendering other players

const GHOST_WIDTH = 40;
const GHOST_HEIGHT = 60;

/**
 * Represents another player in the game world.
 * Uses entity interpolation (lerp) to smooth out the 20Hz server updates.
 */
export class GhostPlayer {
    constructor(playerId, playerName) {
        this.playerId = playerId;
        this.playerName = playerName || `Player${playerId}`;

        // Current interpolated position (what we render)
        this.x = 0;
        this.y = 0;

        // Target position from server
        this.targetX = 0;
        this.targetY = 0;

        // Interpolation amount (0.3 = smooth, 1.0 = instant)
        this.lerpAmount = 0.3;

        // Create sprite container
        this.container = new PIXI.Container();

        // Create ghost sprite (semi-transparent version of player)
        this.sprite = new PIXI.Graphics();
        this.updateSprite();
        this.container.addChild(this.sprite);

        // Create name label
        this.createNameLabel();
    }

    createNameLabel() {
        const nameStyle = new PIXI.TextStyle({
            fontFamily: 'Courier New',
            fontSize: 14,
            fill: 0x00f0ff, // Hyper-Cyan
            align: 'center'
        });

        this.nameLabel = new PIXI.Text(this.playerName, nameStyle);
        this.nameLabel.anchor.set(0.5, 1); // Center horizontally, anchor at bottom
        this.nameLabel.position.set(GHOST_WIDTH / 2, -5); // Position above sprite
        this.container.addChild(this.nameLabel);
    }

    updateSprite() {
        this.sprite.clear();

        // Ghost appearance: Semi-transparent with cyan glow
        this.sprite.beginFill(0x301a4b, 0.6); // Dark Purple, semi-transparent
        this.sprite.drawRect(0, 0, GHOST_WIDTH, GHOST_HEIGHT);
        this.sprite.endFill();

        // Cyan visor (Electric Pink tint)
        this.sprite.beginFill(0xff007f, 0.7); // Electric Pink, semi-transparent
        this.sprite.drawRect(5, 10, 30, 15);
        this.sprite.endFill();

        // Cyan outline (brighter for ghost effect)
        this.sprite.lineStyle(2, 0x00f0ff, 0.8); // Hyper-Cyan glow
        this.sprite.drawRect(0, 0, GHOST_WIDTH, GHOST_HEIGHT);
    }

    /**
     * Updates target position from server state.
     * The ghost will smoothly interpolate to this position.
     */
    setTargetPosition(x, y) {
        this.targetX = x;
        this.targetY = y;
    }

    /**
     * Lerp function for smooth interpolation
     */
    lerp(start, end, amount) {
        return (1 - amount) * start + amount * end;
    }

    /**
     * Update ghost position using entity interpolation.
     * Called every frame (60 FPS) to smooth out 20Hz server updates.
     */
    update(deltaTime) {
        // Interpolate position towards target
        this.x = this.lerp(this.x, this.targetX, this.lerpAmount);
        this.y = this.lerp(this.y, this.targetY, this.lerpAmount);

        // Update sprite position
        this.container.position.set(this.x, this.y);
    }

    /**
     * Returns the Pixi container for adding to stage
     */
    getContainer() {
        return this.container;
    }

    /**
     * Removes this ghost from the stage
     */
    destroy() {
        this.container.destroy({ children: true });
    }
}
