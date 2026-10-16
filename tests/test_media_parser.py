import csv
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "server"))

from media_parser import parse_media_filename
from organizer import build_plan, classify


class MediaParserTests(unittest.TestCase):
    def test_generated_corpus_matches_all_expected_metadata(self):
        corpus = Path(__file__).parent.parent / "fixtures" / "nonstandard-media-corpus"
        with (corpus / "manifest.csv").open(encoding="utf-8-sig", newline="") as manifest_file:
            rows = list(csv.DictReader(manifest_file))
        self.assertEqual(len(rows), 1000)
        with tempfile.TemporaryDirectory() as destination:
            items = {item.source: item for item in build_plan(corpus, Path(destination))}
        for row in rows:
            with self.subTest(case_id=row["case_id"]):
                item = items[str(corpus / row["relative_path"])]
                if row["expected_title"] == "未知名称":
                    self.assertEqual(item.status, "skipped")
                    continue
                actual = (item.kind, item.title, item.year, item.season, item.episode, item.episode_end)
                expected = (row["expected_kind"], row["expected_title"], row["expected_year"],
                            int(row["expected_season"]), int(row["expected_episode"]),
                            int(row["expected_episode_end"]))
                self.assertEqual(actual, expected)

    def test_movie_title_year_and_release_tags(self):
        parsed = parse_media_filename(Path("The.Matrix.1999.1080p.BluRay.x265-GROUP.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.year), ("电影", "The Matrix", "1999"))

    def test_chinese_series_with_attached_season_marker_and_range(self):
        parsed = parse_media_filename(Path("测试剧名S01E02-E03.1080p.WEB-DL.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode, parsed.episode_end),
                         ("电视剧", "测试剧名", 1, 2, 3))

    def test_x_episode_format(self):
        parsed = parse_media_filename(Path("Show.Name.2x05.mkv"))
        self.assertEqual((parsed.title, parsed.season, parsed.episode), ("Show Name", 2, 5))

    def test_named_season_episode_format(self):
        parsed = parse_media_filename(Path("Show Name Season 2 Episode 8.mkv"))
        self.assertEqual((parsed.title, parsed.season, parsed.episode), ("Show Name", 2, 8))

    def test_episode_prefix_uses_series_directory(self):
        parsed = parse_media_filename(Path("剧集目录/Episode 08.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "剧集目录", 1, 8))

    def test_chinese_season_episode_format(self):
        parsed = parse_media_filename(Path("测试剧名第2季第8集.mp4"))
        self.assertEqual((parsed.title, parsed.season, parsed.episode), ("测试剧名", 2, 8))

    def test_episode_only_name_uses_season_and_series_directories(self):
        parsed = parse_media_filename(Path("剧集目录/Season 02/第03集.mkv"))
        self.assertEqual((parsed.title, parsed.season, parsed.episode), ("剧集目录", 2, 3))

    def test_numeric_episode_uses_explicit_season_directory(self):
        parsed = parse_media_filename(Path("剧集目录/Season 02/03.mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "剧集目录", 2, 3))

    def test_absolute_episode_uses_matching_series_and_season_directories(self):
        parsed = parse_media_filename(Path("剧集目录/Season 02/剧集目录 - 003 [1080p].mkv"))
        self.assertEqual((parsed.kind, parsed.title, parsed.season, parsed.episode),
                         ("电视剧", "剧集目录", 2, 3))

    def test_multiple_numbered_files_under_series_folder_are_inferred_as_episodes(self):
        with tempfile.TemporaryDirectory() as directory:
            source = Path(directory) / "source"
            series = source / "剧名"
            series.mkdir(parents=True)
            (series / "01.mkv").touch()
            (series / "02.mkv").touch()
            items = build_plan(source, Path(directory) / "destination")
        self.assertEqual([(item.kind, item.title, item.episode) for item in items],
                         [("电视剧", "剧名", 1), ("电视剧", "剧名", 2)])

    def test_single_numeric_file_without_series_context_is_skipped(self):
        parsed = parse_media_filename(Path("some-folder/03.mkv"))
        self.assertEqual((parsed.kind, parsed.title), ("电影", "未知名称"))

    def test_season_episode_only_name_uses_series_directory(self):
        parsed = parse_media_filename(Path("剧集目录/Season 02/S02E03.mkv"))
        self.assertEqual((parsed.title, parsed.season, parsed.episode), ("剧集目录", 2, 3))

    def test_movie_folder_supplies_year_when_file_name_omits_it(self):
        parsed = parse_media_filename(Path("Arrival (2016)/Arrival.mkv"))
        self.assertEqual((parsed.title, parsed.year), ("Arrival", "2016"))

    def test_movie_folder_does_not_assign_its_year_to_unrelated_files(self):
        parsed = parse_media_filename(Path("Arrival (2016)/Behind the Scenes.mkv"))
        self.assertEqual((parsed.title, parsed.year), ("Behind the Scenes", ""))

    def test_movie_classification_preserves_override_and_movie_layout(self):
        kind, destination, metadata = classify(
            Path("Arrival.2016.2160p.HDR.mkv"), Path("/library"), {"category": "英语电影"}
        )
        self.assertEqual(kind, "电影")
        self.assertEqual(metadata["title"], "Arrival")
        self.assertEqual(destination, Path("/library/电影/英语电影/Arrival (2016)/Arrival (2016).mkv"))

    def test_uncalibrated_media_defaults_to_pending_categories(self):
        movie_kind, movie_path, movie_metadata = classify(
            Path("Arrival.2016.mkv"), Path("/library")
        )
        series_kind, series_path, series_metadata = classify(
            Path("Show.S01E02.mkv"), Path("/library")
        )

        self.assertEqual((movie_kind, movie_metadata["category"]), ("电影", "待确认"))
        self.assertIn("/电影/待确认/", movie_path.as_posix())
        self.assertEqual((series_kind, series_metadata["category"]), ("电视剧", "待确认"))
        self.assertIn("/电视剧/待确认/", series_path.as_posix())

    def test_series_plan_keeps_multi_episode_range(self):
        kind, destination, metadata = classify(Path("Show.S01E02-E04.mkv"), Path("/library"))
        self.assertEqual(kind, "电视剧")
        self.assertEqual(metadata["episode_end"], 4)
        self.assertTrue(str(destination).endswith("Show - S01E02-E04.mkv"))


if __name__ == "__main__":
    unittest.main()
