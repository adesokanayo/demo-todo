# Demo TODO API

A small Go HTTP API for todos. Todos are stored in Postgres. Interactive docs are served with Swagger UI.

## Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/` | Service name and a link to the docs |
| `GET` | `/health` | API and Postgres health |
| `GET` | `/todos` | List todos |
| `POST` | `/todos` | Create a todo |
| `GET` | `/todos/{id}` | Fetch one todo |
| `POST` | `/todos/{id}` | Update the title, the done flag, or both |
| `GET` | `/swagger/index.html` | Swagger UI |

`POST /todos` body:

```json
{"title": "Buy milk"}
```

`POST /todos/{id}` body. Send only the fields you want to change:

```json
{"title": "Buy oat milk", "done": true}
```

## Configuration

| Variable | Required | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | yes | Postgres connection string |
| `PORT` | no | HTTP port. Defaults to `8080`. BreezeCloud sets this. |

## Run locally

```bash
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/todo?sslmode=disable
go run ./cmd/demo-todo
```

Open `http://localhost:8080/swagger/index.html`.
