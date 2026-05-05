#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
package_root='dist/package-windows'
runtime_out="$package_root/runtime"
rm -rf "$package_root"
mkdir -p "$runtime_out"
install -m 0644 'dist/维尼量子节点.exe' "$package_root/维尼量子节点.exe"
install -m 0644 'dist/runtime/sing-box.exe' "$runtime_out/sing-box.exe"
python3 - <<'PY'
import os, zipfile
root = os.path.join(os.getcwd(), 'dist', 'package-windows')
out = os.path.join(os.getcwd(), 'dist', '维尼量子节点-轻量图标版.zip')
with zipfile.ZipFile(out, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as zf:
    for base, _, files in os.walk(root):
        for name in files:
            full = os.path.join(base, name)
            arc = os.path.relpath(full, root)
            zf.write(full, arc)
print(out)
PY
