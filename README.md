# work-tracker

![logo](./docs/images/wt-logo.png)

Work Tracker (wt) — a full-stack website (React + go) for tracking work.

## Develop Environment

| DevOpts | Version |
| - | - |
| OS | Ubuntu 25.04 |
| go | 1.26.2 |
| nodejs | v22.23.3 |
| yarn | 1.22.22 |

## Make

| Type | Command |
| - | - |
| Make all | `make` |
| Backend | `make backend` |
| Frontend | `make frontend` |
| Run | `make run` |
| Tidy | `make tidy` |
| Lint | `make lint` |
| Generate frontend openapi | `make openapi` |
| Docker image | `make docker` |
| Clean (build + db) | `make clean` |

`make` builds the backend binary and the frontend resources under the `build` directory.

## Execute

Setup the configuration file: [config.yaml](./config.yaml)

And run:

```bash
make run
```

## Install - Docker Compose

1. Check the config

    Modify work tracker settings at `./docker/config.yaml`, e.g. the default login credential:

    ```yaml
    username: "admin"
    password: "0000"
    ```

    For other settings, make sure the change is reflected in `docker-compose.yaml` too.

2. Up the compose

    ```bash
    cd docker
    docker compose up -d
    ```

    The default db is stored at `/var/lib/wt/db`, mounted in the compose file.

3. Down the compose

    ```bash
    docker compose down
    ```

## DB

The backend stores its data through the `DbIf` interface (`web/backend/internal/context/db.go`), so the database can be swapped by adding a new implementation. The implementation is selected in the config:

```yaml
db:
  type: "bbolt"          # currently the only supported type
  path: "/tmp/wt.db"     # file path of the bbolt database
```

### bbolt

Every bucket stores one record per key, with the value encoded as JSON. Buckets are created automatically when the db is opened.

#### `account`

Users of the system. There is no self-registration: users are added by an admin through `POST /api/users`. A new user's initial password is its account; users can change it after signing in.

| Key | Value |
| - | - |
| account name in upper case (string, e.g. `ALICE`) | JSON of `model.Account` (`web/backend/model/account.go`) |

| Field | Type | Description |
| - | - | - |
| `account` | string | Login account, same as the key. Stored in upper case; login ignores case. Cannot be renamed. |
| `name` | string | Display name. The system admin defaults to its account. |
| `password` | string | bcrypt hash of the upper-cased password, never the plain text. Passwords are not case-sensitive. |
| `role` | string | `admin` or `default`. |
| `i18n` | string | UI language of the user: `zh-TW` or `en`. |
| `isSystem` | bool | `true` only for the system admin defined by `backend.username` / `backend.password` in the config. |
| `createdAt` | string | RFC 3339 creation time. Days before it are not reported as missed entries. It is the zero time (`0001-01-01T00:00:00Z`) for accounts created before this field existed; those are checked over the whole range. |

Example value:

```json
{
  "account": "ALICE",
  "name": "Alice Wang",
  "password": "$2a$10$...",
  "role": "default",
  "i18n": "en",
  "isSystem": false
}
```

On every startup the backend first renames any account stored in lower or mixed case to upper case, together with the `account` field of its `work` records and todos. If two accounts would end up with the same name, startup stops with an error.

It then syncs the system admin from the config into this bucket:

- It creates the account if it is missing, with `name` set to the account.
- It re-applies the config password and forces `role: admin` and `isSystem: true`. The stored `i18n` is kept.
- If the configured name changed, the previous system admin keeps its record but loses `isSystem`.

#### `category` / `project`

Task categories and projects for the work table, managed by admins on the Work Settings page. Both buckets share the same format. Deleting an option also clears its ID from every `work` record and `todo` in the same transaction; those entries become unassigned.

| Key | Value |
| - | - |
| ID (zero-padded sequence, e.g. `00000000000000000001`) | JSON of `model.WorkOption` (`web/backend/model/work.go`) |

| Field | Type | Description |
| - | - | - |
| `id` | string | Same as the key. IDs sort in creation order. |
| `name` | string | Display name, unique within the bucket (case-insensitive). |
| `active` | bool | Inactive options are hidden from the dropdowns for new entries, but existing records still show their name. |
| `color` | string | Categories only: a palette name (`blue`, `sky`, `teal`, `green`, `lime`, `amber`, `orange`, `red`, `pink`, `purple`). New categories get the least used color and admins can change it. Omitted for projects, and for categories stored before colors existed, which get a stable color derived from their ID. |

```json
{ "id": "00000000000000000001", "name": "Meeting", "active": true, "color": "blue" }
```

#### `work`

Work records. The UI lists them one week (Monday to Sunday) at a time. The newest date is listed first, and records on the same day are ordered latest-created first.

