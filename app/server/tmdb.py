"""Small TMDB search client used to confirm media titles before organizing."""
from __future__ import annotations

import json
import threading
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode, urlsplit
from urllib.request import Request, urlopen

TMDB_QPS = 0.8
REQUEST_INTERVAL_SECONDS = 1 / TMDB_QPS
RATE_LIMIT_LOCK = threading.Lock()
NEXT_REQUEST_AT = 0.0


def _wait_for_request_slot() -> None:
    global NEXT_REQUEST_AT
    with RATE_LIMIT_LOCK:
        delay = NEXT_REQUEST_AT - time.monotonic()
        if delay > 0:
            time.sleep(delay)
        NEXT_REQUEST_AT = time.monotonic() + REQUEST_INTERVAL_SECONDS


def validate_tmdb_proxy_url(value: str) -> str:
    api_base_url = value.strip()
    if not api_base_url:
        return ""
    try:
        parsed = urlsplit(api_base_url)
        parsed.port
    except ValueError as exc:
        raise ValueError("TMDB API 代理地址格式无效，请填写代理服务的 HTTPS 或 HTTP 地址") from exc
    path = parsed.path.rstrip("/")
    if (
        parsed.scheme.lower() not in {"http", "https"}
        or not parsed.hostname
        or parsed.username
        or parsed.password
        or parsed.query
        or parsed.fragment
        or (path and path != "/3" and not path.startswith("/3/"))
        or len(api_base_url) > 512
    ):
        raise ValueError("TMDB API 代理地址无效，请填写域名或 TMDB API 地址，例如 https://tmdb.liubei.org")
    return f"{parsed.scheme.lower()}://{parsed.netloc}"


def search_titles(api_key: str, kind: str, query: str, year: str = "", api_base_url: str = "") -> list[dict]:
    if kind not in {"电影", "电视剧"}:
        raise ValueError("TMDB 搜索类型无效")
    query = query.strip()
    if not query or len(query) > 120:
        raise ValueError("请输入 1 到 120 个字符的片名")
    endpoint = "movie" if kind == "电影" else "tv"
    params = {"api_key": api_key, "query": query, "language": "zh-CN", "include_adult": "false", "page": 1}
    if year.isdigit() and len(year) == 4:
        params["year" if kind == "电影" else "first_air_date_year"] = year
    base_url = api_base_url.rstrip("/") if api_base_url else "https://api.themoviedb.org"
    request = Request(
        f"{base_url}/3/search/{endpoint}?{urlencode(params)}",
        headers={"Accept": "application/json", "User-Agent": "MediaFileRenaming/1.0"},
    )
    try:
        _wait_for_request_slot()
        with urlopen(request, timeout=12) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except HTTPError as exc:
        if exc.code in {401, 403}:
            raise ValueError("TMDB API Key 无效或未获准访问") from exc
        raise ValueError(f"TMDB 请求失败（HTTP {exc.code}）") from exc
    except (URLError, TimeoutError, json.JSONDecodeError) as exc:
        if api_base_url:
            raise ValueError("无法通过 TMDB API 代理连接，请检查代理域名、Cloudflare Worker 状态和 NAS 网络") from exc
        raise ValueError("无法连接 TMDB，请检查 NAS 网络、TMDB API Key 或配置 TMDB API 代理") from exc
    results = payload.get("results", [])
    normalized = []
    for result in results[:10]:
        date = result.get("release_date" if kind == "电影" else "first_air_date", "")
        normalized.append({
            "id": result.get("id"),
            "title": result.get("title" if kind == "电影" else "name", ""),
            "original_title": result.get("original_title" if kind == "电影" else "original_name", ""),
            "year": date[:4],
            "overview": result.get("overview", ""),
            "original_language": result.get("original_language", ""),
            "origin_country": result.get("origin_country") or [],
            "poster_path": result.get("poster_path") or "",
        })
    return normalized
