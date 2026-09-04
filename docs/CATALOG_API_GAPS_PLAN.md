# Catalog API Integration Plan

This plan records the remaining work needed to turn the catalog browser into a
complete, shareable, production catalog. The current implementation can load a
taxonomy node and descendant-inclusive part results, but several API and UI
contracts still need to be completed.

## Decisions

- Category descriptions, images, and related presentation metadata should live
  with taxonomy definitions.
- Catalog URLs should be human-readable. They should use taxonomy label paths
  for navigation while retaining explicit taxonomy context and stable node
  identity where needed to avoid ambiguity.
- The catalog should use taxonomy node metadata for navigation and effective
  attributes, and part search for descendant-inclusive results.
- Values within one filter group are ORed. Different filter groups are ANDed.

## Required API Contract Work

### 1. Multi-value exact filters

The frontend serializes repeated filters, for example:

```text
attribute=material=Steel&attribute=material=Aluminum
```

The server currently stores filters as `map[AttributeDefID]string`, causing
repeated values to overwrite one another.

Change the query model to represent multiple values per attribute:

```go
map[AttributeDefID][]string
```

Matching should use OR semantics within an attribute and AND semantics across
attributes.

### 2. Numeric range filters

The frontend design includes range controls, but the server only supports exact
string equality. Define and implement an explicit query contract, for example:

```text
attributeMin.resistance=100
attributeMax.resistance=10000
```

Range values must be parsed and compared numerically. URL state must represent
minimum and maximum values independently.

### 3. Facets and counts

The API does not currently return:

- Part counts for child categories.
- Counts for enum filter values within the current scope.

Add a catalog facet response, preferably alongside the node metadata response or
as a dedicated endpoint. It should account for the current taxonomy scope,
search query, and active filters as appropriate.

Avoid issuing one count request per child category.

### 4. Status filtering

Parts expose status, but search has no status parameter. Decide and implement
the catalog policy:

- Hide archived parts by default and support an explicit include option; or
- Show archived parts with a visible status; or
- Exclude archived parts permanently from the catalog.

The chosen policy must be reflected in the API query and URL state.

### 5. Filter validation

Unknown attribute filters currently produce no matches without explaining why.
The server should validate filters against the selected taxonomy's effective
attributes and enum vocabularies, or return a structured validation response.

Invalid URL filters should be removed or clearly surfaced by the frontend.

### 6. Consistent part-search response