| Key | Value |
| - | - |
| ID (zero-padded sequence) | JSON of `model.WorkRecord` |

| Field | Type | Description |
| - | - | - |
| `id` | string | Same as the key. |
| `account` | string | Owner (`account` bucket key, upper case). Records are kept when the user is deleted. |
| `date` | string | Work date, `YYYY-MM-DD`. |
| `categoryId` | string | `category` bucket key. Required for records; `""` when a todo has none. |
| `description` | string | What was done. May be empty. |
| `hours` | number | Time spent in half-hour steps (0.5, 1, 1.5, …), at most 24. Required for records; `0` when a todo has none. |
| `projectId` | string | `project` bucket key. Optional: `""` when not set. |
| `createdAt` | string | RFC 3339 creation time. |

```json
{
  "id": "00000000000000000001",
  "account": "ALICE",
  "date": "2026-09-29",
  "categoryId": "00000000000000000001",
  "description": "Weekly meeting",
  "hours": 1.5,
  "projectId": "00000000000000000002",
  "createdAt": "2026-09-29T10:02:03.123456+08:00"
}
```

#### `todo`

Todos, stored as JSON of `model.Todo`. The fields are the same as in `work`, but only `date` is required. `categoryId`, `description`, `hours` and `projectId` may be empty.

Completing a todo deletes it from this bucket and creates a `work` record dated on the completion day, in a single transaction. If the todo has no category or hours, the user must fill them in when completing it.

#### `setting`

System-wide settings.

| Key | Value |
| - | - |
| `work` | JSON of `model.WorkSetting` |
| `holidaySync` | `{ "lastSyncedAt": "<RFC 3339 time>" }`, the last successful holiday sync |

| Field | Type | Description |
| - | - | - |
| `allowViewAll` | bool | Whether every user can view everyone's work table (and the missed-entries report). Admins can always view it. Defaults to `false` when the key is missing. |
| `startDate` | string | Optional `YYYY-MM-DD` date when the team started logging. The missed-entries report never checks days before it: it covers the 30 days up to yesterday, or less while the start date is within those 30 days. Omitted when unset. |

```json
{ "allowViewAll": false, "startDate": "2026-09-01" }
```

#### `holiday`

Exceptions to the Monday-Friday work week, used for the hours each week requires (8 per workday). Only exceptions are stored: weekdays off and working weekend days. A government entry and an admin's entry can exist for the same date; the admin's (`manual`) entry wins.

| Key | Value |
| - | - |
| `<date>#<source>` (e.g. `2026-10-09#gov`) | JSON of `model.Holiday` (`web/backend/model/holiday.go`) |

| Field | Type | Description |
| - | - | - |
| `date` | string | `YYYY-MM-DD`. |
| `name` | string | Holiday or makeup workday name, e.g. `中秋節`. |
| `type` | string | `holiday` (a weekday off) or `workday` (a working weekend day). |
| `source` | string | `gov`: synced from the government office calendar. `manual`: set by an admin on the Work Settings page. |

```json
{ "date": "2026-09-25", "name": "中秋節", "type": "holiday", "source": "gov" }
```

