# Greenfield Reimplementation Plan

This plan turns the schema pipeline design into a sequence of small,
reviewable implementation steps. The target is a greenfield reimplementation,
not a blind copy of accidental behavior in the current code.

The semantic model to preserve is described in `SCHEMA_PIPELINE.md`. The
current implementation is reference material only. Existing behavior should
be preserved when it expresses a domain invariant and deliberately changed
when it is unfinished, accidental, or inconsistent with the design.

## Working Protocol

Each step follows the same loop:

```text
1. Select one unchecked step.
2. Launch an implementation agent with a narrow scope.
3. Agent changes code and adds or updates focused tests.
4. Review the diff for design fit, regressions, and scope creep.
5. Run the relevant focused tests and the full suite.
6. Accept or request corrections.
7. On acceptance, commit only that step.
8. Mark the step complete and select the next dependency-ready step.
```

An implementation step should be small enough that its behavior can be
explained in one paragraph and reviewed in one diff. Agents must not commit;
the coordinating review accepts and commits changes.

## Baseline

The pre-reimplementation code baseline is tagged in Git as:

```text
reimplementation/baseline
```

The tag points to the existing committed code. It does not include uncommitted
worktree changes or this plan and the schema sketch until those documents are
accepted and committed separately.

## Design Decisions

These decisions are the foundation for the implementation steps:

- `SeqDef` defines structural validity and construction.
- Parsing produces structured captured fields.
- Taxonomy classifies parsed fields and refines meaning along a tree path.
- A child taxonomy predicate is cumulative with its ancestors.
- Taxonomy predicates may express exact values and numeric ranges.
- A schema composes SeqDef, taxonomy, ordering where applicable, and policy.
- Syntactic validity and classification are distinct results.
- Root-only classification may be allowed or rejected by schema policy.
- `next` produces an ordered candidate; allocation finds an unused candidate.
- The store is authoritative for uniqueness and persistence.
- Definition versions are immutable once referenced by a part number.
- Generated values for a taxonomy context must classify to that context or an
  allowed descendant.
- Ambiguous sibling classifications are rejected by default; intentional
  priority must be explicit.

## Implementation Steps

### 0. Accepted Documentation Baseline

- Commit `SCHEMA_PIPELINE.md` and this plan.
- Preserve the reimplementation tag as the pre-code-change reference.

Acceptance:

- Both design documents are committed.
- The baseline tag remains available and points to the pre-reimplementation
  commit.

### 1. Domain Value And Error Foundations

Define shared domain errors, identifiers, field names, and small value types
needed by the new implementation. Establish whether errors are classified by
kind rather than by string matching.

Acceptance:

- Public identifiers and error categories have documented semantics.
- No SeqDef or taxonomy behavior changes yet.
- Focused unit tests cover error identity and value validation.

### 2. SeqDef AST And Structural Validation

Reimplement the SeqDef node model and validation independently of parsing,
generation, taxonomy, and persistence.

Acceptance:

- The AST represents literals, concatenation, choice, bind, places, ranges,
  and range-radix structure.
- Invalid definitions fail deterministically.
- Validation does not contain domain vocabulary or taxonomy concerns.

### 3. SeqDef Parsing And Captured Fields

Implement full-input parsing with choice backtracking and structured captures.
Retain source text and normalized value where those differ.

Acceptance:

- Parsing distinguishes syntax failure from successful parsing.
- Failed alternatives cannot leak captures.
- Captures retain enough type information for taxonomy predicates.
- Focused tests cover ambiguity, duplicate fields, ranges, padding, and
  canonical representation.

### 4. SeqDef Generation And Ordering

Implement generation from bindings and define ordering/successor behavior only
for sequence shapes that support it. Do not introduce allocation or store
lookups here.

Acceptance:

- Generation validates bindings against the AST.
- Canonical rendering is deterministic.
- Successor behavior and exhaustion are explicit.
- Definitions without a meaningful ordering fail clearly when succession is
  requested.

### 5. Taxonomy Predicates And Tree Validation

Implement taxonomy nodes, exact predicates, numeric inclusive ranges, path
validation, and predicate validation. Taxonomy validation may receive or be
performed through a schema context so binding names can be checked against the
SeqDef.

Acceptance:

