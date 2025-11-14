// Ground.js - Ground platform

export class Ground {
    constructor(width, y) {
        this.width = width;
        this.height = 20;
        this.y = y;

        // Create placeholder ground sprite
        this.sprite = new PIXI.Graphics();
        this.updateSprite();
    }

    updateSprite() {
        this.sprite.clear();

        // Dark base
        this.sprite.beginFill(0x1a1a2e); // Deep Indigo
        this.sprite.drawRect(0, this.y, this.width, this.height);
        this.sprite.endFill();

        // Glowing cyan top edge
        this.sprite.lineStyle(3, 0x00f0ff); // Hyper-Cyan, thicker line
        this.sprite.moveTo(0, this.y);
        this.sprite.lineTo(this.width, this.y);
    }

    getY() {
        return this.y;
    }
}
