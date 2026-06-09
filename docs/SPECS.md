# OtterBoard – Product Specifications

## Goal

A self-hosted project management tool (Jira-like) built in Go. Designed to manage software development projects within the Otter Labs ecosystem and to be AI-agent-friendly from the start.

---

## Terminology

| Term | Meaning |
|------|---------|
| **Workspace** | Top-level organizational unit; contains projects and members |
| **Project** | A collection of issues within a workspace |
| **Issue** | Primary entity — any unit of work tracked in a project |
| **Type** | Kind of issue: `bug`, `task`, `story`, `epic` |
| **Status** | Lifecycle state of an issue: e.g. To Do, In Progress, Done |

> "Task" refers to an issue type, not the top-level entity. Use "issue" consistently in code, routes, and docs.

---

## Workspace Model

- Each workspace has its own members, projects, settings, and API keys.
- A user can belong to multiple workspaces.
- Workspace is the auth and permission boundary.

---

## Roles and Permissions

Every workspace member has one of three roles: `guest`, `member`, `administrator`.

The workspace creator is automatically added as `administrator`.

### Permission matrix

| Action | Guest | Member | Administrator |
|--------|-------|--------|---------------|
| GET workspace details | ✓ | ✓ | ✓ |
| PATCH / DELETE workspace | ✗ | ✗ | ✓ |
| GET members list | ✓ | ✓ | ✓ |
| PATCH / DELETE member | ✗ | ✗ | ✓ |
| POST invite | ✗ | ✗ | ✓ |
| GET / POST / DELETE api-keys | ✗ | ✗ | ✓ |
| GET projects | ✓ | ✓ | ✓ |
| POST / PATCH project | ✗ | ✓ | ✓ |
| DELETE project | ✗ | ✗ | ✓ |
| GET issues | ✓ | ✓ | ✓ |
| POST / PATCH issue | ✗ | ✓ | ✓ |
| DELETE issue | ✗ | ✗ | ✓ |

### Enforcement architecture

Authorization is enforced at the HTTP layer via two Fiber middleware functions in `internal/middleware/`:

- **`RequireWorkspaceMember`** — applied at the workspace route group level (`/workspaces/:workspaceId/*`). Verifies the authenticated user is a member of the workspace. Stores their role in `c.Locals("workspaceRole")`. Returns 403 if not a member.
- **`RequireRole(roles ...string)`** — applied per-route or per-route-group. Reads the role from `c.Locals` and returns 403 if it is not in the allowed list.

Services do not perform authorization checks. They receive no `RequestorID` and operate purely as business logic.

---

## UI Structure

### Sidebar

- Workspaces (switcher)
- Home
- My Issues
- Members
- Projects
- Settings

### Home View

- Stats row: Total Projects, Total Issues, Assigned Issues, Completed Issues, Overdue Issues
- List of issues assigned to the current user

### My Issues View

**Table**
- Columns: Issue Name, Project, Assignee, Due Date, Status — all sortable
- Filters: Status, Assignee, Project, Due Date

**Kanban**
- Drag-and-drop issue cards across status columns

**Calendar**
- Issues placed by due date

### Individual Issue Page

- Breadcrumb: Project Name > Issue Name
- Fields: Overview (editable), Assignee, Due Date, Status, Type
- Actions: Save changes, Delete issue

### Members View

- List all workspace members
- Assign roles: Guest, Member, Administrator

### Project View

- Stats row: Total Issues, Assigned Issues, Incomplete Issues, Complete Issues, Overdue Issues
- Same table/kanban/calendar views as My Issues, scoped to the project
- Edit project: name, image
- Danger zone: delete project (irreversible)

### Settings View

- Edit workspace name and icon
- Invite members via shareable link
- Manage API keys (create, revoke)
- Danger zone: delete workspace (irreversible)

---

## Core Features (v1 Scope)

- [ ] OAuth + email/password authentication
- [ ] API key management (create, revoke, scope to workspace)
- [ ] Workspace creation and management
- [ ] Member invites and role assignment
- [ ] Projects and boards
- [ ] Issue types: bug, task, story, epic
- [ ] Issue table view with filters and sorting
- [ ] Kanban board with drag-and-drop
- [ ] Calendar view
- [ ] Individual issue pages (edit overview, assign, set due date / status / type)
- [ ] Home dashboard with stats
- [ ] Webhooks (register endpoints, receive issue/project events)
- [ ] Real-time updates (board changes reflected live)
- [ ] Notifications (in-app + email)

---

## Nice to Have / Maybe

Not in v1 scope but not discarded:

- [ ] Scrum board view (alongside Kanban)
- [ ] Sprints and backlogs
- [ ] Comments and activity log
- [ ] Attachments
- [ ] Issue hierarchy beyond basic (epic → story → task nesting)
- [ ] MCP server (wrapper around the REST API for Claude Code integration)
- [ ] WebSocket upgrade (only if a feature requires browser → server streaming)
