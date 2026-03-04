# User CRUD API (Go + SQLite)

A lightweight REST API for user management built with Go, using real SQLite persistence through `zombiezen.com/go/sqlite`.

## Overview

- In-memory storage was replaced with SQLite persistence.
- API contract remains the same: endpoints, request payloads, and response shapes were preserved.
- SQL queries use parameter binding (`?1`, `?2`, etc.) to prevent SQL injection.

## Tech Stack

- Go
- `net/http` (native Go 1.22+ routing)
- `github.com/google/uuid`
- `zombiezen.com/go/sqlite`

## Database

- SQLite file: `users.db`
- SQLite URI: `file:users.db`
- The database file is created in the process working directory.
- In debug (`go run main.go` from project root), `users.db` is created next to `main.go`.
- For a compiled binary, `users.db` is created in the directory where the binary is executed.

### Startup Schema

```sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    biography TEXT NOT NULL
);
```

## Getting Started

1. Install dependencies:

```bash
go mod tidy
```

2. Run the API:

```bash
go run main.go
```

Server URL: `http://localhost:8080`

## API Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `POST` | `/api/users` | Create a user |
| `GET` | `/api/users` | List all users |
| `GET` | `/api/users/{id}` | Get a user by ID |
| `PUT` | `/api/users/{id}` | Update a user by ID |
| `DELETE` | `/api/users/{id}` | Delete a user by ID |

## Quick cURL Test Flow

Use the commands below to test the API end-to-end.

1. Create a user:

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "biography": "Software developer currently learning Go and building REST APIs."
  }'
```

2. List users:

```bash
curl http://localhost:8080/api/users
```

3. Get one user by ID (replace `USER_ID`):

```bash
curl http://localhost:8080/api/users/USER_ID
```

4. Update a user (replace `USER_ID`):

```bash
curl -X PUT http://localhost:8080/api/users/USER_ID \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Johnny",
    "last_name": "Doe",
    "biography": "Software developer focused on Go, SQLite, and clean API design."
  }'
```

5. Delete a user (replace `USER_ID`):

```bash
curl -X DELETE http://localhost:8080/api/users/USER_ID
```

## Example Payload

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "biography": "Software developer currently learning Go."
}
```

## Validation Rules

- `first_name`: 2 to 20 characters
- `last_name`: 2 to 20 characters
- `biography`: 20 to 450 characters
