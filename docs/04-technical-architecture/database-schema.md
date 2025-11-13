# Database Schema

**Version:** 1.5 **Date:** 2025-11-13

## Overview

Vibe Runner uses a two-tier database approach:

1. **Redis** - In-memory cache for real-time data (live leaderboard, session data)
2. **PostgreSQL** - Persistent storage for historical data (all-time scores)

## Redis (In-Memory Cache)

Redis is used for high-speed reads/writes during active gameplay.

### Data Structures

#### Live Leaderboard (Sorted Set)

The live leaderboard uses a Redis **Sorted Set** to maintain real-time rankings.

**Key:** `leaderboard:current`

**Operations:**

```redis
# Add a player's score (or update if exists)
ZADD leaderboard:current 120.5 "VibeKing"

# Get top 10 players with scores
ZREVRANGE leaderboard:current 0 9 WITHSCORES

# Get a player's rank (0-indexed)
ZREVRANK leaderboard:current "VibeKing"

# Get total number of players
ZCARD leaderboard:current

# Remove a player
ZREM leaderboard:current "VibeKing"

# Clear entire leaderboard (e.g., daily reset)
DEL leaderboard:current
```

**Example Data:**
```
Score: 180.5 → Player: "VibeKing"
Score: 156.2 → Player: "CodeRunner"
Score: 142.8 → Player: "Glitch"
Score: 138.4 → Player: "NeonDream"
```

**Notes:**
- Scores are stored as floats (time survived in seconds)
- Higher scores are better (ZREVRANGE returns highest first)
- Player names must be unique per session
- Leaderboard persists until server restart or manual reset

---

#### Player Session Data (Hash)

Active player information stored during their session.

**Key Pattern:** `player:<playerID>`

**Operations:**

```redis
# Set player data
HMSET player:12345 name "VibeKing" score 120.5 status "alive"

# Get player name
HGET player:12345 name

# Get all player data
HGETALL player:12345

# Update player status
HSET player:12345 status "dead"

# Delete player session (on disconnect)
DEL player:12345

# Set expiration (auto-cleanup after 1 hour)
EXPIRE player:12345 3600
```

**Fields:**
- `name` (string): Player nickname
- `score` (float): Current/final score
- `status` (string): "alive" | "dead"
- `joinedAt` (timestamp): When player joined

**Example:**
```
player:12345 → {
  name: "VibeKing",
  score: 120.5,
  status: "alive",
  joinedAt: 1678886400000
}
```

---

#### Server Statistics (String/Hash)

Track server-wide metrics.

**Keys:**
- `stats:totalPlayers` (string): Total players who ever connected
- `stats:currentPlayers` (string): Currently active players
- `stats:peakPlayers` (string): Peak concurrent players

**Operations:**

```redis
# Increment total players
INCR stats:totalPlayers

# Set current player count
SET stats:currentPlayers 247

# Get all stats
MGET stats:totalPlayers stats:currentPlayers stats:peakPlayers
```

---

### Redis Configuration

**Connection Settings:**
```
Host: localhost (development) | redis.server.com (production)
Port: 6379
Database: 0
Password: <secure-password> (production only)
```

**Performance Settings:**
```
maxmemory: 256mb (adjust based on player count)
maxmemory-policy: allkeys-lru (evict least recently used)
```

---

## PostgreSQL (Persistent Storage)

PostgreSQL stores historical data for long-term analysis and all-time leaderboards.

### Schema

#### `scores` Table

Stores every player's death/score event permanently.

**Definition:**

```sql
CREATE TABLE scores (
  id BIGSERIAL PRIMARY KEY,
  player_name VARCHAR(30) NOT NULL,
  score FLOAT NOT NULL,
  achieved_at TIMESTAMPTZ DEFAULT NOW(),
  session_id VARCHAR(50),
  INDEX idx_score DESC (score),
  INDEX idx_achieved_at DESC (achieved_at)
);
```

**Fields:**
- `id`: Unique identifier for each score entry
- `player_name`: Player's nickname (sanitized)
- `score`: Time survived in seconds
- `achieved_at`: Timestamp when the score was achieved (UTC)
- `session_id`: Optional session identifier for analytics

**Indexes:**
- `idx_score`: Optimize queries for top scores
- `idx_achieved_at`: Optimize queries for recent scores

---

#### `players` Table (Optional - Future Enhancement)

Store registered player accounts (if authentication is added).

**Definition:**

```sql
CREATE TABLE players (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(30) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  total_games INT DEFAULT 0,
  best_score FLOAT DEFAULT 0.0
);
```

**Note:** This table is not required for MVP (anonymous gameplay).

---

### Common Queries

#### Insert Score on Death

```sql
INSERT INTO scores (player_name, score, session_id)
VALUES ($1, $2, $3);
```

**Go Implementation:**

