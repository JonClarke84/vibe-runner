// Obstacle.js - Static obstacles (firewalls)

export class Obstacle {
    constructor(x, y, width, height) {
        this.x = x;
        this.y = y;
        this.width = width;
        this.height = height;

        // Create placeholder obstacle sprite
        this.sprite = new PIXI.Graphics();
        this.updateSprite();
    }

    updateSprite() {
        this.sprite.clear();

        // Glitch Red rectangle with flicker effect
        this.sprite.beginFill(0xff003c); // Glitch Red
        this.sprite.drawRect(this.x, this.y, this.width, this.height);
        this.sprite.endFill();

        // Add outline for visibility
        this.sprite.lineStyle(2, 0xff007f); // Electric Pink outline
        this.sprite.drawRect(this.x, this.y, this.width, this.height);
    }

    // AABB collision bounds
    getBounds() {
        return {
            x: this.x,
            y: this.y,
            width: this.width,
            height: this.height
        };
    }
}
