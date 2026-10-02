"""Unit tests of skills/wt-log-work/scripts/wt.py against a fake API server.

Run from the repo root: python3 -m unittest discover -s skills/tests -v
"""

import contextlib
import importlib.util
import io
import json
import os
import pathlib
import stat
import tempfile
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlparse

SCRIPT = pathlib.Path(__file__).resolve().parents[1] / "wt-log-work" / "scripts" / "wt.py"
spec = importlib.util.spec_from_file_location("wt", SCRIPT)
wt = importlib.util.module_from_spec(spec)
spec.loader.exec_module(wt)

TOKEN = "wt_test-token"


class FakeApi:
    """In-memory stand-in for the Work Tracker API."""

    def __init__(self):
        self.categories = [
            {"id": "c1", "name": "Development", "active": True, "color": "blue"},
            {"id": "c2", "name": "Meeting", "active": True, "color": "green"},
            {"id": "c3", "name": "Old", "active": False, "color": "red"},
        ]
        self.projects = [
            # each kind has its own ID sequence, like the real API
            {"id": "c1", "name": "Work Tracker", "active": True},
            {"id": "c3", "name": "Legacy", "active": False},
        ]
        self.holidays = {}  # year -> [holiday]
        self.records = []
        self.requests = []

    def handle(self, method, path, query, body, auth):
        self.requests.append((method, path, query, body))
        if auth != "Bearer " + TOKEN:
            return 401, {"message": "Unauthorized"}
        if method == "GET" and path == "/api/me":
            return 200, {"message": "", "user": {"account": "SE0001", "name": "Amy", "role": "default"}}
        if method == "GET" and path == "/api/work/options":
            return 200, {"message": "", "categories": self.categories, "projects": self.projects, "allowViewAll": False}
        if method == "GET" and path == "/api/holidays":
            return 200, {"message": "", "holidays": self.holidays.get(query["year"][0], [])}
        if method == "GET" and path == "/api/me/work-records":
            start, end = query["from"][0], query["to"][0]
            records = [r for r in self.records if start <= r["date"] <= end]
            return 200, {"message": "", "records": records, "total": len(records),
                         "totalHours": sum(r["hours"] for r in records)}
        if method == "POST" and path == "/api/me/work-records":
            record = dict(body, id=str(len(self.records) + 1), account="SE0001", createdAt="2026-10-01T09:00:00Z")
            record.setdefault("projectId", "")
            self.records.append(record)
            return 200, {"message": "", "record": record}
        return 404, {"message": "Not found"}


class WtTestCase(unittest.TestCase):
    def setUp(self):
        self.api = FakeApi()
        api = self.api

        class Handler(BaseHTTPRequestHandler):
            def _serve(self):
                url = urlparse(self.path)
                length = int(self.headers.get("Content-Length") or 0)
                body = json.loads(self.rfile.read(length)) if length else None
                status, payload = api.handle(self.command, url.path, parse_qs(url.query), body,
                                             self.headers.get("Authorization"))
                data = json.dumps(payload).encode()
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

            do_GET = do_POST = _serve

            def log_message(self, *args):
                pass

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        threading.Thread(target=self.server.serve_forever, daemon=True).start()
        self.url = "http://127.0.0.1:%d" % self.server.server_port

        self.tmp = tempfile.TemporaryDirectory()
        self.config_dir = pathlib.Path(self.tmp.name) / "wt"
        os.environ["WT_CONFIG_DIR"] = str(self.config_dir)

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()
        self.tmp.cleanup()
        os.environ.pop("WT_CONFIG_DIR", None)

    def run_wt(self, *args):
        out = io.StringIO()
        with contextlib.redirect_stdout(out):
            code = wt.main(list(args))
        return code, json.loads(out.getvalue())

    def configure(self):
        code, result = self.run_wt("setup", "--url", self.url + "/", "--token", TOKEN)
        self.assertEqual(code, 0, result)


