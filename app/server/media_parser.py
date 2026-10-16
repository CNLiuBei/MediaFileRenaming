"""Parse common movie and TV naming conventions without external packages."""
from __future__ import annotations

import re
import unicodedata
from dataclasses import dataclass
from pathlib import Path

YEAR_PATTERN = re.compile(r"(?<!\d)((?:19|20)\d{2})(?!\d)")
EPISODE_PATTERNS = (
    re.compile(r"(?i)^(?P<title>.*?)S(?P<season>\d{1,2})[ ._-]*E(?P<episode>\d{1,3})(?:[ ._-]*(?:E|[-~+&])[ ._-]*(?P<end>\d{1,3}))?"),
    re.compile(r"(?i)^(?P<title>.*?)(?P<season>\d{1,2})x(?P<episode>\d{1,3})(?:-(?P<end>\d{1,3}))?"),
    re.compile(r"(?i)^(?P<title>.*?)Season[ ._-]*(?P<season>\d{1,2})[ ._-]*Episode[ ._-]*(?P<episode>\d{1,3})"),
    re.compile(r"(?i)^(?P<title>.*?)(?:Episode\b|Ep\b|(?:^|[ ._-])E)[ ._-]*(?P<episode>\d{1,3})\b"),
    re.compile(r"^(?P<title>.*?)第\s*(?P<season>\d{1,2})\s*季\s*第\s*(?P<episode>\d{1,3})\s*集"),
    re.compile(r"^(?P<title>.*?)第\s*(?P<episode>\d{1,3})\s*集"),
)
SEASON_DIRECTORY = re.compile(r"(?i)^(?:season|series|s)[ ._-]*(\d{1,2})$")
CHINESE_SEASON_DIRECTORY = re.compile(r"^(?:第\s*)?(\d{1,2})\s*季$")
SEASON_SUFFIX = re.compile(r"(?i)^(?P<title>.+?)[ ._-]+S(?P<season>\d{1,2})$")
RELEASE_PREFIX = re.compile(
    r"(?i)^\s*(?:【[^】]*(?:发布|发布页|www\.|\.com)[^】]*】|"
    r"\[[^\]]*(?:发布|发布页|www\.|\.com)[^\]]*\])\s*"
)
RELEASE_SITE_TRAILER = re.compile(
    r"(?i)(?:^|\s)(?:电影港\s*)?(?:地址发布页|发布页|收藏不迷路|"
    r"www\.[a-z0-9.-]+\.(?:com|net|org|me|tv))\b.*$"
)
BROKEN_METADATA_TAG = re.compile(
    r"(?i)[\[【（(]\s*(?:全\s*\d+\s*集|第\s*\d+\s*季|\d+\s*帧率|"
    r"国语|粤语|中文字幕|简繁英字幕|简体中字|繁体中字|杜比视界|无字幕源).*$"
)
HAN_CHARACTER = re.compile(r"[\u3400-\u9fff]")
NOISE_PATTERN = re.compile(
    r"(?i)(?:^|\s)(?:(?:hd)?(?:2160|1080|720|576|480|4320)[pi]|"
    r"2160p|1080p|720p|576p|480p|4320p|4k|8k|uhd|fhd|hd|"
    r"blu[ ._-]?ray|b[dr]rip|web[ ._-]?(?:dl|rip)|hdtv|pdtv|remux|dvd(?:rip)?|"
    r"x26[45]|h[ ._-]?26[45]|hevc|av1|xvid|divx|hdr10?(?:\+)?|dv|dolby[ ._-]?vision|"
    r"aac\d*(?:[ ._-]?\d+\.\d)?|ddp?\d?(?:[ ._-]?\d+\.\d)?|dts(?:[ ._-]?hd)?|"
    r"eac3|ac3|flac|truehd|atmos|chs(?:[ ._-]+eng)?|cht|简体中字|繁体中字|简繁中字|简繁英字幕|"
    r"中文字幕|国语配音|无字幕源|中字|双语|内封|外挂字幕|"
    r"proper|repack|extended|remastered|limited|internal|readnfo|10bit|8bit|amzn|nf|dsnp|hmax|"
    r"cr|uncut)(?=\s|$).*$"
)


@dataclass(frozen=True)
class ParsedMedia:
    kind: str
    title: str
    year: str = ""
    season: int = 0
    episode: int = 0
    episode_end: int = 0


def _normalize(value: str) -> str:
    return unicodedata.normalize("NFKC", value).strip()


def _prefer_native_title(value: str) -> str:
    """Prefer the Chinese title when a release also appends its English alias."""
    words = value.split()
    has_han = False
    for index, word in enumerate(words):
        if has_han and re.match(r"^[A-Za-z][A-Za-z0-9'’:_-]*$", word):
            return " ".join(words[:index]).strip()
        has_han = has_han or bool(HAN_CHARACTER.search(word))
    return value


def _trim_repeated_season_number(title: str, raw_title: str, season: int) -> str:
    """Drop a Chinese title's trailing season digit when the release repeats it as Sxx."""
    if season < 2 or not HAN_CHARACTER.search(raw_title) or not re.search(r"[A-Za-z]{2,}", raw_title):
        return title
    return re.sub(rf"(?<=[\u3400-\u9fff]){season}$", "", title).strip()


