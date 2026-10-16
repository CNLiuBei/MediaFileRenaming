import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "server"))

from guessit_adapter import parse_with_guessit
from media_parser import parse_media_filename


class GuessItParserTests(unittest.TestCase):
    def test_guessit_dependency_parses_release_metadata(self):
        parsed = parse_with_guessit("The.Office.US.S02E04.HDTV.mkv")
        if parsed is None:
            self.skipTest("GuessIt 依赖尚未安装")
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "The Office", 2, 4))

    def test_guessit_recognizes_episode_number_in_matching_series_folder(self):
        parsed = parse_media_filename(Path("Show Name/Show.Name.03.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "Show Name", 1, 3))

    def test_guessit_removes_country_code_from_series_title(self):
        parsed = parse_media_filename(Path("The.Office.US.S02E04.HDTV.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "The Office", 2, 4))

    def test_year_and_episode_number_are_not_misread_as_a_season(self):
        parsed = parse_media_filename(Path("Show Name 2025 03.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.year),
                         ("电影", "Show Name 03", "2025"))


if __name__ == "__main__":
    unittest.main()
