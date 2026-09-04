# App Update Compatibility QA

## Contract

Production deployments must set `UBOM_BUILD_ID` to one stable release identity.
Compose passes that value to the Go API and to the Vite build as
`VITE_APP_BUILD_ID`. The two values must be identical. If unset, both local
builds use the deterministic `dev` identity.

The API exposes the identity at `/api/version` and in the `X-UBOM-Build-ID`
response header. `/api/events` is an SSE stream. Its first event is the named
`app-version` event, followed by comment heartbeats approximately every 30
seconds.

The frontend performs one `/api/version` probe during bootstrap before opening
SSE. A missing, non-success, or invalid version response is treated as an
incompatible backend. Ordinary API failures, including a valid catalog scope
404, remain ordinary API errors.

## Procedure

### 1. Start a deployment with an explicit identity

Run:

```sh
UBOM_BUILD_ID=qa-update-1 docker compose up -d --build
```

Pass: `curl -i http://localhost:8080/api/version` returns `200`, JSON
`{"buildId":"qa-update-1"}`, and `X-UBOM-Build-ID: qa-update-1`.

Fail: either service uses a different identity, the endpoint is unavailable,
or the frontend cannot load.

### 2. Verify the open-app update notice

Open the frontend in a browser, then rebuild/restart the deployment with a new
identity:

```sh
UBOM_BUILD_ID=qa-update-2 docker compose up -d --build
```

Pass: the already-open page shows one accessible notice reading `The app has
updated.` with a `Refresh` button. The page does not reload automatically.

Fail: no notice appears, multiple duplicate notices appear, or the page reloads
without the user activating `Refresh`.

### 3. Verify mismatch is explicit

Build the frontend with `VITE_APP_BUILD_ID=qa-update-1` while serving an API
with `UBOM_BUILD_ID=qa-update-2`, then open a catalog page.

Pass: the notice says `The frontend and API versions do not match.` and the
failed API operation is not presented as an ordinary catalog scope `404`.

Fail: the UI only displays a generic catalog `404`, or no refresh action is
available.

This step also covers an older backend that has neither `/api/events` nor
`X-UBOM-Build-ID`: its failed `/api/version` probe must still produce the
actionable notice before the initial catalog request is allowed to render, and
the frontend must not repeatedly reconnect to SSE. The page may show the
explicit compatibility error, but must not show only a catalog scope `404`.

If the `/api/version` probe cannot connect because of a transient network
failure, the frontend should continue normal route loading. That case must not
produce the version-mismatch notice.

## Completion

Record both build identities, the browser URL, and the API response headers for
any failure. Run this playbook after frontend/API image, proxy, or deployment
configuration changes.
