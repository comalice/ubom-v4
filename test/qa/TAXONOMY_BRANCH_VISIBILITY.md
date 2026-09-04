# Branch-Aware Taxonomy Editor Phase 1

## Scope

This runbook verifies frontend branch visibility and safe field-only targeting. It does not verify persisted branch identity; Phase 1 intentionally does not add it to the taxonomy API.

## Checks

1. Open a sequence definition with two choice alternatives containing different captured fields or numeric domains.
   - Expected pass: both alternatives have distinct labels and summaries, and their captured fields are visible.
   - Expected fail: the editor presents one global field lane with no alternative context.
2. Open a choice where the same binding name occurs in multiple alternatives.
   - Expected pass: occurrences retain distinct grammar paths and numeric domains when available.
   - Expected fail: a persisted field predicate silently moves to the first occurrence.
3. Open a literal-only alternative.
   - Expected pass: the alternative is visible and marked literal-only/not targetable; no category drop zone is offered.
   - Expected fail: a field-based category can be dropped directly onto the literal branch.
4. Open nested choices.
   - Expected pass: nested alternatives retain visible ancestry and do not collapse into a flat indistinguishable field list.
5. Reload a taxonomy containing a predicate whose field maps to multiple grammar occurrences.
   - Expected pass: the category remains present, is marked ambiguous/unresolved, and its original field predicate remains unchanged.
   - Expected fail: the target is silently assigned to `grammar-root` or the first matching occurrence.
6. Add a category to a uniquely resolved binding and save it.
   - Expected pass: the category is targetable and the outgoing taxonomy payload contains only the existing field predicate fields.
   - Expected fail: frontend path, alternative, or branch metadata appears in the request payload.

## Automated checks

From `frontend/`:

```sh
npm test
npm run typecheck
npm run build
```
