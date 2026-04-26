# AGENTS.md

**Project**: Cloud Drive - Full-stack file storage application  
**Stack**: Vue 3 + TypeScript (frontend), Go + Gin (backend), pnpm workspaces  
**Repo**: `/Users/shi/study/frontend/projects/cloud-drive`

## Quick Start

```bash
# Install dependencies
pnpm install

# Start backend (port 9000)
cd apps/cloud-drive-backend && go run cmd/server/main.go

# Start frontend (port 5173, proxies /api to backend)
cd apps/cloud-drive-frontend && pnpm dev

# Build for production
pnpm build                    # Root: typecheck only
pnpm --filter cloud-drive-frontend build   # Frontend
```

## Architecture

**Monorepo** with pnpm workspaces (`apps/*`, `packages/*`):

```
cloud-drive/
├── apps/
│   ├── cloud-drive-frontend/    # Vue 3 SPA with auth + routing
│   └── cloud-drive-backend/     # Go HTTP API (Gin + GORM)
├── packages/                     # EMPTY - no shared packages yet
└── data/                         # Runtime file storage (gitignored)
```

**Communication**: Vite proxy routes `/api/*` → `http://localhost:9000/*`

## Key Facts

### Auth Flow
- Login → JWT token stored in Pinia + localStorage (pinia-plugin-persistedstate)
- Router guard calls `GET /api/auth/check` on protected routes
- Unauthenticated users redirected to `/require-login`
- Token auto-added to axios requests via interceptor

### Frontend Stack
- Vue 3.5 + TypeScript strict mode
- Vue Router 5 with `beforeEach` auth guard
- Pinia 3 + persistedstate
- Tailwind CSS 4 (custom colors: primary #10b674)
- Axios (request/response interceptors)
- No test framework configured

### Backend Stack  
- Go 1.25 + Gin framework
- GORM (MySQL)
- JWT auth (golang-jwt/jwt/v5)
- Swaggo for OpenAPI docs (`/swagger/index.html`)
- Hardcoded paths: reads/writes to `/Users/shi/study/frontend/projects/cloud-drive/data`

## Commands Reference

| Command | Where | What |
|---------|-------|------|
| `pnpm dev` | frontend/ | Vite dev server + HMR |
| `pnpm build` | frontend/ | vue-tsc + vite build |
| `go run cmd/server/main.go` | backend/ | Start backend server |
| `pnpm lint` | root | ESLint on apps/**/*.{ts,tsx,js,jsx} |
| `pnpm typecheck` | root | tsc -b --pretty false |

## Critical Warnings ⚠️

1. **Hardcoded DB credentials** in `backend/internal/database/db.go` - Must extract to env vars
2. **JWT secret is "your_secret_key"** in `backend/internal/utils/jwt.go` - Change before production
3. **Hardcoded storage paths** in `backend/internal/router/router.go` - Use environment variables
4. **Debug console.log** in `frontend/src/router/index.ts:21` - Remove before production

## Conventions

- **ESLint**: Flat config (v9), TypeScript recommended rules only
- **TypeScript**: Strict mode enabled, unused vars/params = errors
- **Indent**: 2 spaces (EditorConfig enforced)
- **Line endings**: LF (Unix-style)
- **Dark mode**: Class-based toggle (`dark` class on root)
- **Styling**: Tailwind with custom primary color (#10b674)

## Where to Look

| Task | Location |
|------|----------|
| Add API endpoint | `backend/internal/router/*.go` → handler → service → repository |
| Add frontend page | `frontend/src/pages/*.vue` → add route in `router/route.ts` |
| Auth logic | `frontend/src/stores/user.ts`, `backend/internal/middleware/auth.go` |
| API client | `frontend/src/services/apis/*.ts`, `frontend/src/services/request.ts` |
| Database models | `backend/internal/model/*.go` |
| File upload logic | `backend/internal/service/file.go`, `frontend/src/pages/UploadFile.vue` |

## Anti-Patterns Here

- No `TODO`/`FIXME`/`HACK` comments found (clean codebase)
- No eslint-disable comments
- No `@ts-ignore` directives
- Generated swagger docs have "DO NOT EDIT" warning (expected)

## CI/CD

**NONE** - No GitHub Actions, Docker, or deployment configs exist.

## Testing

**NONE** - No test framework configured. Add Vitest to frontend, use `go test` for backend.

## Notes

- `tsconfig.json` references `./packages/shared` and `./apps/web` which don't exist (stale config)
- `packages/` directory is empty despite workspace declaration
- Backend runs on port 9000 by default (override with `PORT` env var)
- Frontend dev server proxies `/api` to backend automatically
- File chunks stored in `data/` directory by content hash (content-addressable storage)
