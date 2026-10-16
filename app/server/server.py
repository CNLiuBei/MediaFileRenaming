"""HTTP service and optional periodic organizer for fnOS."""
from __future__ import annotations

import json
import os
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

from api_helpers import list_authorized_directories, requested_overrides, tmdb_search_for_file
from activity_log import recent_activity, record_activity
from organizer import apply_plan, build_plan, list_authorized_roots, validate_folder
from tmdb import validate_tmdb_proxy_url

APP_ROOT = Path(os.environ.get("TRIM_APPDEST", Path(__file__).resolve().parents[1]))
WEB_ROOT = APP_ROOT / "ui" / "www"
DATA_ROOT = Path(os.environ.get("TRIM_PKGVAR", Path.home() / ".media-file-renaming"))
CONFIG_PATH = DATA_ROOT / "config.json"
DEFAULT_CONFIG = {"source_dir": "", "destination_dir": "", "watch_enabled": False, "interval_minutes": 15, "tmdb_api_key": "", "tmdb_proxy_url": ""}
CONFIG_LOCK = threading.Lock()


def read_config() -> dict:
    try:
        stored = json.loads(CONFIG_PATH.read_text(encoding="utf-8"))
        return {**DEFAULT_CONFIG, **stored}
    except (OSError, json.JSONDecodeError):
        return DEFAULT_CONFIG.copy()


def write_config(config: dict) -> None:
    DATA_ROOT.mkdir(parents=True, exist_ok=True)
    temp_path = CONFIG_PATH.with_suffix(".tmp")
    temp_path.write_text(json.dumps(config, ensure_ascii=False, indent=2), encoding="utf-8")
    os.chmod(temp_path, 0o600)
    temp_path.replace(CONFIG_PATH)


def authorized_config(config: dict) -> tuple[Path, Path]:
    roots = list_authorized_roots()
    source = validate_folder(str(config.get("source_dir", "")), roots)
    destination = validate_folder(str(config.get("destination_dir", "")), roots)
    if source == destination or source in destination.parents:
        raise ValueError("目标目录不能与源目录相同或位于源目录内部")
    return source, destination


def run_once(config: dict) -> dict:
    source, destination = authorized_config(config)
    plan = build_plan(source, destination)
    return apply_plan(plan)


