# OtterBoard – Webhook Schema

## Overview

Webhooks deliver event notifications to registered endpoints via HTTP POST. Each event shares a common envelope with an event-specific `payload`.

---

## Delivery

- Method: `POST`
- Content-Type: `application/json`
- Timeout: 10 seconds per attempt
- Retry policy: 3 attempts with exponential backoff (5s, 30s, 5min)
- A delivery is considered successful when the endpoint returns a `2xx` status code
- After all retries are exhausted the delivery is marked as failed and dropped — no dead-letter queue in v1; the webhook owner can inspect failures via logs

---

## Signature Verification

Every delivery includes an HMAC signature header so receivers can verify the request came from OtterBoard:

```
X-OtterBoard-Signature: sha256=<hmac-hex>
```

The signature is computed as:

```
HMAC-SHA256(signing_key, request_body)
```

The signing key is set per webhook at registration time and stored securely. Receivers should reject requests where the signature does not match.

---

## Envelope

Every webhook delivery shares this top-level structure:

```json
{
  "id": "uuid",
  "event": "issue.status_changed",
  "workspace_id": "uuid",
  "occurred_at": "2024-01-15T10:30:00Z",
  "payload": { ... }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique delivery ID |
| `event` | string | Event type (see below) |
| `workspace_id` | UUID | Workspace where the event occurred |
| `occurred_at` | timestamp | ISO 8601 |
| `payload` | object | Event-specific data |

---

## Issue Events

### `issue.created`

```json
{
  "issue": {
    "id": "uuid",
    "project_id": "uuid",
    "title": "Fix login bug",
    "type": "bug",
    "status": "backlog",
    "assignee_id": null,
    "created_by": "uuid",
    "due_date": null,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### `issue.deleted`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid"
}
```

### `issue.assigned`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "assigned_to": "uuid",
  "assigned_by": "uuid"
}
```

### `issue.status_changed`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "previous_status": "todo",
  "current_status": "in_progress",
  "changed_by": "uuid"
}
```

### `issue.type_changed`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "previous_type": "task",
  "current_type": "bug",
  "changed_by": "uuid"
}
```

### `issue.due_date_changed`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "previous_due_date": "2024-01-20",
  "current_due_date": "2024-01-25",
  "changed_by": "uuid"
}
```

### `issue.title_changed`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "previous_title": "Old title",
  "current_title": "New title",
  "changed_by": "uuid"
}
```

### `issue.overview_changed`

```json
{
  "issue_id": "uuid",
  "project_id": "uuid",
  "changed_by": "uuid"
}
```

> Overview content is intentionally omitted from the payload — it can be large. Agents should fetch it via `GET /api/v1/workspaces/:workspaceId/projects/:projectId/issues/:issueId` if needed.

---

## Project Events

### `project.created`

```json
{
  "project": {
    "id": "uuid",
    "workspace_id": "uuid",
    "name": "OtterBoard",
    "image_url": null,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### `project.updated`

```json
{
  "project_id": "uuid",
  "workspace_id": "uuid",
  "changes": {
    "name": "New Name"
  }
}
```

### `project.deleted`

```json
{
  "project_id": "uuid",
  "workspace_id": "uuid"
}
```

---

## Webhook Registration

When registering a webhook, the caller specifies the URL and events to subscribe to:

```json
{
  "url": "https://example.com/webhook",
  "events": [
    "issue.created",
    "issue.assigned",
    "issue.status_changed"
  ]
}
```

An empty `events` array is invalid — at least one event type must be specified.

OtterBoard generates the signing key and returns it once in the registration response. It is never retrievable again — treat it like an API key.

```json
{
  "id": "uuid",
  "url": "https://example.com/webhook",
  "events": ["issue.created", "issue.assigned", "issue.status_changed"],
  "signing_key": "raw-secret-shown-once",
  "created_at": "2024-01-15T10:30:00Z"
}
```

Subsequent `GET /webhooks` responses omit `signing_key`.
