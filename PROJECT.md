# HexSlayer — Requirements Specification

> Idle geo-based monster hunting demo  
> Stack: Go + React | H3 Geospatial | SQLite | WebSocket

---

## 1. Project Overview

HexSlayer is a browser-based idle game where players deploy characters onto a H3 hexagonal map. Characters automatically hunt and fight monsters within their deployed zone. The game runs continuously server-side — players just watch and manage their characters.

---

## 2. Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend language | Go |
| Web framework | Gin |
| WebSocket | Gorilla WebSocket |
| ORM | GORM |
| Database | SQLite (WAL mode) |
| Frontend | React + Vite + Tailwind CSS |
| Map rendering | React-Leaflet + OpenStreetMap (free, no API key) |
| H3 (backend) | uber/h3-go |
| H3 (frontend) | h3-js |

### Project Structure

```
hexslayer/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── db/          ← database init, migrations
│   │   ├── models/      ← GORM models
│   │   ├── game/        ← engine, tick loop, combat
│   │   ├── handlers/    ← HTTP + WebSocket handlers
│   │   └── middleware/  ← rate limit, session validation
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Map.jsx
│   │   │   ├── ZoneHex.jsx
│   │   │   ├── CombatLog.jsx
│   │   │   └── CharacterPanel.jsx
│   │   ├── hooks/
│   │   │   └── useGameSocket.js
│   │   ├── game/
│   │   │   └── session.js  ← player init, localStorage
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
│
└── README.md
```

**Running locally:**
```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm install && npm run dev
```

---

## 3. H3 Spatial Design

### Resolution Layers

| Layer | Resolution | Purpose | Approx Cell Size |
|-------|-----------|---------|-----------------|
| Zone | 6 | Map regions, goroutine boundary, UI hex render | ~36 km² |
| Entity | 12 | Exact position of monsters and characters | ~300 m² |

### Map Area

- Center: Bangkok (`lat: 13.7563, lng: 100.5018`)
- Zones: `h3.GridDisk(center, 2)` → ~19 res-6 zones
- Monsters placed at random res-12 children cells within each zone

### Zone–Entity Relationship

```
res-6 zone
└── contains thousands of res-12 children
    ├── monster @ res-12 cell A
    ├── monster @ res-12 cell B
    └── character @ res-12 cell C
```

---

## 4. Database Schema

### 4.1 `monster_types` — seeded once, never mutated at runtime

| Column | Type | Notes |
|--------|------|-------|
| id | uint PK | auto increment |
| name | string | e.g. Slime, Orc, Dragon |
| base_damage | int | |
| damage_amp | float | multiplier |
| damage_reduction | float | 0.0–1.0 |
| crit_chance | float | 0.0–1.0 |
| crit_multiplier | float | e.g. 1.5 |
| max_hp | int | |
| icon | string | emoji or asset name |

**Seed data (5 types minimum):**

| Name | base_damage | damage_amp | damage_reduction | crit_chance | crit_multiplier | max_hp |
|------|-------------|-----------|-----------------|-------------|----------------|--------|
| Slime | 15 | 1.0 | 0.05 | 0.05 | 1.3 | 80 |
| Goblin | 25 | 1.1 | 0.08 | 0.10 | 1.5 | 120 |
| Orc | 35 | 1.2 | 0.15 | 0.12 | 1.6 | 200 |
| Troll | 45 | 1.3 | 0.20 | 0.15 | 1.8 | 300 |
| Dragon | 60 | 1.5 | 0.25 | 0.20 | 2.0 | 500 |

---

### 4.2 `map_monsters` — monster instances on the map

| Column | Type | Notes |
|--------|------|-------|
| id | string UUID PK | |
| h3_zone | string | res-6 cell index |
| h3_index | string | res-12 cell index (exact position) |
| monster_type_id | uint FK | references monster_types |
| current_hp | int | |
| is_alive | bool | |
| respawn_at | timestamp | nullable, set on death |
| created_at | timestamp | |

---

### 4.3 `players` — one per browser session

