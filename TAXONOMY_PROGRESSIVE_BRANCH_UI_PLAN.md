# Taxonomy Progressive Branch UI Plan

## Objective

Make the taxonomy tree the single production authoring surface. Sequence
grammar branches should be revealed progressively inside the relevant
taxonomy category context instead of appearing as a separate competing view.

The editor should let a user apply taxonomy categories to grammar branches by
dropping categories into branch-specific drop zones. A category created from a
branch carries that branch condition and becomes the parent context for any
later field or nested-branch authoring.

## Current Problem

The editor currently exposes two distinct concepts:

- The taxonomy-first category tree.
- A separate grammar-oriented branch/structure presentation.

This makes the grammar appear to be a second authoring surface and forces the
user to reconcile grammar structure with taxonomy structure. Branch choices are
also presented globally rather than in the category or grammar context where
they are being used.

## Target Experience

The taxonomy tree remains primary:

```text
Taxonomy tree
  Parts
    Category: Numeric parts
      Grammar context: family
        [field drop zone]
        [existing categories]

        Branches
          Alternative A: family 100..999
            [drop category for this branch]
          Alternative B: family 0..99
            [drop category for this branch]
```

Branch information is shown only where it is relevant to the active grammar
context. There is no separate production grammar editor view.

## Core Semantics

### Field drop

Dropping a category onto a captured field creates or moves a category with a
field predicate:

```text
family = 100..199
```

Existing field-target behavior remains unchanged.

### Branch drop

Dropping a category onto a grammar alternative creates a category with a
persisted branch condition:

```json
{
  "branch": {
    "path": ["root/0/choice[0]"]
  }
}
```

The branch path is the canonical path already used by Phase 2 validation and
classification. Frontend display IDs must not replace or modify this path.

### Branch plus field condition

A category may carry both a branch condition and a field predicate:

```text
branch: root/0/choice[0]
family = 100..199
```

The branch condition scopes the category to a grammar alternative. The field
predicate narrows the captured value within that alternative.

### Category hierarchy

A branch drop inserts the new category as a child of the currently active
taxonomy node. It must not silently insert a detached root-level category just
because the grammar branch belongs to the sequence root.

The new category becomes the context for progressive authoring of:

- Captured fields available within the selected branch.
- Nested choices below that branch.
- Child taxonomy categories and rules.
- Attributes and inherited taxonomy state.

## Progressive Disclosure

### Category context

Each taxonomy category should expose only the grammar context relevant to its
target. If the context contains a choice, show a compact `Branches` section
inside that category block.

### Branch context

Each visible branch should include:

- Alternative label, such as `Alternative A`.
- Human-readable sequence summary.
- Canonical path in a compact secondary treatment where useful.
- Captured fields available under the branch.
- Numeric domains for directly bound ranges when available.
- A category drop zone.
- A nested branch disclosure when the alternative contains another choice.

Literal-only alternatives remain targetable in Phase 2 through branch
conditions even when they expose no captured field.

### Nested choices

Nested choices appear inside the selected alternative, not in a global branch
list:

```text
Alternative A
  Nested choice
    Alternative A1
      [drop category]
    Alternative A2
      [drop category]
```

The visible ancestry must make the resulting canonical branch path clear.

## UI Changes

### Keep

- The taxonomy-first tree and category blocks.
- Field context lanes and existing field drop behavior.
- Category editing, attributes, rules, removal, and drag/drop behavior.
- Canonical branch paths and persisted `branch` conditions from Phase 2.

### Change

- Move branch drop zones from the global `TaxonomyChoiceBranches` panel into
  the relevant taxonomy category/grammar context.
- Render branch context beneath the category or grammar node that owns the
  choice.
- Make branch sections collapsible where nested grammar structures become
  large, while keeping the first relevant disclosure visible.
- Remove the redundant standalone grammar-structure view from the production
  taxonomy editor.
- Retain any grammar-tree components used by playgrounds or diagnostics only;
  do not expose them as a second production authoring workflow.
- Update copy so the user understands whether a drop creates a field
  predicate, a branch condition, or both.

## Interaction Rules

- Palette category drops are accepted by branch drop zones.
- Existing category moves into branch zones should be supported only if the
  resulting branch condition can be updated without losing the category's
  existing field predicates or descendants.
- Rule drops remain valid only when the category has a usable captured-field
  target or an explicitly selected field context.
- A branch-only category may have child categories and attributes even without
  a captured field target.
- Ambiguous legacy field targets remain visibly ambiguous and cannot be used as
  real grammar targets until the user explicitly reassigns them.
- Branch paths must be validated against the exact immutable sequence
  definition before save.
- Branch metadata must not leak into ordinary field predicate fields.

## Data Flow

1. The loaded sequence definition produces a path-aware grammar model.
2. The active taxonomy category determines the visible grammar context.
3. A choice in that context produces nested branch drop zones.
4. A branch drop creates a draft node with:
   - `targetId` set to the branch context used by the editor.
   - `branch.path` set to the canonical backend path.
   - Existing category placement metadata.
5. Field drops and rule edits add field predicates without replacing the branch
   condition.
6. Draft serialization emits the persisted `branch` condition and any field
   predicates, excluding frontend-only display metadata.
7. Server validation checks the branch path against the referenced immutable
   sequence definition.

## Acceptance Criteria

- The production taxonomy editor has one primary authoring view.
- The standalone grammar-structure view is absent from the production flow.
- A category block reveals branch choices only when its relevant grammar
  context contains a choice.
- A category can be dropped into each visible alternative.
- A branch drop inserts the category under the active taxonomy node.
- The created category persists the canonical branch path.
- A branch-only category can contain attributes and descendants.
- A branch category can additionally receive field predicates from captured
  fields exposed within that branch.
- Nested choices reveal nested branch drop zones with visible ancestry.
- Existing field-based taxonomy authoring continues to work unchanged.
- Existing Phase 2 branch validation and classification continue to pass.
- No frontend display path or component-only metadata appears in API payloads.
- The UI remains usable on narrow screens and with large/nested grammars.

## Test Plan

### Frontend model tests

- Resolve direct choice branches beneath a grammar target.
- Resolve nested choice branches and canonical paths.
- Create a branch draft node under a selected taxonomy parent.
- Preserve branch conditions while adding field predicates and rules.
- Serialize branch-only and branch-plus-field categories.
- Preserve branch conditions through reload adaptation.
- Reject invalid branch drop targets and ambiguous field targets.

### Component/browser tests

- Confirm only the taxonomy tree is presented in the production editor.
- Confirm branch sections appear inside the relevant category context.
- Confirm palette category drops create children under the active category.
- Confirm literal-only branches offer branch drops but no field drops.
- Confirm nested branch disclosure and ancestry.
- Confirm branch and field metadata in the outgoing save payload.
- Confirm collapse/expand preserves category labels, branch conditions, and
  field input.

### Backend regression tests

- Validate branch paths against the exact sequence definition.
- Classify branch-only and branch-plus-field categories.
- Reject stale or contradictory nested branch paths.
- Preserve field-only taxonomy behavior.
- Preserve PostgreSQL, SQLite, and memory-store round trips.

## Non-Goals

- Do not introduce a second production grammar editor.
- Do not infer branch conditions from legacy field-only taxonomies.
- Do not silently retarget categories when sequence-definition versions change.
- Do not hide branch identity in an opaque frontend-only marker.
