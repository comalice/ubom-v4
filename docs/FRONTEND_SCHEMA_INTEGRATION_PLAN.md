# Frontend Schema Integration Plan

## Purpose

Turn Settings > Grammar Editor from a local interaction prototype into the
frontend authoring surface for the server's explicit SeqDef and taxonomy
contracts.

The frontend should reflect the backend model directly at the authoring
boundary. The server remains authoritative for validation, interpretation,
generation, persistence, and immutable definition identity.

## Contract Decisions

The editor model should be normalized around the available backend node types:

- `literal`
- `concat`
- `choice`
- `range`
- `range_radix`
- `bind`
- `places`

Specific decisions:

- `chars` is represented as `places`, because its intended meaning is a
  one-position selection from a supplied character set.
- `pad` is a character. Normalize the frontend and HTTP contract to represent
  it as a one-character string, which is sufficient for current needs.
- `alpha` is not a separate backend concept. It is a constrained form of
  `range_radix`, and the editor should represent it using the range-radix
  structure rather than inventing a separate wire node.
- The editor must not display blocks that cannot be serialized to a valid
  backend AST.
- Local parsing and generation may provide immediate interaction feedback, but
  server preview is authoritative.

## Current State

The Grammar Editor is available at:

- `/settings/sequence-definitions/grammar-editor`
- `/playground/grammar-editor`

It provides local drag/drop editing, nested bindings, choices, local parsing
and generation, and a structure map. The editor model is normalized to backend
node kinds and covered by frontend tests. Structural validation reports paths
and displays visible errors; incomplete Binding editor states no longer crash.
Strict AST conversion is retained for valid documents only, and saves/previews
are gated by structural validity.

The authoritative Grammar Editor flow is complete candidate validation via:

```text
POST /api/seq-defs/preview/candidate
{root, value}
```

Local Test parse remains available. The obsolete binding-generation preview is
not the current flow.

SeqDef lifecycle support is complete for the current scope. Settings has a
dedicated list at `/settings/sequence-definitions`, with New at
`/settings/sequence-definitions/new` and existing definitions at
`/settings/sequence-definitions/:id`. The flow supports full-definition GET,
dirty state, create, duplicate/conflict display, read-only immutable versions,
and save-as-new-version. Save-as-new-version has been implemented and its entry
point was exercised in the browser; the complete save interaction is not being
claimed as fully user-tested here.

The existing API client and Settings flow now support:

- Creating SeqDefs.
- Candidate validation for an unsaved SeqDef.
- Listing SeqDef summaries.
- Loading full SeqDef definitions.
- Creating new immutable versions.
- Loading bindable fields.

Taxonomy definition and part-authoring APIs remain available for their existing
server-backed flows but are not part of the completed frontend integration.

The existing part-authoring flow is already server-backed and should become the
downstream consumer of the integrated SeqDef and taxonomy authoring flows.

## Completed Through SeqDef Versioning

The following work is complete:

- Editor model normalization to backend kinds, including AST conversion tests.
- Structural validation with document paths and visible errors.
- Candidate validation through `POST /api/seq-defs/preview/candidate` with
  `{root,value}`; Local Test parse remains separate.
- Settings list, new, and existing SeqDef routes.
- Full-definition GET at `/api/seq-defs/{id}`.
- Dirty state, create, duplicate/conflict display, read-only immutable versions,
  and save-as-new-version implementation.
- Binding-state crash handling and structural-validity gates for saves/previews.

## Implementation Order

### 1. Normalize the editor model (Completed)

The UI model and palette use the backend node vocabulary, and conversion is
tested for the supported editor states. Keep the existing tests as regression
coverage when making further editor changes.

- Replace the current `alpha` and `chars` editor-only kinds with backend-aligned
  `range_radix` and `places` representations.
- Keep editor-only IDs and layout state separate from the AST.
- Conversion to `SeqDefNode` covers every supported visible block type.
- `range_radix` children are represented by the normalized editor structure.
- Range padding is normalized to a one-character string at the frontend/API
  boundary.
- Unsupported blocks are excluded from the palette.

Completed frontend model coverage includes:

- Every supported node kind.
- Nested `bind` nodes.
- Multi-expression `choice` alternatives.
- `range_radix` structures.
- `places` character sets.
- Empty and invalid editor states.
- Padding serialization.

### 2. Add editor validation and trustworthy state (Completed)

Model-level validation reports paths into the editor document, for example:

```text
root[1].bind
root[2].choice[0]
```

At minimum, report:

- Empty choices.
- Missing bind children.
- Empty binding names.
- Duplicate binding names.
- Invalid range bounds.
- Invalid widths.
- Invalid character/radix definitions.
- Unsupported or incomplete nodes.

The prototype-only state has been replaced with:

- Dynamic block counts.
- Dirty state.
- Preview status.
- Structural validation status.
- Server validation status.
- Save/conflict errors.

### 3. Add authoritative candidate validation (Completed)

Connect Settings > Grammar Editor to:

```text
POST /api/seq-defs/preview/candidate
```

The page should:

- Submit `{root,value}` after local structural validation passes.
- Display candidate validation results and visible errors.
- Distinguish editor-structure errors from server/domain errors.
- Retain Local Test parse as a separate local interaction.

