# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Work Tracker (`wt`): a department web app where employees log their work and managers view everyone's work and hours. Go (gin) backend in `web/backend`, React + TypeScript (Vite) frontend in `web/frontend`. In production the Go binary serves both the API and the built frontend.

## Commands

Run from the repo root unless noted.

| Task | Command |
| - | - |
| Build backend + frontend into `build/` | `make` (or `make backend` / `make frontend`) |
| Run (uses `./config.yaml`) | `make run` → `./build/wt -c config.yaml`, listens on `:8888` |
| Go lint (golangci-lint v2, config `web/backend/.golangci.yaml`) | `make lint` |
| Go tests | `cd web/backend && go test ./...` |
| Single Go test | `cd web/backend && go test ./internal/processor -run TestName -v` |
| Regenerate frontend API client (needs Docker) | `make openapi` |
| Frontend dev server | `cd web/frontend && yarn dev` (talks to the backend at `<host>:8888`, or `VITE_API_BASE_URL`) |
| Frontend type-check + build | `cd web/frontend && yarn build` |
| Docker image / clean build + db | `make docker` / `make clean` |

- `yarn lint` does not run at the moment: `eslint.config.js` uses `reactHooks.configs['recommended-latest']`, and eslint-plugin-react-hooks v7 no longer supports that key.
- There are no Go tests yet. CI (`.github/workflows`) runs `go build`, `go test`, `make`, golangci-lint and `yarn build`.
- Commit messages must follow Conventional Commits (`feat:`, `fix:`, `style:`, `chore:`…). CI checks this.
- `DB_PATH` in `Makefile` must stay in sync with `backend.db.path` in `config.yaml`. The Docker setup uses `docker/config.yaml` instead.

## API development workflow

Every new endpoint goes through all of these steps. The frontend never calls the API with hand-written axios/fetch.

1. **Backend**: add the model, processor method, handler and route (see the backend conventions below).
2. **`web/openapi.yaml`**: add the path. Its `operationId` becomes the TS method name, and request/response schemas go under `components/schemas`.
   - **Endpoints behind the auth middleware must declare `security: [ { bearerAuth: [] } ]`.** The generated client only attaches the `Authorization: Bearer` header to operations that declare it.
   - `web/work-tracker.postman_collection.json` is not kept in sync (its login body still uses `username`). If the user works from Postman, the `postman-to-openapi` skill converts a collection into openapi.yaml.
3. **`make openapi`**: runs `web/openapi-generator-docker.sh` (the openapi-generator `typescript-axios` generator) and rewrites `web/frontend/src/api/`.
   - Never hand-edit files in `src/api/`; they are overwritten on the next generation.
4. **Frontend**: call the generated method through the shared client:
   ```ts
   import { api, extractErrorMessage } from '../../apiClient'
   const response = await api.login({ username, password })   // typed request/response
   ...
   catch (error: unknown) { addError(extractErrorMessage(error, t('login.failed'))) }
   ```
   `src/apiClient.ts` handles the rest:
   - It builds a single `DefaultApi` with `basePath` = `VITE_API_BASE_URL` or `<protocol>//<host>:8888`. The `servers` URL in openapi.yaml is ignored.
   - It supplies the bearer token from `localStorage.token`.
   - On any 401 it clears the token and hard-redirects to `/login`.

## Backend architecture (`web/backend`, Go module name `backend`)

Request flow: `main.go` → `cmd/wt.go` (cobra; loads YAML config via `util.LoadFromYaml`, builds the logger, `internal.NewBackend`, `Start`, waits for SIGINT/SIGTERM, `Stop`) → gin router → handler → `Processor` → `SystemContext` → DB.

