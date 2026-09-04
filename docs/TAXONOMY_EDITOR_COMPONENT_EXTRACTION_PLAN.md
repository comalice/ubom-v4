# Taxonomy Editor Component Extraction Plan

## Purpose

Extract the hard-coded taxonomy editor visuals into reusable components so the
taxonomy editor can be formalized as a UI surface before any taxonomy backend
integration begins.

The first stage is visual only. Components may accept display data and
callbacks, but they should not load, validate, classify, save, or submit
taxonomy definitions.

## Reference Pattern

Follow the grammar editor's division of responsibilities:

- `SequenceExampleTile.svelte` owns the palette tile visual and drag affordance.
- `GrammarEditorBlock.svelte` owns the rendered block shell and block-specific
  editing surface.
- `GrammarEditor.svelte` composes the palette, canvas, and editor state.

Taxonomy should use the same separation:

- A taxonomy palette tile component owns the palette representation.
- Taxonomy block components own block-specific visual editing surfaces.
- The taxonomy editor composes those components around the immutable grammar
  tree.

## First Extraction Slice

### `TaxonomyExampleTile.svelte`

Create a taxonomy equivalent of `SequenceExampleTile.svelte` for palette and
component-showcase tiles.

It should own:

- Taxonomy type marker/color.
- Title.
- Palette description.
- Palette styling and, if retained, the drag affordance.
- Optional remove affordance if the same tile is used in a dropped state.

It should accept a small typed tile model, for example:

```text
type: category | rule | attribute
title
description
```

The palette metadata should not be duplicated between the taxonomy editor and
the component showcase.

### `TaxonomyCategoryBlock.svelte`

Extract the category block body currently embedded in
`TaxonomyEditorComponentsPlayground.svelte` and repeated in
`TaxonomyGrammarTree.svelte`.

It should compose with `BlockShell` and own the visual fields for:

- Category label.
- Node ID.
- Predicate summary when present.
- Add-child-category affordance.

The first version can use static values or props. It should not mutate a
taxonomy document yet.

The same component should be used for:

- The category block catalogue example.
- A filled taxonomy grammar drop target.
- Future taxonomy editor nodes.

### `TaxonomyRuleBlock.svelte`

Extract the rule block body currently embedded in
`TaxonomyEditorComponentsPlayground.svelte`.

It should compose with `BlockShell` and own the visual controls for:

- Grammar field.
- Predicate kind.
- Exact value.
- Numeric-range representation when the predicate kind is numeric range.
- Predicate summary.

The component should support both documented predicate forms:

```text
category = RES
number in 100..199
```

It should remain local and visual. Server validation of field names, ancestor
compatibility, and sibling overlap is deferred.

## Deferred Attribute UI

Do not create `TaxonomyAttributeBlock.svelte` in this extraction slice.

Attributes are expected to be edited through an attribute table surfaced by a
button, likely as a modal or popout. The table may also be represented in the
grammar tree later, but that is a separate design step.

The current attribute block example in
`TaxonomyEditorComponentsPlayground.svelte` should therefore be treated as a
temporary visual reference, not as the beginning of a reusable block API.

The eventual attribute surface will likely need components such as:

- `TaxonomyAttributePanel.svelte` or `TaxonomyAttributePopover.svelte`.
- `TaxonomyAttributeTable.svelte`.
- `TaxonomyAttributeRow.svelte`.

Those should be designed after the interaction model for opening, positioning,
editing, dismissing, and applying the table is agreed.

## Grammar Tree Extraction

### `TaxonomyGrammarNode.svelte`

Extract recursive grammar-node rendering from `TaxonomyGrammarTree.svelte`.

The current tree manually renders root, child, and grandchild levels. The
taxonomy editor will need arbitrary depth, so the node renderer should be
recursive and receive a node plus its visual state.

It should own:

- Grammar node label and summary.
- Child-node recursion.
- Composition with the taxonomy drop target.

### `TaxonomyGrammarDropTarget.svelte`

Create a taxonomy-specific wrapper around the shared `DropZone` component.

It should own taxonomy presentation such as:

- Target label.
- Inactive, new, move, or filled visual state.
- Whether the target contains a taxonomy overlay.
- Composition of a filled target with `TaxonomyCategoryBlock`.

The generic `DropZone.svelte` should continue to own marker rendering and drag
state. Taxonomy-specific meaning belongs in this wrapper.

The existing duplicated category/high-power markup in
`TaxonomyGrammarTree.svelte` should be removed in favor of the extracted block
and target components.

## Demo-Only Root Drop Zone

The standalone root/sibling drop-zone visualization in
`TaxonomyEditorPlayground.svelte` is not part of the first settings editor
surface.

It may remain in the taxonomy editor components playground as a static demo,
analogous to the grammar editor's playground-only examples. It should not drive
the structure of the eventual Settings taxonomy editor and does not need to be
promoted to a reusable component at this stage.

## Page Composition After Extraction

The taxonomy component catalogue should compose:

```text
TaxonomyEditorComponentsPlayground
  -> TaxonomyGrammarTree
  -> TaxonomyExampleTile
  -> TaxonomyCategoryBlock
  -> TaxonomyRuleBlock
  -> TaxonomyExampleTile (attribute palette/reference tile only)
  -> attribute table reference only, until its popout/modal design is settled
```

The future taxonomy editor should compose:

```text
TaxonomyEditor
  -> TaxonomyPalette
      -> TaxonomyExampleTile
  -> TaxonomyGrammarTree
      -> TaxonomyGrammarNode
          -> TaxonomyGrammarDropTarget
              -> TaxonomyCategoryBlock
              -> TaxonomyRuleBlock
  -> attribute palette action, with its destination deferred to the attribute
     table/popout design
```

The attribute table/popout is intentionally omitted from this composition
until its interaction design is defined.

## Fixture Data

Sample content should eventually move out of component markup into typed local
fixtures or props:

- Grammar tree sample data currently in `TaxonomyGrammarTree.svelte`.
- Palette item metadata currently in `TaxonomyEditorPlayground.svelte`.
- Category and rule example values currently in the component showcase.

This is data extraction, not backend integration. Fixtures should remain local
until the taxonomy editor has a defined server contract and lifecycle.

## Out Of Scope

This plan does not include:

- Taxonomy API calls.
- Loading or submitting taxonomy definitions.
- Server-side predicate validation.
- Classification preview.
- Immutable taxonomy versions.
- Part-authoring integration.
- Attribute table modal/popout implementation.
- Attribute display inside the grammar tree.
- Production root/sibling drop-zone behavior.

## Recommended Order

1. Extract `TaxonomyExampleTile.svelte` and use it for palette tiles and the
   component catalogue.
2. Extract `TaxonomyCategoryBlock.svelte` and reuse it in filled grammar targets.
3. Extract `TaxonomyRuleBlock.svelte`.
4. Extract recursive `TaxonomyGrammarNode.svelte`.
5. Extract `TaxonomyGrammarDropTarget.svelte` around `DropZone`.
6. Move taxonomy sample data into typed local fixtures.
7. Reassess the taxonomy editor page composition before designing the attribute
   modal/popout.