Persistence is gated by local structural validity. Candidate validation remains
the authoritative on-demand check for a complete example; requiring a successful
candidate-preview round trip before saving is future hardening work.

### 4. Introduce the SeqDef lifecycle (Completed for current scope)

The Settings page-level integration wrapper and dedicated list page are in
place.

Suggested routes:

- `/settings/sequence-definitions/new`
- `/settings/sequence-definitions/:id`

The lifecycle needs to represent:

- New unsaved definition.
- Working local document.
- Server-previewed document.
- Created immutable definition.
- Read-only existing version.
- Save-as-new-version.
- Dirty and clean state.
- Duplicate ID and conflict errors.

The full-definition GET is available at `/api/seq-defs/{id}` and existing
versions can be opened read-only. Saving creates a new version through
`POST /api/seq-defs/{id}/versions`.

The version contract is:

- Each version has record ID `id`, stable `definitionId`, positive `version`,
  and immediate `parentId`.
- Generated IDs use `<definitionId>-v<version>`.
- Legacy records default to `definitionId=record ID`, `version=1`, and no
  parent.
- Duplicate `(definitionId, version)` values are rejected.

### 5. Build taxonomy authoring around a SeqDef (Deferred)

After SeqDef polish and hardening, add taxonomy authoring as a separate surface.

The flow should support:

- Selecting a SeqDef.
- Loading bindable fields from `/api/seq-defs/{id}/fields`.
- Defining attributes.
- Building recursive taxonomy nodes.
- Adding exact predicates.
- Adding numeric-range predicates.
- Assigning inherited and required attributes.
- Previewing classification with example values.
- Submitting an immutable taxonomy definition.

The frontend should rely on server validation for cross-node invariants such as
unknown fields, conflicting ancestor predicates, and overlapping siblings.

### 6. Connect taxonomy to part authoring (Deferred)

The target user flow is:

```text
SeqDef -> Taxonomy -> Taxonomy Node -> Part Draft -> Allocation -> Release
```

The existing `Authoring.svelte` flow should be extended rather than replaced.
It should use the selected taxonomy node to show effective attributes, accept
bindings and attributes, preview a candidate, allocate within the taxonomy
context, and release the part.

### 7. Improve reusable editor UX (Next polish/hardening)

Keep the Playground route as a lightweight demonstration of editor mechanics,
but use the Settings route for lifecycle and server integration.

Separate responsibilities over time:

- `GrammarEditorModel.ts`: pure document operations, AST conversion, and
  validation.
- `GrammarEditor.svelte`: reusable editing surface.
- `SeqDefEditorPage.svelte`: route, lifecycle, preview, save, and errors.
- `SeqDefPreviewPanel.svelte`: authoritative server result.
- `TaxonomyEditor.svelte`: recursive taxonomy authoring.
- `TaxonomyPredicateEditor.svelte`: field-aware predicate controls.

Also address:

- Keyboard alternatives to drag/drop.
- Focusable structure-map navigation.
- Accessible preview and validation announcements.
- Readable text and input sizes for dense authoring.
- Mobile behavior for nested editors.

## Backend/API Follow-Ups

SeqDef contract work for the current scope is complete. Remaining backend/API
follow-ups are primarily for deferred taxonomy and future hardening:

- Expose full taxonomy documents for loading existing definitions.
- Keep typed taxonomy predicates as the only normal authoring representation;
  retain legacy `Matches` only for compatibility and migration.
- Continue SeqDef polish/hardening around error handling, accessibility, and
  lifecycle edge cases before starting taxonomy authoring.

## Verification Strategy

Completed verification includes:

- 19 frontend tests covering normalized model/AST conversion, validation, and
  relevant API/lifecycle behavior.
- Frontend build/typecheck.
- Backend Go tests.
- Diff check.
- Firefox browser verification of list/new/existing navigation, candidate
  invalid and success states, save-as-new-version entry, direct navigation
  between existing definitions, and a clean pass with no console errors.

Remaining verification should focus on SeqDef polish/hardening. In particular,
browser entry into save-as-new-version was exercised, but this plan does not
claim the full save-as-new-version interaction was user-tested.

The backend contract tests remain the wire-level authority. Existing end-to-end
schema coverage should continue to verify the complete flow through both stores
and HTTP as the lifecycle is hardened.

## Deferred: Taxonomy

Taxonomy remains a planned, separate effort after SeqDef polish/hardening. Do
not treat the taxonomy items below as completed or as the next active slice:

- Taxonomy definition loading and authoring.
- Recursive taxonomy nodes and typed predicates.
- Classification preview and immutable taxonomy submission.
- Connecting taxonomy nodes to part authoring.

## Recommended First Slice

The next active slice is SeqDef polish/hardening, not taxonomy. Prioritize:

1. Harden SeqDef lifecycle edge cases and conflict/error handling.
2. Exercise the complete save-as-new-version flow in the browser.
3. Improve accessible status, validation, and read-only/version presentation.
4. Preserve coverage with frontend tests, build/typecheck, backend Go tests,
   browser verification, and diff checks.

Only after this hardening is stable should taxonomy authoring begin.
