"""Validation and lookup helpers for the HTTP API routes."""
from __future__ import annotations

import os
import re
from pathlib import Path

from organizer import is_within, list_authorized_roots
from tmdb import search_titles

POSTER_PATH_PATTERN = re.compile(r"/[A-Za-z0-9_-]+\.(?:jpe?g|png|webp)", re.IGNORECASE)


def requested_overrides(payload: dict, source: Path) -> dict[str, dict]:
    raw = payload.get("overrides", {})
    if not isinstance(raw, dict) or len(raw) > 5000:
        raise ValueError("识别修正数据格式无效")
    allowed = {"kind", "category", "title", "year", "season", "episode", "tmdb_id", "poster_path"}
    result: dict[str, dict] = {}
    for raw_path, values in raw.items():
        candidate = Path(raw_path).resolve(strict=True)
        if not is_within(candidate, [source]) or not candidate.is_file():
            raise ValueError("修正的文件不在待整理目录中")
        if not isinstance(values, dict) or set(values) - allowed:
            raise ValueError("识别修正字段无效")
        tmdb_id = values.get("tmdb_id", 0)
        if tmdb_id and (not str(tmdb_id).isdigit() or not 1 <= int(tmdb_id) <= 999_999_999):
            raise ValueError("TMDB 条目编号无效")
        poster_path = values.get("poster_path", "")
        if poster_path and (not isinstance(poster_path, str) or not POSTER_PATH_PATTERN.fullmatch(poster_path)):
            raise ValueError("TMDB 海报路径无效")
        result[str(candidate)] = values
    return result


def list_authorized_directories(requested_path: str) -> dict:
    roots = list_authorized_roots()
    if not requested_path:
        return {"current": "", "parent": None, "roots": [str(root) for root in roots], "directories": []}
    candidate = Path(requested_path).expanduser().resolve(strict=True)
    if not candidate.is_dir() or not is_within(candidate, roots):
        raise ValueError("只能浏览 fnOS 已授权的目录")
    parent = candidate.parent if candidate.parent != candidate and is_within(candidate.parent, roots) else None
    directories = []
    with os.scandir(candidate) as entries:
        for entry in entries:
            if entry.is_symlink() or not entry.is_dir(follow_symlinks=False):
                continue
            child = Path(entry.path).resolve(strict=True)
            if is_within(child, roots):
                directories.append({"name": entry.name, "path": str(child)})
    directories.sort(key=lambda item: item["name"].casefold())
    return {"current": str(candidate), "parent": str(parent) if parent else None, "roots": [], "directories": directories}


def tmdb_search_for_file(payload: dict, config: dict, source: Path) -> list[dict]:
    api_key = str(config.get("tmdb_api_key", ""))
    if not api_key:
        raise ValueError("请先在目录设置中填写 TMDB API Key")
    media_file = Path(str(payload.get("source", ""))).resolve(strict=True)
    if not media_file.is_file() or not is_within(media_file, [source]):
        raise ValueError("只能校准待整理目录中的文件")
    return search_titles(
        api_key,
        str(payload.get("kind", "")),
        str(payload.get("query", "")),
        str(payload.get("year", "")),
        str(config.get("tmdb_proxy_url", "")),
    )
