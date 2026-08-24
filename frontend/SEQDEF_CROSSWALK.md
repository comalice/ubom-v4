# SeqDef UI Cross-Walk

The sequence builder uses the reference application's component vocabulary for
interaction. UBOM persists a typed sequence AST. The UI must adapt to that AST
without creating a second grammar.

## Identity

| Reference UI | UBOM model | Notes |
| --- | --- | --- |
| Sequence name | `SeqDef.ID` | The name is the durable definition ID. It is immutable after creation for now. |
| Active component order | `ConcatNode.Children` | The top-level canvas is always a concatenation. |
| Inactive component zone | None | Transient editor state. Inactive items are not persisted. |
| Selected component/options | None | Transient editor state. |

## Frontend State

All editor state is transient: active and inactive component lists, selection,
draft dialog and drag state, name, bindings, and request status/errors. Only the
name and active components are persisted: the name is `SeqDef.ID`, and active
components serialize to the AST. Bindings are request input for preview and
authoring, not persisted sequence-definition state.

## Components

| Reference UI | UBOM AST | Status |
| --- | --- | --- |
| Literal | `LiteralNode{Text}` | Direct mapping. |
| Named literal | `BindNode{Name, Child: LiteralNode}` | Direct mapping. |
| Numeric range | `RangeNode{Min, Max, WidthValue}` | Direct mapping when numeric. The editor names these `min`, `max`, and `width`; its UI labels remain reference-friendly. |
| Named numeric range | `BindNode{Name, Child: RangeNode}` | Direct mapping and usable by part drafts. |
| Select | `ChoiceNode` of `LiteralNode` children | The reference label is Select; the editor kind is `choice` and its options are adapted to AST children. `ChoiceNode` is the sole AST representation for alternatives. |
| Named select | `BindNode` around the `ChoiceNode` | Direct structural mapping. |
| Places | `PlaceSequenceNode{Alphabets}` | Direct mapping. The editor exposes one alphabet per line. |
| Sequence | No equivalent | Omitted from the builder until cycling semantics have a persistent model. |
| Character range | No equivalent | Omitted from the builder; `RangeNode` is numeric only. |
| Step | No equivalent | Omitted from the builder. |
| Description | No equivalent | UI-only at present. |
| Option descriptions | No equivalent | UI-only at present. |

## Lifecycle

```text
edit component canvas
  -> serialize ConcatNode AST
  -> backend validates SeqDef
  -> backend persists SeqDef
  -> backend exposes BindNode fields
  -> part draft uses those fields
```

The frontend serializes the editor's active components and sends bindings with
preview or authoring requests. The backend remains authoritative for validation,
generation, parsing, and uniqueness.

## Required Follow-Up Slices

1. Add a persisted representation for component metadata only if descriptions
   and option descriptions become product requirements.
2. Model character ranges, likely as a new typed AST node, before exposing them
   as a real persisted option.
3. Add SeqDef detail/update semantics before presenting existing definitions as
   editable.
