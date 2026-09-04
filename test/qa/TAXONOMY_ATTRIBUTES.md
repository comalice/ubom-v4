# Taxonomy Attributes QA

## Dependency

Run `test/qa/TAXONOMY_FIRST_EDITOR.md` first, or run both playbooks in one `make dev` session.

Use its current fixture:

- Sequence: `qa-part-number-v1`
- Taxonomy lineage: `qa-part-number-v1-taxonomy`
- Taxonomy tree: `QA Parts -> US Components -> Resistors -> Small Resistors`

## Procedure

### 1. Open the attributes table

Open attributes on `QA Parts`, `Resistors`, and `Small Resistors`.

Pass: each action opens the slide-out for the selected node, with the correct title and table.

Fail: an action is missing, opens the wrong node, or produces a browser error.

### 2. Add a root attribute

On `QA Parts`, add an attribute assignment. Rename it `Voltage`, set its type to `decimal`, and mark it required.

Pass: `Voltage` appears as required, the draft becomes dirty, and the taxonomy remains valid.

Fail: the assignment cannot be added or edited, required state is lost, or validation becomes incorrect.

### 3. Verify inheritance

Open attributes on `Resistors` and `Small Resistors`.

Pass: both nodes show `Voltage` as inherited from `QA Parts`; inherited rows are visibly distinct and not directly editable.

Fail: the attribute disappears, has the wrong source, or can be edited as a local assignment.

### 4. Override the inherited attribute

On `Resistors`, override `Voltage` and change it from required to optional.

Pass: the row remains singular, its source changes to `Resistors`, it is marked overridden, and `Small Resistors` inherits the optional state.

Fail: a duplicate row appears, the source is wrong, or descendants retain the old state.

### 5. Add a local attribute

On `Resistors`, add `Resistance`, set its type to `decimal`, and mark it required.

Pass: `Resistance` appears on `Resistors` and `Small Resistors` but not on `QA Parts` or unrelated branches.

Fail: the assignment is attached to the wrong node, does not inherit, or loses its type/required state.

### 6. Exercise invalid state

Temporarily create an invalid attribute definition, such as a duplicate label. Then correct it.

Pass: the table displays a validation error, save is disabled while invalid, and correcting the definition clears the error and restores save eligibility.

Fail: invalid state is hidden, an invalid taxonomy can be saved, or correction does not clear validation.

### 7. Save and reload

Save the taxonomy as a new current revision. Reload it and reopen the attribute tables on all three nodes.

Pass: `Voltage`, its override, and `Resistance` persist with the correct source, type, and required state; the new revision is current.

Fail: any assignment, inheritance relationship, or required state changes after reload.

### 8. Verify historical read-only behavior

Open the previous taxonomy revision and open its attributes table.

Pass: the table is readable, but required toggles, definition editing, override, add, and remove controls are unavailable; the historical revision cannot be saved in place.

Fail: the table cannot open, historical data is editable, or the historical revision can be mutated.

## Completion

The combined QA run passes only when the taxonomy-first editor runbook and all eight attribute steps pass. Do not release a part as part of this playbook.
