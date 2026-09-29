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
| `description` | string | What was done. |
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

Todos, stored as JSON of `model.Todo`. The fields are the same as in `work`, but only `date` and `description` are required. `categoryId`, `hours` and `projectId` may be empty.

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
