# Branch-Aware Taxonomy Editor: Phase 1 and Phase 2 Plan

## Objective

Make sequence-definition choice branches visible and understandable in the
taxonomy editor without changing the persisted taxonomy API or backend
classification model.

Phase 1 is a safe UI/model improvement. It must not claim to persist branch
identity that the current taxonomy contract cannot represent.

## Scope

### Included

- Preserve choice and alternative metadata in the frontend taxonomy grammar
  model.
- Expose branch labels, summaries, paths, and captured fields in the taxonomy
  authoring surface.
- Distinguish repeated binding occurrences across alternatives, including their
  numeric domains when available.
- Keep categories targetable only through the existing field-based taxonomy
  contract.
- Make literal-only branches visible but explicitly non-targetable.
- Detect persisted field targets that map to multiple grammar occurrences and
  show them as ambiguous instead of silently resolving them to the root or the
  first matching occurrence.
- Preserve existing field-based serialization and compatibility with existing
  taxonomy revisions.
- Add frontend model/component coverage for choices, repeated bindings,
  literal-only alternatives, nested choices, and ambiguous reload behavior.
- Add a QA runbook under `test/qa/` covering branch visibility and safe
  targeting behavior.

### Excluded

- No backend schema changes.
- No new persisted grammar-path or branch-condition fields.
- No changes to server parsing or classification semantics.
- No direct taxonomy targeting of literal-only alternatives.
- No automatic migration or reinterpretation of existing taxonomy revisions.

## Design Constraints

- Existing taxonomy predicates remain field-based (`field`, `kind`, `value` or
  numeric bounds).
- Frontend grammar paths are display and targeting metadata only; they are not
  sent to the API.
- A category target must resolve to a captured field occurrence that is valid
  under the current field-based contract.
- If a persisted predicate cannot be mapped uniquely to a grammar occurrence,
  the editor must preserve the predicate but expose the target as unresolved or
  ambiguous.
- Existing unique-field taxonomies must retain their current behavior.
- Numeric-domain disambiguation may continue to resolve a repeated field only
  when the match is provably unique.

## Proposed Changes

### Grammar metadata

Extend frontend-only grammar metadata to represent:

- Stable traversal path for the current loaded sequence definition.
- Choice/alternative ancestry.
- Alternative label and summary.
- Captured binding name, where present.
- Numeric domain, where directly available.
- Whether a branch is literal-only or otherwise lacks a targetable binding.

The metadata should be derived from the immutable sequence definition and
should not alter the API representation.

### Taxonomy editor presentation

- Render choice alternatives as visibly distinct branch sections or lanes.
- Show branch summaries before the fields/categories within each branch.
- Keep existing progressive context-lane behavior for captured fields.
- Display repeated fields as separate occurrences rather than one silently
  deduplicated field where the UI can do so safely.
- Label literal-only alternatives as visible but unavailable for category drops.
- Provide an explicit ambiguous-target state for persisted predicates that do
  not map uniquely.

### Target resolution

Replace silent name-only assumptions with a resolution result that can express:

- Unique target.
- Multiple candidate occurrences.
- No candidate.
- Literal-only/non-targetable branch.

Existing numeric-domain inference remains valid only for a unique candidate.
Ambiguous candidates must not silently become `grammar-root` or the first
traversal match.

### Serialization

Keep taxonomy API serialization unchanged. Field-based predicates continue to
be emitted exactly as before. UI-only branch metadata must never be included
in `TaxonomyDefinitionInput` during Phase 1.

## Acceptance Criteria

- A sequence definition with two choice alternatives renders both alternatives
  in the taxonomy editor with distinct labels and summaries.
- A repeated binding name appears with enough branch context to distinguish its
  occurrences when the domains differ.
- A literal-only alternative is visible and clearly not a valid field target.
- A category can still be added to a uniquely resolved captured field.
- A persisted predicate with multiple possible target occurrences is surfaced
  as ambiguous and is not silently moved to the root.
- Existing unique-field taxonomy editing, serialization, and save behavior are
  unchanged.
- Nested choices retain visible ancestry and do not collapse into an
  indistinguishable global field list.
- Frontend tests cover the model and relevant rendered editor behavior.
- The new `test/qa/` runbook covers the above cases and records expected pass
  and fail outcomes.
- Full frontend tests, typecheck, build, and backend tests pass.

## Implementation Sequence

1. Update frontend grammar metadata and target-resolution types.
2. Update taxonomy editor rendering and drop-target eligibility.
3. Update draft loading/ambiguous-target presentation without changing API
   serialization.
4. Add and run frontend regression tests.
5. Review the implementation against this plan and acceptance criteria.
6. Commit the accepted Phase 1 implementation.
7. Add the Phase 1 QA runbook under `test/qa/`.
8. Launch a separate agent to work through the runbook against the local app.
9. Review the QA agent report and address any discovered defects before a
   follow-up commit.

