$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir

Write-Host "==> [1/3] Building frontend assets..." -ForegroundColor Cyan
Set-Location "$RootDir/frontend"
bun install
bun run build

Write-Host "==> [2/3] Syncing frontend distribution to Go embed directory..." -ForegroundColor Cyan
Set-Location $RootDir
if (Test-Path "cmd/app/frontend_dist") {
    Remove-Item -Recurse -Force "cmd/app/frontend_dist"
}
Copy-Item -Recurse -Force "frontend/dist" "cmd/app/frontend_dist"

Write-Host "==> [3/3] Compiling native desktop binary..." -ForegroundColor Cyan
if (-not (Test-Path "dist/windows-portable")) {
    New-Item -ItemType Directory -Force "dist/windows-portable" | Out-Null
}

$env:CGO_ENABLED = "0"
go build -ldflags="-H windowsgui -s -w" -o "$RootDir/dist/windows-portable/TrafficAnalyzer.exe" ./cmd/app


Write-Host "==> Packaging portable zip archive..." -ForegroundColor Cyan
Compress-Archive -Path "$RootDir/dist/windows-portable/*" -DestinationPath "$RootDir/dist/TrafficAnalyzer-windows-amd64-portable.zip" -Force
Write-Host "==> Build complete: $RootDir/dist/windows-portable/TrafficAnalyzer.exe" -ForegroundColor Green
