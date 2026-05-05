#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p dist
scripts/generate_proxy_node_studio_wails_manifest_syso.sh
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
go build -tags production -trimpath -ldflags='-s -w -H windowsgui' -o 'dist/维尼量子节点.exe' ./cmd/proxy-node-studio-wails
x86_64-w64-mingw32-strip 'dist/维尼量子节点.exe' || true
if [ -f 'dist/runtime/sing-box.exe' ]; then
  x86_64-w64-mingw32-strip 'dist/runtime/sing-box.exe' || true
fi
printf 'Built %s\n' "$(pwd)/dist/维尼量子节点.exe"
