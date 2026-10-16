#!/usr/bin/env python3
"""Build the x86 fnOS package for the TMDB sync station."""
from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PROJECT_ROOT = ROOT.parent


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--fnpack", default=shutil.which("fnpack"), help="fnpack executable")
    parser.add_argument("--out", type=Path, default=ROOT / "dist")
    args = parser.parse_args()
    if not args.fnpack:
        parser.error("找不到 fnpack，请使用 --fnpack 指定可执行文件")

    subprocess.run(["npm", "ci"], cwd=ROOT / "frontend", check=True)
    subprocess.run(["npm", "run", "build"], cwd=ROOT / "frontend", check=True)
    with tempfile.TemporaryDirectory(prefix="tmdb-sync-station-") as temporary:
        stage = Path(temporary) / "TmdbSyncStation"
        shutil.copytree(ROOT / "fnos-package", stage)
        shutil.copy2(PROJECT_ROOT / "ICON.PNG", stage / "ICON.PNG")
        shutil.copy2(PROJECT_ROOT / "ICON_256.PNG", stage / "ICON_256.PNG")
        app_bin = stage / "app" / "bin"
        app_bin.mkdir(parents=True)
        subprocess.run(
            ["go", "build", "-trimpath", "-ldflags=-s -w", "-o", str(app_bin / "tmdb-sync-station"), "./cmd/server"],
            cwd=ROOT,
            check=True,
            env={**os.environ, "GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0"},
        )
        shutil.copytree(ROOT / "frontend" / "dist", stage / "app" / "ui" / "www")
        icon_dir = stage / "app" / "ui" / "images"
        icon_dir.mkdir(parents=True)
        shutil.copy2(PROJECT_ROOT / "ICON.PNG", icon_dir / "icon_64.png")
        shutil.copy2(PROJECT_ROOT / "ICON_256.PNG", icon_dir / "icon_256.png")
        for script in (stage / "cmd").iterdir():
            script.chmod(script.stat().st_mode | 0o755)
        subprocess.run([args.fnpack, "build"], cwd=stage, check=True)
        packages = list(stage.glob("*.fpk"))
        if not packages:
            raise RuntimeError("fnpack 未生成 FPK")
        args.out.mkdir(parents=True, exist_ok=True)
        output = args.out / packages[0].name
        shutil.copy2(packages[0], output)
        print(f"已生成：{output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
