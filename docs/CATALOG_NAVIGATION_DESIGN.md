# UBOM Catalog Navigation Design

## Status

Exploratory design. This document describes the intended catalog browsing and
filtering experience. It is not an implementation commitment. The next step is
to build a playground sketch and refine the interaction and data contracts
before changing the application shell or production pages.

This document complements `INITIAL_RELEASE_PLAN.md`. It refines the home-page,
taxonomy-navigation, and attribute-filtering portions of that plan.

## Design Reference

The primary visual and interaction reference is the McMaster catalog:

- https://www.mcmaster.com/products/measuring-inspecting/
- https://www.mcmaster.com/products/ph-testers/

The prior Datum implementation is the primary structural reference:

- Local checkout: `~/git/datum`
- Category workspace: `frontend/src/components/PartTable.tsx`
- Taxonomy API behavior: `server/api/server.go`
- Scope result presentation: `frontend/src/api/taxonomyView.ts`

Datum is especially useful because it treats taxonomy navigation as a
server-derived scope rather than having the frontend infer categories from
part-number strings.

## Product Direction

UBOM should present a catalog-like home page rather than a part-log dashboard.
The home page is the entry point for navigating the taxonomy and finding parts.
The part log remains available as a result and search surface, but it should
not dominate the initial view.

The experience should be dense, structured, and practical:

- A persistent left panel provides navigation and filtering.
- The main view presents category groups and tiles.
- Category pages can expose multiple levels of the taxonomy.
- Any taxonomy node may be part-bearing, including nodes with children.
- Attributes are contextual to the current taxonomy scope.
- URLs preserve the selected taxonomy context and filters.

## Taxonomy Semantics

The existing taxonomy tree is the initial browse tree. Taxonomy nodes can carry
classification predicates, attributes, children, or any combination of these.
The model must not assume that a part-bearing node is a leaf.

These are independent concepts:

```text
node has classification predicates
node is navigable
node is part-bearing
node has children
node has attributes
```

For example:

```text
Fastening                         [navigable]
  Fasteners                       [navigable, part-bearing]
    Bolts                         [navigable, part-bearing]
    Screws                        [navigable, part-bearing]
  Adhesives                       [navigable, part-bearing]
```

A node with children can therefore have both:

- Parts classified directly at the current node.
- Descendant categories containing additional parts.

The catalog must expose both without requiring an unnecessary drill-down.

### Classification and navigation

Classification remains driven by taxonomy predicates and the parsed sequence
fields. Navigation is driven by the resolved taxonomy tree and persisted part
membership. The frontend must not parse or split canonical part-number values
to construct taxonomy paths.

The server owns:

- Taxonomy path resolution.
- Classification and branch applicability.
- Immediate child categories.
- Direct and descendant part counts.
- Effective attributes for the current path.
- Facet values and counts once filtering is implemented.

## Left Panel

The left panel has two concerns and they must remain visually and semantically
distinct.

```text
Current scope
  Breadcrumbs
  Immediate taxonomy categories

Filter by
  Effective attributes from ancestor scopes and the current scope
```

### Category navigation

Category navigation is derived from the immediate scope:

```text
categories = children(currentNode)
```

The panel should not flatten every descendant category. It should show the
current scope and its immediate children, with the selected branch clearly
highlighted.

At the root, the category section lists all top-level taxonomy nodes. On a
deeper page, it lists the children of the selected node and provides a way to
return to the parent or start over.

### Attribute filtering

Attribute filters are derived from the entire resolved path:

```text
filters = effectiveAttributes(root -> currentNode)
```

Attributes assigned at ancestors are available in descendant scopes. A deeper
assignment of the same attribute ID overrides the shallower assignment using
the current taxonomy precedence rule.

For example:

```text
Components
  lifecycle

Passive Components
  technology

Resistors
  resistance
  tolerance
```

The `Resistors` filter panel should expose all four attributes:

```text
Lifecycle
Technology
Resistance
Tolerance
```

Category navigation and attribute filtering must not be conflated. Categories
come from immediate children; filters come from inherited effective attributes.

### Panel states

The desktop panel should support:

