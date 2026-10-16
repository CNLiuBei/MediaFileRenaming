"""Filename based movie and TV episode organization."""
from __future__ import annotations

import os
import re
import shutil
from collections import defaultdict
from dataclasses import asdict, dataclass
from pathlib import Path

from media_parser import parse_media_filename

VIDEO_EXTENSIONS = {".mkv", ".mp4", ".avi", ".mov", ".wmv", ".m4v", ".ts", ".m2ts", ".webm", ".mpg", ".mpeg"}
MOVIE_CATEGORIES = {"待确认", "华语电影", "英语电影", "日语电影", "韩语电影", "其他语言电影"}
TV_CATEGORIES = {"待确认", "国产剧", "美剧", "日剧", "韩剧", "港剧", "台剧", "英剧", "其他地区剧"}
DEFAULT_CATEGORIES = {"电影": "待确认", "电视剧": "待确认"}
MAX_SOURCE_DIRECTORIES = 50
MAX_OPERATION_DETAILS = 20


@dataclass
class PlanItem:
    source: str
    destination: str
    kind: str
    status: str
    reason: str = ""
    title: str = ""
    year: str = ""
    season: int = 0
    episode: int = 0
    episode_end: int = 0
    category: str = ""
    tmdb_id: int = 0
    poster_path: str = ""

    def as_dict(self) -> dict:
        return asdict(self)


def safe_title(value: str) -> str:
    title = re.sub(r'[\\/:*?"<>|]', " ", value)
    title = re.sub(r"\s+", " ", title).strip(" .")
    if not title or len(title) > 160 or title in {".", ".."}:
        raise ValueError("片名不能为空，且不能超过 160 个字符")
    return title


def classify(
    path: Path,
    destination_root: Path,
    override: dict | None = None,
    *,
    series_directory: bool = False,
) -> tuple[str, Path, dict] | None:
    override = override or {}
    requested_kind = override.get("kind")
    if requested_kind not in (None, "电影", "电视剧"):
        raise ValueError("媒体类型只能是电影或电视剧")
    parsed = parse_media_filename(path, series_directory=series_directory)
    kind = requested_kind or parsed.kind
    categories = {"电影": MOVIE_CATEGORIES, "电视剧": TV_CATEGORIES}
    category = str(override.get("category") or DEFAULT_CATEGORIES[kind])
    if category not in categories[kind]:
        raise ValueError("二级分类与一级媒体类型不匹配")
    media_year = str(override.get("year", parsed.year))
    if media_year and not re.fullmatch(r"(?:19|20)\d{2}", media_year):
        raise ValueError("年份格式无效")
    title = safe_title(str(override.get("title") or parsed.title))
    if title == "未知名称":
        return None
    if kind == "电视剧":
        default_season = parsed.season if parsed.kind == "电视剧" else 1
        season = int(override.get("season", default_season))
        episode = int(override.get("episode", parsed.episode or 1))
        episode_end = parsed.episode_end
        if not 0 <= season <= 99 or not 0 <= episode <= 999:
            raise ValueError("季号或集号超出有效范围")
        if episode_end and (episode_end < episode or episode_end > 999):
            raise ValueError("多集范围无效")
        season_label = f"Season {season:02d}"
        episode_suffix = f"-E{episode_end:02d}" if episode_end else ""
        name = f"{title} - S{season:02d}E{episode:02d}{episode_suffix}{path.suffix.lower()}"
        series_folder = f"{title} ({media_year})" if media_year else title
        return "电视剧", destination_root / "电视剧" / category / series_folder / season_label / name, {
            "title": title, "year": media_year, "season": season, "episode": episode,
            "episode_end": episode_end, "category": category,
            "tmdb_id": int(override.get("tmdb_id", 0) or 0),
            "poster_path": str(override.get("poster_path", "")),
        }
    folder = f"{title} ({media_year})" if media_year else title
    return "电影", destination_root / "电影" / category / folder / f"{folder}{path.suffix.lower()}", {
        "title": title, "year": media_year, "season": 0, "episode": 0,
        "episode_end": 0, "category": category,
        "tmdb_id": int(override.get("tmdb_id", 0) or 0),
        "poster_path": str(override.get("poster_path", "")),
    }


def is_within(path: Path, roots: list[Path]) -> bool:
    resolved = path.resolve()
    return any(resolved == root or root in resolved.parents for root in roots)


def list_authorized_roots() -> list[Path]:
    raw_values = [os.environ.get("TRIM_DATA_ACCESSIBLE_PATHS", ""), os.environ.get("TRIM_DATA_SHARE_PATHS", "")]
    roots: list[Path] = []
    for raw in raw_values:
        for value in raw.split(":"):
            if value.strip():
                path = Path(value.strip()).resolve()
                if path.exists() and path.is_dir() and path not in roots:
                    roots.append(path)
    return roots


