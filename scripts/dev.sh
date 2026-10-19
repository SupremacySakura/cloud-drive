#!/usr/bin/env bash
# Native development: host backend + Vite, using an existing local MySQL.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/apps/zhixu-backend"
FRONTEND_DIR="$ROOT_DIR/apps/zhixu-frontend"
ENV_FILE="$ROOT_DIR/.env.development"

BACKEND_PID=""
FRONTEND_PID=""
BUILD_DIR=""

log() {
  printf '[dev] %s\n' "$1"
}

stop_tree() {
  local pid="$1" child
  for child in $(pgrep -P "$pid" || true); do
    stop_tree "$child"
  done
  kill "$pid" 2>/dev/null || true
}

shutdown() {
  local status=$?
  trap - INT TERM EXIT
  [ -z "$FRONTEND_PID" ] || stop_tree "$FRONTEND_PID"
  [ -z "$BACKEND_PID" ] || stop_tree "$BACKEND_PID"
  wait 2>/dev/null || true
  [ -z "$BUILD_DIR" ] || rm -rf "$BUILD_DIR"
  exit "$status"
}
trap shutdown EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if [ ! -f "$ENV_FILE" ]; then
  log "Missing .env.development; run: cp .env.development.example .env.development"
  exit 1
fi
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  log "Missing frontend dependencies; run: pnpm install"
  exit 1
fi

# This local file uses shell-compatible KEY=value assignments.
set -a
source "$ENV_FILE"
set +a
export ZHIXU_ENV_FILE="$ENV_FILE"
export PORT="${PORT:-9000}"
export DEV_FRONTEND_PORT="${DEV_FRONTEND_PORT:-5173}"
export DEV_BACKEND_URL="http://127.0.0.1:$PORT"

for port in "$PORT" "$DEV_FRONTEND_PORT"; do
  if lsof -iTCP:"$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
    log "Port $port is occupied; update .env.development or stop its owner."
    exit 1
  fi
done
if [ "$PORT" = "$DEV_FRONTEND_PORT" ]; then
  log "PORT and DEV_FRONTEND_PORT must differ."
  exit 1
fi

if [ "${DB_HOST:-127.0.0.1}" != "127.0.0.1" ] && [ "${DB_HOST:-}" != "localhost" ] && [ "${DB_HOST:-}" != "::1" ]; then
  log "Run dev requires a local MySQL; set DB_HOST to 127.0.0.1 or localhost."
  exit 1
fi
log "using local MySQL (:${DB_PORT:-3306}); start it and create DB_NAME before running dev."

BUILD_DIR="$(mktemp -d "${TMPDIR:-/tmp}/zhixu-dev.XXXXXX")"
log "building backend"
(cd "$BACKEND_DIR" && go build -o "$BUILD_DIR/server" ./cmd/server)
log "starting backend ($DEV_BACKEND_URL)"
(cd "$BACKEND_DIR" && exec "$BUILD_DIR/server") &
BACKEND_PID=$!

ready=0
for ((attempt = 0; attempt < 60; attempt++)); do
  if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    wait "$BACKEND_PID"
    exit 1
  fi
  if curl --noproxy '*' --fail --silent --max-time 1 "$DEV_BACKEND_URL/healthz" >/dev/null; then
    ready=1
    break
  fi
  sleep 1
done
if [ "$ready" != "1" ]; then
  log "Backend did not become healthy."
  exit 1
fi

log "starting frontend (http://localhost:$DEV_FRONTEND_PORT)"
(cd "$FRONTEND_DIR" && exec pnpm exec vite) &
FRONTEND_PID=$!

# 任一进程退出时，结束另一个并退出，避免留下孤儿进程
while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
  sleep 1
done

if kill -0 "$BACKEND_PID" 2>/dev/null; then
  log "frontend exited; shutting down backend"
  wait "$FRONTEND_PID"
else
  log "backend exited; shutting down frontend"
  wait "$BACKEND_PID"
fi
