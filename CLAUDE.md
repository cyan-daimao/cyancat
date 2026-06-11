# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Dev mode (hot-reload backend + frontend webview)
wails dev

# Production build (binary → build/bin/cyancat)
wails build

# Frontend only (usually called by wails, rarely standalone)
cd frontend && npm run build      # tsc && vite build → frontend/dist

# Go side
go build ./...
go vet ./...
go test ./...                     # no *_test.go files yet
go mod tidy

# Regenerate Wails JS/TS bindings after changing Go API signatures
wails generate module
```

Runtime environment:
- `CYANCAT_MASTER_KEY` (64-char hex) or `~/.cyancat/master.key` (32 raw bytes) — AES master key for password encryption. Dev fallback is hard-coded with a warning.
- Local SQLite at `~/.cyancat/cyancat.db` (auto-created).

## Architecture

Wails v2 desktop app: Go backend exposes struct methods that the React/Vite frontend calls as async JS functions. Follows **DDBD 四层架构** (`adapter → application → domain → infra`).

### Layers

| Layer | Path | Role |
|-------|------|------|
| **adapter** | `internal/adapter/` | Wails-bound API structs (`app.go` + `http/*_api.go`), DTO/Request + conversion funcs (`dto/*_dto.go`) |
| **application** | `internal/application/<biz>service/` | Service interfaces + impls, Cmd/Query/BO objects, orchestration logic |
| **domain** | `internal/domain/<biz>/` | Rich domain entities with behavior methods, Repository interfaces |
| **infra** | `internal/infra/` | Repository impls (GORM/SQLite), driver abstraction, session manager, eventbus, crypto, keychain, logger |

Three vertical slices currently: `connectionservice`, `queryservice`, `schemaservice`.

### Conversion chains (explicit `ToXxx` funcs at every boundary, no reflection)

```
Read:  DO → Domain → BO → DTO
Write: Request → Cmd → Domain → Repository → DO
```

- `adapter/dto`: Request↔Cmd, BO→DTO
- `application/<biz>service`: Cmd→Domain, Domain→BO
- `infra/db/<biz>repo`: DO↔Domain

### Key architectural components

- **`infra/driver`** — Abstracts MySQL/Postgres behind `Driver`/`Conn`/`Dialect`/`Inspector`/`RowStream` interfaces. Global `Registry` keyed by `DriverType`. Concrete drivers in `infra/driver/mysql/` and `infra/driver/postgres/` registered from `main.go`.
- **`infra/session`** — Runtime map of `connID → driver.Conn` (opened long-lived connections). Distinct from `domain/connection` which is the *configuration* entity. Deliberately in infra, not domain.
- **`infra/eventbus`** — Wraps Wails `runtime.EventsEmit`. `Init(ctx)` called in `OnStartup` (ctx not available before then). Events: `query:rows`, `query:done`, `query:error`, `connection:state`.
- **`infra/db/connectionrepo`** — Passwords encrypted via `crypto.Encrypt(masterKey)` before storage in `password_encrypted` column. Soft delete via `WHERE deleted_at IS NULL`.
- **`infra/api`** — `Response[T]`, `Page[T]`, `Code`. Adapters use `api.Success`/`api.Fail`/`api.Error`.

### Wiring (main.go, in order)

1. Logger init → 2. Register drivers (`mysql.New()`, `postgres.New()`) → 3. SQLite init → 4. AutoMigrate per repo → 5. Keychain init → 6. Manual DI: repos → session manager → eventbus → services → `adapter.NewApp(...)` → 7. `wails.Run` with `OnStartup`/`OnShutdown` hooks.

### Frontend (React 18 + Vite + TS + Tailwind + shadcn/ui)

- `src/lib/api/*.ts` — Wraps Wails bindings (`wailsjs/go/http/*`), unwraps `{code, message, data}`, throws on non-200.
- `src/stores/*.ts` — Zustand stores (`connection`, `query`, `schema`).
- Components by feature: `connection/`, `data-table/`, `object-tree/`, `sql-editor/`, `layout/`, `ui/` (shadcn primitives).
- Path alias `@` → `frontend/src`.

## DDBD Project-Specific Conventions

1. **No Gin / no HTTP routes.** Wails replaces HTTP transport. `adapter/http` is a naming convention; methods are bound by reflection, not routed. The DDBD rule "URI must start with `/api/`" doesn't apply.
2. **Only `connection` has a domain package.** `query` and `schema` are driver-execution concerns with no persisted aggregate. Query *history* is an infra read model with `HistoryRepository` interface declared in `application/queryservice`, not domain. Don't force-create empty domain packages.
3. **`session.Manager` is infra**, not domain — it's a runtime connection pool.
4. **`connectionrepo.Update` uses `map[string]interface{}`** Updates call so zero-valued fields (e.g., `ssl=false`) are written — intentional, don't switch to struct-based Updates.
5. **Wails-generated bindings in `frontend/wailsjs/` are auto-generated** — do not edit; they regenerate via `wails generate module`.
6. **All `To*BOs`/`To*DTOs` must return empty slices** (`make([]*X, 0)`) not `nil` — nil serializes to JSON `null` which crashes frontend `.map()` calls.
7. **Field naming between Go and TS**: backend uses `durationMs` (json tag), frontend type must match exactly. Check `wailsjs/go/models.ts` after `wails generate module` for actual generated names.

## Reference Docs

- `doc/01-product-overview.md` — Product vision, V1.0/V1.5/V2.0 feature planning
- `doc/02-tech-architecture.md` — Original architecture proposal (directory layout slightly outdated; actual follows DDBD as described above)