| Column | Type | Notes |
|--------|------|-------|
| id | string UUID PK | |
| session_token | string | unique, validated on every WS message |
| name | string | default "Adventurer", not unique |
| created_at | timestamp | |

---

### 4.4 `characters` — deployed fighters (max 2 alive per player)

| Column | Type | Notes |
|--------|------|-------|
| id | string UUID PK | |
| player_id | string FK | references players |
| name | string | random name on deploy |
| h3_zone | string | res-6 zone deployed to |
| h3_index | string | res-12 current position |
| base_damage | int | randomized on deploy |
| damage_amp | float | randomized on deploy |
| damage_reduction | float | randomized on deploy |
| crit_chance | float | randomized on deploy |
| crit_multiplier | float | randomized on deploy |
| hp | int | current HP |
| max_hp | int | randomized on deploy |
| is_alive | bool | |
| deployed_at | timestamp | |
| died_at | timestamp | nullable |

---

### 4.5 `character_engagements` — active combat pairings

Tracks which character is fighting which monster across ticks. Kept separate from `characters` to preserve clean persistent state.

| Column | Type | Notes |
|--------|------|-------|
| id | string UUID PK | |
| character_id | string FK | references characters — **unique constraint** |
| monster_id | string FK | references map_monsters |
| engaged_at | timestamp | when combat started |

**Lifecycle:**

| Event | Action |
|-------|--------|
| Character finds a monster | INSERT row |
| Monster dies | DELETE WHERE monster_id = ? |
| Character dies | DELETE WHERE character_id = ? |
| Character wanders (no monster found) | DELETE WHERE character_id = ? |

> One character can only fight one monster at a time (unique constraint on `character_id`). Two characters cannot fight the same monster (enforced in-memory via `engaged` map during tick).

---

## 5. Character Stats

### Character Stats (randomized on each deploy)

Each character gets fresh random stats when deployed. The player is just a controller — no combat stats on the player.

```
base_damage:      80  – 120
damage_amp:       1.0 – 1.5
damage_reduction: 0.05 – 0.20
crit_chance:      0.10 – 0.25
crit_multiplier:  1.5 – 2.0
max_hp:           250 – 350
```

**Design target:** A character should survive approximately 3 monster kills worth of incoming damage before dying.

### Combat Formula

```
raw_damage   = attacker.base_damage * attacker.damage_amp
roll         = rand(0.0, 1.0)
is_crit      = roll < attacker.crit_chance
crit_damage  = is_crit ? raw_damage * attacker.crit_multiplier : raw_damage
final_damage = crit_damage * (1.0 - defender.damage_reduction)
```

Both character and monster attack each other every tick (2 seconds).

---

## 6. Game Rules

### 6.1 Character Deployment

- Player may deploy a **maximum of 2 characters** simultaneously
- Deploy action: player clicks a res-6 zone on the map → character spawns at center res-12 cell of that zone
- Character name is randomly generated on each deploy
- Character stats are randomized on each deploy (not inherited from player)

### 6.2 Character Death

- When `hp <= 0` → mark `is_alive = false`, set `died_at`
- Server sends `character_died` WS event to player
- Player may deploy a new character (up to the 2-character cap)
- No respawn limit — idle loop continues indefinitely

### 6.3 Monster Spawn

**On zone load (synchronous, via `GET /api/map/zones`):**
1. FE sends user's current coordinate (lat/lng) to `GET /api/map/zones?lat=X&lng=Y`
2. BE computes the res-6 H3 zone from the coordinate
3. Count living monsters in that zone
4. If count < `ZoneMonsterCap` (default: 300) → spawn `cap - count` new monsters
5. Each monster spawns at a random res-12 child cell within the zone
6. Monster type is random from `monster_types` table
7. Return all monsters in the zone
8. FE shows a loading indicator during this call (monster creation may take time on first load)

**Respawn (async, via game engine tick):**
- Dead monsters respawn after 30 seconds (set `respawn_at`, engine checks each tick)
- Engine counts living monsters per zone each tick; if below cap, spawns new ones

### 6.4 Zone Monster Cap

