@echo off
echo ==========================================
echo  Building WhatsApp Reminder (Full Stack)
echo ==========================================

cd /d "%~dp0"

where node >nul 2>nul
if %errorlevel% equ 0 (
    node -e "const [M,m,p]=process.versions.node.split('.').map(Number); const ok=(M===22&&(m>23||(m===23&&p>=4)))||(M===24&&(m>22||(m===22&&p>=1)))||(M>=26&&(m>12||(m===12&&p>=1))); if(!ok){console.log('[PERINGATAN KEAMANAN] Node.js ' + process.version + ' rentan terhadap CVE Oktober 2026. Disarankan update ke v24.22.1+ / v22.23.4+ / v26.12.1+.');} else {console.log('[Keamanan: Node.js ' + process.version + ' Sesuai Rilis Keamanan Oktober 2026]');}"
)

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
