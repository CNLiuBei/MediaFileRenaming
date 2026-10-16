"""Persistent, bounded activity history for the app's Logs page."""
from __future__ import annotations

import json
import os
import threading
from datetime import datetime
from pathlib import Path

DATA_ROOT = Path(os.environ.get("TRIM_PKGVAR", Path.home() / ".media-file-renaming"))
LOG_PATH = DATA_ROOT / "activity.jsonl"
LOG_LOCK = threading.Lock()
MAX_LOGS = 500
MAX_LOG_DETAILS = 20


def record_activity(
    action: str,
    message: str,
    *,
    total: int = 0,
    ready: int = 0,
    moved: int = 0,
    skipped: int = 0,
    failed: int = 0,
    source_dirs: list[str] | None = None,
    destination_dir: str = "",
    details: list[dict] | None = None,
    details_total: int | None = None,
) -> None:
    safe_details = []
    for detail in (details or [])[:MAX_LOG_DETAILS]:
        if not isinstance(detail, dict):
            continue
        safe_details.append({
            "status": str(detail.get("status", ""))[:30],
            "source": str(detail.get("source", ""))[:500],
            "destination": str(detail.get("destination", ""))[:500],
            "message": str(detail.get("message", ""))[:300],
        })
    entry = {
        "time": datetime.now().astimezone().isoformat(timespec="seconds"),
        "action": action,
        "message": message[:500],
        "total": max(0, int(total)),
        "ready": max(0, int(ready)),
        "moved": moved,
        "skipped": skipped,
        "failed": failed,
        "source_dirs": [str(path)[:500] for path in (source_dirs or [])[:50]],
        "destination_dir": destination_dir[:500],
        "details": safe_details,
        "details_total": max(0, int(details_total if details_total is not None else len(safe_details))),
    }
    try:
        DATA_ROOT.mkdir(parents=True, exist_ok=True)
        with LOG_LOCK:
            existing = LOG_PATH.read_text(encoding="utf-8").splitlines() if LOG_PATH.exists() else []
            existing = existing[-(MAX_LOGS - 1):]
            with LOG_PATH.open("w", encoding="utf-8") as stream:
                stream.write("\n".join([*existing, json.dumps(entry, ensure_ascii=False)]) + "\n")
            os.chmod(LOG_PATH, 0o600)
    except OSError:
        pass


def recent_activity(limit: int = 100) -> list[dict]:
    try:
        lines = LOG_PATH.read_text(encoding="utf-8").splitlines()
    except OSError:
        return []
    entries = []
    for line in reversed(lines[-MAX_LOGS:]):
        try:
            entry = json.loads(line)
            if isinstance(entry, dict):
                entries.append(entry)
        except json.JSONDecodeError:
            continue
        if len(entries) >= max(1, min(limit, 200)):
            break
    return entries
