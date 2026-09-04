# Development Endpoint Stress QA

## Scope

This harness starts an isolated `make dev` stack, sends repeated sequential
requests through the Vite proxy to frontend and API endpoints, prints timing
statistics after every request, and verifies that the dev stack is cleaned up
afterward.

## Run

From the repository root:

```sh
./scripts/stress-dev-endpoints.sh 100
```

The argument is the number of iterations. Each iteration requests the frontend
root/settings/catalog routes and representative version, taxonomy, root,
parts, node, and facets API routes. The default is 100 iterations, for 1,000
sequential requests.

The harness uses a temporary SQLite database and does not modify `dev.db`.

## Output

Each request prints:

- Iteration and endpoint.
- HTTP status.
- Curl exit status.
- Connection time.
- Time to first response byte.
- Total request time.

Example:

```text
iteration=1 endpoint=/api/version curl_status=0 status=200 connect=0.0001s start=0.0004s total=0.0004s
```

Investigate any request with a large `start` or `total` time, especially when
the same endpoint is fast in adjacent iterations. A `000` status indicates a
connection or timeout failure.

## Pass Criteria

- All services become ready within the harness wait period.
- All expected requests return HTTP 200.
- No curl request fails or times out.
- No listeners remain on ports 5173 or 8080 after the script exits.

Do not use fixed timing thresholds as correctness gates; record the output and
compare repeated runs under similar machine load.
