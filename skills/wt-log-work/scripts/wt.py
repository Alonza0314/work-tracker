#!/usr/bin/env python3
"""Work Tracker (wt) API helper for Claude skills. Python 3 standard library only.

Every command prints one JSON object. Exit code 0 = ok, 1 = error, 2 = not configured.

  setup [--url URL] [--token TOKEN]   save the server URL and API token (prompts when omitted)
  check                               verify the saved config
  options                             active category and project names
  log [--date D] [--category C] [--hours H] [--description D] [--project P] [--dry-run]
  records --from DATE --to DATE       my work records in a date range
  last-workday [--before DATE]        the workday before DATE (default today)
  daily [--date DATE | --today DATE]  my records of DATE, or of the workday before --today

The config lives in ~/.wt/config.json (override the folder with WT_CONFIG_DIR).
"""

import argparse
import datetime
import getpass
import json
import os
import pathlib
import sys
import urllib.error
import urllib.parse
import urllib.request

SETUP_HINT = "Ask the user to run in Claude Code: ! python3 ~/.claude/skills/wt-log-work/scripts/wt.py setup"
HOURS_STEP = 0.5
MAX_HOURS = 24
LOOKBACK_DAYS = 31


class WtError(Exception):
    def __init__(self, message, code=1, **extra):
        super().__init__(message)
        self.code = code
        self.extra = extra


# ---------------------------------------------------------------- config

def config_dir():
    return pathlib.Path(os.environ.get("WT_CONFIG_DIR") or pathlib.Path.home() / ".wt")


def config_file():
    return config_dir() / "config.json"


def load_config():
    path = config_file()
    try:
        config = json.loads(path.read_text())
    except FileNotFoundError:
        raise WtError("Work Tracker is not configured", code=2, configured=False, hint=SETUP_HINT)
    except (OSError, ValueError) as err:
        raise WtError("cannot read %s: %s" % (path, err), code=2, configured=False, hint=SETUP_HINT)
    if not config.get("url") or not config.get("token"):
        raise WtError("%s lacks url or token" % path, code=2, configured=False, hint=SETUP_HINT)
    return config


def save_config(config):
    directory = config_dir()
    directory.mkdir(mode=0o700, parents=True, exist_ok=True)
    os.chmod(directory, 0o700)
    path = config_file()
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, "w") as f:
        json.dump(config, f)
    os.chmod(path, 0o600)


def normalize_url(url):
    url = url.strip().rstrip("/")
    if not url:
        raise WtError("the server URL is empty")
    if "://" not in url:
        url = "http://" + url
    return url


# ---------------------------------------------------------------- API

class Client:
    def __init__(self, url, token):
        self.url = url
        self.token = token

    def request(self, method, path, query=None, body=None):
        url = self.url + path
        if query:
            url += "?" + urllib.parse.urlencode(query)
        data = json.dumps(body).encode() if body is not None else None
        req = urllib.request.Request(url, data=data, method=method)
        req.add_header("Authorization", "Bearer " + self.token)
        req.add_header("Accept", "application/json")
        if data is not None:
            req.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(req, timeout=15) as resp:
                return json.loads(resp.read() or b"{}")
        except urllib.error.HTTPError as err:
            try:
                message = json.loads(err.read()).get("message") or err.reason
            except ValueError:
                message = err.reason
            if err.code == 401:
                raise WtError("the API token is invalid, expired or revoked", hint=SETUP_HINT)
            raise WtError("%s %s failed (%d): %s" % (method, path, err.code, message), status=err.code)
        except (urllib.error.URLError, OSError) as err:
            reason = getattr(err, "reason", err)
            raise WtError("cannot connect to %s: %s" % (self.url, reason))
        except ValueError:
            raise WtError("%s did not answer JSON; is the URL the Work Tracker site?" % self.url)

    def me(self):
        return self.request("GET", "/api/me").get("user") or {}

    def options(self):
        return self.request("GET", "/api/work/options")

    def records(self, start, end):
        return self.request("GET", "/api/me/work-records", {"from": start, "to": end}).get("records") or []

    def holidays(self, year):
        return self.request("GET", "/api/holidays", {"year": str(year)}).get("holidays") or []


def client():
    config = load_config()
    return Client(normalize_url(config["url"]), config["token"])


# ---------------------------------------------------------------- helpers

def parse_date(value, what="date"):
    try:
        return datetime.date.fromisoformat(value)
    except (TypeError, ValueError):
        raise WtError("invalid %s %r; use YYYY-MM-DD" % (what, value))


def parse_hours(value):
    try:
        hours = float(value)
    except (TypeError, ValueError):
        raise WtError("invalid hours %r" % value)
    if not 0 <= hours <= MAX_HOURS or (hours / HOURS_STEP) != int(hours / HOURS_STEP):
        raise WtError("hours must be in [0, %d] in steps of %s, got %s" % (MAX_HOURS, HOURS_STEP, value))
    return int(hours) if hours == int(hours) else hours


def resolve(options, value, kind):
    """Returns the ID of the active option named value (case-insensitive) or with that ID."""
    active = [o for o in options if o.get("active")]
    wanted = value.strip().casefold()
    for option in active:
        if option["name"].strip().casefold() == wanted or option["id"] == value:
            return option
    raise WtError("no active %s named %r" % (kind, value), candidates=[o["name"] for o in active])


