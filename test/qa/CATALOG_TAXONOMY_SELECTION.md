# Grouped Catalog Root QA

## Fixture

Use the seeded local application with its default sample data. The seed should
provide these two current taxonomies:

- `gtd-pn-taxonomy-v1` (`gtd-pn-v1`)
- `sample-taxonomy-v1` (`sample-pn-v1`)

The `sample-taxonomy-v1` root is `Components`, with `Resistors` and `Capacitors`
as child categories.

## Procedure

### 1. Start the application

Run `make dev` and open the local application.

Pass: the frontend and API are available without startup errors, and
`/api/catalog/root` returns a JSON object with both current taxonomy groups.

Fail: either service exits, the frontend cannot load, or the aggregate endpoint
returns `404` or another unexpected error.

### 2. Open the unscoped catalog

Open:

```text
/catalog
```

Pass: the page displays one sidebar group and one main tile section for each
current taxonomy. It does not display a taxonomy selector or a combined filter
panel, and the network log shows one `/api/catalog/root` request.

Fail: a taxonomy is selected silently, a generic 404 is shown, a group is
missing, or separate taxonomy/category requests are needed to render the root.

### 3. Verify groups, metadata, counts, and links

Inspect the `sample-taxonomy-v1` group and its category tiles.

Pass: the group identifies `taxonomyDef=sample-taxonomy-v1` and its `seqDef`,
shows the `Components` root and its metadata, and shows `Resistors` and
`Capacitors` with active-part counts. Group headings link to
`/catalog?taxonomyDef=<id>` and child links use paths relative to the taxonomy
root, with the same explicit taxonomy definition.

Fail: counts are combined across taxonomies, archived parts are included by
default, links omit taxonomy context, or root metadata is absent.

### 4. Verify scoped navigation and reload

Open the `Components` group heading, then reload. Open `Resistors`, return to
`/catalog`, and reload again.

Pass: scoped URLs preserve `taxonomyDef=sample-taxonomy-v1`; reloads keep the
catalog on the same taxonomy and scope; returning to `/catalog` restores all
groups. Scoped pages still load their existing parts and facet requests.

Fail: a valid scoped URL 404s, taxonomy context is lost, or the root does not
restore all current groups.

### 5. Verify archived policy and invalid paths

Open the following directly:

```text
/catalog?includeArchived=true
/api/catalog/root?includeArchived=true
/catalog/resistors?taxonomyDef=sample-taxonomy-v1
/catalog/not-a-real-category?taxonomyDef=sample-taxonomy-v1
```

Pass: archived parts appear only with explicit inclusion, the aggregate API
returns archived-inclusive counts, valid scoped URLs load, and the invalid path
reports a clear catalog-scope 404 without corrupting application state.

Fail: archived parts are included by default, the aggregate endpoint accepts
taxonomy-specific filters, or an invalid scope renders unrelated data.

## Completion

The grouped catalog root pass passes only when all five steps pass. Record the
frontend URL, API URL, selected taxonomy ID, and failing request status for any
failure. Run this playbook whenever catalog routing, aggregate API handlers,
taxonomy fixtures, or deployment proxy rules change.
