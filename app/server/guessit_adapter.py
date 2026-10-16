"""Optional GuessIt integration with safe fallbacks for filename parsing."""
from __future__ import annotations

import logging
import re
import sys
from functools import lru_cache
from pathlib import Path
from typing import Callable

from media_types import GuessItMedia, ParsedMedia

LOGGER = logging.getLogger(__name__)
VENDOR_DIRECTORY = Path(__file__).with_name("vendor")
HAN_CHARACTER = re.compile(r"[\u3400-\u9fff]")
RELEASE_PREFIX = re.compile(
    r"(?i)^\s*(?:【[^】]*(?:发布|发布页|www\.|\.com)[^】]*】|"
    r"\[[^\]]*(?:发布|发布页|www\.|\.com)[^\]]*\])\s*"
)
QUALITY_TRAILER = re.compile(
    r"(?i)(?:^|[ ._-])(?:hd)?(?:4320|2160|1080|720|576|480)[pi](?:[ ._-]|$).*$"
)
EPISODE_MARKER = re.compile(
    r"(?i)(?:\bS\d{1,2}[ ._-]*E\d{1,3}|\b\d{1,2}x\d{1,3}|"
    r"\bSeason[ ._-]*\d{1,2}[ ._-]*Episode[ ._-]*\d{1,3}|"
    r"\b(?:Episode|Ep)[ ._-]*\d{1,3}|(?:第\s*\d+\s*季)?\s*第\s*\d+\s*集)"
)

if VENDOR_DIRECTORY.is_dir() and str(VENDOR_DIRECTORY) not in sys.path:
    sys.path.insert(0, str(VENDOR_DIRECTORY))

try:
    from guessit import guessit
except ImportError:
    guessit = None


def _number(value: object) -> int | None:
    if isinstance(value, (list, tuple)):
        values = [item for item in value if isinstance(item, int)]
        return values[0] if values else None
    return value if isinstance(value, int) else None


@lru_cache(maxsize=16384)
def parse_with_guessit(filename: str) -> GuessItMedia | None:
    """Return validated GuessIt fields, or None when the optional library is absent."""
    if guessit is None:
        return None
    normalized_name = RELEASE_PREFIX.sub("", filename)
    normalized_name = QUALITY_TRAILER.sub("", normalized_name)
    try:
        result = guessit(normalized_name)
    except Exception:
        LOGGER.debug("GuessIt could not parse %r", filename, exc_info=True)
        return None

    result_type = result.get("type")
    if result_type not in {"movie", "episode"}:
        return None
    year = result.get("year")
    year_string = str(year) if isinstance(year, int) and 1900 <= year <= 2099 else ""
    seasons = result.get("season")
    season = _number(seasons)
    if season is not None and not 1 <= season <= 99:
        season = None
    episodes = result.get("episode")
    episode = _number(episodes)
    episode_end = 0
    if isinstance(episodes, (list, tuple)) and len(episodes) > 1:
        last = episodes[-1]
        if isinstance(last, int) and episode is not None and episode < last <= 999:
            episode_end = last
    if episode is not None and not 1 <= episode <= 999:
        episode = None

    title = result.get("title")
    alternative = result.get("alternative_title")
    country_code = getattr(result.get("country"), "alpha2", "")
    return GuessItMedia(
        kind="电视剧" if result_type == "episode" else "电影",
        title=title.strip() if isinstance(title, str) else "",
        alternative_title=alternative.strip() if isinstance(alternative, str) else "",
        year=year_string,
        country_code=country_code if isinstance(country_code, str) else "",
        season=season,
        episode=episode,
        episode_end=episode_end,
    )


def _preferred_title(guess: GuessItMedia, current_title: str, clean_title: Callable[[str], str]) -> str:
    candidates = (guess.title, guess.alternative_title)
    cleaned = [clean_title(candidate) for candidate in candidates if candidate]
    valid = [candidate for candidate in cleaned if candidate != "未知名称"]
    if not valid:
        return current_title
    country_suffix = f" {guess.country_code}" if guess.country_code else ""
    if country_suffix and current_title.casefold().endswith(country_suffix.casefold()):
        prefix = current_title[: -len(country_suffix)].rstrip()
        if prefix.casefold() in {candidate.casefold() for candidate in valid}:
            return next(candidate for candidate in valid if candidate.casefold() == prefix.casefold())
    current_has_han = bool(HAN_CHARACTER.search(current_title))
    native_title = next((candidate for candidate in valid if HAN_CHARACTER.search(candidate)), "")
    if current_title != "未知名称" and current_has_han:
        return current_title
    if native_title:
        return native_title
    return valid[0] if current_title == "未知名称" else current_title


def refine_with_guessit(
    parsed: ParsedMedia,
    path: Path,
    *,
    series_title: str,
    series_directory: bool,
    directory_season: int | None,
    clean_title: Callable[[str], str],
) -> ParsedMedia:
    """Use GuessIt conservatively, keeping the existing parser authoritative on conflicts."""
    guess = parse_with_guessit(path.name)
    if guess is None:
        return parsed
    guess_title = _preferred_title(guess, parsed.title, clean_title)
    candidate_title = _preferred_title(guess, "未知名称", clean_title)

    if parsed.kind == "电视剧" and guess.kind == "电视剧":
        episode_matches = guess.episode in {None, parsed.episode}
        season_matches = guess.season in {None, parsed.season}
        if episode_matches and season_matches and guess_title != "未知名称":
            return ParsedMedia(
                parsed.kind, guess_title, season=parsed.season, episode=parsed.episode,
                episode_end=parsed.episode_end,
            )
        return parsed

    if parsed.kind == "电影" and guess.kind == "电视剧":
        title_context_matches = bool(series_title) and candidate_title.casefold() == series_title.casefold()
        has_episode_context = bool(EPISODE_MARKER.search(path.stem)) or directory_season is not None
        if not (guess.episode and (has_episode_context or series_directory or title_context_matches)):
            return parsed
        if guess.season is None and directory_season is None and not (
            series_directory or title_context_matches or has_episode_context
        ):
            return parsed
        title = series_title if title_context_matches else candidate_title
        if title == "未知名称":
            return parsed
        return ParsedMedia(
            "电视剧", title, season=guess.season or directory_season or 1,
            episode=guess.episode, episode_end=guess.episode_end,
        )

    if parsed.kind == "电影" and guess.kind == "电影":
        title = parsed.title
        candidate_has_han = bool(HAN_CHARACTER.search(candidate_title))
        if title == "未知名称" or (candidate_has_han and not HAN_CHARACTER.search(title)):
            title = candidate_title
        year = parsed.year or guess.year
        if title != parsed.title or year != parsed.year:
            return ParsedMedia("电影", title, year=year)
    return parsed