def describe(records, options):
    # categories and projects have separate ID sequences, so the same ID can name one of each
    categories = {o["id"]: o["name"] for o in options["categories"]}
    projects = {o["id"]: o["name"] for o in options["projects"]}
    return [{
        "id": r["id"],
        "date": r["date"],
        "category": categories.get(r.get("categoryId"), ""),
        "project": projects.get(r.get("projectId"), ""),
        "hours": r.get("hours", 0),
        "description": r.get("description", ""),
    } for r in records]


def total_hours(records):
    total = sum(r.get("hours", 0) for r in records)
    return int(total) if total == int(total) else total


def last_workday(api, before):
    calendars = {}
    day = before
    for _ in range(LOOKBACK_DAYS):
        day -= datetime.timedelta(days=1)
        if day.year not in calendars:
            calendars[day.year] = {h["date"]: h["type"] for h in api.holidays(day.year)}
        kind = calendars[day.year].get(day.isoformat())
        if kind == "workday" or (kind != "holiday" and day.weekday() < 5):
            return day
    raise WtError("no workday in the %d days before %s" % (LOOKBACK_DAYS, before))


# ---------------------------------------------------------------- commands

def cmd_setup(args):
    url = args.url or input("Work Tracker URL (e.g. http://wt.example.com): ")
    token = args.token or getpass.getpass("API token (wt_..., input hidden): ")
    url, token = normalize_url(url), token.strip()
    if not token:
        raise WtError("the API token is empty")
    user = Client(url, token).me()
    save_config({"url": url, "token": token})
    return {"configured": True, "url": url, "account": user.get("account"), "name": user.get("name"),
            "config": str(config_file())}


def cmd_check(args):
    api = client()
    user = api.me()
    return {"configured": True, "url": api.url, "account": user.get("account"), "name": user.get("name")}


def cmd_options(args):
    options = client().options()
    return {
        "categories": [o["name"] for o in options["categories"] if o.get("active")],
        "projects": [o["name"] for o in options["projects"] if o.get("active")],
    }


def cmd_log(args):
    day = parse_date(args.date) if args.date else datetime.date.today()
    hours = parse_hours(args.hours)
    description = (args.description or "").strip()

    api = client()
    options = api.options()
    body = {"date": day.isoformat(), "description": description, "hours": hours}
    if args.category:
        body["categoryId"] = resolve(options["categories"], args.category, "category")["id"]
    if args.project:
        body["projectId"] = resolve(options["projects"], args.project, "project")["id"]

    if args.dry_run:
        return {"dryRun": True, "record": describe([dict(body, id="")], options)[0]}
    record = api.request("POST", "/api/me/work-records", body=body)["record"]
    return {"record": describe([record], options)[0]}


def cmd_records(args):
    start, end = parse_date(args.start, "from"), parse_date(args.end, "to")
    api = client()
    records = api.records(start.isoformat(), end.isoformat())
    return {"from": start.isoformat(), "to": end.isoformat(), "records": describe(records, api.options()),
            "totalHours": total_hours(records)}


def cmd_last_workday(args):
    before = parse_date(args.before, "before") if args.before else datetime.date.today()
    day = last_workday(client(), before)
    return {"date": day.isoformat(), "weekday": day.strftime("%A")}


def cmd_daily(args):
    api = client()
    if args.date:
        day = parse_date(args.date)
    else:
        day = last_workday(api, parse_date(args.today, "today") if args.today else datetime.date.today())
    records = api.records(day.isoformat(), day.isoformat())
    records.sort(key=lambda r: (r.get("createdAt", ""), r["id"]))
    return {"date": day.isoformat(), "weekday": day.strftime("%A"), "records": describe(records, api.options()),
            "totalHours": total_hours(records)}


def parser():
    p = argparse.ArgumentParser(prog="wt.py", description="Work Tracker API helper")
    sub = p.add_subparsers(dest="command", required=True)

    s = sub.add_parser("setup", help="save the server URL and API token")
    s.add_argument("--url")
    s.add_argument("--token")
    s.set_defaults(run=cmd_setup)

    sub.add_parser("check", help="verify the saved config").set_defaults(run=cmd_check)
    sub.add_parser("options", help="active category and project names").set_defaults(run=cmd_options)

    s = sub.add_parser("log", help="add one work record")
    s.add_argument("--date", help="YYYY-MM-DD, default today")
    s.add_argument("--category", help="may be omitted")
    s.add_argument("--project")
    s.add_argument("--hours", default="0", help="default 0")
    s.add_argument("--description", default="", help="may be empty")
    s.add_argument("--dry-run", action="store_true", help="validate without saving")
    s.set_defaults(run=cmd_log)

    s = sub.add_parser("records", help="my work records in a date range")
    s.add_argument("--from", dest="start", required=True)
    s.add_argument("--to", dest="end", required=True)
    s.set_defaults(run=cmd_records)

    s = sub.add_parser("last-workday", help="the workday before a date")
    s.add_argument("--before", help="YYYY-MM-DD, default today")
    s.set_defaults(run=cmd_last_workday)

    s = sub.add_parser("daily", help="my records of the last workday")
    group = s.add_mutually_exclusive_group()
    group.add_argument("--date", help="report this date")
    group.add_argument("--today", help="report the workday before this date, default today")
    s.set_defaults(run=cmd_daily)
    return p


def main(argv=None):
    args = parser().parse_args(argv)
    try:
        result, code = args.run(args), 0
    except WtError as err:
        result, code = dict({"error": str(err)}, **err.extra), err.code
    except (KeyboardInterrupt, EOFError):
        result, code = {"error": "cancelled"}, 1
    print(json.dumps(result, ensure_ascii=False, indent=2))
    return code


if __name__ == "__main__":
    sys.exit(main())
