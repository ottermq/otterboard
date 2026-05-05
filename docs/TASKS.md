# OtterBoard – Tasks & Milestones

## M0 — Design

Define the foundations before any code.

- [x] Data model: workspaces, projects, issues, users, roles
- [x] API contract: resource naming, endpoint structure, query parameters
- [x] Webhook schema: event types, payload format, delivery guarantees
- [x] Auth flow: OAuth providers, email/password, API key structure
- [x] SSE event schema: what events get pushed and in what format
- [x] Folder structure and Go module setup

---

## M1 — Auth + Workspace + Members

- [ ] OAuth login (at least one provider)
- [x] Email/password login and registration
- [ ] API key creation and revocation
- [x] Workspace creation
- [x] Workspace read, update, delete (owner only)
- [ ] Member invite via shareable link
- [ ] Role assignment (Guest, Member, Administrator)
- [ ] Permission checks across all workspace-scoped operations

---

## M2 — Projects + Issues CRUD

- [ ] Project create, read, update, delete
- [ ] Issue create, read, update, delete
- [ ] Issue fields: title, overview, type, status, assignee, due date
- [ ] Basic REST API with proper status codes and error responses
- [ ] Basic frontend: project list, issue list, create/edit forms

---

## M3 — Issue Table View

- [ ] Table view with columns: Issue Name, Project, Assignee, Due Date, Status
- [ ] Sorting by any column
- [ ] Filters: Status, Assignee, Project, Due Date
- [ ] My Issues view (scoped to current user)
- [ ] Project view (scoped to selected project)

---

## M4 — Kanban Board

- [ ] Kanban view with status columns
- [ ] Drag-and-drop issue cards
- [ ] Status update on drop (REST PATCH + optimistic UI)
- [ ] Available in both My Issues and Project views

---

## M5 — Calendar View + Home Dashboard

- [ ] Calendar view: issues placed by due date
- [ ] Home dashboard stats: Total Projects, Total Issues, Assigned Issues, Completed Issues, Overdue Issues
- [ ] Home dashboard: list of issues assigned to current user

---

## M6 — Webhooks + Async Jobs

- [ ] Webhook endpoint registration (create, list, delete)
- [ ] Event dispatch on issue create, update, delete, assign
- [ ] Event dispatch on project create, update, delete
- [ ] OtterMQ integration for async delivery
- [ ] Retry logic for failed webhook deliveries
- [ ] In-app notifications
- [ ] Email notifications

---

## M7 — Real-time Updates

- [ ] SSE endpoint (`GET /stream`)
- [ ] Real-time gateway subscribing to OtterMQ
- [ ] Board updates reflected live (issue moved, created, deleted)
- [ ] Notification badge updates in real time
- [ ] Multi-client fanout (multiple browser tabs)

---

## Final

- [ ] End-to-end testing of core flows
- [ ] API documentation
- [ ] Self-hosting guide (Docker / docker-compose)
- [ ] README with setup instructions

---

## Tech Debt & Cleanup

Open tasks identified during PR review. Not tied to a specific milestone — pick up whenever the surrounding code is being touched.

- [ ] **Structured logger** — replace `log.Printf` / `log.Fatalf` with a structured logger (`slog` or `zap`). Currently scattered across `main.go` and `common/errors.go`. Do in one pass so the call-site style is consistent.
- [ ] **Input validation** — add `go-playground/validator/v10` to the auth package. Define a `BindAndValidate[T]` helper in `internal/common` and apply it in all auth handlers. Enforce `required,email` on email, `required,min=8,max=72` on password (bcrypt truncates at 72 bytes).
- [ ] **Connection pool** — replace bare `pgx.Connect` in `main.go` with `pgxpool.New`. Single connection serializes all DB requests under any load.
- [ ] **Redis unavailability** — in `AuthMiddleware`, distinguish `redis.Nil` (session not found → 401) from other errors (store down → 503). Avoids silently turning infrastructure failures into auth failures.
