#!/usr/bin/env python3
"""Build an fnOS package without including Vue sources or local dependencies."""
from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--fnpack", default=shutil.which("fnpack"), help="fnpack executable path")
    args = parser.parse_args()
    if not args.fnpack:
        parser.error("找不到 fnpack，请通过 --fnpack 指定可执行文件")

    with tempfile.TemporaryDirectory(prefix="media-file-renaming-") as temporary:
        stage = Path(temporary) / "source"
        stage.mkdir()
        for name in ("manifest", "ICON.PNG", "ICON_256.PNG", "app", "cmd", "config", "wizard"):
            source = ROOT / name
            if not source.exists():
                continue
            if source.is_dir():
                shutil.copytree(
                    source,
                    stage / name,
                    ignore=shutil.ignore_patterns("__pycache__", "*.pyc", ".DS_Store", "vendor"),
                )
            else:
                shutil.copy2(source, stage / name)
        requirements = stage / "app" / "server" / "requirements.txt"
        if requirements.exists():
            vendor = stage / "app" / "server" / "vendor"
            subprocess.run(
                [
                    sys.executable,
                    "-m",
                    "pip",
                    "install",
                    "--disable-pip-version-check",
                    "--no-compile",
                    "--no-warn-script-location",
                    "--target",
                    str(vendor),
                    "--requirement",
                    str(requirements),
                ],
                check=True,
            )
        subprocess.run([args.fnpack, "build"], cwd=stage, check=True)
        packages = sorted(stage.glob("*.fpk"), key=lambda item: item.stat().st_mtime, reverse=True)
        if not packages:
            raise RuntimeError("fnpack 构建完成但未生成 FPK 文件")
        output = ROOT / f"{packages[0].stem}.fpk"
        shutil.copy2(packages[0], output)
        print(f"已生成：{output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
