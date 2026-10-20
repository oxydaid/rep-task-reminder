#!/usr/bin/env bash
set -e

echo "=========================================="
echo " Building WhatsApp Reminder (Full Stack) "
echo "=========================================="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

# 1. Build Frontend
echo ""
echo "--> 1. Building Frontend (Vue 3 + Vite + Tailwind + PWA)..."
cd "$ROOT_DIR/frontend"
if [ ! -d "node_modules" ]; then
    npm install
fi
npm run build

# 2. Build Backend
echo ""
echo "--> 2. Building Backend (Go + Fiber)..."
cd "$ROOT_DIR/backend"
go mod tidy

mkdir -p "$ROOT_DIR/bin"
mkdir -p "$ROOT_DIR/data/backups"

OUTPUT_BIN="$ROOT_DIR/bin/whatsapp-reminder"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    OUTPUT_BIN="$OUTPUT_BIN.exe"
fi

go build -ldflags="-s -w" -o "$OUTPUT_BIN" ./cmd/server

echo ""
echo "=========================================="
echo " Build Completed Successfully! "
echo " Binary created: $OUTPUT_BIN"
echo " Web dist: $ROOT_DIR/frontend/dist"
echo "=========================================="
echo ""
echo "To run with native VPS:"
echo "  ./bin/whatsapp-reminder"
echo ""
echo "To run with PM2:"
echo "  pm2 start ecosystem.config.cjs"
echo ""
