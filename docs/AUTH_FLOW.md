# OtterBoard – Auth Flow

## Overview

Three auth strategies, each for a different actor:

| Strategy | Actor | Session |
|----------|-------|---------|
| Email/password | Human users | Server-side session (GoodiesDB) |
| OAuth | Human users | Server-side session (GoodiesDB) |
| API keys | Agents / integrations | Stateless — no session |

Sessions are stored in GoodiesDB and delivered to the browser via an httpOnly cookie. This enables instant revocation on logout and protects the session ID from JavaScript access.

---

## Email / Password

### Register

```
POST /api/v1/auth/register
{
  "name": "John",
  "email": "user@example.com",
  "password": "..."
}
```

1. Validate input
2. Check email is not already registered
3. Hash password with bcrypt
4. Create user record
5. Create session in GoodiesDB
6. Set httpOnly session cookie
7. Return user object

### Login

```
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "..."
}
```

1. Look up user by email
2. Compare password against bcrypt hash
3. Create session in GoodiesDB
4. Set httpOnly session cookie
5. Return user object

### Logout

```
POST /api/v1/auth/logout
```

1. Read session ID from cookie
2. Delete session from GoodiesDB
3. Clear cookie
4. Return `204 No Content`

---

## OAuth

### Initiate

```
GET /api/v1/auth/oauth/:provider
```

1. Generate a random `state` parameter (CSRF protection), store in GoodiesDB with short TTL
2. Redirect browser to provider authorization URL with `client_id`, `redirect_uri`, `scope`, and `state`

### Callback

```
GET /api/v1/auth/oauth/:provider/callback?code=...&state=...
```

1. Validate `state` against the value stored in GoodiesDB — reject if missing or mismatched
2. Exchange `code` for provider access token
3. Fetch user profile from provider (email, name, avatar)
4. Look up `user_identity` by `(provider, provider_user_id)`
   - If found: load the associated user
   - If not found: create user + user_identity (or link to existing user with same email)
5. Create session in GoodiesDB
6. Set httpOnly session cookie
7. Redirect to frontend

---

## API Keys

API keys are used by agents and integrations. There is no session — each request is authenticated independently.

### Creation

```
POST /api/v1/workspaces/:workspaceId/api-keys
{
  "name": "Claude Code"
}
```

1. Generate a cryptographically random token
2. Hash the token (SHA-256)
3. Store the hash + metadata in DB
4. Return the raw token **once** — it is never stored and cannot be retrieved again

### Request Authentication

```
Authorization: Bearer <raw-token>
```

1. Extract token from `Authorization` header
2. Hash the token (SHA-256)
3. Look up hash in `api_keys` table
4. Verify `revoked_at` is null
5. Load the associated workspace and user — attach to request context

### Revocation

```
DELETE /api/v1/workspaces/:workspaceId/api-keys/:keyId
```

1. Set `revoked_at` to current timestamp
2. All subsequent requests with that token are rejected

---

## Session Middleware

All protected endpoints run through session middleware:

1. Read session cookie
2. Look up session in GoodiesDB
3. If valid: attach user to request context and proceed
4. If invalid or missing: check for `Authorization: Bearer` header (API key path)
5. If neither: return `401 Unauthorized`

---

## CSRF Protection

Session cookies are set with `SameSite=Strict`, which prevents them from being sent on cross-site requests in all modern browsers. This is the primary CSRF defence for cookie-authenticated endpoints.

API key requests are not vulnerable to CSRF — the `Authorization` header cannot be set by cross-site forms.

---

## Rate Limiting

`POST /auth/login` and `POST /auth/register` must be rate limited to prevent brute-force attacks. Implementation details (middleware vs. GoodiesDB counter) are deferred to M1, but rate limiting is a hard requirement before these endpoints are exposed.
