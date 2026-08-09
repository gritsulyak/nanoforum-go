# AGENTS.md

Single-module Go app (`github.com/gritsulyak/nanoforum-go`, go 1.25): a minimal SQLite-backed forum with bcrypt auth. Entrypoints: `cmd/forum` (web server), `cmd/createuser` (interactive CLI user registration). Packages in `internal/`: `auth`, `config`, `db`, `handlers`, `models`, `repository`. One template: `web/templates/index.html` (client-side RU/EN i18n via `data-i18n`).

## Build / verify

- CI gate is `go vet ./...` + `golangci-lint run ./...` (equivalent to `task check`). Tests cover `internal/` at 100% (`go test -cover ./internal/...`); run `go test ./...`.
- `golangci-lint` v2 is required (`golangci-lint version`).
- README's `task run` and `task createuser` do **not exist** in `Taskfile.yml`. Real commands: `go run ./cmd/forum` (serves :8080) and `go run ./cmd/createuser`.
- The server loads `web/templates/index.html` via a relative path, so `go run ./cmd/forum` must be run from the repo root.

## Gotchas

- **Lint config is active.** `.golangci.yml` (golangci-lint v2) enables extra linters: bodyclose, contextcheck, errname, errorlint, gosec, noctx, unconvert, unparam. The repo currently has violations (e.g. gosec G114/G124 on `cmd/forum/main.go` and `internal/auth/auth.go`, noctx on `internal/db` and `internal/repository`) — `golangci-lint run ./...` is not clean.
- **Docker serves under `/forum`.** Both `docker/docker-compose.yml` and `docker-compose-prod.yml` set `BASE_PATH=/forum` (and compose also sets `PAGE_SIZE=15`), so the container is reachable at `http://localhost:8084/forum`, not `:8084/`. The container itself listens on :8080.
- Env config (README documents only `DB_PATH`): `DB_PATH` (default `./forum.db`), `BASE_PATH` (URL prefix, empty by default), `PAGE_SIZE` (default 10).

## Architecture notes

- SQLite via `modernc.org/sqlite` (pure Go, no CGo — builds fine with `CGO_ENABLED=0`). `internal/db.New` creates the schema on startup (users, posts); no migration files.
- Session auth is just a `session_user` cookie (24h, HttpOnly, SameSite=Lax); no CSRF protection. Passwords are bcrypt; users can only be created via the CLI, UI registration is closed.
- `docker-publish.yml` builds to `ghcr.io/gritsulyak/nanoforum-go` on `v*` tags (branch trigger is commented out). Backup/restore: `task db-backup` / `task db-restore -- <path>`.
