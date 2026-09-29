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
| account name (string, e.g. `alice`) | JSON of `model.Account` (`web/backend/model/account.go`) |

| Field | Type | Description |
| - | - | - |
| `account` | string | Login account, same as the key. Cannot be renamed. |
| `name` | string | Display name. The system admin defaults to its account. |
| `password` | string | bcrypt hash of the password, never the plain text. |
| `role` | string | `admin` or `default`. |
| `i18n` | string | UI language of the user: `zh-TW` or `en`. |
| `isSystem` | bool | `true` only for the system admin defined by `backend.username` / `backend.password` in the config. |

Example value:

```json
{
  "account": "alice",
  "name": "Alice Wang",
  "password": "$2a$10$...",
  "role": "default",
  "i18n": "en",
  "isSystem": false
}
```

On every startup the backend syncs the system admin from the config into this bucket:

- It creates the account if it is missing, with `name` set to the account.
- It re-applies the config password and forces `role: admin` and `isSystem: true`. The stored `i18n` is kept.
- If the configured name changed, the previous system admin keeps its record but loses `isSystem`.
