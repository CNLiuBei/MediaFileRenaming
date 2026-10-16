import json
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "server"))

import activity_log


class ActivityLogTests(unittest.TestCase):
    def test_activity_stores_counts_paths_and_capped_details(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            previous_root = activity_log.DATA_ROOT
            previous_path = activity_log.LOG_PATH
            activity_log.DATA_ROOT = root
            activity_log.LOG_PATH = root / "activity.jsonl"
            try:
                activity_log.record_activity(
                    "手动整理",
                    "整理完成",
                    total=27,
                    ready=20,
                    moved=20,
                    skipped=6,
                    failed=1,
                    source_dirs=["/media/a", "/media/b"],
                    destination_dir="/library",
                    details=[{"status": "failed", "source": "/media/a/bad.mkv", "message": "权限不足"}] * 25,
                    details_total=25,
                )

                entries = activity_log.recent_activity()
                stored = json.loads(activity_log.LOG_PATH.read_text(encoding="utf-8").splitlines()[0])
            finally:
                activity_log.DATA_ROOT = previous_root
                activity_log.LOG_PATH = previous_path

        self.assertEqual(entries[0]["total"], 27)
        self.assertEqual(entries[0]["ready"], 20)
        self.assertEqual(entries[0]["source_dirs"], ["/media/a", "/media/b"])
        self.assertEqual(entries[0]["destination_dir"], "/library")
        self.assertEqual(entries[0]["details_total"], 25)
        self.assertEqual(len(entries[0]["details"]), activity_log.MAX_LOG_DETAILS)
        self.assertEqual(stored["failed"], 1)


if __name__ == "__main__":
    unittest.main()
