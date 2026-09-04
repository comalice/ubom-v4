# UBOM

UBOM is a single-user part-number and taxonomy manager. It defines the language
used by part numbers, organizes that language into a taxonomy, captures typed
attributes, and provides a searchable catalog of released parts.

The initial release deliberately focuses on durable part-number workflows. BOM
authoring, artifact integrations, and quality-system workflows are later work.

## Current Capabilities

- Define sequence grammars using literals, concatenation, choices, ranges,
  range-radix nodes, bindings, and character sets.
- Validate and preview sequence definitions in the browser.
- Save immutable sequence-definition revisions and open historical revisions
  read-only.
- Build taxonomy trees bound to sequence definitions.
- Add exact and numeric-range taxonomy predicates.
- Define typed string, decimal, and enum attributes.
- Inherit and override taxonomy attributes at descendant nodes.
- Save immutable taxonomy revisions with a server-owned current revision.
- Create, preview, allocate, and release part-number drafts against a taxonomy
  node.
- Search and browse parts by taxonomy scope, part-number value, attributes, and
  status.
- Browse current taxonomies through a grouped catalog home page with
  human-readable taxonomy URLs.
- Use multi-value filters, numeric ranges, facets, pagination, sorting, and
  archived-part filtering.
- Persist deployed data in Postgres, with SQLite and memory stores for local
  development and tests.

## Domain Model

Sequence definitions describe arbitrary string grammars. They support parsing
existing values and generating new values. Saved definitions are immutable
revisions; changes create a new revision rather than changing the definition
used by existing parts.

A taxonomy applies meaning to sequence fields. Its nodes can carry predicates,
attributes, children, and directly classified parts independently. Taxonomy
revisions are immutable, and a part retains the exact sequence and taxonomy
revisions used when it was created.

Part-number values are globally unique canonical values. Drafts do not reserve a
number; allocation happens transactionally during release. Released parts are
not silently deleted and can be archived instead.

## Local Development

Run the API and frontend with SQLite:

```sh
make dev
```

The frontend is available at `http://localhost:5173` and proxies `/api` to the
API at `http://localhost:8080`.

For local development, `make dev` seeds a SQLite database at
`./.data/dev.db`. To run the local API against the Compose Postgres service
without reseeding it, use:

```sh
make dev-postgres
```

The Postgres data directory defaults to `./.data/postgres`, matching the
`POSTGRES_DATA_DIR` bind mount used by Compose. Override it consistently when
using another directory:

```sh
POSTGRES_DATA_DIR=/srv/ubom/postgres docker compose up -d postgres
POSTGRES_DSN='postgresql://ubom:ubom-dev-only@localhost:5432/ubom?sslmode=disable' \
POSTGRES_DATA_DIR=/srv/ubom/postgres make dev-postgres
```

Useful checks:

```sh
go test ./...  # run from server/
make -C frontend test
make -C frontend typecheck
make -C frontend build
```

The development endpoint stress harness exercises the API and Vite server with
per-request timing output:

```sh
./scripts/stress-dev-endpoints.sh 100
```

To time representative frontend HTML page requests repeatedly:

```sh
./scripts/measure-frontend-pages.sh 20
```

This measures the frontend response only. `curl` does not run page JavaScript or
wait for API calls started by the browser; see `test/qa/FRONTEND_PAGE_LOAD.md`
for the distinction and browser-level measurement guidance.

## Compose Deployment

`docker compose up --build` starts the frontend, API, and durable Postgres
bind-mounted data directory. The directory defaults to `./.data/postgres` and
can be changed with `POSTGRES_DATA_DIR`. Open `http://localhost:8080`.

The API runs embedded, versioned Postgres migrations before serving. The API
provides `/healthz` for process health and `/readyz` for database readiness.
The frontend proxies both endpoints and the `/api` routes to the API service.

The repository supports Docker installations using the classic builder. The
Compose build configuration uses the host network for build-time dependency
downloads because some classic-builder installations cannot resolve external
package registries through the default builder network. Runtime service
networking is unchanged.

For a production-like startup, provide deployment credentials rather than the
development defaults:

```sh
POSTGRES_PASSWORD='use-a-secret' docker compose up -d --build
docker compose ps
docker compose logs -f ubom
```

Set `POSTGRES_USER`, `POSTGRES_DB`, and optionally `UBOM_HTTP_PORT` through the
deployment environment or a secrets manager. Do not expose the development
password outside local evaluation.

Set `UBOM_BUILD_ID` to the same immutable release identifier for the API and
frontend images, for example `UBOM_BUILD_ID=v0.0.1`. Compose passes it to the
API as `UBOM_BUILD_ID` and to the Vite build as `VITE_APP_BUILD_ID`. If it is
unset, local builds use `dev`. The API publishes this identity at
`/api/version`, in `X-UBOM-Build-ID`, and through the `/api/events` SSE stream.
The frontend checks the identity at startup and shows a refresh notice when the
running application and API no longer match.

The deployment boundary is explicitly single-user. There is no login,
authorization, tenant isolation, or audit actor identity. Put the deployment
behind an authenticated network boundary before exposing it to other users.

## Migrations And Backups

Postgres migrations are embedded in `server/store/postgres_migrations.go` and
applied automatically in order at API startup. A failed migration prevents the
API from starting. Back up before upgrades; migrations do not have automatic
rollback.

Create a logical backup from the running Compose database:

```sh
docker compose exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  --format=custom' > ubom-$(date +%Y%m%d-%H%M%S).dump
```

Restore into a stopped application and an empty database:

```sh
docker compose stop frontend ubom
cat backup.dump | docker compose exec -T postgres sh -c 'pg_restore -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" --clean --if-exists'
docker compose start ubom frontend
```

The Postgres data directory is persistent, but it is not a backup. Artifact
snapshot files, when configured, live outside Postgres and must be backed up
separately. Expect data loss since the last successful backup if both the
directory and its backup are lost.

## Release Scope

Included in the initial release:

- Sequence definition authoring and immutable revisions.
- Taxonomy authoring, predicates, attributes, and immutable revisions.
- Context-sensitive part draft creation, preview, allocation, and release.
- Searchable catalog browsing with filters, facets, pagination, and sorting.
- SQLite, memory, and Postgres persistence adapters.
- Docker Compose deployment with health and readiness checks.

Deferred:

- BOM authoring and multi-BOM support.
- Artifact sources, snapshots, and file integrations.
- Release approvals and formal ECO/ECR change control.
- CAD integrations.
- Advanced reporting and analytics.
- Fine-grained permissions and multi-user tenancy.

Planning and design notes live under `docs/`. Operational QA runbooks live
under `test/qa/`.