def _clean_title(value: str) -> str:
    value = _normalize(value)
    value = RELEASE_PREFIX.sub("", value)
    value = re.sub(r"(?i)www\.[a-z0-9.-]+\.(?:com|net|org|me|tv)\b.*$", " ", value)
    value = re.sub(r"\[[^\]]*\]|【[^】]*】|\{[^}]*\}|（[^）]*）", " ", value)
    value = re.sub(r"[._]+", " ", value)
    value = RELEASE_SITE_TRAILER.sub(" ", value)
    value = BROKEN_METADATA_TAG.sub(" ", value)
    value = re.sub(r"[\[【（(].*$", " ", value)
    value = NOISE_PATTERN.sub("", value)
    value = re.sub(r"\s+-[A-Za-z0-9]{2,}$", "", value)
    value = re.sub(r"\s+", " ", value).strip(" ._-[](){}【】（）")
    value = _prefer_native_title(value)
    return value or "未知名称"


def _series_directory(path: Path) -> str:
    parent = path.parent.name
    season_match = SEASON_DIRECTORY.fullmatch(parent)
    if season_match or CHINESE_SEASON_DIRECTORY.fullmatch(parent):
        parent = path.parent.parent.name
    else:
        season_suffix = SEASON_SUFFIX.fullmatch(parent)
        if season_suffix:
            raw_title = season_suffix.group("title")
            title = _clean_title(raw_title)
            return _trim_repeated_season_number(
                title, raw_title, int(season_suffix.group("season"))
            )
    return _clean_title(parent)


def _directory_season(path: Path) -> int | None:
    parent = path.parent.name
    match = SEASON_DIRECTORY.fullmatch(parent) or CHINESE_SEASON_DIRECTORY.fullmatch(parent)
    if not match:
        match = SEASON_SUFFIX.fullmatch(parent)
        if match:
            return int(match.group("season"))
    return int(match.group(1)) if match else None


def _parse_episode(stem: str, path: Path) -> ParsedMedia | None:
    normalized = _normalize(stem)
    for pattern in EPISODE_PATTERNS:
        match = pattern.search(normalized)
        if not match:
            continue
        raw_title = match.group("title")
        title = _clean_title(raw_title)
        if title == "未知名称":
            title = _series_directory(path)
        episode = int(match.group("episode"))
        season_value = match.groupdict().get("season")
        directory_season = _directory_season(path)
        season = int(season_value) if season_value else (directory_season if directory_season is not None else 1)
        title = _trim_repeated_season_number(title, raw_title, season)
        episode_end = int(match.groupdict().get("end") or 0)
        if title == "未知名称" or season > 99 or episode > 999 or episode_end > 999:
            return None
        return ParsedMedia("电视剧", title, season=season, episode=episode, episode_end=episode_end)
    season = _directory_season(path)
    numeric_episode = re.fullmatch(r"(?i)(?:episode|ep|e)?[ ._-]*(\d{1,3})", normalized)
    if season is not None and numeric_episode:
        return ParsedMedia(
            "电视剧", _series_directory(path), season=season, episode=int(numeric_episode.group(1))
        )
    if season is not None:
        absolute_match = re.fullmatch(r"(.+?)[ ._-]+(\d{1,3})", _clean_title(normalized))
        if absolute_match and _clean_title(absolute_match.group(1)).casefold() == _series_directory(path).casefold():
            return ParsedMedia(
                "电视剧", _series_directory(path), season=season, episode=int(absolute_match.group(2))
            )
    return None


def parse_media_filename(path: Path, *, series_directory: bool = False) -> ParsedMedia:
    """Infer media type and metadata from a file name and useful parent folders."""
    stem = _normalize(path.stem)
    episode = _parse_episode(stem, path)
    if episode:
        return episode
    if re.fullmatch(r"\d{1,3}", stem):
        if series_directory:
            season = _directory_season(path)
            return ParsedMedia(
                "电视剧", _series_directory(path), season=season if season is not None else 1, episode=int(stem)
            )
        directory_year = YEAR_PATTERN.search(_normalize(path.parent.name))
        if directory_year:
            directory_title = _clean_title(path.parent.name[: directory_year.start()])
            if directory_title == stem:
                return ParsedMedia("电影", directory_title, year=directory_year.group(1))
        return ParsedMedia("电影", "未知名称")
    year_matches = list(YEAR_PATTERN.finditer(stem))
    year_match = year_matches[-1] if year_matches else None
    directory_match = YEAR_PATTERN.search(_normalize(path.parent.name))
    title_source = stem[: year_match.start()] + " " + stem[year_match.end() :] if year_match else stem
    title = _clean_title(title_source)
    generic_names = {"movie", "film", "video", "unknown", "未知名称"}
    if directory_match:
        directory_title = _clean_title(path.parent.name[: directory_match.start()])
        use_directory = title.casefold() == directory_title.casefold() or title.casefold() in generic_names
        if use_directory:
            title = directory_title
    else:
        use_directory = False
    year = year_match.group(1) if year_match else directory_match.group(1) if use_directory else ""
    return ParsedMedia("电影", title, year=year)