## Review Checklist

- No branch metadata leaks into persisted taxonomy input.
- No existing unique-field behavior regresses.
- No silent fallback from ambiguous target to root or first candidate.
- Literal-only branches cannot receive invalid field-based categories.
- Nested and repeated alternatives remain intelligible.
- Existing dirty-worktree changes are preserved and kept separate from the
  Phase 1 implementation.

## Phase 2: Persisted Branch-Aware Taxonomy

### Objective

Make grammar branch identity a durable part of taxonomy authoring,
validation, classification, and revision history. Phase 2 addresses cases
that Phase 1 intentionally leaves unresolved, including literal-only
alternatives, repeated binding names, nested choices, and branches that
produce indistinguishable field values.

### Scope

Phase 2 may change the domain model, API contracts, persistence formats,
parser results, and classification behavior. It must retain compatibility
with existing field-only taxonomy revisions.

### Domain model

- Give sequence-definition AST nodes and choice alternatives stable identities.
- Define a typed taxonomy condition model that can represent at least:
  - Exact captured-field conditions.
  - Numeric captured-field ranges.
  - Grammar-path or alternative identity conditions.
  - Literal/path conditions for branches with no capture.
- Allow a taxonomy node to combine a branch condition with ordinary field
  predicates.
- Define whether conditions are conjunctive by default and document any
  future `all`/`any`/`not` composition separately.
- Keep legacy `field` predicates valid and semantically unchanged.

### Parser and classification

- Extend parser results to retain the selected sequence grammar path,
  including nested choice alternatives.
- Preserve parser choice ordering and existing first-success behavior unless
  the new validation explicitly rejects ambiguous sequence definitions.
- Make classification evaluate both grammar-path conditions and captured-field
  conditions.
- Include selected grammar-path information in classification results when it
  is useful for diagnostics, authoring, or audit output.
- Ensure branch conditions work when alternatives contain only literals and
  when repeated field names occur in different alternatives.

### API and persistence

- Extend taxonomy authoring request/response types with branch-aware
  conditions while preserving legacy predicate fields for existing data.
- Persist stable sequence-node/alternative identities with sequence
  definitions or in a versioned AST representation.
- Persist taxonomy conditions without losing legacy field-only predicates.
- Add equivalent PostgreSQL and SQLite migrations or versioned JSON changes.
- Update memory, SQLite, and PostgreSQL stores together and extend store
  contract tests.
- Validate every persisted grammar-path condition against the exact immutable
  sequence definition referenced by the taxonomy.

### Revision and migration policy

- Existing taxonomy revisions remain readable and classify using their
  existing field-only semantics.
- Existing taxonomies must not receive inferred branch conditions unless the
  inference is provably unique.
- A taxonomy revision remains tied to its immutable sequence-definition
  version.
- Copying a taxonomy to a new sequence-definition version requires explicit
  path mapping and validation.
- Reordering alternatives or changing stable node identity must not silently
  retarget an existing taxonomy.
- Invalid, stale, or ambiguous branch references must produce visible,
  actionable validation errors.

### Phase 2 acceptance criteria

- Literal-only alternatives can be represented and classified when explicitly
  targeted.
- Repeated binding names across alternatives can be targeted without frontend
  heuristics.
- Nested choice paths survive save, reload, and classification.
- A branch condition and a field condition can coexist on one taxonomy node.
- PostgreSQL, SQLite, memory, API, and parser behavior is consistent.
- Existing field-only taxonomy definitions continue to load, validate, save as
  new revisions, and classify without behavior changes.
- Sequence-definition and taxonomy revisions reject stale or invalid path
  references.
- Tests cover parser paths, API round trips, migrations, store contracts,
  classification, and editor reload behavior.

### Phase 2 implementation sequence

1. Specify stable sequence AST identity and branch-condition semantics.
2. Update domain types and parser results.
3. Add API and persistence representations with legacy compatibility.
4. Add PostgreSQL and SQLite migration/round-trip support.
5. Update validation and classification.
6. Update frontend authoring, reload, and branch-aware serialization.
7. Add end-to-end tests across stores, API, parser, and UI.
8. Add migration and compatibility QA coverage.
9. Review behavior on old revisions and changed sequence-definition versions.

### Phase 2 risks and decisions

- Decide whether stable AST identities are generated at creation time and
  immutable, or deterministically derived from structural identity.
- Decide how branch conditions behave if two alternatives are semantically
  equivalent or produce the same field values.
- Decide whether a taxonomy may target a grammar path without any captured
  field.
- Decide how taxonomy copies map paths between sequence-definition versions.
- Preserve exact-text semantics for values such as `001`; do not coerce them
  to numeric predicates accidentally.
- Keep branch identity separate from ordinary field predicates so the current
  one-predicate-per-field validation rule does not prevent valid branch plus
  field conditions.
