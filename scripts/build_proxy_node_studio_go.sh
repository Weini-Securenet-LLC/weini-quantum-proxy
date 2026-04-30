#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p dist
CGO_ENABLED=0 go build -o dist/ProxyNodeStudio-go ./cmd/proxy-node-studio
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/ProxyNodeStudio-go.exe ./cmd/proxy-node-studio
printf 'Built:\n  %s\n  %s\n' "$(pwd)/dist/ProxyNodeStudio-go" "$(pwd)/dist/ProxyNodeStudio-go.exe"