`GET /api/parts` currently returns a bare array with no query parameters and a
paginated envelope when any query parameter is present. Use one response shape
for all calls:

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "pageSize": 50
}
```

This simplifies the client and avoids compatibility behavior leaking into new
catalog code.

## Taxonomy Definition Enhancements

### 7. Category metadata

Taxonomy children currently expose only an ID and label. Add optional category
metadata to taxonomy definitions and node views:

- Description.
- Image or image reference.
- Optional display/order metadata if needed.

This metadata belongs to the taxonomy definition because it describes the
catalog taxonomy itself and should be versioned with that taxonomy revision.

The API should expose the metadata in both the complete taxonomy tree and the
selected node's child views.

### 8. Human-readable taxonomy URLs

Catalog links should be readable, for example:

```text
/catalog/fasteners?taxonomyDef=components-v3
```

The implementation must still resolve nodes using stable taxonomy IDs. Options
include:

- Resolve each path segment against the selected taxonomy definition.
- Include a stable `node` query parameter when labels are not unique.
- Use explicit taxonomy revision context in the URL.

The canonical URL policy should be deterministic so copied links do not depend
on the currently selected taxonomy revision.

### 9. Taxonomy selection

When no taxonomy is specified, the frontend currently chooses the first summary
marked current. This is ambiguous when multiple sequence definitions or
taxonomy lineages exist.

Choose one of:

- Require `taxonomyDef` in catalog URLs.
- Add a catalog taxonomy selector.
- Configure a default catalog taxonomy.
- Add a server endpoint representing the default catalog taxonomy.

The selected taxonomy revision must be part of a shareable URL when it affects
the rendered result.

### 10. Node result semantics

The current API has two different behaviors:

- The taxonomy node endpoint returns parts assigned exactly to the node.
- Part search with `taxonomyNode` includes the node and descendants.

Keep the distinction explicit in API documentation and tests. The catalog should
continue using node responses for structure/attributes and part search for the
main descendant-inclusive result set.

Consider separating node metadata from part results in the API so clients do
not accidentally consume an unpaginated exact-node list.

## Frontend Interaction Work

### 11. Pagination

The URL and API client support `page` and `pageSize`, but the catalog has no
pagination controls. Add:

- Previous/next controls.
- Current page and total-page state.
- Disabled states.
- Page-size selection if needed.
- URL updates when paging.

Choose one canonical page-size default. The current catalog route defaults to
25 while the server defaults to 50.

### 12. Sorting

The URL and API client support sorting, but the catalog has no sort control.
Expose only server-supported choices initially, currently value and ID, or
extend the server before exposing additional fields.

Sorting changes must be represented in the shareable URL.

### 13. URL-state validation and canonicalization

Define behavior for stale or invalid URLs, including:

- Deleted taxonomy revisions.
- Unknown taxonomy nodes or path segments.
- Filters no longer available in the selected scope.
- Out-of-range pages.
- Unsupported sort/order values.

Prefer a useful error/not-found state for invalid taxonomy identity and
canonicalization or removal of invalid optional query state.

### 14. Search semantics

The current `q` query searches only part-number values. The UI placeholder says
"Part number or description", but parts have no description field.

Either change the UI wording or add description metadata to the part model and
search contract. Category descriptions are separate taxonomy metadata and do
not solve part-description search.

### 15. Catalog component coverage

Add frontend tests for:

- Taxonomy resolution.
- Root and child node loading.
- Effective attribute filter rendering.
- Search and filter interactions.
- Empty and error states.
- Stale request protection.
- URL updates and browser back/forward behavior.

## Server Performance Work

### 16. Move search filtering toward storage

`SearchPartNumberViews` currently loads all parts and filters in memory. As the
catalog grows, move filtering and pagination toward the store layer and add
appropriate indexes.

Likely areas include:

- Taxonomy definition and node filtering.
- Part status.
- Searchable part value.
- Attribute filtering.
- Stable sorting.

The lack of a materialized taxonomy ancestor/path table should be addressed if
descendant queries become expensive.

### 17. Paginated descendant result API

The current node view embeds all direct-node parts without pagination. The
catalog does not use this field, but a clearer API should separate:

- Node metadata and navigation.
- A paginated, descendant-inclusive part result query.

This also prevents accidental large responses for categories with many direct
parts.

### 18. Explicit server query types

Evolve `PartSearchQuery` to explicitly model catalog behavior:

- Multi-value exact filters.
- Numeric ranges.
- Status.
- Supported sort fields and direction.
- Pagination limits.
- Optional facet requests.

Avoid encoding new semantics only through loosely parsed query strings.

## Verification and Operations

### 19. Server and frontend integration tests

Use a seeded taxonomy and parts dataset to verify:

- Taxonomy selection.
- Human-readable path resolution.
- Node navigation.
- Descendant-inclusive results.
- Multi-value filters.
- Numeric ranges.
- Pagination and sorting.
- Archived-part behavior.

These tests should cover memory and durable stores where practical.

### 20. Live browser validation

Provide a reproducible local seeded server/database setup so the catalog can be
validated in Firefox against real API responses, including copied URLs and
back/forward navigation.

## Recommended Implementation Order

1. Fix multi-value exact filters in the server.
2. Decide and unify the page-size default.
3. Implement human-readable taxonomy URL resolution and explicit taxonomy
   selection behavior.
4. Add pagination and sorting controls.
5. Add URL-state validation and canonicalization.
6. Add taxonomy category descriptions and image metadata.
7. Add numeric range filtering.
8. Add status filtering.
9. Add facets and category/filter counts.
10. Add component and server/frontend integration coverage.
11. Move search and descendant filtering toward the storage layer as scale
    requires.
