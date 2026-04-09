# AGENTS.md - Frontend

**App**: cloud-drive-frontend  
**Stack**: Vue 3.5 + TypeScript + Vite + Tailwind CSS  
**Location**: `/Users/shi/study/frontend/projects/cloud-drive/apps/cloud-drive-frontend`

## Entry Points

| File | Purpose |
|------|---------|
| `src/main.ts` | Bootstraps Vue app, Pinia (with persist), router |
| `src/App.vue` | Root component with `<router-view>` |
| `src/router/index.ts` | Router with `beforeEach` auth guard |

## Auth & State Management

**Store**: `src/stores/user.ts`
- Pinia store with `pinia-plugin-persistedstate`
- Token persisted to localStorage automatically
- Store survives page refresh

**Auth Guard**: `src/router/index.ts`
- Calls `checkLogin()` API before protected routes
- Redirects to `/require-login` if not authenticated
- `/home/pickup-codes` is **exempt** (public access)

**API Client**: `src/services/request.ts`
- Axios instance with interceptors
- Auto-adds `Authorization: Bearer {token}` header
- Reads token from localStorage (key: `user`)

## Routing

**Routes**: `src/router/route.ts`

| Path | Component | Auth Required |
|------|-----------|---------------|
| `/login` | Login.vue | No |
| `/register` | Register.vue | No |
| `/require-login` | RequireLogin.vue | No |
| `/home` | HomePage.vue | Yes (parent) |
| `/home/dashboard` | Dashboard.vue | Yes |
| `/home/files` | FileManagement.vue | Yes |
| `/home/pickup-codes` | PickupCodes.vue | **No** |
| `/home/upload` | UploadFile.vue | Yes |
| `/home/share` | ShareFile.vue | Yes |

## Styling

**Tailwind Config**: Custom theme in `tailwind.config.js`
- Primary color: `#10b674` (green)
- Dark mode: class-based (`dark` class on root)
- Custom fonts: `display: Inter, sans-serif`

**Dark Mode Toggle**: Add/remove `dark` class on `<html>` element

## API Integration

**Proxy**: Vite dev server proxies `/api` → `http://localhost:9000`
- Config: `vite.config.ts`
- Strips `/api` prefix before forwarding

**Auth API**: `src/services/apis/auth.ts`
- `login(data)` - POST /api/auth/login
- `register(data)` - POST /api/auth/register
- `checkLogin()` - GET /api/auth/check

## Commands

```bash
pnpm dev        # Start dev server (port 5173)
pnpm build      # Production build (vue-tsc + vite)
pnpm preview    # Preview production build
```

## Critical Notes

1. **Token Storage**: Token stored in localStorage under key `user` (JSON: `{token: "..."}`)
2. **No Tests**: No test framework configured (consider adding Vitest)
3. **Strict TS**: Unused vars/params are errors (tsconfig.app.json)
4. **Debug Code**: `console.log(res)` in router guard (line 21) - remove for production
