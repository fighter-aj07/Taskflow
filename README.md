# TaskFlow

A task management REST API built with Go. TaskFlow provides project and task management with JWT-based authentication, role-based access control (owner/creator enforcement), and paginated list endpoints.

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Router | Chi v5 |
| Database Driver | sqlx |
| Database | PostgreSQL 16 |
| Migrations | golang-migrate |
| Authentication | JWT (golang-jwt/jwt v5) |
| Password Hashing | bcrypt (golang.org/x/crypto) |
| Containerization | Docker / Docker Compose |

---

## Architecture

TaskFlow uses a clean layered architecture:

```
HTTP Request
    |
    v
Handler       (internal/handler/)   - decode/validate request, encode response
    |
    v
Service       (internal/service/)   - business logic, authorization checks
    |
    v
Repository    (internal/repository/) - SQL queries via sqlx
    |
    v
PostgreSQL
```

### Design Decisions

**Chi over Gin**: Chi is stdlib-compatible (`net/http` handler signatures) with no custom context types. Middleware composes naturally. The router is lightweight with zero dependencies beyond the standard library.

**sqlx over GORM**: Explicit SQL keeps queries readable and debuggable. No hidden N+1 queries or magic. sqlx adds just enough convenience (struct scanning, named queries) without obscuring what hits the database.

**slog for structured logging**: Available in the stdlib since Go 1.21. No third-party logging dependency. JSON output by default makes logs easy to parse in production environments.

**Layered architecture**: Each layer has a single responsibility. Handlers do not contain business logic. Services do not know about HTTP. Repositories do not know about authorization. This makes each layer independently testable and replaceable.

---

## Quick Start

```bash
git clone https://github.com/your-name/taskflow.git
cd taskflow
cp .env.example .env
docker compose up --build
```

The API will be available at http://localhost:8080

Migrations run automatically on container start. The seed user (see Test Credentials below) is created on first boot.

---

## Test Credentials

A seed user is created automatically when the database is initialized:

| Field | Value |
|---|---|
| Email | test@example.com |
| Password | password123 |

---

## API Endpoints

All authenticated endpoints require:
```
Authorization: Bearer <token>
Content-Type: application/json
```

---

### POST /auth/register

```json
// Request
{ "name": "Jane Doe", "email": "jane@example.com", "password": "secret123" }

// Response 201
{ "data": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com", "created_at": "..." } }
```

### POST /auth/login

```json
// Request
{ "email": "jane@example.com", "password": "secret123" }

// Response 200
{ "data": { "token": "<jwt>", "user": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com", "created_at": "..." } } }
```

---

### GET /projects

```json
// Response 200
{
  "data": [
    { "id": "uuid", "name": "Website Redesign", "description": "Q2 project", "owner_id": "uuid", "created_at": "..." }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 }
}
```

### POST /projects

```json
// Request
{ "name": "Website Redesign", "description": "Q2 project" }

// Response 201
{ "data": { "id": "uuid", "name": "Website Redesign", "description": "Q2 project", "owner_id": "uuid", "created_at": "..." } }
```

### GET /projects/:id

Returns the project together with all its tasks.

```json
// Response 200
{
  "data": {
    "id": "uuid", "name": "Website Redesign", "description": "Q2 project", "owner_id": "uuid", "created_at": "...",
    "tasks": [
      { "id": "uuid", "title": "Design homepage", "status": "todo", "priority": "high", "project_id": "uuid", "assignee_id": "uuid", "due_date": "2026-04-20", "created_at": "...", "updated_at": "..." }
    ]
  }
}
```

### PATCH /projects/:id

Owner only. All fields optional.

```json
// Request
{ "name": "New Name", "description": "Updated description" }

// Response 200
{ "data": { "id": "uuid", "name": "New Name", "description": "Updated description", "owner_id": "uuid", "created_at": "..." } }
```

### DELETE /projects/:id

Owner only. Returns `204 No Content`.

---

### GET /projects/:id/stats

```json
// Response 200
{
  "data": { "total": 3, "todo": 1, "in_progress": 1, "done": 1 }
}
```

### GET /projects/:id/tasks

Supports query params: `?status=todo|in_progress|done`, `?assignee_id=uuid`, `?page=1`, `?limit=20`

```json
// Response 200
{
  "data": [
    { "id": "uuid", "title": "Design homepage", "status": "todo", "priority": "high", "project_id": "uuid", "assignee_id": "uuid", "due_date": "2026-04-20", "created_at": "...", "updated_at": "..." }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 1, "total_pages": 1 }
}
```

### POST /projects/:id/tasks

```json
// Request
{ "title": "Design homepage", "description": "Create mockups", "priority": "high", "assignee_id": "uuid", "due_date": "2026-04-20" }

// Response 201
{ "data": { "id": "uuid", "title": "Design homepage", "status": "todo", "priority": "high", "project_id": "uuid", "assignee_id": "uuid", "due_date": "2026-04-20", "created_at": "...", "updated_at": "..." } }
```

### PATCH /tasks/:id

All fields optional. `status` must be one of `todo`, `in_progress`, `done`.

```json
// Request
{ "status": "in_progress", "priority": "low", "title": "Updated title" }

// Response 200
{ "data": { "id": "uuid", "title": "Updated title", "status": "in_progress", "priority": "low", "project_id": "uuid", "updated_at": "..." } }
```

### DELETE /tasks/:id

Project owner or task creator only. Returns `204 No Content`.

---

### Error responses

All errors follow a consistent shape:

```json
{ "error": "not found" }
{ "error": "unauthorized" }
{ "error": "invalid request body" }
```

---

## Running Tests

Integration tests require a live PostgreSQL instance. Set up a test database and run migrations before executing tests.

```bash
# Start the database container
docker compose up db -d

# Create the test database
PGPASSWORD=taskflow_secret psql -h localhost -U taskflow -d postgres -c "CREATE DATABASE taskflow_test;"

# Run migrations against the test database
# (adjust the DATABASE_URL to point to taskflow_test)

# Run integration tests
cd backend && POSTGRES_HOST=localhost POSTGRES_DB=taskflow_test go test ./tests/ -v -count=1
```

Tests use a bcrypt cost of 4 (instead of the production default of 12) to keep test execution fast.

---

## API Collection

A Bruno API collection is included at `backend/api-collection/`.

To use it:
1. Install the [Bruno](https://www.usebruno.com/) desktop app
2. Open the collection from `backend/api-collection/`
3. Select the `local` environment
4. Run `Register` or `Login` first — the post-response script automatically saves the token to the `token` environment variable
5. Run `Create Project` — the post-response script saves `project_id`
6. Run `Create Task` — the post-response script saves `task_id`

All subsequent requests use the saved variables automatically.

---

## What I'd Do With More Time

- **Rate limiting**: Per-IP and per-user rate limits on auth endpoints to prevent brute force attacks
- **Refresh tokens**: Short-lived access tokens with longer-lived refresh tokens for better security
- **Request ID correlation**: Inject a unique request ID at the middleware level and include it in all log lines and error responses for easier debugging
- **Cursor-based pagination**: Replace offset pagination with cursor-based pagination for consistent results on large datasets
- **More comprehensive test coverage**: Unit tests for each service and repository layer in addition to the integration tests; table-driven tests for all validation paths
- **CI/CD pipeline**: GitHub Actions workflow for lint, vet, test, and Docker build on every pull request
- **OpenAPI spec generation**: Annotate handlers to auto-generate an OpenAPI 3.0 spec, enabling auto-generated client SDKs and interactive documentation