- Match keys refer to real captured fields.
- Predicate types and bounds are validated.
- Sibling ambiguity is rejected or explicitly prioritized.
- Parent/child predicate relationships are checked where possible.

### 6. Taxonomy Classification And Effective Meaning

Implement classification of parsed values, deepest-path selection, open/closed
classification policy, effective attributes, and inherited predicates.

Acceptance:

- Exact matches and numeric ranges classify correctly.
- Nested predicates are cumulative.
- Sibling overlap behavior is deterministic and intentional.
- Root-only, unclassified, and no-root cases have distinct documented results.
- Effective attributes inherit and override predictably.

### 7. Schema Composition

Introduce the explicit Schema abstraction that binds one SeqDef to one
taxonomy, ordering, and policy. Add interpret/parse operations that return
both captured syntax and classification.

Acceptance:

- Schema validation checks cross-definition invariants.
- Parsing and classification remain separately callable.
- Schema interpretation has one clear result model for valid, unclassified,
  and invalid values.
- A taxonomy cannot be applied to an unrelated SeqDef version.

### 8. PartNumber And Attribute Materialization

Implement part-number construction from an interpreted schema value, including
attribute validation against effective taxonomy assignments.

Acceptance:

- A created part number is reproducibly interpretable from its value and
  immutable definition references.
- Classification and attributes are validated before materialization.
- Part-number structural and semantic validation boundaries are explicit.

### 9. Persistence Model And Store Contract

Reimplement the store around the new domain model. Define persistence for
definitions, classifications, attributes, parts, revisions, drafts, and BOMs.

Acceptance:

- Uniqueness is enforced atomically by the store.
- Reference validation is complete and consistent.
- Memory and SQLite implementations share the intended contract.
- Definition versions are immutable and historical interpretations remain
  reproducible.
- Ownership, defensive-copy, ordering, and transaction behavior are tested
  where they are domain-visible.

### 10. Drafts And Context-Constrained Allocation

Implement draft lifecycle and an allocator separate from preview. Allocation
uses taxonomy constraints, SeqDef generation/successors, and atomic store
creation.

Acceptance:

- Draft context means the resulting classification must be at or below that
  context.
- Candidate generation and classification are revalidated on every attempt.
- Preview does not reserve values unless explicitly designed to do so.
- Release is atomic and cannot release a stale or invalid draft.
- Range exhaustion, no ordering, conflicts, and empty constraint intersections
  have explicit errors.

### 11. Revisions And BOM Graph

Implement part revisions, line items, ownership checks, cycle detection, and
recursive views over the revision graph.

Acceptance:

- Revision values conform to their revision SeqDef.
- Line-item part and revision identities agree.
- Direct and indirect cycles are rejected.
- BOM ordering is preserved.

### 12. Application Services And Views

Build application services around the explicit domain operations. Keep view
projection separate from domain classification and preserve effective path and
attribute semantics.

Acceptance:

- Services do not duplicate schema interpretation logic.
- Views expose stable, intentional JSON shapes.
- Taxonomy queries explicitly distinguish exact-node and subtree membership.
- Draft preview and release use the same allocation semantics.

### 13. HTTP API And Frontend Contract

Reimplement HTTP handlers and frontend-facing request/response types after the
domain and application contracts stabilize.

Acceptance:

- Routes, request validation, response shapes, and status mappings are
  intentional and tested.
- Domain JSON tags are explicit.
- Unsupported operations are either intentionally absent or implemented, not
  accidental.
- Frontend assumptions are represented by API tests rather than hidden
  behavior.

### 14. Artifact Integration

Reimplement artifact metadata and snapshot capture around the stable store and
application interfaces.

Acceptance:

- Source/path security rules remain enforced.
- Snapshot content is digest-addressed and idempotent.
- Size limits and cancellation behavior are tested.
- Artifact behavior remains independent from schema interpretation.

### 15. Cleanup And Migration Review

Review the completed implementation against the schema pipeline and decide
which current behavior was intentionally changed. Remove compatibility shims
that are not needed by the new contract.

Acceptance:

- Every design decision has an implementation or an explicit deferred item.
- Full tests pass.
- No accidental dependency on the old implementation remains.
- The final design and implementation plan reflect the accepted system.

## Review Checklist

For every implementation step, review:

- Does the change preserve the SeqDef/taxonomy boundary?
- Is the new behavior expressed by domain types rather than hidden in a
  handler or store?
