#!/usr/bin/env python3
from __future__ import annotations

import importlib.util
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
APP = ROOT / "scripts" / "proxy_node_studio.py"
DIST = ROOT / "dist"
BUILD = ROOT / "build"


def ensure_pyinstaller() -> None:
    if importlib.util.find_spec("PyInstaller") is None:
        subprocess.check_call([sys.executable, "-m", "pip", "install", "pyinstaller"])


def main() -> int:
    ensure_pyinstaller()
    command = [
        sys.executable,
        "-m",
        "PyInstaller",
        "--noconfirm",
        "--clean",
        "--onefile",
        "--windowed",
        "--name",
        "ProxyNodeStudio",
        "--distpath",
        str(DIST),
        "--workpath",
        str(BUILD),
        str(APP),
    ]
    print("Running:", " ".join(command))
    subprocess.check_call(command, cwd=ROOT)
    print(f"Build complete: {DIST}")
    if sys.platform.startswith("win"):
        print(f"Windows exe: {DIST / 'ProxyNodeStudio.exe'}")
    else:
        print(f"Native executable: {DIST / 'ProxyNodeStudio'}")
        print("Note: Windows .exe must be built on Windows for best compatibility.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
