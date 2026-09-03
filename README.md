# ubom-v4

The fourth draft of my mini BOM manager, which encapsulates my very opionated way of thinking about how parts, bills of materials, and revisions relate to one another.

Unique ideas here are

1. sequences
2. taxonomy trees
3. attributes
4. artifacts

## Sequences

A sequence is a user generated grammar definition for arbitrary string sequences. It supports 'choice' for alternatives within a string, 'range' for values that span ranges such as 0-99 -- as well as padding out those ranges, 'rangeradix' for ranges that require position dependent radices (not radishes :D), and so on.

A sequence definition can parse a value OR generate a new value, something like `mySeq.Parse("a123")` or `mySeq.Next('a123') -> "a124"`.

Saved sequence definitions are immutable revisions; changes are recorded as a
new revision rather than modifying the definition used by existing parts.

## Taxonomy

A taxonomy tree applies labels to a sequence, so we can give some degree of meaning to each segment of a sequence.

This all builds up to `PartNumber`, which possesses a part number literal, and the accompanying taxonomy tree so you can see what categories a given part number occupies.

Part number values are globally unique canonical values across the application.
Taxonomy definitions are also immutable revisions, and a part number retains
the exact sequence and taxonomy revisions used to generate and classify it.

From `PartNumber`, we jump to the recursive DAG that is a BOM: `PartNumber + Revision -> BOM -> []LineItem(PartNumber, Revision)`, and the cycle continues.

## Attributes

An attribute is essentially metadata that can hang off of a taxonomy node (and is applied to all part numbers that the node is applied to). For example, suppose we have a part number schema like `RES-<GGG>-<NNN>` where `GGG` is some arbitrary group number and `NNN` is just a unique ID within a given group. If we hang a 'resistance value' attribute off the `RES` taxonomy node, now all resistors have a 'resistance value' field that can be filled in on creation.

## Artifacts

An artifact is the engineering record that lives behind a part number + revision combination. It could be anything. Right now we treat these as pure byte-streams and keep an 'artifact snapshot' in UBOM for every part number + revision combination.

Our first implementation is simply a file system artifact source. A periodic scan kicks off, changes are ingested, and a new artifact snapshot is recorded internally.

# Roadmap

## Local containers

`docker compose up --build` starts the UBOM API and a durable Postgres instance.
The API currently uses the mounted SQLite database because the server has not
yet gained a Postgres store adapter; the Postgres service is intentionally
present as the target production dependency, not a silently unused claim of
Postgres support. The frontend still runs separately with `make -C frontend dev`.

- strongly typed attributes, can be applied to any node in a taxonomy tree, to a part number, or to a bill of materials -- this allows you to do things like add resistor values to a specific branch of the taxonomy tree. It also enables things like querying parts by attribute, requirements satisfaction (maybe), and so on.
  - We now have attributes, but they aren't strongly typed yet. Soon.
- A user interface. :D McMaster has the right idea here, though I'll also be barrowing ideas from Inventree.
  - We have one in progress. It isn't great. It stinks of GPT-5.6 styling. If you've seen one UI from GPT-5.6, you've seen them all. We'll get it tightened up before MVP release.
- CRUD ops for the user facing stuff (part numbers, sequence grammars, taxonomy trees, etc.)
  - Standard workflows including system standup, core ECO/ECR pipelines, BOM creation, etc.
- A minimal change control workflow; all quality systems subsist on their change control workflows. To build a quality system from scratch, and have traceability, one needs at least an atomic change control workflow.
- Multi BOM, where a part number can have a number of BOMs, one for each unit of responsibility: mech, elec, docs, production flows (depends on workflows, I think), and so on.
- Solidworks integration.
- Altium integration.
- KiCAD integration.
- Others? Is this giving you ideas? Drop an issue and lets discuss!