- A compact category breadcrumb.
- Selected-category highlighting.
- Immediate child category links.
- A `Filter by` heading.
- Collapsible attribute groups.
- Active filter chips or clear controls.
- Counts for available values where available.
- A clear-all action.

On mobile, the same two concerns should become separate drawers or tabs:

- `Browse categories`.
- `Filter`.

## Home Page

The home page uses the root taxonomy scope.

```text
+---------------------------------------------------------------+
| UBOM logo | Search parts | actions                            |
+----------------------+----------------------------------------+
| Browse catalog       | Catalog                               |
|                      |                                        |
| Electrical            | Electrical                             |
| Fastening & Joining  | [Passive Components] [Connectors]    |
| Mechanical            | [Sensors] [Power]                     |
| Materials             |                                        |
| ...                  | Fastening & Joining                   |
|                      | [Fasteners] [Adhesives] [Welding]    |
+----------------------+----------------------------------------+
```

The left rail and main content must use the same taxonomy response. They are
two presentations of one scope, not two independently maintained category
lists.

The main content may render more than one level for a broad scope:

```text
Electrical
  Passive Components
    Resistors
    Capacitors
    Inductors

  Connectors
    Board-to-Board
    Wire-to-Board
    Circular
```

Tiles may contain:

- A category label.
- An optional image or icon.
- An optional short description.
- A released-part count.
- A clear link to the category scope.

The first playground should use representative fixture data and should not
depend on production taxonomy content being present.

## Category Pages

A category route such as:

```text
/parts/electrical/
```

should render:

1. Breadcrumbs and the current category.
2. Immediate child categories in the left panel.
3. Effective inherited attributes in the filter section.
4. Child category sections and tiles in the main view.
5. Direct parts assigned to the current node, when present.
6. A way to view all descendant parts in the current scope.

The main view may group child tiles into subcategory sections, as McMaster does
for broad categories. The grouping must come from explicit taxonomy structure,
not from label parsing or arbitrary depth assumptions.

## Parts At Intermediate Nodes

The API and UI must distinguish direct membership from descendant membership.
The preferred response terminology is:

```text
directParts
descendantPartCount
```

This avoids ambiguity when a node is both part-bearing and has children.

The default category page behavior should be:

- Show child categories as navigation.
- Show direct parts in a clearly labeled section.
- Offer `View all parts` for the complete descendant scope.

The complete descendant result set should include parts classified at the
current node and all descendant nodes.

Filters applied while viewing all parts should apply to that complete
descendant scope. Filters applied to a direct-parts section should apply only
to direct membership.

The exact visual treatment of these two result sets belongs in the playground.

## Attribute Model

The current UBOM model is the foundation:

```text
AttributeDef
  defines ID, label, type, unit, and enum vocabulary

AttributeAssignment
  assigns an attribute to a taxonomy node and marks requiredness

PartNumberAttribute
  stores the value on a released part number
```

Definitions and values remain separate. Attribute values belong to part
numbers, not revisions.

Effective attribute resolution is path-based:

1. Start at the root.
2. Apply assignments at each node in path order.
3. Preserve the first-seen ordering for new attributes.
4. Replace the earlier assignment when a deeper node reassigns the same ID.
5. Use the final assignment to determine label, type, unit, and requiredness.

This behavior should be made explicit in the catalog contract and tests.

## Filter Types

The current API supports exact string equality. The catalog will need typed
filters based on `AttributeDef.ValueType`.

Initial target types:

- Enum: one or more selected values.
- String: exact or contains matching, with exact matching preferred for
  controlled values.
- Decimal: minimum, maximum, or inclusive range.
- Boolean: true, false, or any.

Units must be treated as part of the attribute definition. Numeric comparisons
must be numeric, not lexical string comparisons.

The backend query model should eventually move from:

```go
map[AttributeDefID]string
```

to a typed filter representation capable of equality, inclusion, and numeric
range operations.

The URL representation should remain shareable and deterministic. The exact
encoding is open, but it must support multiple active filters and preserve the
taxonomy scope.

## Facets

The filter panel needs more than attribute definitions. It needs values that
are available in the current result scope.