class Handler(BaseHTTPRequestHandler):
    server_version = "MediaOrganizer/1.0"

    def log_message(self, fmt: str, *args: object) -> None:
        print("%s - %s" % (self.log_date_time_string(), fmt % args), flush=True)

    def send_json(self, payload: dict, status: int = 200) -> None:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def read_json(self) -> dict:
        if self.headers.get_content_type() != "application/json":
            raise ValueError("请求格式必须是 JSON")
        length = int(self.headers.get("Content-Length", "0"))
        if length < 1 or length > 64_000:
            raise ValueError("请求数据长度无效")
        value = json.loads(self.rfile.read(length))
        if not isinstance(value, dict):
            raise ValueError("请求内容格式无效")
        return value

    def do_GET(self) -> None:
        request = urlparse(self.path)
        route = request.path
        if route == "/api/status":
            roots = [str(path) for path in list_authorized_roots()]
            config = read_config()
            public_config = {key: value for key, value in config.items() if key != "tmdb_api_key"}
            public_config["tmdb_key_configured"] = bool(config.get("tmdb_api_key"))
            self.send_json({"config": public_config, "authorized_roots": roots, "service": "running"})
            return
        if route == "/api/logs":
            try:
                limit = int(parse_qs(request.query).get("limit", ["100"])[0])
            except ValueError:
                limit = 100
            self.send_json({"items": recent_activity(limit)})
            return
        if route == "/api/directories":
            requested_path = parse_qs(request.query).get("path", [""])[0].strip()
            try:
                self.send_json(list_authorized_directories(requested_path))
            except (OSError, RuntimeError, ValueError) as exc:
                self.send_json({"error": str(exc) or "目录无法访问"}, 400)
            return
        relative = "index.html" if route == "/" else route.lstrip("/")
        target = (WEB_ROOT / relative).resolve()
        if WEB_ROOT.resolve() not in target.parents or not target.is_file():
            self.send_error(404)
            return
        content_types = {".html": "text/html; charset=utf-8", ".js": "text/javascript; charset=utf-8", ".css": "text/css; charset=utf-8"}
        content_type = content_types.get(target.suffix, "application/octet-stream")
        body = target.read_bytes()
        self.send_response(200)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_POST(self) -> None:
        route = urlparse(self.path).path
        try:
            payload = self.read_json()
            config = read_config()
            if route == "/api/preview":
                source, destination = authorized_config({**config, **payload})
                overrides = requested_overrides(payload, source)
                items = build_plan(source, destination, overrides)
                ready = sum(item.status == "ready" for item in items)
                record_activity("扫描", f"扫描 {len(items)} 个视频，{ready} 个可整理")
                self.send_json({"items": [item.as_dict() for item in items], "count": len(items)})
                return
            if route == "/api/tmdb/search":
                source, _ = authorized_config(config)
                results = tmdb_search_for_file(payload, config, source)
                self.send_json({"results": results})
                return
            if route == "/api/run":
                source, destination = authorized_config(config)
                overrides = requested_overrides(payload, source)
                items = build_plan(source, destination, overrides)
                result = apply_plan(items)
                result["total"] = len(items)
                record_activity("手动整理", f"整理完成：移动 {result['moved']} 个，跳过 {result['skipped']} 个，失败 {result['failed']} 个",
                                moved=result["moved"], skipped=result["skipped"], failed=result["failed"])
                self.send_json(result)
                return
            self.send_error(404)
        except (ValueError, OSError) as exc:
            record_activity("错误", f"{route}: {exc}")
            self.send_json({"error": str(exc)}, 400)
        except Exception as exc:  # Return a useful error instead of silently failing the request.
            record_activity("错误", f"{route}: {exc}")
            self.send_json({"error": f"整理失败：{exc}"}, 500)

    def do_PUT(self) -> None:
        if urlparse(self.path).path != "/api/config":
            self.send_error(404)
            return
        try:
            payload = self.read_json()
            previous_config = read_config()
            config = {**DEFAULT_CONFIG, **previous_config, **payload}
            if not str(payload.get("tmdb_api_key", "")).strip():
                config["tmdb_api_key"] = previous_config.get("tmdb_api_key", "")
            if len(str(config["tmdb_api_key"])) > 256:
                raise ValueError("TMDB API Key 长度无效")
            config["tmdb_proxy_url"] = validate_tmdb_proxy_url(str(config.get("tmdb_proxy_url", "")))
            interval = int(config["interval_minutes"])
            if not 5 <= interval <= 1440:
                raise ValueError("自动整理间隔必须在 5 到 1440 分钟之间")
            if config["watch_enabled"]:
                authorized_config(config)
            config["interval_minutes"] = interval
            config["watch_enabled"] = bool(config["watch_enabled"])
            with CONFIG_LOCK:
                write_config(config)
            public_config = {key: value for key, value in config.items() if key != "tmdb_api_key"}
            public_config["tmdb_key_configured"] = bool(config.get("tmdb_api_key"))
            self.send_json({"config": public_config})
        except (ValueError, OSError, TypeError) as exc:
            self.send_json({"error": str(exc)}, 400)


def watch_loop() -> None:
    last_signature = None
    next_run = 0.0
    while True:
        config = read_config()
        signature = (
            config.get("watch_enabled"),
            config.get("source_dir"),
            config.get("destination_dir"),
            config.get("interval_minutes"),
        )
        if signature != last_signature:
            last_signature = signature
            next_run = time.monotonic() if config.get("watch_enabled") else 0.0
        if config.get("watch_enabled") and time.monotonic() >= next_run:
            try:
                result = run_once(config)
                print(f"Automatic organization: {result}", flush=True)
                record_activity("自动整理", f"自动整理完成：移动 {result['moved']} 个，跳过 {result['skipped']} 个，失败 {result['failed']} 个",
                                moved=result["moved"], skipped=result["skipped"], failed=result["failed"])
            except Exception as exc:
                print(f"Automatic organization failed: {exc}", flush=True)
                record_activity("错误", f"自动整理失败：{exc}")
            next_run = time.monotonic() + max(5, int(config.get("interval_minutes", 15))) * 60
        time.sleep(5)


def main() -> None:
    DATA_ROOT.mkdir(parents=True, exist_ok=True)
    threading.Thread(target=watch_loop, daemon=True, name="organizer-watch").start()
    port = int(os.environ.get("TRIM_SERVICE_PORT", "8787"))
    server = ThreadingHTTPServer((os.environ.get("MEDIA_ORGANIZER_HOST", "0.0.0.0"), port), Handler)
    print(f"Media organizer listening on port {port}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