- **`internal/backend.go`**: defines the `backend` struct, which embeds `processor.Processor` and `*logger.BackendLogger`, so handlers call `b.Processor.X(...)` and `b.AccLog...`. It also contains:
  - **`addServices`**: registers routes on three groups:
    - `apiGroup` (`/api`, from `constant.API_PREFIX`): public routes.
    - `authGroup`: runs `addAuthMiddleware`, which validates the JWT (via `github.com/free-ran-ue/util`), loads the account named by the `sub` claim from the DB through `Processor.Authenticate`, and stores it in the gin context. Handlers read it with `currentAccount(c)`.
    - `adminGroup`: nested in `authGroup`, and additionally requires `role == admin`.
  - **`NoRoute(returnPages())`**: serves the SPA. It returns files from `frontendFilePath` and falls back to `index.html`.
- **`internal/api_<domain>.go`**: HTTP layer for one domain.
  - `func (b *backend) get<Domain>Routes() util.Routes` returns `util.Routes{{Name, Method, Pattern, HandlerFunc}}`.
  - Every `HandlerFunc` is wrapped: `withLogging("Login", b.AccLog, b.handleLogin)`. `withLogging` (in `api.go`) logs once after the handler, with the level derived from the status code.
  - Handlers are named `handle<Action>`.
- **`internal/processor/`**: business logic, one file per domain.
  - Methods take `*model.Request<Action>` and return `(*model.Response<Action>, *model.ErrorDetail)`.
  - `ErrorDetail{HttpStatus, Detail}` carries the HTTP status up to the handler, so processors decide status codes. They never touch gin.
- **`internal/context/`**: `SystemContext` → `dbContext` → `DbIf`.
  - `DbIf` (`db.go`) is the storage abstraction. It is composed of per-domain interfaces such as `AccountDbIf`, plus `Release()`. Implementations return the package's sentinel errors (`ErrAccountNotFound`, `ErrAccountExists`).
  - `newDb` switches on `db.type`; only `"bbolt"` exists (`dbBbolt.go`, one bucket per domain, JSON values).
  - `dbContext` embeds `DbIf`, so its methods are promoted up to `Processor` (`p.GetAccount(...)`).
  - Adding storage for a new domain: add an `XxxDbIf`, embed it in `DbIf`, implement it on `bboltDb` (create the bucket in `newBboltDb`), and test it in `dbBbolt_test.go`. Also document the bucket's key and value fields in the README's `## DB` section.
- **`model/`**: request/response DTOs per domain (`model/<domain>.go`) plus `error.go`.
- **`logger/`**: `BackendLogger` holds one tagged logger per area (`CfgLog`, `AccLog`, `BckLog`, `ProcLog`, `GinLog`, `CtxLog`, `DbLog`). Tags live in `logger/tag.go`. A new domain gets a new tag constant plus a field wired up in `NewBackendLogger`.
- **Shutdown chain**: `backend.Stop` → `Processor.Release` → `SystemContext.Release` → `dbContext.release` → `DbIf.Release`. New resources hook into this chain.
- **`config/config.go`**: mirrors `config.yaml`. Structs are suffixed `IE` and fields carry `yaml:"..." valid:"required"` tags.

### Accounts

- **Stored users**: users live in the `account` bucket as `model.Account` (`account`, `name`, bcrypt `password`, `role` `admin`/`default`, `i18n` `zh-TW`/`en`, `isSystem`). There is no self-registration; admins create users through `/api/users`. A new user's initial password is its account, and the user can change it via `PUT /api/me/password`.
- **System admin**: `backend.username/password` in the config is the system admin. On startup, `Processor.InitSystemAdmin` creates or updates it in the DB with `isSystem=true`, applies the config password and keeps its i18n. It clears the flag on a previous system admin if the configured name changed.
  - The system admin cannot be updated or deleted via the API and cannot change its own password.
  - Admins cannot delete themselves.
- **JWT**: tokens carry `sub` (account), `name`, `role` and `i18n` claims. The frontend reads the display name, UI language and role from them. `PUT /api/me` returns a fresh token after an i18n change.
  - Authorization always uses the DB account loaded by the middleware, never the claims.

### Backend coding style (follow it exactly)

