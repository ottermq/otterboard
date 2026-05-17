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
| updated_at | TIMESTAMP | not null |

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
| owner_id | UUID | FK → users.id, not null |
| icon_url | VARCHAR | nullable |
| created_at | TIMESTAMP | not null |
| updated_at | TIMESTAMP | not null |

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
| updated_at | TIMESTAMP | not null |

---

## issues

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| project_id | UUID | FK → projects.id |
| title | VARCHAR | not null |
| overview | TEXT | nullable |
| type | VARCHAR | `bug`, `task`, `story`, `epic` — CHECK constraint + app-level validation |
| status | VARCHAR | `backlog`, `todo`, `in_progress`, `in_review`, `done` — CHECK constraint + app-level validation |
| position | FLOAT | ordering within a status column for Kanban; insert between two items by averaging their positions |
| assignee_id | UUID | FK → users.id, nullable |
| created_by | UUID | FK → users.id, not null |
| due_date | DATE | nullable |
| created_at | TIMESTAMP | not null |
| updated_at | TIMESTAMP | not null |

### CHECK constraints

```sql
CONSTRAINT issues_type_check CHECK (type IN ('bug', 'task', 'story', 'epic')),
CONSTRAINT issues_status_check CHECK (status IN ('backlog', 'todo', 'in_progress', 'in_review', 'done'))
```

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

---

## invites

Shareable invite links for joining a workspace. A link is single-use — `used_at` is set on acceptance.

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| workspace_id | UUID | FK → workspaces.id, ON DELETE CASCADE |
| created_by | UUID | FK → users.id, nullable — SET NULL if creator is deleted |
| token | TEXT | unique random token; included in the invite URL |
| created_at | TIMESTAMPTZ | not null |
| expires_at | TIMESTAMPTZ | not null |
| used_at | TIMESTAMPTZ | nullable — null means not yet accepted |

---

## Indexes

Documented here for reference — to be created in migrations.

| Table | Column(s) | Reason |
|-------|-----------|--------|
| issues | `project_id` | most issue queries are project-scoped |
| issues | `assignee_id` | My Issues view, workspace-scoped issue list |
| issues | `status` | filtering and Kanban column queries |
| issues | `due_date` | calendar view, overdue filters |
| api_keys | `key_hash` | authentication hot path — every agent request |
| workspace_members | `workspace_id` | member list, permission checks |
