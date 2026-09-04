# Frontend Page-Load Timing

## Run

Start the local development stack first:

```sh
make dev
```

Then, from the repository root, run one pass over the representative frontend
routes:

```sh
./scripts/measure-frontend-pages.sh
```

To reproduce a page-load stall while repeatedly requesting the pages:

```sh
./scripts/measure-frontend-pages.sh 20
```

The optional second argument overrides the frontend URL:

```sh
./scripts/measure-frontend-pages.sh 20 http://localhost:5173
```

`CURL_MAX_TIME` controls the per-page timeout and defaults to 30 seconds:

```sh
CURL_MAX_TIME=60 ./scripts/measure-frontend-pages.sh 20
```

## Output

Each line includes the iteration, page, HTTP status, curl status, DNS lookup,
connection, time to first byte, and total response time. The `total` value is
the time for the frontend HTML request, not the time for the browser to finish
running the page.

## API Requests

`curl` does not execute the returned HTML or JavaScript, so it cannot observe or
wait for API calls started by a browser. A frontend page request completing only
means that the Vite server returned `index.html`.

Use the existing
[`stress-dev-endpoints.sh`](../../scripts/stress-dev-endpoints.sh) harness to
time API endpoints directly. To measure the actual browser page completion,
use browser automation such as Playwright and wait until the page has no
outstanding requests matching `/api/` (or wait for a specific application-ready
condition). That is the appropriate follow-up if the HTML timings stay low but
the UI still takes 15–20 seconds to become usable.
