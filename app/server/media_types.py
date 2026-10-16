"""Shared media parser result types."""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class ParsedMedia:
    kind: str
    title: str
    year: str = ""
    season: int = 0
    episode: int = 0
    episode_end: int = 0


@dataclass(frozen=True)
class GuessItMedia:
    kind: str
    title: str
    alternative_title: str = ""
    year: str = ""
    country_code: str = ""
    season: int | None = None
    episode: int | None = None
    episode_end: int = 0
