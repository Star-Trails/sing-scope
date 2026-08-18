#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

export PATH="${PATH}:${HOME}/.local/bin:${HOME}/go/bin"

echo "==> Generating Go protobuf files from third_party/singbox-api/started_service.proto..."

mkdir -p "${ROOT_DIR}/internal/singboxapi/gen"

INCLUDE_DIR="${HOME}/.local/include"
EXTRA_FLAGS=()
if [ -d "${INCLUDE_DIR}" ]; then
  EXTRA_FLAGS+=("-I" "${INCLUDE_DIR}")
fi

protoc \
  -I "${ROOT_DIR}/third_party/singbox-api" \
  "${EXTRA_FLAGS[@]}" \
  --go_out="${ROOT_DIR}/internal/singboxapi/gen" \
  --go_opt=module=sing-scope/internal/singboxapi/gen \
  --go_opt=Mstarted_service.proto=sing-scope/internal/singboxapi/gen \
  --go-grpc_out="${ROOT_DIR}/internal/singboxapi/gen" \
  --go-grpc_opt=module=sing-scope/internal/singboxapi/gen \
  --go-grpc_opt=Mstarted_service.proto=sing-scope/internal/singboxapi/gen \
  "${ROOT_DIR}/third_party/singbox-api/started_service.proto"

echo "==> Protobuf generation complete."
