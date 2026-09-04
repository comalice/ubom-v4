# Server Configuration

The server uses nested configuration structs. Each configurable module owns a
config struct, and its consumer supplies that config explicitly.

The executable composes the top-level configuration:

```text
Config
  HTTP
  Database
    DSN
    Path
  Artifacts
    SnapshotRoot
    LocalSource
      ID
      Root
```

Configuration is not part of the domain model. Domain types describe UBOM
data; configuration describes how modules run.

## Ownership

- `HTTPConfig` controls the HTTP listener.
- `DatabaseConfig` controls metadata persistence.
- `ArtifactConfig` controls artifact capture.
- `LocalFileAdapterConfig` controls a local source adapter.
- `FileSnapshotStoreConfig` controls durable snapshot content storage.

The executable sets defaults and reads command-line flags. Modules validate the
config they receive and do not read flags or environment variables directly.

## Artifact Capture

Artifact capture is opt-in. It is enabled only when all of these values are
set:

- `artifact-snapshot-root`
- `local-source-id`
- `local-source-root`

The source identified by `local-source-id` must already exist in metadata. The
executable then constructs the local adapter, durable snapshot store, and
snapshot service, and passes the service to the API.

If all artifact settings are empty, the API remains available for artifact
metadata but snapshot capture is not configured. Partial artifact settings are
an invalid startup configuration.

## Command-Line Flags

- `-listen`: HTTP listen address. Defaults to `127.0.0.1:8080`.
- `-database-url`: Postgres connection URL. Defaults to `UBOM_DATABASE_URL` when set.
- `-db`: SQLite database path. Used when no Postgres URL is configured; empty uses the in-memory sample.
- `-artifact-snapshot-root`: filesystem root for retained snapshot content.
- `-local-source-id`: existing local artifact source ID.
- `-local-source-root`: filesystem root exposed by that source.

Do not add source-specific connection details to domain types. Add them to the
configuration for the relevant adapter instead.

The API exposes `/healthz` for process health and `/readyz` for database
readiness. Postgres-backed startup runs embedded, versioned migrations before
the listener is opened. SQLite remains the local-development option.

## HTTP Diagnostics

The HTTP server emits structured JSON request diagnostics through `log/slog`.
Each request produces `http request started`, `http response started`, and
`http request completed` records. The records include the method, path,
request ID, status, response bytes, and elapsed time in milliseconds. The
request ID is also returned as `X-Request-ID`, so a browser Network-panel
request can be matched directly to server logs.

For a slow request, look for a start record followed by its completion record:

```sh
docker compose logs --no-log-prefix ubom | grep '"request_id"' | jq 'select(.request_id == "...")'
```

The `/api/events` endpoint is a deliberate long-lived SSE connection. Its
completion record is emitted only when the browser closes the connection, so
do not treat the absence of its completion record as a stuck handler. Its
response-start record confirms that the initial SSE response was sent.
