"""Private, user-provided TMDB catalog storage and lookup."""
from __future__ import annotations

import json
import os
import sqlite3
import threading
import unicodedata
from pathlib import Path

CATALOG_PATH = Path(os.environ.get("TRIM_PKGVAR", Path.home() / ".media-file-renaming")) / "catalog.sqlite"
INFO_PATH = CATALOG_PATH.with_suffix(".json")
MAX_UPLOAD_BYTES = 2_000_000_000
CATALOG_LOCK = threading.RLock()
REQUIRED_COLUMNS = {
    "titles": {"tmdb_id", "kind", "title", "original_title", "english_title", "year", "overview", "popularity", "original_language", "origin_countries", "poster_path"},
    "aliases": {"tmdb_id", "kind", "name_key"},
}


def catalog_info() -> dict:
    try:
        info = json.loads(INFO_PATH.read_text(encoding="utf-8"))
        if CATALOG_PATH.is_file():
            return {"available": True, **info}
    except (OSError, json.JSONDecodeError, TypeError):
        pass
    return {"available": False, "titles": 0, "aliases": 0, "episodes": 0, "size_bytes": 0}


def _inspect_catalog(path: Path) -> dict:
    connection = sqlite3.connect(f"{path.resolve().as_uri()}?mode=ro", uri=True, timeout=10)
    try:
        tables = {row[0] for row in connection.execute("SELECT name FROM sqlite_master WHERE type='table'")}
        if not REQUIRED_COLUMNS.keys() <= tables:
            raise ValueError("数据库不是受支持的 TMDB 离线片库（缺少 titles 或 aliases 表）")
        for table, required in REQUIRED_COLUMNS.items():
            columns = {row[1] for row in connection.execute(f"PRAGMA table_info({table})")}
            if not required <= columns:
                raise ValueError(f"数据库结构无效：{table} 表缺少必要字段")
        counts = {table: connection.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0] for table in ("titles", "aliases")}
        if counts["titles"] < 1:
            raise ValueError("数据库中没有影视条目")
        counts["episodes"] = connection.execute("SELECT COUNT(*) FROM episodes").fetchone()[0] if "episodes" in tables else 0
        return counts
    except sqlite3.DatabaseError as exc:
        raise ValueError("上传文件不是有效的 SQLite 数据库") from exc
    finally:
        connection.close()


def import_catalog(handler) -> dict:
    try:
        length = int(handler.headers.get("Content-Length", "0"))
    except ValueError as exc:
        raise ValueError("上传文件长度无效") from exc
    if length < 1 or length > MAX_UPLOAD_BYTES:
        raise ValueError("数据库文件必须大于 0 且不超过 2 GB")
    if handler.headers.get_content_type() not in {"application/octet-stream", "application/vnd.sqlite3"}:
        raise ValueError("请上传 SQLite 数据库文件")
    CATALOG_PATH.parent.mkdir(parents=True, exist_ok=True)
    temporary_path = CATALOG_PATH.with_suffix(".upload")
    try:
        remaining = length
        with temporary_path.open("wb") as output:
            while remaining:
                chunk = handler.rfile.read(min(1024 * 1024, remaining))
                if not chunk:
                    raise ValueError("数据库上传中断，请重试")
                output.write(chunk)
                remaining -= len(chunk)
            output.flush()
            os.fsync(output.fileno())
        counts = _inspect_catalog(temporary_path)
        with CATALOG_LOCK:
            temporary_path.replace(CATALOG_PATH)
            info = {"titles": counts["titles"], "aliases": counts["aliases"], "episodes": counts["episodes"], "size_bytes": length}
            info_temp = INFO_PATH.with_suffix(".tmp")
            info_temp.write_text(json.dumps(info), encoding="utf-8")
            os.chmod(info_temp, 0o600)
            info_temp.replace(INFO_PATH)
        return {"available": True, **info}
    finally:
        temporary_path.unlink(missing_ok=True)


def _name_key(value: str) -> str:
    normalized = unicodedata.normalize("NFKC", value).lower()
    return "".join(char for char in normalized if not char.isspace() and not unicodedata.category(char).startswith(("P", "Z")))


def search_catalog(kind: str, query: str, year: str = "") -> list[dict]:
    if kind not in {"电影", "电视剧"}:
        raise ValueError("TMDB 搜索类型无效")
    key = _name_key(query.strip())
    if not key or len(query) > 120:
        raise ValueError("请输入 1 到 120 个字符的片名")
    if not CATALOG_PATH.is_file():
        return []
    db_kind = "movie" if kind == "电影" else "tv"
    year_value = int(year) if year.isdigit() and len(year) == 4 else -1
    with CATALOG_LOCK:
        connection = sqlite3.connect(f"{CATALOG_PATH.resolve().as_uri()}?mode=ro", uri=True, timeout=10)
        connection.row_factory = sqlite3.Row
        try:
            rows = connection.execute(
                """SELECT t.tmdb_id, t.title, t.original_title, t.english_title, t.year, t.overview,
                          t.original_language, t.origin_countries, t.poster_path,
                          (SELECT name FROM aliases a WHERE a.tmdb_id = t.tmdb_id AND a.kind = t.kind AND a.name_key = ?) AS matched_alias
                   FROM titles t WHERE t.kind = ? AND t.tmdb_id IN (
                     SELECT tmdb_id FROM aliases WHERE kind = ? AND
                       (name_key = ? OR (? >= 3 AND name_key LIKE ? || '%') OR (? >= 4 AND name_key LIKE '%' || ? || '%'))
                   ) ORDER BY CASE WHEN t.year = ? THEN 0 ELSE 1 END, t.popularity DESC LIMIT 40""",
                (key, db_kind, db_kind, key, len(key), key, len(key), key, year_value),
            ).fetchall()
        finally:
            connection.close()
    results = []
    for row in rows:
        try:
            countries = json.loads(row["origin_countries"] or "[]")
        except (json.JSONDecodeError, TypeError):
            countries = []
        results.append({"id": row["tmdb_id"], "title": row["title"], "original_title": row["original_title"],
                        "english_title": row["english_title"], "matched_alias": row["matched_alias"] or "",
                        "year": str(row["year"] or ""), "overview": row["overview"],
                        "original_language": row["original_language"], "origin_country": countries,
                        "poster_path": row["poster_path"]})
    return results