An enum facet could look like:

```json
{
  "id": "tolerance",
  "label": "Tolerance",
  "valueType": "enum",
  "values": [
    { "value": "1%", "count": 18 },
    { "value": "5%", "count": 31 }
  ]
}
```

A decimal facet could look like:

```json
{
  "id": "resistance",
  "label": "Resistance",
  "valueType": "decimal",
  "unit": "ohm",
  "min": "10",
  "max": "1000000"
}
```

Facet counts should be calculated within the current taxonomy scope. Once
multiple filters are active, a facet should preferably calculate its counts
with its own active filter removed so valid alternative choices do not vanish.

The first playground may use fixture counts. Production aggregation and
indexing can follow after the interaction contract is accepted.

## Result Table

The part result view should be context-aware rather than globally fixed.
Columns should be derived from the effective attributes of the selected scope:

```text
Part Number | Resistance | Tolerance | Power | Package | Revision
```

Core columns remain stable. Contextual attribute columns are added only when
the selected scope makes them relevant. Missing values should render clearly.

The result view should show:

- Result count.
- Active scope.
- Active filters.
- Part number as the primary identity.
- Taxonomy context.
- Relevant attributes.
- Revision summary.
- Archive/status state where applicable.

## Proposed Scope Response

The production API should evolve toward a response conceptually similar to:

```json
{
  "taxonomyDefId": "components-v3",
  "node": {
    "id": "resistors",
    "label": "Resistors",
    "partBearing": true
  },
  "path": [
    { "id": "components", "label": "Components" },
    { "id": "passives", "label": "Passive Components" },
    { "id": "resistors", "label": "Resistors" }
  ],
  "categories": [
    {
      "id": "fixed-resistors",
      "label": "Fixed Resistors",
      "partBearing": true,
      "partCount": 42
    }
  ],
  "directParts": [],
  "descendantPartCount": 42,
  "attributes": [],
  "facets": []
}
```

The exact schema is intentionally not locked by this document. The important
contract is that categories, direct parts, descendant scope, effective
attributes, and filter facets are distinct concepts.

## Playground-First Workflow

No production catalog rewrite should begin until the playground has been
reviewed interactively.

The playground should include fixture states for:

- Root home page with several top-level categories.
- Broad category with nested subcategory sections.
- Current node with both children and direct parts.
- Current node with inherited and locally overridden attributes.
- Enum, decimal, string, and boolean filter controls.
- Active filters and clear/reset behavior.
- Empty category scope.
- No direct parts but descendant parts.
- No children but direct parts.
- Mobile category and filter drawers.
- Long labels, many categories, and large result counts.

The playground should allow us to settle:

- Density and tile sizing.
- Whether direct parts appear above or below child categories.
- How category navigation and filters share vertical space.
- Whether filters apply immediately or use an Apply action.
- How selected filters are represented.
- How intermediate part-bearing nodes are labeled.
- How deep taxonomy paths remain understandable.
- How the design behaves at mobile widths.

## Open Decisions

These decisions remain intentionally open for the design conversation:

- Whether the home page shows direct root-level parts, if any.
- Whether category pages show direct parts inline or behind a separate tab.
- Whether the left panel expands the selected branch or stays one level deep.
- Whether category counts represent direct parts or all descendant parts.
- Whether filters apply to all descendants by default or require `View all
  parts` first.
- Whether attribute groups are ordered by taxonomy placement, editorial order,
  or usage.
- Whether attribute overrides may change type or unit, or only requiredness and
  presentation.
- Whether free-text attributes should be filterable in the first production
  version.
- Whether facet counts update immediately or after an Apply action.
- Whether a separate browse taxonomy becomes necessary after real taxonomy data
  is loaded.

## Non-Goals For This Iteration

- Replacing the existing taxonomy editor.
- Changing part-number generation or classification semantics.
- Introducing a separate browse taxonomy before the current model is tested.
- Implementing production facet indexing before the interaction is accepted.
- Copying McMaster branding, assets, or exact visual styling.
- Collapsing taxonomy navigation and BOM navigation into one tree component.
