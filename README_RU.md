# nanoforum-go

Минималистичный форум на Go с авторизацией через bcrypt и хранением в SQLite. Внешняя СУБД не требуется.

Этот проект был создан с использованием комбинации помощи opencode ИИ, человеческой разработки и экспериментальных чат-взаимодействий - выступая одновременно в роли эксперимента в области ИИ- assisted development и оптимизированного инструмента для небольших сайтов.

## Возможности

- Регистрация пользователей через консольный скрипт
- Авторизация через сессионные куки
- Хэширование паролей bcrypt
- SQLite (чистый Go-драйвер, без CGo)
- Готовность к Docker с персистентным volume
- Скрипты бэкапа/восстановления базы

## Структура проекта
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

## Требования

- Go 1.25+
- [Task](https://taskfile.dev) (альтернатива make)
- [golangci-lint](https://golangci-lint.run) v2.12+
- Docker и Docker Compose (для запуска в контейнере)

## Локальная сборка и запуск

```bash
go mod tidy
task vet             # go vet ./...
task lint            # golangci-lint run ./...
task check           # vet + lint

task createuser      # интерактивная регистрация пользователя (использует ./forum.db)
task run             # собрать и запустить сервер на :8080
```

Без Task можно использовать обычные команды Go:

```bash
go run ./cmd/createuser
go run ./cmd/forum
```

## Конфигурирование

Приложение читает одну переменную окружения:

| Переменная | По умолчанию   | Описание                                 |
|------------|----------------|-------------------------------------------|
| `DB_PATH`  | `./forum.db`   | Путь к файлу базы данных SQLite           |

Локально задавать её обычно не нужно — значение по умолчанию подходит и серверу, и CLI-утилите, если оба указывают на один файл.

## Запуск в Docker

```bash
task docker-build
task docker-createuser    # зарегистрировать пользователя внутри контейнера
task docker-up            # форум доступен на хост-порту 8084
task docker-logs
task docker-down
```

Внутри контейнера сервер слушает `:8080`; `docker-compose.yml` публикует его на хост-порт `8084`. База лежит на именованном volume `forum-data`, смонтированном в `/app/data`, поэтому данные не теряются при пересоздании контейнера.

## Просмотр и правка данных в базе

Через sqlite3 CLI:

```bash
# локально
sqlite3 ./forum.db

# внутри контейнера
docker exec -it nanoforum sqlite3 /app/data/forum.db
```

Полезные команды:

```sql
.tables
.schema users
SELECT id, username FROM users;
UPDATE posts SET content = 'исправленный текст' WHERE id = 5;
DELETE FROM users WHERE username = 'old_login';
```

Для GUI-варианта используйте [DB Browser for SQLite](https://sqlitebrowser.org/) — откройте `.db`-файл, редактируйте строки во вкладке Browse Data и нажмите Write Changes.

Пароли нельзя редактировать прямо через SQL, так как хранится только bcrypt-хэш; для смены пароля пересоздайте пользователя через `task createuser`.

## Бэкап и восстановление

```bash
task db-backup                                        # создаёт ./backups/forum_<timestamp>.db
task db-restore -- ./backups/forum_20260730_150000.db
```

## Nginx reverse proxy

Пример конфига для проксирования контейнера через домен:

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

## Публикация Docker-образа в GitHub Container Registry

В репозитории есть `.github/workflows/docker-publish.yml`, который собирает образ и пушит его в `ghcr.io/gritsulyak/nanoforum-go` при каждом push в `main` и при создании тега версии, авторизуясь встроенным `GITHUB_TOKEN` [web:94][web:99].

Ручная сборка и публикация (если хочется сделать это локально):

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u gritsulyak --password-stdin

docker build -f docker/Dockerfile -t ghcr.io/gritsulyak/nanoforum-go:latest .
docker push ghcr.io/gritsulyak/nanoforum-go:latest
```

### Скачивание и запуск опубликованного образа

```bash
docker pull ghcr.io/gritsulyak/nanoforum-go:latest

docker run -d \
  --name nanoforum \
  -p 8084:8080 \
  -v forum-data:/app/data \
  -e DB_PATH=/app/data/forum.db \
  ghcr.io/gritsulyak/nanoforum-go:latest
```

Или через продакшен-compose файл:

```bash
docker compose -f docker/docker-compose-prod.yml up -d
```

## Лицензия

MIT
