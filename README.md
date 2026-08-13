# ubom-v4

The fourth draft of my mini BOM manager, which encapsulates my very opionated way of thinking about how parts, bills of materials, and revisions relate to one another.

Unique ideas here are

1. sequences
2. taxonomy trees

where a sequence is a user generated grammar definition for arbitrary string sequences. It supports 'choice', for selectable literals within a string, 'branch' for a grammar that requires more than one definition, 'range' for values that span ranges such as 0-99 -- as well as padding out those ranges, 'rangeradix' for ranges that require position dependent radices (not radishes :D), and so on.

A sequence definition can parse a value OR generate a new value, something like `mySeq.Parse("a123")` or `mySeq.Next('a123') -> "a124"`.

A taxonomy tree applies labels to a sequence, so we can give some degree of meaning to each segment of a sequence.

This all builds up to `PartNumber`, which possesses a part number literal, and the accompanying taxonomy tree so you can see what categories a given part number occupies.

From `PartNumber`, we jump to the recursive DAG that is a BOM: `PartNumber + Revision -> BOM -> []LineItem(PartNumber, Revision)`, and the cycle continues.

# Roadmap

- strongly typed attributes, can be applied to any node in a taxonomy tree, to a part number, or to a bill of materials -- this allows you to do things like add resistor values to a specific branch of the taxonomy tree. It also enables things like querying parts by attribute, requirements satisfaction (maybe), and so on.
- A user interface. :D McMaster has the right idea here, though I'll also be barrowing ideas from Inventree.
- CRUD ops for the user facing stuff (part numbers, sequence grammars, taxonomy trees, etc.)
- A minimal change control workflow; all quality systems subsist on their change control workflows. To build a quality system from scratch, and have traceability, one needs at least an atomic change control workflow.
- Part number artifacts and artifact sources; I want to be able to point a part number at a git repo and have the revision bump with new release tags, ingest bills of materials (the git repo is the source of truth in this case), and so on.
- Multi BOM, where a part number can have a number of BOMs, one for each unit of responsibility: mech, elec, docs, production flows (depends on workflows, I think), and so on.
- Solidworks integration.
- Altium integration.
- KiCAD integration.
- Others? Is this giving you ideas? Drop an issue and lets discuss!