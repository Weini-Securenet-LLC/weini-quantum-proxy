#!/usr/bin/env python3
import argparse
import json
import os
import shutil
import sys
import tarfile
import tempfile
import urllib.request
import zipfile
from pathlib import Path

API_URL = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"


def fetch_json(url: str):
    req = urllib.request.Request(url, headers={"User-Agent": "ProxyNodeStudio/1.0"})
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.load(resp)


def download(url: str, target: Path):
    req = urllib.request.Request(url, headers={"User-Agent": "ProxyNodeStudio/1.0"})
    with urllib.request.urlopen(req, timeout=120) as resp, target.open("wb") as out:
        shutil.copyfileobj(resp, out)


def main() -> int:
    parser = argparse.ArgumentParser(description="Download latest sing-box release asset")
    parser.add_argument("--os", required=True, dest="target_os")
    parser.add_argument("--arch", required=True)
    parser.add_argument("--output-dir", required=True)
    args = parser.parse_args()

    release = fetch_json(API_URL)
    archive_suffixes = []
    if args.target_os == "windows":
        archive_suffixes = [f"{args.target_os}-{args.arch}.zip"]
    else:
        archive_suffixes = [
            f"{args.target_os}-{args.arch}.tar.gz",
            f"{args.target_os}-{args.arch}.zip",
        ]
    asset = None
    for candidate in release.get("assets", []):
        name = candidate.get("name", "")
        if any(name.endswith(suffix) for suffix in archive_suffixes):
            asset = candidate
            break
    if asset is None:
        print(f"No sing-box asset found for {args.target_os}-{args.arch}", file=sys.stderr)
        return 1

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory() as tmpdir:
        archive = Path(tmpdir) / asset["name"]
        download(asset["browser_download_url"], archive)
        if archive.name.endswith(".zip"):
            with zipfile.ZipFile(archive) as zf:
                zf.extractall(tmpdir)
        elif archive.name.endswith(".tar.gz"):
            with tarfile.open(archive, "r:gz") as tf:
                tf.extractall(tmpdir)
        else:
            print(f"Unsupported archive format: {archive.name}", file=sys.stderr)
            return 1
        extracted_root = None
        for child in Path(tmpdir).iterdir():
            if child.is_dir() and child.name.startswith("sing-box"):
                extracted_root = child
                break
        if extracted_root is None:
            print("Extracted sing-box directory not found", file=sys.stderr)
            return 1
        binary_name = "sing-box.exe" if args.target_os == "windows" else "sing-box"
        source_binary = extracted_root / binary_name
        if not source_binary.exists():
            print(f"Binary not found in archive: {source_binary}", file=sys.stderr)
            return 1
        target_binary = out_dir / binary_name
        shutil.copy2(source_binary, target_binary)
        os.chmod(target_binary, 0o755)
        print(target_binary)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
