#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p dist/runtime
if ! python3 scripts/fetch_sing_box.py --os linux --arch amd64 --output-dir dist/runtime; then
  printf 'Skipping local sing-box bundle (no matching Linux asset found).\n'
fi
CGO_ENABLED=1 go build -tags production -o dist/ProxyNodeStudio-wails ./cmd/proxy-node-studio-wails
printf 'Built %s\n' "$(pwd)/dist/ProxyNodeStudio-wails"
if [ -f dist/runtime/sing-box ]; then
  printf 'Bundled runtime %s\n' "$(pwd)/dist/runtime/sing-box"
fi
printf 'For Windows, run scripts/build_proxy_node_studio_wails.bat on a Windows machine.\n'
