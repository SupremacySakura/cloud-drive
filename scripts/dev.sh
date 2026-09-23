#!/usr/bin/env bash
# 一键启动本地开发环境：后端 (Go/Gin, :9000) + 前端 (Vite, :5173)
# 用法：在仓库根目录执行 `pnpm run dev`
set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/apps/cloud-drive-backend"
FRONTEND_DIR="$ROOT_DIR/apps/cloud-drive-frontend"

BACKEND_PID=""
FRONTEND_PID=""

log() {
  printf '[dev] %s\n' "$1"
}

shutdown() {
  trap - INT TERM EXIT
  [ -n "$FRONTEND_PID" ] && kill "$FRONTEND_PID" 2>/dev/null
  [ -n "$BACKEND_PID" ] && kill "$BACKEND_PID" 2>/dev/null
  wait 2>/dev/null
  exit 0
}
trap shutdown INT TERM

log "starting backend  (http://localhost:9000)"
(cd "$BACKEND_DIR" && go run cmd/server/main.go) &
BACKEND_PID=$!

log "starting frontend (http://localhost:5173)"
(cd "$FRONTEND_DIR" && pnpm dev) &
FRONTEND_PID=$!

# 任一进程退出时，结束另一个并退出，避免留下孤儿进程
while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 1
done

if kill -0 "$BACKEND_PID" 2>/dev/null; then
  log "frontend exited; shutting down backend"
else
  log "backend exited; shutting down frontend"
fi
shutdown