- Are syntax validity, classification, ordering, and availability kept
  distinct?
- Are errors meaningful and testable without string matching?
- Are immutable definition references sufficient to reproduce meaning?
- Are ambiguous or incomplete cases explicit?
- Is the diff limited to the selected step?
- Are focused tests and the full suite green?

## Next Phase

The core reimplementation is complete. The next phase makes the new schema
model authoritative throughout the system and then improves taxonomy-directed
allocation.

### 16. Migrate Internal Callers To Schema

Move all internal interpretation and part-number construction through the
explicit `Schema` abstraction.

Targets:

- Seed data and fixtures
- Store release paths
- API part creation
- Application views
- Tests that construct or classify part numbers

Acceptance:

- No internal production path duplicates SeqDef parsing plus taxonomy
  classification.
- Schema validation is the single cross-definition validation boundary.
- Existing semantic tests remain green.

### 17. Migrate Taxonomy Data To Typed Predicates

Replace internal uses of legacy `Matches` taxonomy data with typed
`TaxonomyPredicate` values.

Acceptance:

- Seed data uses `ExactPredicate` and `NumericRangePredicate` as appropriate.
- SQLite and MemoryStore fixtures exercise typed predicates.
- Exact matching, numeric ranges, nested refinement, and inherited attributes
  remain covered by tests.
- No new taxonomy behavior depends on a second predicate representation.

### 18. Remove Legacy Taxonomy Projection Paths

After all internal callers use the explicit classification API, remove or
quarantine compatibility paths that duplicate interpretation.

Candidates:

- `TaxonomyNode.Matches`
- `Taxonomy.Project`
- `Taxonomy.ProjectNode`
- Legacy `NewPartNumber`
- Binding-only taxonomy classification paths

Acceptance:

- The remaining public API has one authoritative classification path.
- Any retained compatibility adapter is isolated, documented, and unreferenced
  by production code.
- No behavior depends on declaration-order ambiguity that typed validation
  should reject.

### 19. Implement Reverse Taxonomy Constraints

Add the reverse direction of taxonomy interpretation so a taxonomy context can
constrain candidate bindings:

```text
taxonomy context
    -> effective predicates
    -> candidate binding constraints
    -> SeqDef.Generate
```

The operation must combine predicates along the taxonomy path, including
numeric range intersection and exact-value constraints.

Acceptance:

- `Taxonomy.Constraints(nodeID)` returns effective constraints from root to the
  selected node.
- Exact and numeric constraints can be combined on different fields.
- Intersecting constraints produce the narrower valid domain.
- Empty intersections fail clearly.
- Conflicting exact constraints fail clearly.
- Constraints do not render strings or check store availability.

### 20. Make Allocation Taxonomy-Directed

Use reverse taxonomy constraints before candidate generation while retaining
classification as the final semantic check.

Acceptance:

- Allocation for a taxonomy context starts within that context's constrained
  candidate space.
- Every generated candidate is classified back to the requested node or an
  allowed descendant.
- Successor ordering remains a SeqDef concern.
- Store uniqueness remains authoritative.
- Range exhaustion, no ordering, conflicts, and unavailable constrained
  candidates have distinct errors.

### 21. Add End-To-End Schema Acceptance Scenarios

Exercise the complete pipeline with a representative hierarchical schema:

```text
SeqDef:
    category: three letters
    group: three digits
    number: three digits

Taxonomy:
    component
    |- resistor, category = RES
    |  `- high-power, group = 100..199
    `- capacitor, category = CAP
```

Acceptance:

- A value parses into structured captures.
- The value classifies to the expected taxonomy path.
- Effective attributes are inherited and overridden correctly.
- A taxonomy-context allocation uses the expected constrained range.
- The candidate persists and reloads with the same interpretation.
- Duplicate allocation is rejected by the store.
- The API and application views expose the same path and meaning.

### 22. Frontend Integration Review

Run the frontend against the completed schema API and decide whether frontend
changes belong in this reimplementation or in a separate project.

Acceptance:

- Frontend schema authoring can create and preview typed taxonomy predicates.
- Draft preview and release use the same binding and classification semantics.
- API field names and status handling are intentional and tested.
- Any remaining frontend work is recorded as a separate scoped plan.
