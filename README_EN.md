# nanoforum-go

A minimal forum application in Go with bcrypt-based authentication and SQLite storage. No external database server required.

This project was built using a combination of opencode AI assistance, human development, and experimental chat interactions - serving as both an experiment in AI-assisted development and an optimized tool for small sites.

## Features

- User registration via CLI script
- Cookie-based session authentication
- Bcrypt password hashing
- SQLite storage (pure Go driver, no CGo)
- Docker-ready with persistent volume
- Backup/restore scripts for the database

## Project structure
```
nanoforum-go/
├── cmd/
│ ├── forum/ # main web server
│ └── createuser/ # CLI user registration
├── internal/
│ ├── auth/ # bcrypt + session cookies
│ ├── config/ # env-based configuration
│ ├── db/ # SQLite connection & migrations
│ ├── handlers/ # HTTP handlers
│ ├── models/ # data structures
│ └── repository/ # DB access layer
├── web/templates/ # HTML templates
├── docker/
│ ├── Dockerfile
│ ├── docker-compose.yml
│ ├── docker-compose-prod.yml
│ └── scripts/
│ ├── backup.sh
│ └── restore.sh
├── .github/workflows/
│ ├── ci.yml
│ └── docker-publish.yml
├── Taskfile.yml
├── .golangci.yml
└── go.mod
```

## Requirements

- Go 1.25+
- [Task](https://taskfile.dev) (make alternative)
- [golangci-lint](https://golangci-lint.run) v2.12+
- Docker & Docker Compose (for containerized run)

## Local build & run

```bash
go mod tidy
task vet             # go vet ./...
task lint            # golangci-lint run ./...
task check           # vet + lint

task createuser      # register a new user interactively (uses ./forum.db)
task run             # build and start the server on :8080
```

Without Task, plain Go commands work too:

```bash
go run ./cmd/createuser
go run ./cmd/forum
```

## Configuration

The application reads a single environment variable:

| Variable  | Default        | Description                                   |
|-----------|----------------|------------------------------------------------|
| `DB_PATH` | `./forum.db`   | Path to the SQLite database file               |

Locally you usually don't need to set it — the default works for both the server and the CLI tool, as long as both point to the same file.

## Running in Docker

```bash
task docker-build
task docker-createuser    # register a user inside the container
task docker-up            # forum exposed on host port 8084
task docker-logs
task docker-down
```

Internally the container listens on `:8080`; `docker-compose.yml` maps it to host port `8084`. The database lives on a named volume `forum-data` mounted at `/app/data`, so data survives container recreation.

## Accessing and editing the database

Using the sqlite3 CLI:

```bash
# locally
sqlite3 ./forum.db

# inside the container
docker exec -it nanoforum sqlite3 /app/data/forum.db
```

Useful commands:

```sql
.tables
.schema users
SELECT id, username FROM users;
UPDATE posts SET content = 'fixed text' WHERE id = 5;
DELETE FROM users WHERE username = 'old_login';
```

For a GUI alternative, use [DB Browser for SQLite](https://sqlitebrowser.org/) — open the `.db` file, edit rows in the Browse Data tab, and click Write Changes.

Passwords cannot be edited directly in SQL since only bcrypt hashes are stored; recreate the user via `task createuser` instead.

## Backup & restore

```bash
task db-backup                                        # creates ./backups/forum_<timestamp>.db
task db-restore -- ./backups/forum_20260730_150000.db
```

## Nginx reverse proxy

Example config to expose the container behind Nginx on a domain:

```nginx
server {
    listen 80;
    server_name your-domain.example;

    location / {
        proxy_pass http://127.0.0.1:8084;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
sudo ln -s /etc/nginx/sites-available/nanoforum /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Publishing the Docker image to GitHub Container Registry

The repository includes `.github/workflows/docker-publish.yml`, which builds the image and pushes it to `ghcr.io/gritsulyak/nanoforum-go` on every push to `main` and on version tags, authenticating with the built-in `GITHUB_TOKEN` [web:94][web:99].

Manual build & push (if you prefer to do it locally):

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u gritsulyak --password-stdin

docker build -f docker/Dockerfile -t ghcr.io/gritsulyak/nanoforum-go:latest .
docker push ghcr.io/gritsulyak/nanoforum-go:latest
```

### Pulling and running the published image

```bash
docker pull ghcr.io/gritsulyak/nanoforum-go:latest

docker run -d \
  --name nanoforum \
  -p 8084:8080 \
  -v forum-data:/app/data \
  -e DB_PATH=/app/data/forum.db \
  ghcr.io/gritsulyak/nanoforum-go:latest
```

Or with the production compose file:

```bash
docker compose -f docker/docker-compose-prod.yml up -d
```

## License

MIT