class SetupTest(WtTestCase):
    def test_check_without_config_asks_for_setup(self):
        code, result = self.run_wt("check")
        self.assertEqual(code, 2)
        self.assertFalse(result["configured"])
        self.assertIn("setup", result["hint"])

    def test_setup_saves_private_config(self):
        self.configure()
        config = self.config_dir / "config.json"
        self.assertEqual(stat.S_IMODE(self.config_dir.stat().st_mode), 0o700)
        self.assertEqual(stat.S_IMODE(config.stat().st_mode), 0o600)
        self.assertEqual(json.loads(config.read_text()), {"url": self.url, "token": TOKEN})

        code, result = self.run_wt("check")
        self.assertEqual(code, 0)
        self.assertEqual(result["account"], "SE0001")
        self.assertNotIn(TOKEN, json.dumps(result))

    def test_setup_rejects_invalid_token_and_saves_nothing(self):
        code, result = self.run_wt("setup", "--url", self.url, "--token", "wt_wrong")
        self.assertEqual(code, 1)
        self.assertIn("token", result["error"])
        self.assertFalse((self.config_dir / "config.json").exists())

    def test_check_reports_expired_token(self):
        self.configure()
        config = self.config_dir / "config.json"
        config.write_text(json.dumps({"url": self.url, "token": "wt_revoked"}))
        code, result = self.run_wt("check")
        self.assertEqual(code, 1)
        self.assertIn("setup", result["hint"])

    def test_check_reports_unreachable_server(self):
        self.configure()
        self.server.shutdown()
        self.server.server_close()
        code, result = self.run_wt("check")
        self.assertEqual(code, 1)
        self.assertIn("connect", result["error"])


class LogTest(WtTestCase):
    def setUp(self):
        super().setUp()
        self.configure()

    def test_options_lists_active_names_only(self):
        code, result = self.run_wt("options")
        self.assertEqual(code, 0)
        self.assertEqual(result["categories"], ["Development", "Meeting"])
        self.assertEqual(result["projects"], ["Work Tracker"])

    def test_log_resolves_names_case_insensitively(self):
        code, result = self.run_wt("log", "--date", "2026-10-01", "--category", "development",
                                   "--project", "work tracker", "--hours", "1.5",
                                   "--description", "API token")
        self.assertEqual(code, 0, result)
        body = self.api.requests[-1][3]
        self.assertEqual(body, {"date": "2026-10-01", "categoryId": "c1", "projectId": "c1",
                                "hours": 1.5, "description": "API token"})
        self.assertEqual(result["record"]["category"], "Development")
        self.assertEqual(result["record"]["project"], "Work Tracker")

    def test_log_without_project(self):
        code, _ = self.run_wt("log", "--date", "2026-10-01", "--category", "Meeting",
                              "--hours", "1", "--description", "daily")
        self.assertEqual(code, 0)
        self.assertNotIn("projectId", self.api.requests[-1][3])

    def test_log_without_description(self):
        for extra in ((), ("--description", "  ")):
            code, result = self.run_wt("log", "--date", "2026-10-01", "--category", "Meeting",
                                       "--hours", "1", *extra)
            self.assertEqual(code, 0, result)
            self.assertEqual(self.api.requests[-1][3]["description"], "")
            self.assertEqual(result["record"]["description"], "")

    def test_log_unknown_or_inactive_name_lists_candidates(self):
        for args in (("--category", "Testing"), ("--category", "Old")):
            code, result = self.run_wt("log", "--date", "2026-10-01", *args, "--hours", "1",
                                       "--description", "x")
            self.assertEqual(code, 1)
            self.assertEqual(result["candidates"], ["Development", "Meeting"])
        code, result = self.run_wt("log", "--date", "2026-10-01", "--category", "Meeting",
                                   "--project", "Legacy", "--hours", "1", "--description", "x")
        self.assertEqual(code, 1)
        self.assertEqual(result["candidates"], ["Work Tracker"])
        self.assertFalse(any(r[0] == "POST" for r in self.api.requests))

    def test_log_rejects_bad_hours_and_dates(self):
        for hours in ("0", "1.2", "24.5", "abc"):
            code, result = self.run_wt("log", "--date", "2026-10-01", "--category", "Meeting",
                                       "--hours", hours, "--description", "x")
            self.assertEqual(code, 1, hours)
            self.assertIn("hours", result["error"])
        code, result = self.run_wt("log", "--date", "2026/10/01", "--category", "Meeting",
                                   "--hours", "1", "--description", "x")
        self.assertEqual(code, 1)
        self.assertIn("date", result["error"])

    def test_log_dry_run_posts_nothing(self):
        code, result = self.run_wt("log", "--dry-run", "--date", "2026-10-01", "--category", "Meeting",
                                   "--hours", "2", "--description", "x")
        self.assertEqual(code, 0)
        self.assertTrue(result["dryRun"])
        self.assertFalse(any(r[0] == "POST" for r in self.api.requests))


