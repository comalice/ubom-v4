# Catalog Taxonomy Selection QA

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
`/api/catalog/facets` responds to a valid catalog query.

Fail: either service exits, the frontend cannot load, or the facets endpoint
returns `404` or another unexpected error.

### 2. Open the unscoped catalog

Open:

```text
/catalog
```

Pass: the page displays `Choose a catalog taxonomy`, and the selector contains
both current taxonomy definitions. No catalog data request is made without a
taxonomy choice when multiple current taxonomies exist.

Fail: the page selects a taxonomy silently, shows a generic 404, or the
selector is absent or incomplete.

### 3. Select a taxonomy and verify the root

Select `sample-taxonomy-v1`.

Pass: the browser URL becomes `/catalog?taxonomyDef=sample-taxonomy-v1`; the
page loads the `Components` root, shows `Resistors` and `Capacitors`, and does
not display `404 Not Found`. The browser console has no errors, and both the
parts search and catalog facets requests succeed.

Fail: selection navigates to a 404, the root is missing, the query loses the
selected taxonomy definition, or either catalog API request fails.

### 4. Verify root navigation and reload

From the `Components` root, activate the `Catalog` breadcrumb, then reload the
page. Open `Resistors`, return to `Catalog`, and reload again.

Pass: root navigation preserves `taxonomyDef=sample-taxonomy-v1`; reloads keep
the catalog on the same taxonomy and scope; child navigation loads the
corresponding category without a full-page request for an unknown route.

Fail: the root breadcrumb drops the taxonomy definition, reloads show the
selector or a 404 unexpectedly, or a valid child link cannot resolve.

### 5. Verify direct links and invalid paths

Open the following directly:

```text
/catalog?taxonomyDef=sample-taxonomy-v1
/catalog/components/resistors?taxonomyDef=sample-taxonomy-v1
/catalog/not-a-real-category?taxonomyDef=sample-taxonomy-v1
```

Pass: the first URL loads `Components`, the second loads `Resistors`, and the
third reports a clear catalog-scope 404 without corrupting the application
state.

Fail: a valid URL 404s, an invalid scope renders unrelated data, or the error
is hidden while requests continue indefinitely.

## Completion

The catalog selection run passes only when all five steps pass. Record the
frontend URL, API URL, selected taxonomy ID, and failing request status for any
failure. Run this playbook after taxonomy fixtures are created and whenever
catalog routing, taxonomy selection, API handlers, or deployment proxy rules
change.
