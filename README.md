# nanoforum-go
Nano forum with every feature exterminated.

Lightweight forum with authentication, written in Go and SQLite.

This project was built using a combination of opencode AI assistance, human development, and experimental chat interactions - serving as both an experiment in AI-assisted development and an optimized tool for small sites.

- 🇬🇧 [English documentation](README_EN.md)
- 🇷🇺 [Документация на русском](README_RU.md)

## Quick start

```bash
git clone https://github.com/gritsulyak/nanoforum-go.git
cd nanoforum-go
task docker-build
task docker-up
```

Forum will be available at `http://localhost:8084`.

## Load testing

See [LOAD_TESTING.md](LOAD_TESTING.md) for k6 load test targets, results, and CPU usage analysis.