// WebSocketClient.js - Handles WebSocket connection to game server

/**
 * WebSocketClient manages the WebSocket connection to the Vibe Runner server.
 * It handles the join handshake, receives game state updates, and sends player actions.
 *
 * Message Protocol:
 * - All messages use format: { e: "event", d: data }
 * - Join: { e: "join", d: { n: playerName } }
 * - Welcome: { e: "welcome", d: { id, seed, serverTime } }
 * - State: { e: "state", d: { t: timestamp, p: [players] } }
 * - Jump: { e: "jump", d: { t: timestamp } }
 */
export class WebSocketClient {
    /**
     * Creates a new WebSocket client.
     *
     * @param {string} playerName - The player's chosen display name
     */
    constructor(playerName) {
        this.playerName = playerName || 'Player';
        this.playerId = null;
        this.seed = null;
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // ms

        // Callbacks (set by main.js)
        this.onWelcome = null; // Called when welcome message received
        this.onStateUpdate = null; // Called when state message received
        this.onDisconnect = null; // Called when connection closes
    }

    /**
     * Connects to the WebSocket server.
     * Initiates the connection and sets up event handlers.
     *
     * @param {string} serverUrl - WebSocket server URL (default: ws://localhost:8080/ws)
     */
    connect(serverUrl = 'ws://localhost:8080/ws') {
        console.log(`[WebSocket] Connecting to ${serverUrl}...`);

        try {
            this.ws = new WebSocket(serverUrl);

            this.ws.onopen = () => this.handleOpen();
            this.ws.onmessage = (event) => this.handleMessage(event);
            this.ws.onerror = (error) => this.handleError(error);
            this.ws.onclose = (event) => this.handleClose(event);
        } catch (error) {
            console.error('[WebSocket] Connection error:', error);
            this.attemptReconnect(serverUrl);
        }
    }

    /**
     * Handles WebSocket connection opening.
     * Sends the join message to authenticate with the server.
     */
    handleOpen() {
        console.log('[WebSocket] Connected to server');
        this.isConnected = true;
        this.reconnectAttempts = 0; // Reset reconnect counter

        // Send join message
        this.sendJoinMessage();
    }

    /**
     * Sends the join message to the server.
     * This initiates the handshake and assigns a player ID.
     */
    sendJoinMessage() {
        const joinMsg = {
            e: 'join',
            d: {
                n: this.playerName
            }
        };

        console.log('[WebSocket] Sending join message:', joinMsg);
        this.send(joinMsg);
    }

    /**
     * Handles incoming WebSocket messages.
     * Routes messages to appropriate handlers based on event type.
     *
     * @param {MessageEvent} event - WebSocket message event
     */
    handleMessage(event) {
        try {
            const message = JSON.parse(event.data);

            // Debug logging
            console.log(`[WebSocket] Received:`, message.e);

            // Route message based on event type
            switch (message.e) {
                case 'welcome':
                    this.handleWelcome(message.d);
                    break;
                case 'state':
                    this.handleState(message.d);
                    break;
                case 'death':
                    this.handleDeath(message.d);
                    break;
                case 'chunk':
                    this.handleChunk(message.d);
                    break;
                default:
                    console.warn('[WebSocket] Unknown message type:', message.e);
            }
        } catch (error) {
            console.error('[WebSocket] Failed to parse message:', error);
        }
    }

    /**
     * Handles the welcome message from the server.
     * This is received after a successful join.
     *
     * @param {Object} data - Welcome data { id, seed, serverTime }
     */
    handleWelcome(data) {
        this.playerId = data.id;
        this.seed = data.seed;
        const serverTime = data.serverTime;

        console.log(`[WebSocket] Welcome received!`);
        console.log(`  Player ID: ${this.playerId}`);
        console.log(`  Seed: ${this.seed}`);
        console.log(`  Server Time: ${serverTime}`);

        // Notify main game
        if (this.onWelcome) {
            this.onWelcome(this.playerId, this.seed, serverTime);
        }
    }

    /**
     * Handles state update messages from the server.
     * These are sent at 20Hz and contain all player positions.
     *
     * @param {Object} data - State data { t: timestamp, p: [players] }
     */
    handleState(data) {
        const timestamp = data.t;
        const players = data.p;

        // Find this player's data
        const myPlayerData = players.find(p => p.i === this.playerId);

        if (myPlayerData) {
            // Notify main game of position update
            if (this.onStateUpdate) {
                this.onStateUpdate(myPlayerData.x, myPlayerData.y, players);
            }
        }
    }

    /**
     * Handles death message from the server.
     * Sent when the player collides with an obstacle.
     *
     * @param {Object} data - Death data { s: finalScore }
     */
    handleDeath(data) {
        const finalScore = data.s;
        console.log(`[WebSocket] Death! Final score: ${finalScore}`);
        // Will be implemented in later phases
    }

    /**
     * Handles chunk message from the server.
     * Sent when a new level chunk is generated.
     *
     * @param {Object} data - Chunk data { id, obs: [obstacles] }
     */
    handleChunk(data) {
        console.log(`[WebSocket] Chunk ${data.id} received with ${data.obs.length} obstacles`);
        // Will be implemented in later phases
    }

    /**
     * Handles WebSocket errors.
     *
     * @param {Event} error - WebSocket error event
     */
    handleError(error) {
        console.error('[WebSocket] Error:', error);
    }

    /**
     * Handles WebSocket connection closing.
     * Attempts reconnection if not explicitly closed.
     *
     * @param {CloseEvent} event - WebSocket close event
     */
    handleClose(event) {
        console.log(`[WebSocket] Connection closed (code: ${event.code})`);
        this.isConnected = false;

        // Notify main game
        if (this.onDisconnect) {
            this.onDisconnect();
        }

        // Attempt reconnect unless explicitly closed (code 1000)
        if (event.code !== 1000) {
            this.attemptReconnect();
        }
    }

    /**
     * Attempts to reconnect to the server after a delay.
     *
     * @param {string} serverUrl - WebSocket server URL
     */
    attemptReconnect(serverUrl = 'ws://localhost:8080/ws') {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('[WebSocket] Max reconnect attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * this.reconnectAttempts;

        console.log(`[WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

        setTimeout(() => {
            this.connect(serverUrl);
        }, delay);
    }

    /**
     * Sends a jump action to the server.
     * Called when the player presses spacebar.
     */
    sendJump() {
        if (!this.isConnected) {
            console.warn('[WebSocket] Cannot send jump - not connected');
            return;
        }

        const jumpMsg = {
            e: 'jump',
            d: {
                t: Date.now()
            }
        };

        console.log('[WebSocket] Sending jump');
        this.send(jumpMsg);
    }

    /**
     * Sends a message to the server.
     *
     * @param {Object} message - Message object to send
     */
    send(message) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.warn('[WebSocket] Cannot send message - not connected');
            return;
        }

        this.ws.send(JSON.stringify(message));
    }

    /**
     * Closes the WebSocket connection.
     */
    disconnect() {
        if (this.ws) {
            console.log('[WebSocket] Disconnecting...');
            this.ws.close(1000, 'Client disconnect');
            this.ws = null;
        }
    }
}
