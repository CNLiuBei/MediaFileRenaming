#!/usr/bin/env python3
"""Add a built FPK release to the FnDepot V2 source index."""
from __future__ import annotations

import argparse
import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--tag", required=True)
    parser.add_argument("--artifact", required=True, type=Path)
    parser.add_argument("--index", default="fnpack.json", type=Path)
    args = parser.parse_args()

    index = json.loads(args.index.read_text(encoding="utf-8"))
    app = index["apps"]["MediaFileRenaming"]
    release = app.setdefault("releases", {}).setdefault(args.version, {})
    content = args.artifact.read_bytes()
    release["updated_at"] = datetime.now(timezone.utc).isoformat(timespec="seconds")
    release["packages"] = {
        "all": {
            "download_url": (
                "https://github.com/CNLiuBei/MediaFileRenaming/releases/download/"
                f"{args.tag}/{args.artifact.name}"
            ),
            "sha256": hashlib.sha256(content).hexdigest(),
            "size": len(content),
        }
    }
    args.index.write_text(
        json.dumps(index, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
