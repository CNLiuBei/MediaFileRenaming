#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

export DATABASE_URL="${DATABASE_URL:-postgres://localhost:5432/tmdb_sync_station?sslmode=disable}"

go build -o .dev-server ./cmd/server
./.dev-server &
API_PID=$!
npm --prefix frontend run dev &
WEB_PID=$!
cleanup() {
  kill "$API_PID" "$WEB_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM
wait
