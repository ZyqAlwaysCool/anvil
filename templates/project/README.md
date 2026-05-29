# {{.ProjectName}}

Agent platform starter generated from the platform scaffold. This project includes an HTTP server, optional background worker, and a sample agent under `internal/agent/biz/example`.

## Quick start

```bash
cp configs/.env.example .env
make run
```

The server listens on the port configured in `.env` (default `8080`). Health check: `GET /health`.

## Development commands

| Command | Description |
|---------|-------------|
| `make build` | Build `bin/server` and `bin/worker` |
| `make run` | Build and start the HTTP server only |
| `make run-worker` | Build and start the task worker only |
| `make test` | Run `go test ./...` |
| `make lint` | Run linters (when configured) |

## Optional: tasks, Redis, and MongoDB

By default, `configs/.env.example` keeps `TASK_ENABLED`, `REDIS_ENABLED`, and `MONGO_ENABLED` set to `false` so you can run HTTP-only locally.

To enable the full async task pipeline:

1. Set `TASK_ENABLED=true`, `REDIS_ENABLED=true`, and `MONGO_ENABLED=true` in `.env`.
2. Point `REDIS_ADDRS` and `MONGO_URI` at running instances.
3. Adjust `TASK_STREAM_KEY`, `TASK_CONSUMER_GROUP`, and `TASK_CONSUMER_NAME` if you run multiple projects against the same Redis.

Run the worker in a second terminal:

```bash
make run-worker
```

Or use Docker Compose (enables Redis, Mongo, server, and worker with task settings overridden):

```bash
docker compose up --build
```

## Project layout

- `cmd/server` — HTTP entrypoint
- `cmd/worker` — task consumer entrypoint
- `internal/platform` — shared config, logging, storage, task queue
- `internal/server` — routes and wiring
- `internal/agent` — register your agents here
- `prompts/` — prompt assets for agents
- `configs/.env.example` — environment template; copy to `.env` at repo root

## Adding an agent

Implement handlers and services under `internal/agent/biz/<your-agent>/`, register routes in that package’s `register.go`, and call your registrar from `internal/agent/register.go`.
