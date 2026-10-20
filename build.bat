@echo off
echo ==========================================
echo  Building WhatsApp Reminder (Full Stack)
echo ==========================================

cd /d "%~dp0"

echo.
echo --^> 1. Building Frontend (Vue 3 + Vite + Tailwind + PWA)...
cd frontend
if not exist "node_modules\" call npm install
call npm run build
if %errorlevel% neq 0 (
    echo Frontend build failed!
    exit /b %errorlevel%
)

echo.
echo --^> 2. Building Backend (Go + Fiber)...
cd ..\backend
call go mod tidy
if not exist "..\bin\" mkdir "..\bin"
if not exist "..\data\backups\" mkdir "..\data\backups"

call go build -o "..\bin\whatsapp-reminder.exe" ./cmd/server
if %errorlevel% neq 0 (
    echo Backend build failed!
    exit /b %errorlevel%
)

echo.
echo ==========================================
echo  Build Completed Successfully!
echo  Binary: bin\whatsapp-reminder.exe
echo  Frontend Dist: frontend\dist
echo ==========================================
echo.
pause
