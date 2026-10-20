@echo off
cd /d "%~dp0"

echo ==========================================
echo  Starting WhatsApp Reminder (Dev Mode)
echo  Backend API:  http://localhost:8080
echo  Frontend Web: http://localhost:5173
echo ==========================================

REM Check if node_modules exists in root
if not exist "node_modules\" (
    echo Installing root dev dependencies...
    call npm install
)

REM Run both concurrently in one window
call npx concurrently -k -p "[{name}]" -n "BACKEND,FRONTEND" -c "blue.bold,green.bold" "cd backend && go run ./cmd/server" "cd frontend && npm run dev"
