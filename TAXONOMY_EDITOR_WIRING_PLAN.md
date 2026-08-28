# Taxonomy Editor Wiring Plan

## Goal

Separate the taxonomy editor from hard-coded demo data, then connect it to the
same Settings workflow used by the grammar editor.

Taxonomies are treated as one-to-one with grammars for now, while taxonomy
definitions themselves are immutable. The enforced invariant is:

> Each immutable sequence-definition version has one taxonomy revision lineage
> and one current taxonomy revision.

This allows multiple immutable taxonomy revisions for one grammar while keeping
one active taxonomy easy to resolve. A new sequence-definition version starts a
separate taxonomy revision lineage. Existing taxonomy revisions remain
available for audit and historical records.

## Navigation

The Settings area should expose both sequence definitions and taxonomies:

```text
Settings
|- Sequence definitions
|  |- New sequence definition
|  |- Sequence definition detail
|  `- New/Edit taxonomy
`- Taxonomies
   |- New taxonomy
   `- Taxonomy detail/editor and revision history
```

Proposed routes:

```text
/settings/taxonomies
/settings/taxonomies/new
/settings/taxonomies/:id
```

The sequence-definition detail page should show a contextual action for the
current taxonomy association:

- `New taxonomy` when no taxonomy exists for that grammar.
- `Edit taxonomy` when a current taxonomy exists.

The taxonomy list should show the grammar association, current revision, and
configuration state, for example:

```text
Part number v1    Taxonomy v2 configured    Edit taxonomy · View history
Part number v2    No taxonomy               New taxonomy
```

Historical taxonomy revisions should be reachable from the taxonomy detail
page, but the default list and authoring flow should use the current revision.

`New taxonomy` should require a grammar selection. From a grammar page, the
grammar should be preselected, for example:

```text
/settings/taxonomies/new?seqDef=<id>
```

Do not create a free-standing taxonomy with no associated grammar.

## Editor Boundary

`TaxonomyEditor` should be a reusable editing surface, not a fixture provider.
It should accept:

```ts
export let grammar: TaxonomyGrammar
export let initialAttributeDefinitions: TaxonomyAttributeDefinition[] = []
export let initialDraft: TaxonomyDraftNode
```

The distinction between initial inputs and local state is intentional:

- `grammar` is the immutable grammar structure displayed by the editor.
- `initialAttributeDefinitions` initializes the local definition draft.
- `initialDraft` initializes the local taxonomy working copy.
- `taxonomyDraft` remains internal editor state until persistence is added.

The editor should not import or select from `grammarFixtures`.

Generalize the fixture-specific grammar type into a reusable type such as:

```ts
export type TaxonomyGrammar = {
  id: string
  label: string
  fields: number
  nodes: StructureNode[]
}
```

Fixtures and future API adapters should both produce this shape.

## Playground Fixtures

`TaxonomyEditorPlayground` should own:

- Fixture selection.
- Fixture grammar data.
- Fixture attribute definitions.
- The initial taxonomy draft factory.

Add a factory such as:

```ts
createTaxonomyFixtureDraft(grammar: TaxonomyGrammar): TaxonomyDraftNode
```

The factory must return a fresh draft and should not assume every grammar has
the same fields. It may provide either:

- A minimal empty root draft for every grammar, or
- A seeded example only when the selected grammar contains compatible fields.

The minimal empty root is the clearer default for interaction testing.

Use a keyed editor instance in the playground so changing fixtures resets the
draft explicitly:

```svelte
{#key selectedGrammar.id}
  <TaxonomyEditor
    grammar={selectedGrammar}
    initialAttributeDefinitions={fixtureAttributeDefinitions}
    initialDraft={createTaxonomyFixtureDraft(selectedGrammar)}
  />
{/key}
```

Do not make `TaxonomyEditor` watch a fixture ID or implement fixture-specific
grammar reset logic.

## Attribute Editing

Before adding the real route, finish the local attribute editing contract.

First, make the required-change callback consistent between
`TaxonomyAttributeTableRow` and `TaxonomyEditor`. The callback should pass one
well-defined value, preferably:

```ts
(required: boolean) => void
```

Use a constrained value type instead of `string`:

```ts
type TaxonomyAttributeValueType = 'string' | 'decimal' | 'enum'
```

Add pure validation in `TaxonomyEditorModel.ts` for:

- Empty or whitespace-only attribute labels.
- Duplicate attribute IDs.
- Duplicate labels, if labels are required to be unique.
- `enum` definitions with no values.
- Empty enum values.
- Duplicate enum values.
- Invalid or missing units when units are required for a value type.
- Assignments that reference missing definitions.
- Duplicate assignments on one taxonomy node.

Return structured errors keyed by definition or assignment ID. The row
component should render row-level errors but should not own domain validation.

Defer full enum-value editing until this validation model exists. Otherwise,
the UI will accumulate state without a defined error or serialization model.

## Grammar Changes

For the playground, changing grammar fixtures should reset the draft. Do not
implement compatibility-preserving migration yet.

For the real editor route, a grammar change should not silently reuse a draft
whose targets are invalid for the new grammar. Any future preserve-compatible
behavior should be explicit and validated.

A new immutable sequence-definition version should receive a new taxonomy
revision lineage. Copying an existing taxonomy into the new sequence-definition
version can be added later as an explicit action, and must still produce a new
immutable taxonomy definition.

## API Integration

The current API foundations include:

- `getSeqDefs()`
- `getTaxonomyDefs()`
- `getTaxonomyDef(id)`
- `createTaxonomyDef()`
- `TaxonomyDefinition.seqDef`
- `TaxonomyNodeInput`

The taxonomy list should join taxonomy summaries with sequence-definition
metadata so it can display useful grammar labels and versions. A future API
response may provide this association directly if the client-side join becomes
awkward.

The real editor route should load and adapt backend data into the editor's
reusable `TaxonomyGrammar`, attribute-definition, and draft shapes. Keep API
loading out of the reusable editor component.

Taxonomy definition responses and summaries should expose revision lineage
metadata using names consistent with sequence definitions:

```ts
type TaxonomyDefinition = {
  id: string
  definitionId: string
  version: number
  parentId?: string
  seqDef: string
  attributeDefs?: TaxonomyAttributeDefinition[]
  root: TaxonomyNodeInput
}
```

The API should expose the current taxonomy for a sequence definition and the
revision history for a taxonomy lineage. The client should not calculate the
next revision number.

## Persistence

Taxonomy definitions are immutable, matching sequence-definition behavior. The
editor never updates an existing taxonomy definition in place.

For a new taxonomy:

```text
create taxonomy definition
  seqDef = part-number-v1
  parentId = none
  version = 1
```

When editing the current taxonomy:

```text
create next taxonomy revision
  seqDef = part-number-v1
  parentId = part-taxonomy-v1
  version = 2
```

Saving must:

- Leave the previous taxonomy definition unchanged.
- Validate the new taxonomy against the referenced sequence definition.
- Assign the next revision server-side and transactionally.
- Advance the current taxonomy association for that sequence-definition.
- Return the new taxonomy ID and revision metadata.
- Reject attempts to overwrite an existing taxonomy ID.

The sequence-definition document itself remains immutable. The current taxonomy
pointer must live in taxonomy persistence or be resolved by a server-owned
latest-revision query; it must not be written back into the sequence definition.

The current association should be explicit if the system later needs drafts,
publication, rollback, or staged revisions. A server-owned latest-revision
query is sufficient for the first implementation if the persistence layer can
guarantee ordering and uniqueness.

Historical behavior:

- Existing parts retain the taxonomy revision used when they were created.
- Existing part-number drafts retain their taxonomy revision unless explicitly
  migrated.
- New part-number drafts use the current taxonomy revision.
- Taxonomy revisions cannot be edited or deleted.
- Revision history remains queryable for auditability.

The save request should map the validated local draft to
`TaxonomyNodeInput`/the taxonomy-definition creation shape. Server-side
validation must enforce the revision and current-association invariants.

## Implementation Order

1. Fix and formalize the attribute table callback contract.
2. Introduce the reusable `TaxonomyGrammar` type.
3. Add `createTaxonomyFixtureDraft` and move fixture state to
   `TaxonomyEditorPlayground`.
4. Change `TaxonomyEditor` to accept grammar, initial definitions, and initial
   draft.
5. Reset the playground editor with a keyed component when the grammar changes.
6. Add model-level taxonomy and attribute validation with row-level error
   display.
7. Add `/settings/taxonomies` and the new/detail editor routes.
8. Add the sequence-definition contextual `New/Edit taxonomy` action.
9. Add backend loading and adapters for real grammar and taxonomy data.
10. Add taxonomy revision metadata: `definitionId`, `version`, and optional
    `parentId`.
11. Change taxonomy creation into server-owned next-revision creation without
    allowing in-place updates.
12. Add current-revision resolution and taxonomy revision history APIs.
13. Add a save action that creates a new immutable taxonomy revision and
    advances the current taxonomy association atomically.
14. Verify that old taxonomy revisions remain usable for existing parts and
    drafts, and add server-side invariant tests.

The playground remains fixture-backed throughout. The real Settings route is
the only consumer that should eventually load persisted data.
