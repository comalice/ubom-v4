# Server Configuration

The server uses nested configuration structs. Each configurable module owns a
config struct, and its consumer supplies that config explicitly.

The executable composes the top-level configuration:

```text
Config
  HTTP
  Database
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
- `-db`: SQLite database path. Empty uses the in-memory sample.
- `-artifact-snapshot-root`: filesystem root for retained snapshot content.
- `-local-source-id`: existing local artifact source ID.
- `-local-source-root`: filesystem root exposed by that source.

Do not add source-specific connection details to domain types. Add them to the
configuration for the relevant adapter instead.
