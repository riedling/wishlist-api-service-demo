# Wishlist API Service

Boilerplate Go REST API built with [Gin](https://github.com/gin-gonic/gin).

## Features

- Gin-based HTTP server with graceful shutdown
- Consistent JSON response envelope (`success`, `data`, `error`, `meta`)
- Reusable request-binding/validation helpers (`internal/httputil`)
- Reusable middleware: request ID propagation, structured logging, panic
  recovery, CORS (`internal/middleware`)
- Pagination helper for list endpoints
- Example `/items` CRUD resource demonstrating the extensible pattern
- Environment-based configuration (`internal/config`)

## Project layout

```
cmd/api/                 Application entry point (main.go)
internal/config/         Environment-based configuration loading
internal/server/         HTTP server bootstrap + graceful shutdown
internal/router/         Route registration/wiring
internal/middleware/     Cross-cutting Gin middleware
internal/httputil/       Shared response envelope, binding, pagination helpers
internal/handlers/       Endpoint handlers, grouped by resource
```

## Getting started

```bash
go mod tidy
go run ./cmd/api
```

The server listens on `:8080` by default. Configure via environment
variables:

| Variable    | Default       | Description                         |
|-------------|---------------|--------------------------------------|
| `PORT`      | `8080`        | TCP port to listen on                |
| `APP_ENV`   | `development` | `development` or `production`        |
| `LOG_LEVEL` | `info`        | Reserved for future logger config    |

## Adding a new endpoint

1. Create a handler struct/function in `internal/handlers/` (see
   [items.go](internal/handlers/items.go) for the pattern).
2. Use the helpers in [internal/httputil/response.go](internal/httputil/response.go)
   (`OK`, `Created`, `NotFound`, `BindJSON`, etc.) so every endpoint returns
   the same response shape and error handling.
3. Register the routes in a `register*Routes` function in
   [internal/router/router.go](internal/router/router.go), and call it from
   `New()`.

## Example requests

```bash
curl http://localhost:8080/health

curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Headphones","price":9999}'

curl http://localhost:8080/api/v1/items
```
