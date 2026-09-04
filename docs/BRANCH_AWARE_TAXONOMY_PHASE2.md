# Branch-Aware Taxonomy Phase 2

Phase 2 uses canonical structural paths as branch identities. A path is scoped
by the immutable `SeqDefID` already stored on a taxonomy revision and has the
form `root/0/choice[1]`; nested choices append another `choice[index]`
segment. The path is derived from the persisted AST traversal and is therefore
stable for that immutable sequence-definition version. A changed sequence
definition is a different identity and cannot silently retarget a taxonomy.

`ParsedSequence.Branches` contains every selected choice alternative. A
`TaxonomyNode` may optionally carry `branch: {"path": [...]}` in addition to
the existing `predicates` list. Branch segments are conjunctive, as are field
predicates on the same node. A branch-only condition can target a literal-only
alternative; no capture is required. Field exact predicates continue to
compare source text, while numeric ranges use the normalized range value.

`TaxonomyDef.ValidateAgainst` validates every branch path against the exact
`SeqDef` supplied by the caller. Missing, stale, or malformed paths fail with a
definition error before classification or persistence. Sibling overlap checks
also treat different alternatives of the same choice as disjoint.

The JSON API and both SQL stores add the optional branch field. SQLite adds a
nullable `branch` column to existing databases, and PostgreSQL migration 4
does the equivalent. Existing rows with only `matches` or `predicates` remain
valid and retain their original classification behavior. Memory-store cloning
copies branch paths so callers cannot mutate stored revisions through aliases.

This phase deliberately does not infer branch conditions for legacy revisions.
Copying a taxonomy to another sequence-definition version still requires the
caller to provide paths that validate against the new immutable definition.
