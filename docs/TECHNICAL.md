# TeleOpServer — Technical Reference

> Go backend for a teleoperation system — auth, command ingestion, GPS telemetry, and MQTT message brokering.
> v1.0 · 2026-06-10

---

## Table of Contents

1. [System Architecture](#1-system-architecture)
2. [Technology Stack](#2-technology-stack)
3. [Module Structure](#3-module-structure)
4. [Dependency Injection (Wire)](#4-dependency-injection-wire)
5. [Database](#5-database)
6. [MQTT Integration](#6-mqtt-integration)
7. [Authentication](#7-authentication)
8. [API Endpoints](#8-api-endpoints)
9. [Startup Sequence](#9-startup-sequence)
10. [Local Development](#10-local-development)
11. [Deployment](#11-deployment)

---

## 1. System Architecture

The backend handles persistence and authentication only. Real-time control and live video travel directly between the web client and robot via WebRTC P2P — never through this server. Low-fps telemetry (commands, GPS) flows robot → MQTT broker → backend → TimescaleDB.

```
Web Client ──WebRTC P2P──► Robot (live video + commands, real-time)
    │                         │
    │ REST (auth, queries)     │ MQTT (commands, GPS)
    ▼                         ▼
TeleOpServer            CloudAMQP (RabbitMQ)
    │                         │
    └──────── consumer ────────┘
                 │
           TimescaleDB
```

> **Core design principle:** The backend never sits in the real-time control path. Latency-sensitive traffic (joystick commands during active teleoperation) goes P2P. The backend only receives a copy for audit/replay purposes via the MQTT queue.

---

## 2. Technology Stack

| Concern | Library / Service | Notes |
|---|---|---|
| Language | Go 1.23 | |
| HTTP router | `go-chi/chi v5` | Lightweight, middleware-friendly |
| ORM | `gorm.io/gorm` | With `gorm.io/driver/postgres` |
| Database | TimescaleDB (PostgreSQL 16) | Time-series hypertables for all domain data |
| Message broker | CloudAMQP (RabbitMQ) | MQTT over TLS, port 8883, free tier |
| MQTT client | `eclipse/paho.mqtt.golang` | QoS 1, persistent session, auto-reconnect |
| DI | Google Wire | Compile-time dependency injection |
| Auth | JWT (HS256) via `golang-jwt/jwt v5` | Stateless, 24h expiry |
| Logging | `rs/zerolog` | Structured JSON, console writer in dev |
| API docs | Swagger via `swaggo/swag` | UI at `/swagger/index.html` |
| Config | `joho/godotenv` | Reads `.env` file |

---

## 3. Module Structure

The codebase is organized as vertical slices. Each domain module is fully self-contained — model, repository, service, handler, and consumer in one directory. Modules never import each other.

```
internal/
├── platform/                  # Shared infrastructure, no domain logic
│   ├── config/config.go       # Env-var config struct
│   ├── database/postgres.go   # GORM + TimescaleDB extension setup
│   ├── mqttclient/client.go   # Paho MQTT wrapper (subscribe-then-connect)
│   ├── middleware/
│   │   ├── auth.go            # JWT Bearer middleware
│   │   └── logging.go         # zerolog request logger
│   └── server/server.go       # Chi bootstrap + graceful shutdown
│
├── modules/
│   ├── user/                  # Auth and user management
│   ├── TOCommands/            # Teleop command ingestion and history
│   └── gps/                   # GPS reading ingestion and history
│
└── shared/
    └── events/bus.go          # Inter-module event bus (in-process)
```

| Module | Description |
|---|---|
| `user` | Register, login, JWT issuance, profile read and soft-delete. Regular Postgres table with `gorm.Model`. |
| `TOCommands` | Joystick command ingestion via MQTT and history queries. TimescaleDB hypertable partitioned by `Time`. |
| `gps` | GPS reading ingestion via MQTT and history queries. TimescaleDB hypertable partitioned by `Time`. |

### Files inside each module

| File | Purpose |
|---|---|
| `model.go` | DB struct, request/response types, mapping helpers |
| `repository.go` | Interface + GORM implementation (Save, FindByRobotAndTimeRange) |
| `service.go` | Business logic — sets server-side timestamp, maps to response types |
| `handler.go` | HTTP handlers with Swagger annotations |
| `consumer.go` | MQTT subscriber — parses payload, calls service.Record |
| `module.go` | Wire ProviderSet, Migrate(), RegisterRoutes() |

> **Import rule:** `platform/*` → external only. `modules/*` → platform + shared + external. Modules never import each other — inter-module communication uses `shared/events.Bus`.

---

## 4. Dependency Injection (Wire)

Wire generates a compile-time dependency graph — equivalent to source-generated DI in .NET 8. No reflection at runtime.

```
wire/
├── wire.go         # Injection point (wireinject build tag)
├── wire_gen.go     # Generated — do not edit
└── providers.go    # Composition root: MigrateAll() + provideRouter()
```

### Adding a new module

1. Create `internal/modules/<name>/module.go` with `ProviderSet`, `Migrate()`, `RegisterRoutes()`
2. Add `<name>.Migrate(db)` to `MigrateAll()` in `wire/providers.go`
3. Add `<name>.RegisterRoutes(...)` to `provideRouter()` in `wire/providers.go`
4. Add `<name>.ProviderSet` to `wire.Build()` in `wire/wire.go`
5. Run: `go run github.com/google/wire/cmd/wire ./wire/...`

---

## 5. Database

TimescaleDB (PostgreSQL extension) runs in Docker via `timescale/timescaledb:latest-pg16`. Two table types are used:

| Type | Tables | GORM model | Notes |
|---|---|---|---|
| Regular | `users` | Embeds `gorm.Model` | Soft delete via `DeletedAt` |
| Hypertable | `tele_op_commands`, `gps_readings` | No `gorm.Model` | `Time time.Time` is the partition key |

Hypertables cannot use auto-increment integer PKs because TimescaleDB requires the time column in all unique constraints. The `create_hypertable` call uses `if_not_exists => TRUE` so migrations are safe to re-run.

Each module owns its `Migrate(db)`. `MigrateAll()` in `wire/providers.go` calls them in order at startup before the HTTP server opens.

### gps_readings table

| Column | Type | Notes |
|---|---|---|
| `time` | timestamptz | Server-set UTC. Partition key. |
| `robot_id` | bigint | Indexed |
| `latitude` | float8 | |
| `longitude` | float8 | |

### tele_op_commands table

| Column | Type | Notes |
|---|---|---|
| `time` | timestamptz | Server-set UTC. Partition key. |
| `robot_id` | bigint | Indexed |
| `user_id` | bigint | |
| `command` | text | Command type identifier |
| `msg_id` | bigint | Message sequence number |
| `t1` | bigint | Timing metadata |
| `lx` | float8 | Left stick X |
| `ly` | float8 | Left stick Y |
| `az` | float8 | Angular Z (rotation) |

---

## 6. MQTT Integration

The backend subscribes to robot telemetry topics on CloudAMQP (RabbitMQ with MQTT plugin) over TLS. The integration is built around a deliberate subscribe-before-connect ordering to prevent message loss on startup.

### Subscribe-before-connect pattern

`mqttclient.New()` does **not** open the broker connection. All consumers call `Subscribe()` to register their handlers first, then `Connect()` is called. This ensures that when the broker delivers queued messages on first connect, handlers are already in place and no messages are dropped.

```
// providers.go startup order
gpsConsumer.Register()       // registers handler → c.subs
consumer.Start(ctx)          // registers handler → c.subs, then calls Connect()
  └── mqtt.Subscribe(topic)  // adds to c.subs
  └── mqtt.Connect()         // broker delivers queued msgs → handlers fire ✓
```

### Reconnect durability

The client uses `CleanSession: false` and QoS 1. On reconnect, the `OnConnectHandler` re-subscribes all topics from `c.subs` before any messages are delivered, using a snapshot taken outside the mutex to avoid deadlock with blocking network calls.

### Topics

| Topic | Publisher | Consumer | Payload |
|---|---|---|---|
| `teleopserver/robots/+/commands` | Web client | TOCommands | JSON |
| `teleopserver/robots/+/gps` | Robot | GPS | JSON |

### Environment variables

```
MQTT_BROKER=tls://cow.rmq2.cloudamqp.com:8883
MQTT_USERNAME=<vhost>:<user>
MQTT_PASSWORD=<password>
MQTT_CLIENT_ID=teleopserver
```

---

## 7. Authentication

JWT HS256 via `golang-jwt/jwt v5`. Tokens are stateless — no session store. The `Auth` middleware validates the Bearer token and injects `userID uint` into the request context.

```go
// Retrieve in handlers
userID := middleware.UserIDFromContext(r.Context())
```

Tokens expire after `JWT_EXPIRY_HOURS` (default 24h). Signing key is `JWT_SECRET` from env.

---

## 8. API Endpoints

All routes are under `/api/v1`. Swagger UI at `http://localhost:8080/swagger/index.html`. Time parameters use RFC3339 format (e.g. `2026-06-10T14:00:00Z`).

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | — | Register new user |
| POST | `/auth/login` | — | Login, returns JWT |
| GET | `/users/me` | JWT | Authenticated user profile |
| DELETE | `/users/me` | JWT | Soft-delete authenticated user |
| POST | `/commands` | JWT | Ingest a teleop command (REST fallback) |
| GET | `/commands?robot_id&from&to` | JWT | Command history for a robot in time range |
| GET | `/gps?robot_id&from&to` | JWT | GPS history for a robot in time range |

---

## 9. Startup Sequence

1. **Signal context created** — `main.go` listens for `SIGINT` / `SIGTERM`. Cancelling this context triggers graceful shutdown.
2. **Wire initializes the graph** — Config loaded, Postgres connected, MQTT client created (not yet connected), all module providers wired.
3. **DB migrations run** — `MigrateAll()` runs each module's `Migrate()` in order — AutoMigrate for schema, then `create_hypertable` for time-series tables.
4. **MQTT subscriptions registered** — All consumers call `Subscribe()` — handlers are stored in `c.subs`. No broker connection yet.
5. **MQTT connects, queued messages delivered** — `Connect()` opens the TLS connection. `OnConnectHandler` resubscribes from `c.subs`. Broker flushes any queued messages — all handlers are in place.
6. **HTTP server starts** — Chi router mounts all routes under `/api/v1`. Listens on `:8080`. On `Ctrl+C`: 10s graceful shutdown, MQTT disconnects cleanly.

---

## 10. Local Development

### Prerequisites

Docker Desktop, Go 1.23, Delve (for VS Code debugging).

### Start the database

```bash
docker-compose up -d postgres
```

Uses `timescale/timescaledb:latest-pg16`. Exposed on port **5433** (5432 is reserved for a local native Postgres installation).

### Run the server

```bash
go run ./cmd/api
```

Auto-migrates on startup. Swagger UI opens automatically at `http://localhost:8080/swagger/index.html`.

### Regenerate Wire after provider changes

```bash
go run github.com/google/wire/cmd/wire ./wire/...
```

### Regenerate Swagger after annotation changes

```bash
swag init -g cmd/api/main.go -o docs
```

### Environment (.env)

```
SERVER_PORT=8080
DATABASE_URL=host=localhost port=5433 dbname=teleopserver user=postgres password=postgres1 sslmode=disable
JWT_SECRET=change-me-in-production
JWT_EXPIRY_HOURS=24
MQTT_BROKER=tls://cow.rmq2.cloudamqp.com:8883
MQTT_USERNAME=<vhost>:<user>
MQTT_PASSWORD=<password>
MQTT_CLIENT_ID=teleopserver
```

### Test MQTT manually

Use MQTT Explorer connected to `cow.rmq2.cloudamqp.com:8883` (TLS). Publish to `teleopserver/robots/1/commands` or `teleopserver/robots/1/gps` with QoS 1 to test persistence when the server is offline.

---

## 11. Deployment

The app ships as a Docker image built via a two-stage Dockerfile. The database (TimescaleDB) runs as a separate container. CloudAMQP is already hosted — no broker container needed in production.

### Build overview

The Dockerfile uses a multi-stage build: `golang:1.22-alpine` compiles the binary, then `alpine:3.20` runs it. The final image contains only the compiled binary — no Go toolchain.

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Deployment sequence

1. **Prepare environment** — On the target host, create a `.env` file or inject environment variables. Set production values for `JWT_SECRET`, `DATABASE_URL`, and MQTT credentials. Never commit secrets to version control.
2. **Start the database** — Run TimescaleDB first. The `api` service depends on `postgres` being healthy before it starts.
3. **Build and start all services** — Docker Compose builds the Go binary and launches both containers. In production, `DATABASE_URL` host must be `postgres` (the service name), not `localhost`.
4. **Migrations run automatically** — On first boot the app runs `MigrateAll()`. Subsequent restarts are safe; all migrations are idempotent.
5. **Verify health** — Check logs for `TeleOpServer starting` and absence of MQTT or DB errors. Hit `/swagger/index.html` to confirm the HTTP server is up.

### Commands

```bash
# First deploy — wipe old volume if changing DB password
docker-compose down -v
docker-compose up -d

# Redeploy after code changes
docker-compose build api
docker-compose up -d --no-deps api

# View logs
docker-compose logs -f api
docker-compose logs -f postgres
```

### docker-compose services

| Service | Image | Port | Notes |
|---|---|---|---|
| `api` | Built from `./Dockerfile` | 8080 | Depends on `postgres` health check |
| `postgres` | `timescale/timescaledb:latest-pg16` | 5433 → 5432 | Port 5433 on host to avoid conflict with local Postgres |

> **Production DATABASE_URL note:** When running via `docker-compose up`, the `api` container must use `host=postgres` (the Docker service name) instead of `host=localhost`. Use `localhost` only when running `go run ./cmd/api` directly on the host machine.
