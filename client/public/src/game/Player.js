// Player.js - Player entity with physics and rendering

// Physics constants from docs
const GRAVITY = 1200;          // pixels/second^2
const JUMP_VELOCITY = -600;    // pixels/second (negative = up)
const PLAYER_WIDTH = 40;
const PLAYER_HEIGHT = 60;

export class Player {
    constructor(x, y) {
        this.x = x;
        this.y = y;
        this.width = PLAYER_WIDTH;
        this.height = PLAYER_HEIGHT;

        this.velocityX = 0;
        this.velocityY = 0;

        this.isGrounded = false;
        this.isAlive = true;

        // Create placeholder sprite (colored rectangle)
        this.sprite = new PIXI.Graphics();
        this.updateSprite();
    }

    updateSprite() {
        this.sprite.clear();

        if (this.isAlive) {
            // Player: Dark Purple suit with Cyan outline
            this.sprite.beginFill(0x301a4b); // Dark Purple
            this.sprite.drawRect(0, 0, this.width, this.height);
            this.sprite.endFill();

            // Cyan visor (Electric Pink visor placeholder)
            this.sprite.beginFill(0xff007f); // Electric Pink
            this.sprite.drawRect(5, 10, 30, 15);
            this.sprite.endFill();

            // Cyan outline
            this.sprite.lineStyle(2, 0x00f0ff); // Hyper-Cyan
            this.sprite.drawRect(0, 0, this.width, this.height);
        } else {
            // Death state: Glitch Red
            this.sprite.beginFill(0xff003c, 0.5); // Glitch Red, semi-transparent
            this.sprite.drawRect(0, 0, this.width, this.height);
            this.sprite.endFill();
        }

        this.sprite.position.set(this.x, this.y);
    }

    jump() {
        if (this.isGrounded && this.isAlive) {
            this.velocityY = JUMP_VELOCITY;
            this.isGrounded = false;
        }
    }

    update(deltaTime) {
        if (!this.isAlive) {
            return; // Don't update physics when dead
        }

        // Apply gravity
        this.velocityY += GRAVITY * deltaTime;

        // Update position
        this.y += this.velocityY * deltaTime;

        // Update sprite position
        this.sprite.position.set(this.x, this.y);
    }

    checkGroundCollision(groundY) {
        // Simple ground collision
        if (this.y + this.height >= groundY) {
            this.y = groundY - this.height;
            this.velocityY = 0;
            this.isGrounded = true;
        }
    }

    // AABB collision check
    getBounds() {
        return {
            x: this.x,
            y: this.y,
            width: this.width,
            height: this.height
        };
    }

    die() {
        this.isAlive = false;
        this.updateSprite();
    }
}