```go
func saveScore(db *sql.DB, playerName string, score float64, sessionID string) error {
    _, err := db.Exec(
        "INSERT INTO scores (player_name, score, session_id) VALUES ($1, $2, $3)",
        playerName, score, sessionID,
    )
    return err
}
```

---

#### Get Top 100 All-Time Scores

```sql
SELECT player_name, score, achieved_at
FROM scores
ORDER BY score DESC
LIMIT 100;
```

---

#### Get Scores from Last 24 Hours

```sql
SELECT player_name, score, achieved_at
FROM scores
WHERE achieved_at > NOW() - INTERVAL '24 hours'
ORDER BY score DESC
LIMIT 50;
```

---

#### Get Player's Personal Best

```sql
SELECT MAX(score) as best_score
FROM scores
WHERE player_name = $1;
```

---

#### Get Statistics

```sql
-- Total games played
SELECT COUNT(*) FROM scores;

-- Average score
SELECT AVG(score) FROM scores;

-- Total unique players
SELECT COUNT(DISTINCT player_name) FROM scores;
```

---

### PostgreSQL Configuration

**Connection Settings:**
```
Host: localhost (development) | postgres.server.com (production)
Port: 5432
Database: viberunner
User: viberunner_app
Password: <secure-password>
SSL Mode: require (production)
```

**Performance Settings:**
```
shared_buffers: 256MB
max_connections: 100
work_mem: 4MB
```

---

## Data Flow

### On Player Death

```
1. Server detects collision
2. Calculate final score (time survived)

Parallel writes:
3a. Write to Redis (immediate):
    ZADD leaderboard:current <score> <name>

3b. Write to PostgreSQL (async):
    INSERT INTO scores (player_name, score) VALUES (...)

4. Send death message to client:
   {"e": "death", "d": {"s": 120.5}}

5. Broadcast updated leaderboard to all clients
```

### Leaderboard Fetch

```
1. Client requests leaderboard (or server sends periodically)
2. Server queries Redis:
   ZREVRANGE leaderboard:current 0 9 WITHSCORES
3. Server sends to client:
   {"e": "leaderboard", "d": {"top": [...]}}
```

### All-Time Leaderboard (Future Feature)

```
1. Client navigates to "All-Time" tab
2. Client sends request to HTTP API endpoint (not WebSocket)
3. Server queries PostgreSQL:
   SELECT ... FROM scores ORDER BY score DESC LIMIT 100
4. Server returns JSON response
```

---

## Database Initialization

### Redis Initialization

```bash
# Start Redis
redis-server

# No schema required (schemaless)
# Data structures created on first write
```

### PostgreSQL Initialization

```bash
# Create database
createdb viberunner

# Run schema migration
psql viberunner < schema.sql
```

**schema.sql:**

```sql
-- Create scores table
CREATE TABLE IF NOT EXISTS scores (
  id BIGSERIAL PRIMARY KEY,
  player_name VARCHAR(30) NOT NULL,
  score FLOAT NOT NULL,
  achieved_at TIMESTAMPTZ DEFAULT NOW(),
  session_id VARCHAR(50)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_score ON scores (score DESC);
CREATE INDEX IF NOT EXISTS idx_achieved_at ON scores (achieved_at DESC);

-- Create statistics view
CREATE VIEW player_stats AS
SELECT
  player_name,
  COUNT(*) as games_played,
  MAX(score) as best_score,
  AVG(score) as avg_score
FROM scores
GROUP BY player_name;
```

---

## Backup & Maintenance

### Redis

**Persistence:**
```
# Enable RDB snapshots
save 900 1      # Save if 1 key changed in 15 min
save 300 10     # Save if 10 keys changed in 5 min
save 60 10000   # Save if 10000 keys changed in 1 min
```

**Backup:**
```bash
# Manual backup
redis-cli BGSAVE

# Backup file location
/var/lib/redis/dump.rdb
```

### PostgreSQL

**Backup:**
```bash
# Daily backup
pg_dump viberunner > backup_$(date +%Y%m%d).sql

# Restore
psql viberunner < backup_20231115.sql
```

**Maintenance:**
```sql
-- Vacuum (cleanup)
VACUUM ANALYZE scores;

-- Reindex
REINDEX TABLE scores;
```

---

## Scaling Considerations

### Redis Scaling

- **Vertical:** Increase memory (256MB → 1GB → 4GB)
- **Horizontal:** Redis Cluster for >100k concurrent players
- **Replication:** Redis Sentinel for high availability

### PostgreSQL Scaling

- **Partitioning:** Partition `scores` table by date
- **Read Replicas:** For analytics queries
- **Connection Pooling:** Use pgBouncer for >100 connections

---

## Related Documentation

- Backend Implementation: docs/04-technical-architecture/backend-go.md
- Network Protocol: docs/04-technical-architecture/network-protocol.md
- Phase 5 (Full Game Loop): docs/00-development-phases/phase-5-full-game-loop.md
