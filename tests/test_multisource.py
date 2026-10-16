import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "server"))

from organizer import (
    PlanItem,
    apply_plan,
    build_plan,
    normalize_source_dirs,
    validate_source_folders,
)
from api_helpers import requested_overrides


class MultiSourceTests(unittest.TestCase):
    def test_legacy_single_source_and_multiple_sources_are_normalized(self):
        self.assertEqual(normalize_source_dirs([], "/media/inbox"), ["/media/inbox"])
        self.assertEqual(
            normalize_source_dirs([" /media/a ", "/media/b", "/media/a"]),
            ["/media/a", "/media/b"],
        )

    def test_multiple_sources_build_one_plan(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            first = root / "first"
            second = root / "second"
            destination = root / "library"
            first.mkdir()
            second.mkdir()
            (first / "Arrival.2016.mkv").touch()
            (second / "Show.S01E02.mkv").touch()

            plan = build_plan([first, second], destination)

        self.assertEqual(len(plan), 2)
        self.assertEqual({Path(item.source).parent.name for item in plan}, {"first", "second"})

    def test_source_folders_cannot_overlap(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            parent = root / "inbox"
            child = parent / "series"
            child.mkdir(parents=True)

            with self.assertRaisesRegex(ValueError, "不能重复或互相包含"):
                validate_source_folders([str(parent), str(child)], [root.resolve()])

    def test_override_file_can_belong_to_any_selected_source(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            first = root / "first"
            second = root / "second"
            first.mkdir()
            second.mkdir()
            media_file = second / "Show.S01E01.mkv"
            media_file.touch()

            overrides = requested_overrides(
                {"overrides": {str(media_file): {"title": "Show"}}},
                [first.resolve(), second.resolve()],
            )

        self.assertEqual(overrides[str(media_file.resolve())]["title"], "Show")

    def test_apply_plan_includes_skipped_file_details(self):
        item = PlanItem(
            source="/media/unknown.mkv",
            destination="",
            kind="未识别",
            status="skipped",
            reason="文件名中没有可识别的标题",
        )

        result = apply_plan([item])

        self.assertEqual(result["skipped"], 1)
        self.assertEqual(result["ready"], 0)
        self.assertEqual(result["details_total"], 1)
        self.assertEqual(result["details"][0]["message"], "文件名中没有可识别的标题")


if __name__ == "__main__":
    unittest.main()
