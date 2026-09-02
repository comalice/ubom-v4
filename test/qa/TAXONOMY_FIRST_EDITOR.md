# Taxonomy-First Editor QA

## Fixture

Sequence definition ID: `qa-part-number-v1`

```text
"QA-" + region + "-" + family + "-" + size
```

- `region`: choice `US` or `EU`
- `family`: choice `RES` or `CAP`
- `size`: range `1..999`, width `3`, pad `0`

Expected values include `QA-US-RES-001` and `QA-EU-CAP-042`.

Taxonomy definition lineage: `qa-part-number-v1-taxonomy` (server-generated)

Expected current revision after creation: `qa-part-number-v1-taxonomy-v2`

```text
QA Parts
|- US Components       region = US
|  `- Resistors         family = RES
|     `- Small Resistors size = 1..99
`- EU Components       region = EU
   `- Capacitors        family = CAP
```

## Procedure

### 1. Start the application

Run `make dev` and open the local application.

Pass: the frontend and API are available without startup errors.

Fail: `make dev` exits, the frontend cannot load, or API requests fail.

### 2. Create the sequence definition

Open Settings, create a new sequence definition with ID `qa-part-number-v1`, and build the schema above. Save it.

Pass: the definition saves, receives a revision, and its grammar displays `region`, `family`, and `size` in order.

Fail: validation prevents saving, the revision is not returned, or any binding is missing or reordered.

### 3. Create the taxonomy

From the sequence definition, create a taxonomy. The server generates the taxonomy definition lineage from the sequence definition. Add the five categories and predicates shown above, then save the taxonomy.

Pass: the taxonomy saves as a current revision and the editor shows the taxonomy tree as the primary surface.

Fail: the editor is grammar-first, a category cannot be nested, a predicate cannot be edited, or save reports validation/conflict errors.

### 4. Verify taxonomy-first context lanes

Reload the saved taxonomy. Inspect the context lanes while moving or adding categories.

Pass:

- `region` exposes `family` and `size`.
- `family` exposes `size`.
- `size` exposes no later field.
- Normal child drops remain in `Child categories`.
- Explicit context-lane drops target the selected grammar field.
- Reloading preserves the category tree and predicates.

Fail: a category appears in the wrong lane, a normal child drop changes grammar context, or persisted categories disappear or move after reload.

### 5. Verify production taxonomy routing

Open:

```text
/parts/qa-parts?seqDef=qa-part-number-v1
```

Navigate through `US Components`, `Resistors`, and `Small Resistors`.

Pass: each node loads from the current taxonomy for `qa-part-number-v1`; child links preserve the `seqDef` query parameter.

Fail: the route uses a hard-coded sample taxonomy, loses `seqDef`, returns an unrelated taxonomy, or cannot resolve a child node.

### 6. Create and preview a part draft

Start a draft from `Small Resistors`. Enter bindings that produce `QA-US-RES-042`, then preview the draft.

Pass: the draft carries `qa-part-number-v1`, the current `qa-part-number-v1-taxonomy` revision, and the `Small Resistors` context node; preview is complete and returns `QA-US-RES-042`.

Fail: the draft uses a different sequence or taxonomy, rejects valid bindings, selects the wrong context node, or preview produces a different value/error.

## Completion

The QA run passes only when all six steps pass. Do not release the generated part unless release-path coverage is explicitly requested; successful preview is sufficient for this pipeline.
