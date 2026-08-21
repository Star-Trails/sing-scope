#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> [1/3] Building frontend assets with Bun..."
cd "${ROOT_DIR}/frontend"
bun install
bun run build

echo "==> [2/3] Syncing frontend distribution to Go embed directory..."
rm -rf "${ROOT_DIR}/cmd/app/frontend_dist"
cp -r "${ROOT_DIR}/frontend/dist" "${ROOT_DIR}/cmd/app/frontend_dist"

echo "==> [3/3] Compiling native desktop binary..."
cd "${ROOT_DIR}"
mkdir -p "${ROOT_DIR}/dist/windows-portable"

# Compile native Wails v3 GUI binary
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-H windowsgui -s -w" -o "${ROOT_DIR}/dist/windows-portable/TrafficAnalyzer.exe" ./cmd/app

echo "==> Packaging portable zip archive..."
if command -v zip >/dev/null 2>&1; then
    (cd "${ROOT_DIR}/dist/windows-portable" && zip -r -q "${ROOT_DIR}/dist/TrafficAnalyzer-windows-amd64-portable.zip" .)
fi

echo "==> Build complete: ${ROOT_DIR}/dist/windows-portable/TrafficAnalyzer.exe"
