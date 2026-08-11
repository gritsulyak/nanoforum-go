# AGENTS.md

Single-module Go app (`github.com/gritsulyak/nanoforum-go`, go 1.25): a minimal SQLite-backed forum with bcrypt auth. Entrypoints: `cmd/forum` (web server), `cmd/createuser` (interactive CLI user registration). Packages in `internal/`: `auth`, `config`, `db`, `handlers`, `models`, `repository`. One template: `web/templates/index.html` (client-side RU/EN i18n via `data-i18n`).

## Build / verify

- CI gate is `go vet ./...` + `golangci-lint run ./...` (equivalent to `task check`). Tests cover `internal/` at 100% (`go test -cover ./internal/...`); run `go test ./...`.
- `golangci-lint` v2 is required (`golangci-lint version`, e.g. 2.12.2).
- README's `task run` and `task createuser` do **not exist** in `Taskfile.yml`. Real commands: `go run ./cmd/forum` (serves :8080) and `go run ./cmd/createuser`.
- The server loads `web/templates/index.html` via a relative path, so `go run ./cmd/forum` must be run from the repo root.

## Gotchas

- **Lint is clean.** `.golangci.yml` (golangci-lint v2) enables extra linters: bodyclose, contextcheck, errname, errorlint, gosec, noctx, unconvert, unparam. `golangci-lint run ./...` passes with 0 issues (gosec excludes G404; gosec/noctx are skipped for tests).
- **Docker serves under `/forum`.** Both `docker/docker-compose.yml` and `docker-compose-prod.yml` set `BASE_PATH=/forum` (and compose also sets `PAGE_SIZE=15`), so the container is reachable at `http://localhost:8084/forum`, not `:8084/`. The container itself listens on :8080.
- Env config (README documents only `DB_PATH`): `DB_PATH` (default `./forum.db`), `BASE_PATH` (URL prefix, empty by default), `PAGE_SIZE` (default 10), `PPROF` (enables `/debug/pprof/*`, off by default), `DEBUG` (per-request latency logs for Forum/Login, off by default), `POSTS_CACHE_TTL` (posts list cache TTL as a Go duration, default `1s`).

## Architecture notes

- SQLite via `modernc.org/sqlite` (pure Go, no CGo — builds fine with `CGO_ENABLED=0`). `internal/db.New` creates the schema on startup (users, posts); no migration files. **Posts list is cached** per (limit, offset) with a TTL (`POSTS_CACHE_TTL`, default 1s) in `internal/repository.PostRepo`; the cache is invalidated on every successful `Create`. Profiling once showed `modernc.org/sqlite` internals + syscall/GC dominated CPU under load (~8k RPS cap); the cache raised sustained posts throughput to 20k RPS at p95 ~250µs (verified with k6).
- k6 load tests live in `k6/` (`login.js`, `posts.js`, `post-create.js`). `bash k6/setup.sh` creates the `loadtest` user inside the container; `bash k6/run-load.sh` (or `task k6-load`) runs them against `http://localhost:8084/forum`. `k6/profile.sh` (or `task k6-profile`) runs k6 and captures pprof profiles into `k6/profiles/` — requires the server started with `PPROF=1`; analyze with `go tool pprof -http :9999 k6/profiles/cpu.pprof`.
- Session auth is just a `session_user` cookie (24h, HttpOnly, SameSite=Lax); no CSRF protection. Passwords are bcrypt; users can only be created via the CLI, UI registration is closed.
- `docker-publish.yml` builds to `ghcr.io/gritsulyak/nanoforum-go` on `v*` tags (branch trigger is commented out). Backup/restore: `task db-backup` / `task db-restore -- <path>`.
