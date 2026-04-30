#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
python3 scripts/generate_proxy_node_studio_wails_icon.py >/dev/null
manifest_dir="cmd/proxy-node-studio-wails"
out_file="$manifest_dir/app_windows_amd64.syso"
windres_bin="${WINDRES:-x86_64-w64-mingw32-windres}"
"$windres_bin" \
  --input "$manifest_dir/app.rc" \
  --output-format coff \
  --output "$out_file"
printf 'Generated %s\n' "$(pwd)/$out_file"
