# AGENTS.md - Backend

**App**: cloud-drive-backend  
**Stack**: Go 1.25 + Gin + GORM + MYSQL
**Location**: `/Users/shi/study/frontend/projects/cloud-drive/apps/cloud-drive-backend`

## Entry Point

**File**: `cmd/server/main.go`
- Initializes database: `database.InitDB()`
- Sets up router: `router.SetUpRouter()`
- Runs on port from `PORT` env var (default: 9000)

## Architecture

Standard Go layout:
```
cmd/server/         # Entry point
internal/
  ├── configs/      # Configuration
  ├── database/     # DB connection init
  ├── deployments/  # Deployment configs (empty)
  ├── docs/         # Swagger docs (generated)
  ├── dto/          # Data Transfer Objects
  ├── handler/      # HTTP handlers (controllers)
  ├── middleware/   # Auth middleware, CORS
  ├── model/        # GORM models
  ├── repository/   # Data access layer
  ├── response/     # HTTP response helpers
  ├── router/       # Route registration
  ├── scripts/      # Utility scripts
  ├── service/      # Business logic
  ├── utils/        # JWT, helpers
  └── vo/           # View Objects
```

## Auth

**JWT**: `internal/utils/jwt.go`
- Secret: `your_secret_key` (⚠️ **CHANGE THIS**)
- Tokens parsed from `Authorization: Bearer <token>` header

**Middleware**: `internal/middleware/auth.go`
- Validates JWT, sets `user_id` in context
- Applied to protected routes

**Handler**: `internal/handler/auth.go`
- `POST /auth/login` - Authenticate, return JWT
- `POST /auth/register` - Create account
- `GET /auth/check` - Validate token (used by frontend)

## Database

**Connection**: `internal/database/db.go`
- MySQL via GORM
- DSN is **hardcoded** (⚠️ extract to env vars)
- `root:123456123456@tcp(127.0.0.1:3306)/cloud-drive`

**Models**: `internal/model/`
- User, File, etc.

## File Storage

**Paths**: Hardcoded in `internal/router/router.go`
- Chunk storage: `/Users/shi/study/frontend/projects/cloud-drive/data`
- File storage: `/Users/shi/study/frontend/projects/cloud-drive/data`

**⚠️ Must use environment variables for portability**

## API Documentation

**Swagger**: Auto-generated with swaggo
- URL: `http://localhost:9000/swagger/index.html`
- Generated from comments in handler files
- **Do not edit** `docs/docs.go` (auto-generated)

## Commands

```bash
# Run server
go run cmd/server/main.go

# With custom port
PORT=8080 go run cmd/server/main.go

# Generate swagger docs
cd internal && swag init
```

## Critical Security Issues

1. **Hardcoded DB credentials** in `internal/database/db.go`
2. **Hardcoded JWT secret** in `internal/utils/jwt.go`
3. **Hardcoded file paths** in `internal/router/router.go`

**All three must be moved to environment variables before production.**

## Where to Add Features

| Feature | Files |
|---------|-------|
| New endpoint | `internal/router/*.go` → handler → service → repository |
| New model | `internal/model/*.go` + auto-migrate |
| New middleware | `internal/middleware/*.go` → register in router |

## No Tests

No test files (`*_test.go`) exist. Add tests with standard Go testing:
```bash
go test ./...
```
