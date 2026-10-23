#!/usr/bin/env bash
set -e

echo "=========================================="
echo " Building WhatsApp Reminder (Full Stack) "
echo "=========================================="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

# 0. Check Node.js Security (October 2026 security release compliance)
if command -v node >/dev/null 2>&1; then
    NODE_SECURITY_CHECK=$(node -e '
        const [M, m, p] = process.versions.node.split(".").map(Number);
        const ok = (M === 22 && (m > 23 || (m === 23 && p >= 4))) ||
                   (M === 24 && (m > 22 || (m === 22 && p >= 1))) ||
                   (M >= 26 && (m > 12 || (m === 12 && p >= 1)));
        process.stdout.write(ok ? "OK" : "WARN");
    ' 2>/dev/null || echo "UNKNOWN")

    CURRENT_NODE_VER=$(node -v 2>/dev/null || echo "unknown")
    if [ "$NODE_SECURITY_CHECK" = "OK" ]; then
        echo "--> Node.js $CURRENT_NODE_VER [Keamanan: Terverifikasi Sesuai Rilis Keamanan Oktober 2026]"
    elif [ "$NODE_SECURITY_CHECK" = "WARN" ]; then
        echo "--> [PERINGATAN KEAMANAN] Node.js $CURRENT_NODE_VER rentan terhadap CVE Oktober 2026."
        echo "    Disarankan update ke Node.js v24.22.1+ (LTS) atau v22.23.4+ / v26.12.1+:"
        echo "    nvm install 24.22.1 && nvm use 24.22.1"
    fi
fi

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