class DailyTest(WtTestCase):
    def setUp(self):
        super().setUp()
        self.configure()
        # 2026-10-01 (Thu) is a holiday; 2026-10-03 (Sat) a makeup workday
        self.api.holidays["2026"] = [
            {"date": "2026-10-01", "name": "Holiday", "type": "holiday", "source": "gov"},
            {"date": "2026-10-03", "name": "Makeup", "type": "workday", "source": "manual"},
        ]

    def last_workday(self, before):
        code, result = self.run_wt("last-workday", "--before", before)
        self.assertEqual(code, 0, result)
        return result["date"]

    def test_last_workday(self):
        self.assertEqual(self.last_workday("2026-09-29"), "2026-09-28")  # Tue -> Mon
        self.assertEqual(self.last_workday("2026-09-28"), "2026-09-25")  # Mon -> Fri
        self.assertEqual(self.last_workday("2026-10-02"), "2026-09-30")  # skips the holiday
        self.assertEqual(self.last_workday("2026-10-05"), "2026-10-03")  # makeup Saturday

    def test_last_workday_across_years(self):
        self.api.holidays["2026"] = [{"date": "2026-01-01", "name": "New Year", "type": "holiday", "source": "gov"}]
        self.assertEqual(self.last_workday("2026-01-02"), "2025-12-31")
        self.assertIn(["2025"], [r[2].get("year") for r in self.api.requests])

    def test_daily_groups_records(self):
        self.api.records = [
            {"id": "1", "date": "2026-09-30", "categoryId": "c1", "projectId": "c1", "hours": 3,
             "description": "API", "account": "SE0001", "createdAt": ""},
            {"id": "2", "date": "2026-09-30", "categoryId": "c2", "projectId": "", "hours": 1,
             "description": "sync", "account": "SE0001", "createdAt": ""},
            {"id": "3", "date": "2026-09-29", "categoryId": "c1", "projectId": "", "hours": 8,
             "description": "older", "account": "SE0001", "createdAt": ""},
        ]
        code, result = self.run_wt("daily", "--today", "2026-10-02")
        self.assertEqual(code, 0, result)
        self.assertEqual(result["date"], "2026-09-30")
        self.assertEqual(result["weekday"], "Wednesday")
        self.assertEqual(result["totalHours"], 4)
        self.assertEqual([r["description"] for r in result["records"]], ["API", "sync"])
        self.assertEqual(result["records"][0]["category"], "Development")
        self.assertEqual(result["records"][0]["project"], "Work Tracker")
        self.assertEqual(result["records"][1]["project"], "")

    def test_daily_with_explicit_date(self):
        code, result = self.run_wt("daily", "--date", "2026-09-29")
        self.assertEqual(code, 0)
        self.assertEqual(result["date"], "2026-09-29")
        self.assertEqual(result["records"], [])
        self.assertEqual(result["totalHours"], 0)

    def test_records_range(self):
        self.api.records = [{"id": "1", "date": "2026-09-30", "categoryId": "c3", "projectId": "c3",
                             "hours": 2, "description": "x", "account": "SE0001", "createdAt": ""}]
        code, result = self.run_wt("records", "--from", "2026-09-28", "--to", "2026-10-04")
        self.assertEqual(code, 0)
        self.assertEqual(result["records"][0]["category"], "Old")  # inactive names still resolve
        self.assertEqual(result["records"][0]["project"], "Legacy")


if __name__ == "__main__":
    unittest.main()
