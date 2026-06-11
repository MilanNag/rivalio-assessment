# TaskFlow — Full-Stack Task Management Application

A task management application with a **Go** REST API, a **Next.js** frontend, and **PostgreSQL** persistence.

| Layer    | Tech                                                          |
| -------- | ------------------------------------------------------------- |
| Frontend | Next.js 16 (App Router) · React 19 · Tailwind CSS 4 · TypeScript |
| Backend  | Go 1.26 · chi router · pgx · JWT (golang-jwt) · bcrypt        |
| Database | PostgreSQL 17                                                  |
| Tests    | Go `testing` · Jest + React Testing Library                   |
| Infra    | Docker Compose · GitHub Actions                                |

## Features

- **Task CRUD** — create, list, view, update (PATCH), delete
- **Auth** — signup/login with JWT, bcrypt-hashed passwords, all task routes protected, per-user task isolation
- **Persisted sessions** — page refresh keeps you logged in (token in localStorage, re-validated via `/api/auth/me`)
- **Filtering, search, sort, pagination** — by status, title search, sort by created/due date/priority; all combinable
- **Role-based access** — `admin` role can view (and manage) all users' tasks via an "All users" toggle
- **Real-time updates** — task changes stream live over Server-Sent Events
- **Optimistic UI** — complete/delete update instantly and roll back on failure with a toast
- **Attachments** — upload images/documents per task (type & size validated), download, delete
- **Activity log** — per-task history of changes (who, what, when)
- **Dark mode** — toggle persisted in localStorage, no flash on load
- **Responsive** — works on mobile and desktop
- **Dockerized** — one-command setup; **CI** — GitHub Actions runs vet/lint/tests/build on push

## Quick start (Docker — one command)

```bash
docker compose up --build
```

- Frontend: http://localhost:3100
- API: http://localhost:8090 (health check at `/healthz`)

Host ports are overridable via `BACKEND_PORT`, `FRONTEND_PORT` and `POSTGRES_PORT` (see `.env.example`) if any default clashes with something already running on your machine.

Optionally copy `.env.example` to `.env` first to override defaults (compose ships with working dev defaults).

> The **first user to sign up automatically becomes admin** (bootstrap convenience). Subsequent signups are regular users.

## Local development (without Docker)

Prerequisites: Go 1.26+, Node 22+, PostgreSQL running locally.

```bash
# 1. Database
createdb taskflow

# 2. Backend (migrations run automatically on boot)
cd backend
export DATABASE_URL="postgres://localhost:5432/taskflow?sslmode=disable"
export JWT_SECRET="change-me-to-a-32+-char-secret-in-prod!"
go run ./cmd/server

# 3. Frontend (in another terminal)
cd frontend
npm install
npm run dev          # uses http://localhost:8090 as API by default
```

## Deploying to Render

