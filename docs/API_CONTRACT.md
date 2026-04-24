# OtterBoard – API Contract

## Conventions

- Base path: `/api/v1`
- All request and response bodies are JSON
- Timestamps are ISO 8601 (e.g. `2024-01-15T10:30:00Z`)
- IDs are UUIDs
- Pagination: `?page=1&limit=20` on all list endpoints
- Auth: session cookie (browser) or `Authorization: Bearer <api-key>` (agents)
- No refresh token — sessions are managed server-side via GoodiesDB with rolling expiry; logout invalidates the session immediately
- Errors follow a consistent shape:

```json
{
  "error": "human-readable message"
}
```

### Paginated response envelope

All list endpoints return:

```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

---

## Auth

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register with email + password |
| POST | `/api/v1/auth/login` | Login with email + password |
| POST | `/api/v1/auth/logout` | Invalidate current session |
| GET | `/api/v1/auth/oauth/:provider` | Initiate OAuth flow |
| GET | `/api/v1/auth/oauth/:provider/callback` | OAuth callback |
| GET | `/api/v1/auth/me` | Get current authenticated user |

---

## Workspaces

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces` | Create a workspace |
| GET | `/api/v1/workspaces` | List workspaces for current user |
| GET | `/api/v1/workspaces/:workspaceId` | Get workspace details |
| PATCH | `/api/v1/workspaces/:workspaceId` | Update workspace (name, icon) |
| DELETE | `/api/v1/workspaces/:workspaceId` | Delete workspace (danger zone) |

---

## Members

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/workspaces/:workspaceId/members` | List workspace members |
| PATCH | `/api/v1/workspaces/:workspaceId/members/:userId` | Update member role |
| DELETE | `/api/v1/workspaces/:workspaceId/members/:userId` | Remove member from workspace |

### Invites

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces/:workspaceId/invites` | Generate shareable invite link |
| GET | `/api/v1/invites/:token` | Get invite details (public) |
| POST | `/api/v1/invites/:token/accept` | Accept invite and join workspace |

---

## API Keys

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces/:workspaceId/api-keys` | Create API key (raw key returned once) |
| GET | `/api/v1/workspaces/:workspaceId/api-keys` | List API keys (no raw tokens) |
| DELETE | `/api/v1/workspaces/:workspaceId/api-keys/:keyId` | Revoke API key |

---

## Projects

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces/:workspaceId/projects` | Create project |
| GET | `/api/v1/workspaces/:workspaceId/projects` | List projects in workspace |
| GET | `/api/v1/workspaces/:workspaceId/projects/:projectId` | Get project details |
| PATCH | `/api/v1/workspaces/:workspaceId/projects/:projectId` | Update project (name, image) |
| DELETE | `/api/v1/workspaces/:workspaceId/projects/:projectId` | Delete project (danger zone) |

---

## Issues

### Project-scoped

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces/:workspaceId/projects/:projectId/issues` | Create issue |
| GET | `/api/v1/workspaces/:workspaceId/projects/:projectId/issues` | List issues in project |
| GET | `/api/v1/workspaces/:workspaceId/projects/:projectId/issues/:issueId` | Get issue details |
| PATCH | `/api/v1/workspaces/:workspaceId/projects/:projectId/issues/:issueId` | Update issue |
| DELETE | `/api/v1/workspaces/:workspaceId/projects/:projectId/issues/:issueId` | Delete issue |

### Workspace-scoped (My Issues / cross-project views)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/workspaces/:workspaceId/issues` | List all issues in workspace |

### Query parameters (list endpoints)

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status: `backlog`, `todo`, `in_progress`, `in_review`, `done` |
| `type` | string | Filter by type: `bug`, `task`, `story`, `epic` |
| `assignee` | UUID or `me` | Filter by assignee |
| `project` | UUID | Filter by project (workspace-scoped list only) |
| `due_before` | date | Filter issues due before this date |
| `due_after` | date | Filter issues due after this date |
| `sort` | string | Field to sort by: `title`, `status`, `due_date`, `created_at` |
| `order` | string | `asc` or `desc` (default: `asc`) |
| `page` | int | Page number (default: `1`) |
| `limit` | int | Page size (default: `20`, max: `100`) |

---

## Webhooks

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/workspaces/:workspaceId/webhooks` | Register webhook endpoint |
| GET | `/api/v1/workspaces/:workspaceId/webhooks` | List registered webhooks |
| PATCH | `/api/v1/workspaces/:workspaceId/webhooks/:webhookId` | Update webhook URL or subscribed events |
| DELETE | `/api/v1/workspaces/:workspaceId/webhooks/:webhookId` | Delete webhook |

---

## Real-time

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/workspaces/:workspaceId/stream` | SSE stream for workspace events |

Browser clients authenticate via session cookie (sent automatically). Agents authenticate via `Authorization: Bearer <api-key>` using an HTTP streaming client that supports custom headers.
