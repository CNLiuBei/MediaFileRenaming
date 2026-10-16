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


def record_activity(action: str, message: str, *, moved: int = 0, skipped: int = 0, failed: int = 0) -> None:
    entry = {
        "time": datetime.now().astimezone().isoformat(timespec="seconds"),
        "action": action,
        "message": message[:500],
        "moved": moved,
        "skipped": skipped,
        "failed": failed,
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