The root [`Dockerfile`](Dockerfile) builds both services into a single image: the Go API runs on `127.0.0.1:8090` inside the container and the Next.js server (on Render's `PORT`) proxies `/api/*` and `/healthz` to it. Everything is served from one origin, so no CORS or `NEXT_PUBLIC_API_URL` configuration is needed.

A [`render.yaml`](render.yaml) Blueprint provisions everything (web service + PostgreSQL):

1. Push the repository to GitHub.
2. In the [Render dashboard](https://dashboard.render.com), click **New → Blueprint** and select the repo.
3. Render reads `render.yaml`, creates the `taskflow` web service and the `taskflow-db` PostgreSQL instance, wires `DATABASE_URL`, and generates a random `JWT_SECRET` automatically.
4. Click **Apply** — the first deploy builds the Docker image and the app goes live at `https://taskflow-<hash>.onrender.com`. Database migrations run automatically when the API boots.

Subsequent pushes to the default branch auto-deploy.

Free-tier notes:

- The web service spins down after ~15 minutes of inactivity; the first request afterwards takes ~30–60s (cold start).
- Free PostgreSQL instances expire after 30 days — upgrade the database plan to keep data long-term.
- Uploaded attachments live on the container's local disk and are lost on redeploy; uncomment the `disk` block in `render.yaml` (paid plan) for persistence.

## Environment variables

All variables are listed with descriptions in [`.env.example`](.env.example). Required for the backend: `DATABASE_URL`, `JWT_SECRET` (min 32 chars). Everything else has sensible defaults.

## Running tests

```bash
# Backend
cd backend && go test ./...
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Frontend
cd frontend && npm test
npm run test:coverage
```

## API overview

All endpoints are JSON under `/api`. Authenticated routes require `Authorization: Bearer <token>`.

| Method | Path                          | Description                                              |
| ------ | ----------------------------- | -------------------------------------------------------- |
| POST   | `/api/auth/signup`            | Create account → `{token, user}`                          |
| POST   | `/api/auth/login`             | Login → `{token, user}`                                   |
| GET    | `/api/auth/me`                | Current user (validates persisted session)                |
| POST   | `/api/tasks`                  | Create task (title required; description, status, priority, dueDate optional) |
| GET    | `/api/tasks`                  | List with `status`, `q` (title search), `sort` (`created_at`\|`due_date`\|`priority`), `order`, `page`, `limit`, `all` (admin) |
| GET    | `/api/tasks/{id}`             | Fetch a single task                                       |
| PATCH  | `/api/tasks/{id}`             | Partial update (only provided fields change)              |
| DELETE | `/api/tasks/{id}`             | Delete a task → 204                                       |
| GET    | `/api/tasks/{id}/activity`    | Per-task change history                                   |
| POST   | `/api/tasks/{id}/attachments` | Upload file (multipart field `file`)                      |
| GET    | `/api/tasks/{id}/attachments` | List attachments                                          |
| GET    | `/api/attachments/{id}/download` | Download a file                                        |
| DELETE | `/api/attachments/{id}`       | Delete a file → 204                                       |
| GET    | `/api/events`                 | SSE stream of task events (`?access_token=` auth)         |

**Error envelope** (consistent across all endpoints):

```json
{ "error": { "code": "validation_error", "message": "One or more fields are invalid.", "fields": { "title": "Title is required." } } }
```

Status codes: `201` create, `204` delete, `400` malformed JSON, `401` unauthenticated, `404` not found / not yours, `409` duplicate email, `413` upload too large, `422` validation failure, `500` internal.

## Project structure

```
backend/
  cmd/server/          entry point
  internal/
    auth/              JWT + bcrypt helpers
    config/            env config
    database/          pgx pool + embedded SQL migrations
    httpapi/           handlers, router, middleware, validation (+ tests)
    models/            domain types
    realtime/          SSE pub/sub hub
    store/             store interfaces + postgres implementation
frontend/
  app/                 routes (login, signup, tasks)
  components/          UI components
  lib/                 api client, auth/theme contexts, useTasks hook, validation
  __tests__/           Jest + RTL tests
.github/workflows/     CI pipeline
docker-compose.yml     one-command local setup
```

## Assumptions & trade-offs

- **JWT in localStorage** instead of httpOnly cookies: simpler cross-origin setup between the two dev servers and explicitly demonstrates persisted auth state. An httpOnly cookie + CSRF token would be the hardened production choice.
- **First signup becomes admin**: the spec requires an admin role but no admin-management UI; this bootstrap rule makes the feature testable without seeding scripts.
- **SSE over WebSockets**: the traffic is strictly server→client notifications, so SSE is simpler (plain HTTP, auto-reconnect built into `EventSource`). `EventSource` cannot send headers, so the stream authenticates via an `access_token` query parameter.
- **Live updates trigger a refetch** of the current page rather than patching the list in place — keeps pagination/filter/sort consistency trivially correct at the cost of an extra request.
- **Attachments on local disk** (Docker volume) rather than S3-style object storage; the storage path is random-named to prevent traversal/collisions. Type allow-list and a 10 MiB cap are enforced.
- **Other users' tasks return 404** (not 403) to avoid leaking the existence of task IDs.
- **Migrations run automatically at startup** from embedded SQL files — no separate migration tool needed for a project this size.
- **Search uses `ILIKE`** with escaped wildcards; fine at this scale, would move to trigram/full-text indexes for large datasets.
- **Unit tests use an in-memory fake store** implementing the same `store.Store` interface as the Postgres implementation, so handlers (validation, authz, status codes) are tested without a database. The Postgres layer itself is thin SQL exercised in compose/dev.
