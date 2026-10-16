"""Read-only client for the local TMDB sync station API."""
from __future__ import annotations

import json
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode, urlsplit
from urllib.request import Request, urlopen

REQUEST_TIMEOUT_SECONDS = 3


def validate_sync_station_url(value: str) -> str:
    base_url = value.strip().rstrip("/")
    if not base_url:
        return ""
    try:
        parsed = urlsplit(base_url)
        parsed.port
    except ValueError as exc:
        raise ValueError("TMDB 同步站地址无效，请填写 http://主机:端口") from exc
    if (
        parsed.scheme.lower() not in {"http", "https"}
        or not parsed.hostname
        or parsed.username
        or parsed.password
        or parsed.path not in {"", "/"}
        or parsed.query
        or parsed.fragment
        or len(base_url) > 512
    ):
        raise ValueError("TMDB 同步站地址无效，请填写 http://主机:端口，例如 http://127.0.0.1:8788")
    return f"{parsed.scheme.lower()}://{parsed.netloc}"


def _request_json(base_url: str, path: str, params: dict | None = None) -> dict:
    suffix = f"?{urlencode(params)}" if params else ""
    request = Request(
        f"{base_url}{path}{suffix}",
        headers={"Accept": "application/json", "User-Agent": "MediaFileRenaming/1.0"},
    )
    try:
        with urlopen(request, timeout=REQUEST_TIMEOUT_SECONDS) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except HTTPError as exc:
        if exc.code == 404:
            raise ValueError("TMDB 同步站接口版本较旧，请先重启或更新同步站") from exc
        raise ValueError(f"TMDB 同步站请求失败（HTTP {exc.code}）") from exc
    except (URLError, TimeoutError, json.JSONDecodeError) as exc:
        raise ValueError("无法连接 TMDB 同步站，请检查地址和服务状态") from exc
    if not isinstance(payload, dict):
        raise ValueError("TMDB 同步站返回的数据格式无效")
    return payload


def sync_station_status(value: str) -> dict:
    base_url = validate_sync_station_url(value)
    if not base_url:
        return {"configured": False, "available": False, "message": "尚未配置同步站"}
    try:
        payload = _request_json(base_url, "/api/health")
    except ValueError as exc:
        return {"configured": True, "available": False, "message": str(exc)}
    available = payload.get("state") == "connected"
    return {
        "configured": True,
        "available": available,
        "message": "同步站已连接" if available else "同步站数据库未连接",
    }


def search_sync_station(value: str, kind: str, query: str, year: str = "") -> list[dict]:
    base_url = validate_sync_station_url(value)
    if not base_url:
        return []
    payload = _request_json(
        base_url,
        "/api/catalog/search",
        {"kind": kind, "q": query.strip(), "year": year},
    )
    results = payload.get("results")
    if not isinstance(results, list):
        raise ValueError("TMDB 同步站搜索结果格式无效")
    return [
        {
            "id": item.get("id"),
            "title": item.get("title", ""),
            "original_title": item.get("original_title", ""),
            "english_title": item.get("english_title", ""),
            "matched_alias": item.get("matched_alias", ""),
            "year": str(item.get("year", "")),
            "overview": item.get("overview", ""),
            "original_language": item.get("original_language", ""),
            "origin_country": item.get("origin_country") or [],
            "poster_path": item.get("poster_path", ""),
        }
        for item in results
        if isinstance(item, dict)
    ]