| Config | Value |
|--------|-------|
| ZoneMonsterCap | 300 |
| MonsterRespawnDelay | 30 seconds |
| TickInterval | 2 seconds |

---

## 7. Game Engine

### 7.1 Goroutine Architecture

- 1 goroutine per res-6 zone
- Each goroutine runs an independent tick loop every 2 seconds
- Zones are isolated — no shared monster state between goroutines

```
Engine.Start()
└── for each zone → go runZoneLoop(zoneH3)
    └── every 2s → tickZone(zoneH3)
        ├── check monster respawns
        ├── build engaged map (in-memory, this tick only)
        ├── for each living character in zone
        │   ├── check character_engagements for existing target
        │   │   ├── if monster still alive → combat(), add to engaged map
        │   │   └── if monster dead → DELETE engagement, find new target
        │   ├── if no engagement → findNearestFreeMonster (k-ring, skip engaged)
        │   │   ├── if found → INSERT character_engagements, combat()
        │   │   └── if not found → DELETE engagement if any, wander()
        └── broadcastZoneState() via WebSocket
```

### 7.2 Nearest Monster Search

K-ring expanding search at res-12, capped at k=5:

```
k=1 → check 7 cells
k=2 → check 19 cells
...
k=5 → check 91 cells
if still none → character wanders
```

### 7.3 Character Wandering

- Pick random k=1 neighbor at res-12
- Validate neighbor's res-6 parent == character's zone (stay in zone)
- If at zone edge with no valid neighbors → stay in place
- Update `h3_index` in DB

### 7.4 Zone Entry — Center Cell

When deploying into a zone, find the res-12 cell closest to the res-6 centroid using haversine distance.

---

## 8. Authentication & Player Initialization

### Core Idea: Fingerprint → Token → localStorage

No passwords, no OAuth, no cookies. A single UUID token identifies the player.

- Same tab / new tab = same `localStorage` = same player
- Clear storage / incognito = fresh `localStorage` = new player

### First Visit Flow (Frontend)

```
1. Check localStorage for { playerId, sessionToken }
2. If not found:
   a. POST /api/player/init (no body)
   b. Receive { playerId, sessionToken, name } from server
   c. Save to localStorage
3. If found → skip init, reuse existing token
4. Open WebSocket: ws://.../ws?token=SESSION_TOKEN
```

### Session Token

- Generated server-side (UUID)
- Stored in `players.session_token` column (unique index)
- Sent by client on every request:
  - **HTTP:** `Authorization: Bearer SESSION_TOKEN` header
  - **WebSocket:** `?token=SESSION_TOKEN` query param
- Invalid / missing token → 401 for HTTP, close connection for WS

### Server-Side Auth Middleware

- `SessionAuth()` Gin middleware on all `/api/*` routes except `/api/player/init`
- Extracts token from `Authorization` header
- Looks up `players` row by `session_token`
- Sets player in Gin context (`c.Set("player", player)`)
- Returns 401 if token is missing or not found

---

## 9. API Endpoints

### HTTP

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/player/init` | None | Create player, return session token |
| GET | `/api/map/zones?lat=X&lng=Y` | Bearer token | Compute zone from coordinate, ensure monsters are spawned, return monsters in zone |
| POST | `/api/character/deploy` | Bearer token | Deploy character to a zone |

### WebSocket

**Endpoint:** `ws://.../ws?token=SESSION_TOKEN`

**Zone Subscription**

Each player connection tracks one active zone at a time via `activeZone` field on the server-side `PlayerConn` struct. The engine only broadcasts `tick_update` to players subscribed to that zone. Switching zones updates `activeZone` immediately — player stops receiving old zone ticks.

**Server → Client events:**

