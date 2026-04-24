# OtterBoard – Technical Design

## Tech Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Language | Go | Consistent with the rest of Otter Labs |
| Backend framework | Fiber | Same choices as OtterMQ |
| API style | REST | Simpler for Go CRUD; GraphQL not justified at this scale |
| Frontend | React | Exploring new possibilities beyond Quasar/Vue |
| Database | PostgreSQL | Persistent relational data |
| Real-time (backend) | OtterMQ | Async event routing between services; webhook delivery |
| Real-time (browser) | SSE (Server-Sent Events) | Server → client push; simpler than WebSocket, sufficient for read-only event stream |
| Caching / sessions | GoodiesDB | Session store, hot data cache |
| Auth | OAuth + email/password + API keys | Human auth via OAuth/email; agent/MCP auth via API keys |

---

## Architecture

### Core Components

- **Workspace service** — create, manage, invite members, delete workspace
- **Project service** — CRUD for projects within a workspace
- **Issue service** — issue lifecycle (create, assign, transition, close)
- **User service** — OAuth/email auth, API key management, roles, permissions per workspace
- **Webhook service** — register endpoints, dispatch events on issue/project changes
- **Real-time gateway** — subscribes to OtterMQ, fans out relevant events to connected browser clients via SSE
- **Notification service** — async jobs via OtterMQ (email, in-app alerts, webhook delivery)
- **Frontend SPA** — React, issue views, drag-and-drop Kanban

### Data Flow

```
[Browser / Agent] → [REST API]
                       ↓
                  [Service layer] → [PostgreSQL]
                       ↓
                  [OtterMQ] → [Real-time gateway] → [Browser (SSE)]
                            → [Notification service] → [Email / in-app]
                            → [Webhook service] → [External endpoints / MCP]
```

---

## Real-time Architecture

OtterMQ and SSE are not alternatives — they operate at different layers:

| Layer | Protocol | Direction | Purpose |
|-------|----------|-----------|---------|
| Backend | OtterMQ (AMQP) | Service → Service | Event routing, async jobs, webhook delivery |
| Browser | SSE (HTTP) | Server → Browser | Live board updates, notifications |

The **real-time gateway** is the bridge: it subscribes to OtterMQ topics and pushes relevant events to connected SSE clients.

Browsers cannot speak AMQP directly. They receive a plain HTTP event stream (`text/event-stream`). WebSocket is not needed because the browser only receives events — it sends actions via REST (`PATCH /issues/:id`), which triggers the event pipeline.

---

## Auth Design

Three auth strategies, each for a different actor:

| Strategy | Actor | Mechanism |
|----------|-------|-----------|
| OAuth | Human users | Provider redirect flow (Google, GitHub, etc.) |
| Email/password | Human users | Local credentials, hashed storage |
| API keys | Agents / integrations | Static token, scoped to workspace, no interactive flow |

API keys are a first-class feature — created and revoked via the Settings UI and API. Required for any non-human actor (CI pipelines, Claude Code, MCP servers).

---

## AI / MCP Integration Design

The REST API is the integration surface for AI agents. An MCP server is a thin wrapper that exposes tools backed by REST calls.

### Requirements

- **API keys** — agents need static tokens; OAuth/email flows require human interaction
- **Webhooks** — push direction: OtterBoard notifies agents when issues are assigned or status changes; registered via the API, delivered via OtterMQ async jobs
- **Clean, filterable endpoints** — consistent resource naming so agent queries like "list open bugs in project X assigned to nobody" map directly to REST parameters

### Example Agent Interactions

- Claude Code reviews a codebase and opens a bug issue via `POST /workspaces/:id/projects/:id/issues`
- An issue is assigned to Claude Code → OtterBoard delivers a webhook event to a registered endpoint
- Claude Code queries open issues before starting work via `GET /issues?status=open&assignee=me`

A dedicated MCP server (thin REST wrapper) is planned as a future Nice to Have once the API is stable.

---

## Repository Structure

Folders are created on demand as each domain is implemented — none are pre-scaffolded.

```
otterboard_git/
├── src/
│   ├── backend/                  ← Go application
│   │   ├── cmd/
│   │   │   └── api/
│   │   │       └── main.go       ← entrypoint
│   │   ├── internal/
│   │   │   ├── auth/             ← OAuth, email/password, API key auth
│   │   │   ├── workspace/        ← workspace + member management
│   │   │   ├── project/          ← project CRUD
│   │   │   ├── issue/            ← issue lifecycle
│   │   │   ├── webhook/          ← webhook registration + dispatch
│   │   │   ├── notification/     ← async jobs (email, in-app)
│   │   │   ├── realtime/         ← SSE gateway + OtterMQ bridge
│   │   │   ├── middleware/       ← auth, CORS, etc.
│   │   │   ├── common/           ← shared helpers (responses, errors)
│   │   │   ├── config/           ← app configuration
│   │   │   └── db/
│   │   │       └── migrations/   ← numbered up/down SQL files
│   │   ├── pkg/
│   │   │   └── dtos/             ← request/response structs
│   │   ├── go.mod
│   │   └── go.sum
│   └── frontend/                 ← React SPA (Vite)
└── docs/                         ← architecture docs, API specs, ADRs
```

Each domain package under `internal/` follows the same layout:

```
internal/<domain>/
├── handler.go       ← HTTP handlers (Fiber route functions)
├── service.go       ← business logic
├── repository.go    ← database queries
├── model.go         ← domain structs
└── service_test.go  ← TDD tests
```
