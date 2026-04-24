# OtterBoard – Data Model

## users

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| email | VARCHAR | unique, not null |
| password_hash | VARCHAR | nullable (OAuth-only users have no password) |
| name | VARCHAR | not null |
| avatar_url | VARCHAR | nullable |
| created_at | TIMESTAMP | not null |

---

## user_identities

Stores OAuth provider identities. A user can have multiple (e.g. GitHub + Google).

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| user_id | UUID | FK → users.id |
| provider | VARCHAR | e.g. `github`, `google` |
| provider_user_id | VARCHAR | ID from the provider |
| created_at | TIMESTAMP | not null |

Unique constraint on `(provider, provider_user_id)`.

---

## workspaces

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| name | VARCHAR | not null |
| icon_url | VARCHAR | nullable |
| created_at | TIMESTAMP | not null |

---

## workspace_members

Join table between users and workspaces. Also carries the role.

| Column | Type | Notes |
|--------|------|-------|
| user_id | UUID | FK → users.id |
| workspace_id | UUID | FK → workspaces.id |
| role | VARCHAR | `guest`, `member`, `administrator` |
| joined_at | TIMESTAMP | not null |

PK on `(user_id, workspace_id)`.

---

## projects

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| workspace_id | UUID | FK → workspaces.id |
| name | VARCHAR | not null |
| image_url | VARCHAR | nullable |
| created_at | TIMESTAMP | not null |

---

## issues

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects.id |
| title | VARCHAR | not null |
| overview | TEXT | nullable |
| type | VARCHAR | `bug`, `task`, `story`, `epic` — enforced at app level |
| status | VARCHAR | `backlog`, `todo`, `in_progress`, `in_review`, `done` — enforced at app level |
| assignee_id | UUID | FK → users.id, nullable |
| created_by | UUID | FK → users.id, not null |
| due_date | DATE | nullable |
| created_at | TIMESTAMP | not null |
| updated_at | TIMESTAMP | not null |

---

## api_keys

Workspace-scoped static tokens for agents and integrations. The raw key is shown once on creation — only its hash is stored.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| workspace_id | UUID | FK → workspaces.id |
| user_id | UUID | FK → users.id (creator) |
| name | VARCHAR | human-readable label |
| key_hash | VARCHAR | hashed token — raw key never stored |
| created_at | TIMESTAMP | not null |
| revoked_at | TIMESTAMP | nullable — null means active |

---

## webhooks

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| workspace_id | UUID | FK → workspaces.id |
| url | VARCHAR | delivery endpoint, not null |
| events | VARCHAR[] | subscribed event types, e.g. `issue.created`, `issue.assigned` |
| created_at | TIMESTAMP | not null |