- **Constructors**: `NewXxx(ie *XxxIE) *Xxx`, where `XxxIE` is an exported input struct (unexported `xxxIE` for package-internal types). Config sections also use the `IE` suffix.
- **Struct layout**: fields are grouped by concern with blank lines between groups. The `*logger.BackendLogger` embed always comes last. Struct literals repeat the same grouping and blank lines.
- **Composition**: embed dependencies (`processor.Processor`, `*context.SystemContext`, `*logger.BackendLogger`) instead of naming fields, so their methods and loggers are promoted.
- **Handler shape**, from `handleLogin`:
  1. `c.ShouldBindJSON(&req)`. On error, `b.<Tag>Log.Warnf(... c.ClientIP() ...)` and `c.JSON(400, model.Response<Action>{Message: "Invalid request"})`.
  2. `response, errDetail := b.Processor.<Action>(&req)`. On error, log a warning and return `c.JSON(errDetail.HttpStatus, model.Response<Action>{Message: errDetail.Detail})`.
  3. Otherwise `c.JSON(http.StatusOK, response)`. Use `c.Status(http.StatusNoContent)` for empty responses.
- **Models**: `Request<Action>` / `Response<Action>` with `json` tags. Required inputs get `binding:"required"`. Responses carry `Message string \`json:"message"\``, plus optional data with `omitempty`. Keep the models identical to the openapi.yaml schemas.
- **Naming**: constants are `UPPER_SNAKE` in `constant/`, and log tags are short uppercase (`ACC`, `PROC`). Staticcheck ST1003 is disabled, so names like `HttpStatus` and `API_PREFIX` are intentional.
- **Log levels**: `Debugf` for processor tracing, `Warnf` for client errors, `Errorf` for server failures. `Infoln` brackets lifecycle steps ("Release X..." / "X released").
- **Imports**: stdlib and local `backend/...` imports share one alphabetically sorted group, third-party packages go in a second group. Fixed aliases: `loggergo`, `loggergoModel`, `loggergoUtil`, `sysctx` (for `backend/internal/context` inside `internal`).
- **Errors**: `fmt.Errorf("failed to ...: %v", err)` for internal errors. Only `ErrorDetail` crosses from processor to handler.

## Frontend architecture (`web/frontend/src`)

- **Routing** (`App.tsx`): `/login` is public. Authenticated pages are nested routes under `<RequireAuth><AppLayout/></RequireAuth>`, where `components/layout/AppLayout.tsx` provides the sidebar, the topbar and an `<Outlet/>`.
  - Admin-only pages are also wrapped in `<RequireAdmin>`.
  - Sidebar entries come from `components/layout/navigation.ts` (`adminOnly` hides an entry from default users), which also supplies the topbar title.
- **Auth** (`auth/`): `AuthProvider` holds the session parsed from the JWT in `localStorage.token` (`account`, `name`, `role`, `i18n`, expiry). `useAuth()` returns `{ session, isAdmin, signIn(token), signOut(), changeLocale(locale) }`.
  - `signIn` also switches the UI to the token's i18n.
  - `changeLocale` persists the language through `api.updateMe` when signed in.
- **i18n** (`i18n/`): a custom provider with no library. `useI18n()` returns `{ t, locale, setLocale }`.
  - Every UI string is a key in `i18n/locales/en.ts`, the source of truth. `zh-TW.ts` is typed `Messages`, so a missing key fails `tsc`. Always add both languages.
  - `t('key', { name })` interpolates `{name}`. The chosen locale is persisted in `localStorage.locale`, and the default is `zh-TW`.
- **Styling**: CSS Modules (`*.module.css`) next to each component. All colors, spacing, radius and shadows come from CSS variables in `src/index.css`, whose `--brand-*` variables hold the brand palette. Don't hard-code hex values in components.
- **Icons**: `lucide-react`. The brand mark is `components/logo/Logo.tsx` (`public/wt-favicon.jpg`).
- **Shared components**: `components/`. Modal renders a `<form>` (Enter submits). Also Panel, Field/TextInput/SelectInput, Badge, Button, and notifications.
  - Surface errors and successes with `useNotifications()` + `<NotificationContainer/>`.
  - Map API failures to localized messages with `errorStatus(error)` from `apiClient.ts`; backend messages are English only.