def validate_folder(value: str, roots: list[Path]) -> Path:
    if not value.strip():
        raise ValueError("请填写源目录和目标目录")
    candidate = Path(value).expanduser()
    if not candidate.is_absolute():
        raise ValueError("目录必须使用绝对路径")
    if not candidate.exists() or not candidate.is_dir():
        raise ValueError(f"目录不存在或不可访问：{candidate}")
    resolved = candidate.resolve()
    if not is_within(resolved, roots):
        raise ValueError(f"目录未获 fnOS 授权：{resolved}")
    return resolved


def normalize_source_dirs(values: object, legacy_value: str = "") -> list[str]:
    if values is None:
        candidates: list[object] = []
    elif isinstance(values, list):
        candidates = values
    elif isinstance(values, str):
        candidates = [values]
    else:
        raise ValueError("待整理目录格式无效")
    paths: list[str] = []
    for value in candidates:
        if not isinstance(value, str):
            raise ValueError("待整理目录必须是文件夹路径")
        path = value.strip()
        if path and path not in paths:
            paths.append(path)
    if not paths and legacy_value.strip():
        paths.append(legacy_value.strip())
    if len(paths) > MAX_SOURCE_DIRECTORIES:
        raise ValueError(f"最多可添加 {MAX_SOURCE_DIRECTORIES} 个待整理目录")
    return paths


def validate_source_folders(values: list[str], roots: list[Path]) -> list[Path]:
    if not values:
        raise ValueError("请至少选择一个待整理文件夹")
    sources = [validate_folder(value, roots) for value in values]
    for index, source in enumerate(sources):
        for other in sources[:index]:
            if source == other or source in other.parents or other in source.parents:
                raise ValueError("多个待整理文件夹不能重复或互相包含")
    return sources


def build_plan(
    source: Path | list[Path],
    destination: Path,
    overrides: dict[str, dict] | None = None,
) -> list[PlanItem]:
    source_roots = [source] if isinstance(source, Path) else source
    items: list[PlanItem] = []
    for source_root in source_roots:
        items.extend(_build_plan_for_source(source_root, destination, overrides))
    return items


def _build_plan_for_source(source: Path, destination: Path, overrides: dict[str, dict] | None = None) -> list[PlanItem]:
    items: list[PlanItem] = []
    media_paths: list[Path] = []
    for current, dirnames, filenames in os.walk(source, followlinks=False):
        current_path = Path(current)
        dirnames[:] = [name for name in dirnames if not (current_path / name).is_symlink()]
        for filename in sorted(filenames):
            path = current_path / filename
            if not path.is_symlink() and path.suffix.lower() in VIDEO_EXTENSIONS:
                media_paths.append(path)

    series_directories = {path.parent for path in media_paths if parse_media_filename(path).kind == "电视剧"}
    numeric_files: dict[Path, set[int]] = defaultdict(set)
    for path in media_paths:
        if re.fullmatch(r"\d{1,3}", path.stem):
            numeric_files[path.parent].add(int(path.stem))
    for directory, numbers in numeric_files.items():
        if any(number + 1 in numbers for number in numbers):
            series_directories.add(directory)

    for path in media_paths:
        try:
            classified = classify(
                path,
                destination,
                (overrides or {}).get(str(path)),
                series_directory=path.parent in series_directories,
            )
        except (TypeError, ValueError) as exc:
            items.append(PlanItem(str(path), "", "未识别", "skipped", str(exc)))
            continue
        if classified is None:
            items.append(PlanItem(str(path), "", "未识别", "skipped", "文件名中没有可识别的标题"))
            continue
        kind, target, metadata = classified
        if path.resolve() == target.resolve():
            items.append(PlanItem(str(path), str(target), kind, "unchanged", "文件已在目标位置", **metadata))
        elif target.exists():
            items.append(PlanItem(str(path), str(target), kind, "conflict", "目标文件已存在，不会覆盖", **metadata))
        else:
            items.append(PlanItem(str(path), str(target), kind, "ready", **metadata))
    return items


def apply_plan(items: list[PlanItem]) -> dict:
    result = {
        "ready": sum(item.status == "ready" for item in items),
        "moved": 0,
        "skipped": 0,
        "failed": 0,
        "details": [],
        "details_total": 0,
    }

    def add_detail(item: PlanItem, status: str, message: str) -> None:
        result["details_total"] += 1
        if len(result["details"]) < MAX_OPERATION_DETAILS:
            result["details"].append({
                "status": status,
                "source": item.source[:500],
                "destination": item.destination[:500],
                "message": message[:300],
            })

    for item in items:
        if item.status != "ready":
            result["skipped"] += 1
            add_detail(item, item.status, item.reason or "未执行")
            continue
        source, target = Path(item.source), Path(item.destination)
        try:
            if not source.is_file():
                result["skipped"] += 1
                add_detail(item, "skipped", "源文件不存在或不可访问")
                continue
            if target.exists():
                result["skipped"] += 1
                add_detail(item, "conflict", "目标文件已存在，不会覆盖")
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.move(str(source), str(target))
            result["moved"] += 1
        except OSError as exc:
            result["failed"] += 1
            add_detail(item, "failed", str(exc))
    return result
