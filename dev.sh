#!/usr/bin/env bash

# WhatsApp Reminder - All-in-one Dev Server Runner
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

cleanup() {
    echo ""
    echo "Shutting down dev servers..."
    kill $(jobs -p) 2>/dev/null || true
    exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo "=========================================="
echo " Starting WhatsApp Reminder (Dev Mode)    "
echo " Backend API:  http://localhost:8080      "
echo " Frontend Web: http://localhost:5173      "
echo "=========================================="

if command -v npx &> /dev/null; then
    npx concurrently -k -p "[{name}]" -n "BACKEND,FRONTEND" -c "blue.bold,green.bold" \
        "cd backend && go run ./cmd/server" \
        "cd frontend && npm run dev"
else
    # Fallback to background jobs
    cd "$ROOT_DIR/backend" && go run ./cmd/server &
    cd "$ROOT_DIR/frontend" && npm run dev &
    wait
fi
