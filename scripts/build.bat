@echo off
setlocal enabledelayedexpansion

set "ROOT_DIR=%~dp0.."
cd /d "%ROOT_DIR%"

echo ==^> [1/3] Building frontend assets...
cd /d "%ROOT_DIR%\frontend"
call bun install
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] bun install failed.
    exit /b %ERRORLEVEL%
)
call bun run build
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] bun run build failed.
    exit /b %ERRORLEVEL%
)
echo ==^> [2/3] Syncing frontend distribution to Go embed directory...
cd "%ROOT_DIR%"
if exist "cmd\app\frontend_dist" rd /s /q "cmd\app\frontend_dist"
xcopy /E /I /Q /Y "frontend\dist" "cmd\app\frontend_dist"

echo ==^> [3/3] Compiling native desktop binary...
if not exist "dist\windows-portable" mkdir "dist\windows-portable"
set CGO_ENABLED=0
go build -ldflags="-H windowsgui -s -w" -o "%ROOT_DIR%\dist\windows-portable\TrafficAnalyzer.exe" ./cmd/app
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go build failed.
    exit /b %ERRORLEVEL%
)

echo ==^> Packaging portable zip archive...
powershell -NoProfile -Command "Compress-Archive -Path '%ROOT_DIR%\dist\windows-portable\*' -DestinationPath '%ROOT_DIR%\dist\TrafficAnalyzer-windows-amd64-portable.zip' -Force"

echo ==^> Build complete: %ROOT_DIR%\dist\windows-portable\TrafficAnalyzer.exe