The `gov` entries come from the [TaiwanCalendar](https://github.com/ruyut/TaiwanCalendar) data of the DGPA office calendar. The backend syncs this year and the next on startup and every 24 hours, and admins can also press "Sync now". Each sync replaces that year's `gov` entries and never touches `manual` ones. The sync is configured in `config.yaml`:

```yaml
holiday:
  sync: true
  sourceUrl: "https://cdn.jsdelivr.net/gh/ruyut/TaiwanCalendar/data/{year}.json"
```

A failed sync is only logged; the calendar keeps its last data.

#### `apitoken`

Personal API tokens for scripts and Claude skills. Users create and revoke them on the Profile page. Clients send a token as `Authorization: Bearer wt_...`, and it has the same permissions as its account. The token itself is shown once and never stored.

| Key | Value |
| - | - |
| SHA-256 of the token (hex) | JSON of `model.ApiToken` (`web/backend/model/apitoken.go`) |

| Field | Type | Description |
| - | - | - |
| `id` | string | Zero-padded sequence, used to revoke it. |
| `account` | string | Owner (`account` bucket key). Deleting the account deletes its tokens. |
| `name` | string | Label chosen by the user, e.g. `Claude Code skill`. |
| `prefix` | string | First 10 characters of the token (`wt_` + 7), to recognize it in the list. |
| `hash` | string | Same as the key. |
| `createdAt` / `expiresAt` | string | RFC 3339. Lifetimes are 30, 60, 180 or 365 days (default 365). |
| `lastUsedAt` | string | Last authenticated request, updated at most once a minute; the zero time until first use. |

At most 10 tokens per account.

### Backup, restore and reset

Admins open **System** in the sidebar (above their name):

- **Download backup**: `GET /api/system/backup` returns `work-tracker-backup_<time>.zip`. It contains a `manifest.json` (`{"app": "work-tracker", "version": 1, "createdAt": ...}`) and one JSON file per bucket:
  - `account.json`, `category.json`, `project.json`, `work.json`, `todo.json`, `holiday.json` and `apitoken.json` are arrays of the records described above.
  - `setting.json` is `{"work": <setting.work>, "holidaySyncedAt": <time>}`.

  The format does not depend on the database type. Accounts include their password hashes, so keep backups safe.
- **Restore**: `POST /api/system/restore` with the zip in the multipart field `file` (at most 50 MB). It replaces all data in one transaction.
  - A file that is not a Work Tracker backup (wrong manifest, newer version, broken JSON, duplicate accounts) is refused and nothing changes.
  - The config admin is re-applied afterwards, and new IDs continue after the restored ones.
- **Reset**: `POST /api/system/reset` with `{"confirm": "RESET"}`. It deletes all data like a fresh install: only the config admin is left, and the government holidays sync again.

In the UI, restore and reset need `RESTORE` / `RESET` typed in, and they sign everyone out.

## Integration test

`integration-test/` tests the whole app, the way a user would, over HTTP against the Docker image. The Go tests live in `integration-test/goTest/`, a separate module that only uses the standard library.

### Run

Requirements: Docker with Compose, Go and curl on the host.

```bash
make dockertest                 # build the test image alonza0314/work-tracker:test
cd integration-test
./test.sh list                  # list the test cases
./test.sh TestLogin             # run one test case
./test.sh TestAll               # run every test case
```

Every `Test*` case gets a fresh app:

1. `test.sh` starts `docker-compose.yaml` with the test image on port 18888.
2. The case runs with `go test -run '^TestXxx$'`.
3. The compose is removed before the next case, and on failure its log is printed.

`TestAll` ends with a pass/fail summary and exits non-zero if a case failed.

CI runs the cases in the "Integration test" step of the `Build Check` job in `.github/workflows/docker.yaml`, after `make dockertest`. That step has one `./integration-test/test.sh TestXxx` line per case (not `TestAll`), and the first failing case stops it and fails the job. A new `Test*` case needs its line there as well.

- **Config**: the compose mounts `integration-test/config/config.yaml`.
- **Data**: the db goes to `/tmp/wt-integration-test/<TestName>/` and is deleted after the case, so nothing is written to the repo or the host's own `/tmp/wt.db`. The container runs as the host user, so these files stay removable.
- **Time zone**: the host's time zone is mounted into the container, so "today" is the same day for the app and the tests.
- **Holiday calendar**: `TestHolidays` serves a fake one on host port 18889, and the app reaches it through `host.docker.internal`. No internet access is needed.

| Variable | Default | Meaning |
| - | - | - |
| `WT_TEST_IMAGE` | `alonza0314/work-tracker:test` | image under test |
| `WT_TEST_PORT` | `18888` | host port of the app |
| `WT_TEST_ROOT` | `/tmp/wt-integration-test` | where per-case data go |
| `WT_STRESS_USERS` | `10` | `TestStress`: concurrent users |
| `WT_STRESS_RECORDS` | `40` | `TestStress`: records each user writes |

### Test cases

| Case | File | Covers |
| - | - | - |
| `TestApiTokens` | `apitoken_test.go` | creating a token (shown once, `wt_` prefix, 365 days by default); 30/60/180-day lifetimes; bad lifetime or blank name → 400; using it as a bearer credential; same permissions as the account (admin routes → 403 for a default user); the list shows prefix and last use, never the token; only the owner revokes (others → 404) and a revoked token → 401; at most 10 per account (→ 409); deleting the account invalidates its tokens |
| `TestFrontend` | `frontend_test.go` | `/` serves the app; client routes (`/login`, `/work/me`, `/users`) fall back to `index.html`; static files are served |
| `TestHolidays` | `holidays_test.go` | syncing the (fake) government calendar stores weekday days off and weekend makeup days, and ignores weekend holidays; a failed sync → 502 and keeps the calendar; a synced day off makes the week need 32 hours; an admin entry wins over the government one, and deleting it brings the government entry back; validation (bad date, blank name, bad type, bad year → 400); default users can read but not change the calendar (→ 403) |
| `TestLogin` | `login_test.go` | the config admin signs in and the JWT carries sub/name/role/i18n; account and password ignore case; wrong password or unknown account → 401; missing fields → 400; protected routes need a valid JWT or API token (→ 401); logout → 204 |
| `TestMissingEntries` | `missing_test.go` | workdays without a work record are listed per member, most missing first; todos don't count; the system admin is not checked; days before an account was created are skipped; a day off is not a missed day; checking starts at the work start date; default users need `allowViewAll` (→ 403); bad date → 400 |
| `TestProfile` | `profile_test.go` | reading your profile; changing the UI language returns a token carrying it (unsupported language → 400); changing the password (wrong old password → 403; the new one ignores case); the system admin's password comes from the config (→ 403) |
| `TestBackupRestore` | `system_test.go` | downloading a backup zip (manifest plus one JSON file per table); only admins back up or restore; restoring brings back users, records and API tokens, and new IDs continue after the restored ones; non-zip, another app's zip or a missing file field → 400 without changing data |
| `TestReset` | `system_test.go` | a reset needs `{"confirm": "RESET"}` (→ 400) and an admin (→ 403); it leaves only the config admin and old tokens stop working |
| `TestStress` | `stress_test.go` | load: every user writes records at once while others read everyone's table (no failed request, no lost write, every read is a consistent snapshot whose count never goes back); concurrent updates and deletes; 20 racing completions of one todo create exactly one record (the rest → 404); concurrent logins and API token calls; a backup taken during writes is a valid snapshot. Logs requests per second and p50/p95/max latency. Heavier run: `WT_STRESS_USERS=50 WT_STRESS_RECORDS=400 ./test.sh TestStress` |
| `TestTodos` | `todos_test.go` | a todo needs only a date (the description may be blank); todos are listed by date; completing one creates a work record on the given date; a todo without category or hours needs them to complete (→ 400, todo kept); update and delete; other people's todos → 404 |
| `TestUserManagement` | `users_test.go` | creating a user stores the account in upper case with the account as initial password; duplicates → 409, blank name or unknown role → 400; the list is sorted by account; partial updates and password reset; the system admin can't be changed or deleted and admins can't delete themselves (→ 403); default users can't manage users (→ 403); a deleted user's token → 401 |
| `TestEveryonesWork` | `viewall_test.go` | default users need `allowViewAll` for everyone's records, members and missed entries (→ 403); admins see everyone's records with totals; filters by member (any case), category and project; quarter ranges; the members list; switching `allowViewAll` on and off |
| `TestWeekSummary` | `weeksummary_test.go` | a plain week needs 40 hours, with per-day logged/required hours; a Friday off → 32 hours; a makeup Saturday adds 8; remaining hours never go below zero; bad date → 400 |
| `TestWorkOptions` | `workoptions_test.go` | new categories get the least used colors; names are unique ignoring case and not blank; rename, recolor and deactivate (unknown color → 400, project color → 400, unknown ID → 404); everyone reads the options, only admins change them; deleting a category or project clears it from records and todos; work settings update only the given fields |
| `TestWorkRecords` | `workrecords_test.go` | validation (category and hours required, half-hour steps, dates, unknown or inactive options); the project and the description are optional; multi-line descriptions; listing a range newest first with totals; a range is required and must not be reversed; update and delete; a record keeps a category deactivated after it was chosen; other people's records → 404 |

## Claude skills

`skills/` holds two Claude Code skills that use the API with a personal API token:

| Skill | Use |
| - | - |
| `wt-log-work` | Log work from a plain description ("幫我記今天開發 API 3 小時"). It shows the records for confirmation before saving. |
| `wt-daily-report` | Summarize the previous workday's records for the daily meeting. Weekends and holidays are skipped using the Work Tracker calendar. |

### Install

1. Start Claude Code in `skills/` (or in the repo and open a file under `skills/`) and run `/integrate`. This built-in skill (`skills/.claude/skills/integrate/`) copies every `skills/wt-*` into `~/.claude/skills/`; run it again after pulling updates.
2. Reload Claude Code.
3. Create an API token on the site's Profile page, then save the site URL and the token:

   ```bash
   python3 ~/.claude/skills/wt-log-work/scripts/wt.py setup
   ```

   Inside Claude Code, type `! python3 ~/.claude/skills/wt-log-work/scripts/wt.py setup` so the token never goes into the chat. The token is checked before it is saved to `~/.wt/config.json` (mode 600). Run `setup` again when the token expires.

### Develop

- Both skills share `skills/wt-log-work/scripts/wt.py` (Python 3, standard library only); `wt-daily-report/scripts/wt.py` is a symlink to it, which `/integrate` copies as a real file. Run `python3 skills/wt-log-work/scripts/wt.py --help` for its commands.
- Tests: `python3 -m unittest discover -s skills/tests -v` (a fake API server; CI runs it in the "Skills" workflow).
