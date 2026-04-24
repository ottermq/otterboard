# OtterBoard – SSE Event Schema

## Overview

The real-time stream delivers workspace events to connected browser clients via Server-Sent Events (SSE).

- Endpoint: `GET /api/v1/workspaces/:workspaceId/stream`
- Content-Type: `text/event-stream`
- Scope: all events for the workspace — client ignores what it doesn't need
- Auth: same `Authorization: Bearer <token>` header as REST endpoints

---

## Connection

The browser opens a persistent HTTP connection. The server pushes events as they occur and sends periodic keepalive pings to prevent proxy timeouts.

```
GET /api/v1/workspaces/:workspaceId/stream
Authorization: Bearer <token>
Accept: text/event-stream
```

### Keepalive

A comment line is sent every 30 seconds to keep the connection alive through proxies and load balancers:

```
: keepalive
```

---

## Event Format

SSE events follow the standard format:

```
event: <event-type>
data: <json-payload>

```

Example:

```
event: issue.status_changed
data: {"issue_id":"uuid","project_id":"uuid","current_status":"in_progress"}

```

---

## Payloads

SSE payloads are intentionally slim. The UI uses them to know what to refetch, not to update state directly.

### Issue Events

#### `issue.created`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid"
}
```

#### `issue.deleted`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid"
}
```

#### `issue.assigned`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "assignee_id": "uuid"
}
```

#### `issue.status_changed`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "current_status": "in_progress"
}
```

#### `issue.type_changed`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "current_type": "bug"
}
```

#### `issue.due_date_changed`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "current_due_date": "2024-01-25"
}
```

#### `issue.title_changed`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "current_title": "New title"
}
```

#### `issue.overview_changed`
```json
{
  "issue_id": "uuid",
  "project_id": "uuid"
}
```

### Project Events

#### `project.created`
```json
{
  "project_id": "uuid"
}
```

#### `project.updated`
```json
{
  "project_id": "uuid"
}
```

#### `project.deleted`
```json
{
  "project_id": "uuid"
}
```

---

## Reconnection

SSE reconnects automatically in the browser via the `EventSource` API. The server does not need to replay missed events for v1 — the client refetches on reconnect.