```jsonc
// Every tick (2s per zone)
{
  "type": "tick_update",
  "zone": "8f...",
  "monsters": [
    { "id": "uuid-orc", "type": "Orc", "current_hp": 87, "max_hp": 200, "h3_index": "8f..." },
    { "id": "uuid-slime", "type": "Slime", "current_hp": 200, "max_hp": 200, "h3_index": "8f..." }
  ],
  "characters": [
    { "id": "uuid-hero", "name": "Hero", "hp": 210, "max_hp": 300, "player_id": "...", "h3_index": "8f...", "fighting_monster_id": "uuid-orc" },
    { "id": "uuid-mage", "name": "Mage", "hp": 180, "max_hp": 300, "player_id": "...", "h3_index": "8f...", "fighting_monster_id": null }
  ]
}

// On character death
{ "type": "character_died", "character_id": "uuid", "killed_by": "monster_name" }

// Combat event (sent each tick per combat pair, in-memory only — not persisted)
{ "type": "combat_log", "attacker": "Goblin", "defender": "Hero", "damage": 42, "is_crit": true }

// On monster death
{ "type": "monster_died", "monster_id": "uuid", "killed_by": "character_name", "kills_total": 5 }
```

**Client → Server events:**

```jsonc
// Subscribe to a zone (on map click or deploy)
{ "type": "subscribe_zone", "h3_zone": "8f2830828052d25" }

// Deploy character to zone
{ "type": "deploy_character", "h3_zone": "8f2830828052d25" }

// Keepalive
{ "type": "ping" }
```

---

## 10. Frontend

### Map View

- Render all res-6 zone hexagons using `h3ToGeoBoundary` → Leaflet `Polygon`
- Zone color indicates activity:
  - Grey → no characters, no player interest
  - Yellow → has monsters, no characters
  - Green → character active here
  - Red → character in combat
- Monster count badge on each hex
- Click hex → deploy character modal

### Character Panel

- Show 2 character slots (active / empty / dead)
- Each slot: name, HP bar, stats, current zone, kill count
- "Deploy" button when slot is empty

### Combat Log

- Scrolling feed of recent combat events
- Show: attacker → defender, damage, 💥 if crit

### Player Info Panel

- Show player name
- Total kills counter

---

## 11. Security (Public Demo)

| Concern | Mitigation |
|---------|-----------|
| Spam deploys | Rate limit: max 1 deploy per 3 seconds per session token |
| Forged damage | All combat runs server-side, client never sends damage values |
| Invalid H3 input | Validate all h3_zone/h3_index values server-side before DB write |
| DB corruption | SQLite WAL mode + `PRAGMA synchronous=NORMAL` |
| WS flooding | Rate limit incoming WS messages per connection |
| Anonymous writes | Session token required on all requests — Bearer header for HTTP, query param for WS |
| Admin exposure | No admin endpoints; seeding runs once at startup only |
| CORS | Whitelist deployed domain only |

---

## 12. Seeding

Run once at server startup if `monster_types` table is empty:

1. Seed 5 monster types (fixed data, see Section 4.1)
2. Compute 19 res-6 zones via `GridDisk(bangkokCenter, 2)`
3. For each zone, pick 5 random res-12 children → create `map_monsters`
4. Total: ~95 monsters at startup

---

## 13. Configuration Constants

```go
const (
    ZoneMonsterCap      = 300
    MaxCharactersAlive  = 2
    TickIntervalSeconds = 2
    MonsterRespawnSecs  = 30
    KRingSearchMax      = 5
    BangkokLat          = 13.7563
    BangkokLng          = 100.5018
    ZoneResolution      = 6
    EntityResolution    = 12
    GridDiskRadius      = 2   // zones around Bangkok center
)
```

---

## 14. README Sections (for GitHub)

1. **What is this** — one paragraph demo description
2. **Architecture** — goroutine per zone design, H3 dual resolution
3. **H3 spatial model** — res-6 zones vs res-12 entities explanation
4. **Combat formula** — the math
5. **How to run locally**
   ```bash
   # Backend
   cd backend && go run ./cmd/server

   # Frontend
   cd frontend && npm install && npm run dev
   ```
6. **Tech choices** — why Go, why H3, why SQLite for demo

---

## 15. Out of Scope (for demo)

- Character movement feels natural / pathfinding (placeholder: random wander)
- Player accounts / persistent login
- Leaderboard
- Multiple maps
- Character upgrade system
- Mobile responsiveness
- Redis / external cache