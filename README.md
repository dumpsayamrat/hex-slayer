# HexSlayer

A real-time idle geo-based monster hunting game built with Go, Java (Spring Boot), and React. Characters autonomously wander an H3 hex-grid world, detect nearby monsters, pursue them, and engage in turn-based combat — all streamed live to the browser over WebSocket.

### [Live Demo — hexslayer.dumpsayamrat.com](https://hexslayer.dumpsayamrat.com)

## What It Does

- **Hex-grid world** — The map is divided into zones using Uber's [H3 spatial index](https://h3geo.org/) (resolution 6 for zones, resolution 12 for entity placement). Monsters spawn across inner and outer rings of each zone with weighted distribution.
- **Autonomous characters** — Deploy characters into a zone and watch them act on their own: wander with natural-looking bearing drift, scan for nearby monsters using ring-by-ring search, move toward targets, and engage in combat.
- **Real-time game loop** — A server-side tick engine runs one goroutine per active zone. Each tick processes character state transitions (wandering → scanning → hunting → combat) and broadcasts events to all connected clients.
- **Combat system** — Characters and monsters have stats (base damage, damage amp/reduction, crit chance/multiplier, HP). Combat resolves per-tick with attack rolls, crits, and HP tracking until one side falls.
- **Live map** — A Leaflet-based frontend renders the hex grid, monsters, and characters in real time. Combat logs stream into a side panel.

## Tech Stack

### Backend — Go 1.23
- **Gin** — HTTP router and middleware
- **GORM + SQLite** — ORM with auto-migration and seed data
- **Gorilla WebSocket** — Real-time event streaming with topic-based pub/sub
- **uber/h3-go/v4** — H3 geospatial indexing (CGo bindings)
- **Swagger** (swaggo) — Auto-generated API docs

### Backend — Java 25 (Spring Boot 4)
- **Spring Web** — REST controllers and middleware (interceptors)
- **Spring Data JPA + Hibernate + SQLite** — ORM with auto-migration and seed data
- **Spring WebSocket** — Real-time event streaming with topic-based pub/sub
- **H3-Java** (com.uber:h3) — H3 geospatial indexing
- **Jackson** — JSON serialization with snake_case naming

### Frontend (React 18)
- **Vite** — Dev server with HMR
- **React-Leaflet** — Interactive map rendering
- **h3-js** — Client-side H3 hex boundary computation
- **Tailwind CSS** — Styling

## Project Structure

```
backend/
├── cmd/server/main.go          # Entry point, route setup
├── internal/
│   ├── config/                  # Game constants (tick rate, radii, caps)
│   ├── db/                      # SQLite init, migrations, seeding
│   ├── apperr/                   # Domain error types (Validation, NotFound)
│   ├── game/
│   │   ├── engine.go            # Tick loop, zone goroutine management
│   │   ├── tick.go              # Per-character state machine (combat, hunting, scanning, wandering)
│   │   ├── combat.go            # Damage calculation, attack rolls
│   │   ├── search.go            # Ring-by-ring monster detection, pathfinding
│   │   └── wander.go            # Bearing-based movement with drift
│   ├── dto/                      # Shared response types
│   ├── handlers/                # HTTP + WebSocket handlers
│   ├── middleware/               # Auth (Bearer token), rate limiting
│   ├── models/                  # GORM models (Player, Character, MapMonster, etc.)
│   ├── services/                # Zone monster spawning, character deployment
│   └── ws/                      # WebSocket hub with topic pub/sub
backend-java/
├── src/main/java/com/hexslayer/
│   ├── HexslayerApplication.java   # Entry point, engine bootstrap
│   ├── config/                     # WebSocket, CORS, data seeder
│   ├── controller/                 # REST controllers (Health, Player, Map, Character)
│   ├── dto/                        # Response DTOs
│   ├── exception/                  # Global error handler
│   ├── game/                       # Game engine (tick loop, combat, movement)
│   ├── middleware/                  # Auth interceptor (Bearer token)
│   ├── model/                      # JPA entities (Player, Character, MapMonster, etc.)
│   ├── repository/                 # Spring Data JPA repositories
│   ├── service/                    # Zone spawning, player, character services
│   └── ws/                         # WebSocket hub + handler
frontend/
├── src/
│   ├── components/
│   │   ├── Map.jsx              # Leaflet map with hex overlay
│   │   ├── ZoneHex.jsx          # H3 hex boundary rendering
│   │   ├── CharacterPanel.jsx   # Character stats display
│   │   └── CombatLog.jsx        # Live combat event feed
│   └── hooks/
│       └── useGameSocket.js     # WebSocket hook with auto-reconnect
```

## How to Run

### Prerequisites

- Go 1.23+ (for Go backend)
- Java 25+ and Maven 3.9+ (for Java backend)
- Node.js 18+
- C compiler (gcc/build-essential) — required by uber/h3-go CGo bindings
- CMake

### Backend (Go)

```bash
cd backend
go run ./cmd/server
```

The server starts on `http://localhost:8080`. Swagger docs are available at `/swagger/index.html`.

### Backend (Java)

```bash
cd backend-java
mvn spring-boot:run
```

The server starts on `http://localhost:8080`. Both backends expose the same API and are interchangeable.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

The dev server starts on `http://localhost:5173` and proxies API requests to the backend.

## Deployment

### Build locally

```bash
make build
```

### First-time server setup

```bash
make setup HOST=<ip>
```

Installs nginx, creates the systemd service, and configures the reverse proxy.

### Deploy

```bash
make deploy HOST=<ip>
```

Uploads the Go binary + frontend build, restarts the service.

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| POST | `/api/player/init` | Create a player session |
| GET | `/api/map/zones` | Get zone monsters (spawns if needed) |
| POST | `/api/character/deploy` | Deploy a character into a zone |
| GET | `/ws` | WebSocket for real-time game events |
