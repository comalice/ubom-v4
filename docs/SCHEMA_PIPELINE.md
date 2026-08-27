# Part Number Schema Pipeline

This document sketches the relationship between sequence definitions,
taxonomy, part numbers, and allocation.

## The Schema

A part-number schema is the composition of two independent concerns:

```text
Schema
|- SeqDef: validity
`- Taxonomy: meaning
```

The separation is intentional:

- `SeqDef` answers: what strings are structurally allowed?
- Taxonomy answers: what does a valid string mean?

For example, a sequence definition may describe a category as three letters
and a number as four digits:

```text
category := [A-Za-z]{3}
number   := [0-9]{4}
```

The taxonomy can then assign domain meaning to captured values:

```text
category = RES -> resistor
category = CAP -> capacitor
```

This keeps domain vocabulary out of the sequence definition when that
vocabulary is classification rather than structure. Fixed punctuation,
required widths, character alphabets, and other structural constraints remain
in `SeqDef`.

## Pipeline

The conceptual flow is:

```text
raw value
    -> SeqDef.Parse
    -> ParseResult
    -> Taxonomy.Classify
    -> Classification
    -> PartNumber
```

`SeqDef.Parse` recognizes syntax and captures named segments. A capture is a
segment of the sequence made available for classification. Taxonomy evaluates
the captured values against its tree and returns the most specific matching
classification.

For example:

```text
RES-0042
    -> category = RES, number = 0042
    -> resistor
```

Taxonomy nodes represent progressively more specific meaning:

```text
component
`- resistor
   `- thick-film
```

A part classified as `thick-film` also inherits the meaning of `resistor` and
`component`.

## Two Kinds Of Validity

Once validity and meaning are separated, there are two related questions:

```text
syntactic validity:
    Does the value conform to SeqDef?

classification:
    Does the parsed value map to a taxonomy node?
```

A value can be syntactically valid but not specifically classified. An open
taxonomy may allow such a value to remain at its generic root. A closed
taxonomy may require a more specific classification before release.

That release requirement is schema policy. It should not be hidden inside
`SeqDef` or taxonomy matching.

## Responsibilities

### SeqDef

`SeqDef` owns syntax and structural construction:

```text
Parse(value)
    -> ParseResult or error

Generate(bindings)
    -> canonical value or error

Successor(value)
    -> next syntactically valid value or error
```

`Successor` is only available where the definition has a meaningful ordering.
Parsing and generation do not automatically imply that a general ordering
exists.

### Taxonomy

Taxonomy owns interpretation and classification:

```text
Classify(ParseResult)
    -> taxonomy node or unclassified
```

Taxonomy may eventually provide reverse constraints for generation. This is a
proposed/deferred API, not part of the accepted implementation:

```text
Constraints(node or context)
    -> bindings compatible with that classification
```

When implemented, this would allow taxonomy to own facts such as `resistor
requires category = RES` without making taxonomy responsible for rendering
`RES-0042`. `SeqDef` still constructs the string from bindings.

### Schema

A composed schema coordinates the two layers:

```text
Schema.Parse(value)
    parsed = SeqDef.Parse(value)
    classification = Taxonomy.Classify(parsed)
    return parsed plus classification
```

`Schema` performs this composition explicitly without merging the underlying
responsibilities. `NewPartNumber` is available as a compatibility adapter for
callers that do not supply attributes, while durable creation still requires a
classification node.

### PartNumber

`PartNumber` is a durable classified value:

```text
PartNumber =
    canonical value
    + SeqDef version
    + Taxonomy version
    + taxonomy node
    + attributes
```

The definition IDs are part of the meaning. Changing either definition can
change the interpretation of the same textual value.

The taxonomy node stored on a part number is the materialized classification
result. Bindings can be reconstructed from the value and the referenced
`SeqDef`, so they do not necessarily need to be stored independently.

## Next Versus Allocate

`next` and `allocate` are different operations.

```text
next:
    What value follows this value according to sequence ordering?

allocate:
    What valid, classifiable, unused value should be assigned?
```

For example, if `RES-0001`, `RES-0002`, and `RES-0003` already exist, the
successor of `RES-0001` is still `RES-0002`, while the next available resistor
number may be `RES-0004`.

The responsibilities are therefore:

```text
SeqDef:
    parse, generate, and produce successor candidates

Taxonomy:
    classify candidates; reverse constraints are proposed/deferred

Store:
    enforce uniqueness and persistence

Allocator or application service:
    iterate candidates until one satisfies policy and is unused
```

Taxonomy should not define the mechanics of string succession. The accepted
allocator iterates SeqDef successors and reclassifies each candidate against
the requested context. A future reverse-constraint API could reduce that
search, while the allocator would still coordinate sequence ordering and store
availability.

## Example Allocation Flow

For a released part number:

```text
1. Select a schema and taxonomy context.
2. Request candidate bindings from the allocator.
3. Use SeqDef.Generate to construct a candidate value.
4. Parse or canonicalize the candidate with SeqDef.
5. Classify the candidate with taxonomy.
6. Check that the classification satisfies the requested context.
7. Check store uniqueness.
8. Persist the classified PartNumber.
```

The current allocator obtains its starting bindings from the draft and then
iterates SeqDef successors, reclassifying every candidate. A future allocator
may first request reverse constraints from `Taxonomy.Constraints(context)` to
choose a narrower starting/intermediate binding set; that API is not currently
implemented.

For an existing value:

```text
1. Parse it with SeqDef.
2. Classify it with taxonomy.
3. Display or persist the result.
```

The key invariant is:

```text
SeqDef can determine whether a value is possible.
Taxonomy can determine what a possible value means.
Neither alone determines whether the value is available to assign.
```

## Conceptual Semantics

The schema can be understood as a language together with an interpretation of
that language:

```text
SeqDef language:
    the set of strings accepted by SeqDef

Parser:
    a valid string -> captured fields

Taxonomy:
    captured fields -> a classification path
```

Taxonomy therefore partitions, or partially partitions, the language admitted
by `SeqDef`:

```text
SeqDef admits all structurally valid values.
Taxonomy subdivides those values into meaningful regions.
```

The partition may be:

```text
complete:
    every valid value receives a specific classification

partial:
    some valid values remain generic or unclassified

hierarchical:
    classifications refine one another along taxonomy paths
```

### Effective Predicates

A taxonomy node's meaning is the conjunction of its own predicate and every
predicate along the path to it. For example:

```text
component
`- resistor       category = RES
   `- high-power   group in 100..199
```

The effective meanings are:

```text
resistor:
    category = RES

high-power:
    category = RES AND group in 100..199
```

This is why a more precise meaning belongs below the broader meaning. The
child is a refinement of the parent rather than a competing classification.
Sibling predicates should normally be disjoint. If sibling predicates overlap,
the taxonomy must define whether that is rejected as ambiguous or resolved by
an explicit priority rule.

### Classification And Constraints

Classification and generation constraints are two directions over the same
semantic relationship:

```text
classify:
    parsed fields -> taxonomy node

constrain:
    taxonomy node -> admissible field values
```

For example, if taxonomy says:

```text
resistor:
    category = RES
```

then classification should be true now, while the reverse constraint is an
ideal future capability:

```text
Classify({category: RES}) -> resistor
proposed Constraints(resistor) -> category must equal RES
```

`SeqDef` remains responsible for rendering those constrained fields into a
valid string. The consistency invariant is:

```text
values generated for taxonomy node N
    must classify back to N
    or to an allowed descendant of N
```

### Validity And Ordering

A grammar defines a set of valid values, but does not necessarily define a
canonical ordering over that set:

```text
SeqDef validity != SeqDef ordering
```

Some sequence definitions have an obvious successor operation, such as a
fixed-width numeric range. Others, such as a three-letter field, require an
additional ordering policy if they are to be enumerated.

The ideal schema can therefore be viewed as:

```text
Schema
|- SeqDef: validity
|- Taxonomy: meaning
|- Ordering: succession rules, where applicable
`- Policy: classification and allocation rules
```

Ordering produces candidates. Taxonomy determines whether candidates have the
requested meaning. Allocation and persistence determine whether a candidate is
available to assign.

## Terminology

The following vocabulary keeps the boundaries clear:

```text
SeqDef:
    syntax, grammar, admissible values

ParseResult:
    captured syntax and bindings

Taxonomy:
    classification and interpretation

Schema:
    SeqDef plus Taxonomy plus release policy

PartNumber:
    durable classified value

Successor:
    next value according to sequence ordering

Allocation:
    next unused value satisfying schema policy
```

## Ideal Implementation Sketch

The following is an idealized API, not a claim about the current code. It
makes the composed schema and the boundaries between parsing, classification,
succession, and allocation explicit.

### Core Types

```go
type Schema struct {
    SeqDef   SeqDef
    Taxonomy TaxonomyDef
    Policy   SchemaPolicy
}

type SchemaPolicy struct {
    RequireClassification bool
}

type ParsedSequence struct {
    Value    string
    Bindings map[string]string
}

type Classification struct {
    TaxonomyDefID TaxonomyDefID
    NodeID        TaxonomyNodeID
    Path          []TaxonomyNodeID
}

type InterpretedValue struct {
    Parsed         ParsedSequence
    Classification *Classification
}
```

`ParsedSequence` is the syntax-level result. `Classification` is the
taxonomy-level result. `InterpretedValue` combines them without making either
layer responsible for the other.

### SeqDef Methods

```go
func (d SeqDef) ParseValues(value string) (ParsedSequence, error)
func (d SeqDef) Generate(bindings map[string]string) (string, error)
func (d SeqDef) Successor(value string) (string, error)
```

`Parse` remains a string-returning compatibility operation; schema
interpretation uses `ParseValues` so taxonomy receives structured captures.

`ParseValues` recognizes and captures. `Generate` renders a value from bindings.
`Successor` produces a structural candidate only where the definition has a
well-defined ordering. It does not check store uniqueness or taxonomy policy.

An alternative API could make succession operate on parsed values:

```go
func (d SeqDef) SuccessorOf(parsed ParsedSequence) (ParsedSequence, error)
```

That avoids parsing the same value twice when an allocator already has a
`ParsedSequence`.

### Taxonomy Methods

```go
func (t TaxonomyDef) Classify(parsed ParsedSequence) (*Classification, error)
```

`Classify` maps captured values to a taxonomy node. Application allocation
checks whether the result is at or below a requested taxonomy context and
iterates SeqDef successors; taxonomy does not render strings or reserve
values. There is no separate reverse-direction `Constraints` API in the
accepted implementation.

Taxonomy constraints should not render strings. They supply facts such as:

```text
resistor requires category = RES
```

`SeqDef.Generate` remains responsible for turning those bindings into a
canonical value.

### Schema Methods

```go
func (s Schema) Validate() error
func (s Schema) Parse(value string) (InterpretedValue, error)
func (s Schema) Generate(bindings map[string]string) (InterpretedValue, error)
func (s Schema) Successor(value string) (InterpretedValue, error)
```

The schema methods coordinate the two definitions:

```go
func (s Schema) Parse(value string) (InterpretedValue, error) {
    parsed, err := s.SeqDef.ParseValues(value)
    if err != nil {
        return InterpretedValue{}, err
    }

    classification, err := s.Taxonomy.ClassifyWithPolicy(parsed, s.Policy)
    if err != nil && !s.Policy.RequireClassification {
        return InterpretedValue{Parsed: parsed}, nil
    }
    if err != nil {
        return InterpretedValue{}, err
    }

    return InterpretedValue{
        Parsed:         parsed,
        Classification: classification,
    }, nil
}
```

If open taxonomies are supported, an unclassified value can return a nil
classification. If the policy is closed, classification is part of schema
validity and `Parse` fails when no acceptable taxonomy node is found.

### Part Number Creation

```go
func (s Schema) NewPartNumber(
    value string,
    attributes []PartNumberAttribute,
) (PartNumber, error)
```

The call tree is:

```text
Schema.NewPartNumber
`- Schema.Parse
   |- SeqDef.Parse
   `- TaxonomyDef.Classify
`- TaxonomyDef.EffectiveAttributes
`- ValidatePartNumberAttributes
`- construct PartNumber
```

The resulting `PartNumber` stores the canonical value, the schema definition
IDs, the selected taxonomy node, and the validated attributes. The store still
performs reference and uniqueness checks when persisting it.

### Successor Call Tree

For a structural successor:

```text
Schema.Successor
`- SeqDef.Successor
   `- SeqDef.Parse
   `- advance ordered fields
   `- SeqDef.Generate
`- TaxonomyDef.Classify
`- apply schema classification policy
```

The successor operation may return a syntactically valid value that is not in
the desired taxonomy context. That is a classification result, not an error
in `SeqDef`.

### Allocation Interfaces

Allocation combines candidate generation, interpretation, and persistence:

```go
type PartNumberAllocator interface {
    Allocate(
        schema Schema,
        context TaxonomyNodeID,
        seed map[string]string,
        attributes []PartNumberAttribute,
    ) (PartNumber, error)
}

type PartNumberStore interface {
    GetPartNumber(value string) (PartNumber, error)
    CreatePartNumber(part PartNumber) (PartNumber, error)
}
```

The current allocator's call tree is:

```text
Allocator.Allocate
`- use draft/seed bindings
`- Schema.Generate(bindings)
   |- SeqDef.Generate
   |- SeqDef.Parse
   `- TaxonomyDef.Classify
`- check classification is at or below context
`- Store.GetPartNumber(candidate.Value)
`- if occupied:
   `- Schema.Successor(candidate.Value)
      |- SeqDef.Successor
      |- SeqDef.Parse
      |- SeqDef.Generate
      `- TaxonomyDef.Classify
`- Schema.NewPartNumber(candidate.Value, attributes)
`- Store.CreatePartNumber(part)
```

The proposed future constraint-assisted variant would add the following before
`Schema.Generate`:

```text
Allocator.Allocate
`- proposed Schema.Taxonomy.Constraints(context)
`- merge constraints with seed bindings
`- continue with the current successor/reclassification loop
```

The final store operation must still be authoritative for uniqueness. A
read-then-create sequence can race in a concurrent implementation, so the
store should reject a conflicting insert and the allocator should decide
whether to retry.

## Worked Examples

### Parsing And Classifying An Existing Value

Assume this sequence definition:

```text
SeqDef:
    Bind("category", PlaceSequence("ABCDEFGHIJKLMNOPQRSTUVWXYZ", ... 3 places ...))
    Literal("-")
    Bind("number", Range(0, 9999).Width(4))
```

And this taxonomy:

```text
component
|- resistor       when category = RES
`- capacitor      when category = CAP
```

The call tree for `RES-0042` is:

```text
Schema.Parse("RES-0042")
`- SeqDef.Parse("RES-0042")
   `- ParsedSequence{
          Value: "RES-0042",
          Bindings: {category: "RES", number: "0042"},
      }
`- Taxonomy.Classify(parsed)
   `- test component/resistor predicate
      `- category == RES
   `- Classification{NodeID: "resistor"}
```

The result is syntactically valid and classified as a resistor. `SeqDef` did
not need to know that `RES` means resistor.

### Structurally Valid But Unclassified

For `BAT-0042`, the call tree is:

```text
Schema.Parse("BAT-0042")
`- SeqDef.Parse("BAT-0042")
   `- success: category = BAT, number = 0042
`- Taxonomy.Classify(parsed)
   `- no specific child predicate matches
```

With an open taxonomy, the result can retain the generic `component` meaning.
With `RequireClassification: true`, the schema returns an error. The sequence
definition has not changed; only the schema policy determines whether the
unclassified result is acceptable.

### Allocating A Resistor Number

Assume the requested context is `resistor`, and the taxonomy supplies:

```text
category = RES
```

The current call tree is:

```text
Allocator.Allocate(schema, "resistor", {category: "RES", number: "0042"}, attributes)
`- use seed bindings
   `- {category: "RES", number: "0042"}
`- Schema.Generate(bindings)
   |- SeqDef.Generate
   |  `- "RES-0042"
   |- SeqDef.Parse("RES-0042")
   `- Taxonomy.Classify
      `- resistor
`- Store.GetPartNumber("RES-0042")
   `- not found
`- Schema.NewPartNumber("RES-0042", attributes)
`- Store.CreatePartNumber(part)
```

If `RES-0042` is already occupied, allocation advances according to the
sequence ordering and reclassifies each successor. The successor is a candidate,
not automatically an allocation result:

```text
SeqDef.Successor("RES-0042")
    -> "RES-0043"

Taxonomy.Classify("RES-0043")
    -> resistor

Store.GetPartNumber("RES-0043")
    -> not found
```

If a successor crosses into a value classified outside the requested context,
the allocator skips it rather than asking `SeqDef` to understand taxonomy.

A future constraint-assisted allocator could use the proposed
`Taxonomy.Constraints("resistor")` to derive `{category: "RES"}` before
generation. That would narrow candidate selection but would not replace
classification or store uniqueness checks.

### Nested Range Refinement

When one meaning is a more precise version of another, the more precise
taxonomy node should be a child of the broader node:

```text
component
`- X                 group in 100..199
   `- Y              group in 150..175
```

This is the taxonomy equivalent of saying:

```text
X = 100..199
Y = X and 150..175
```

The taxonomy does not need to repeat the `100..199` predicate on `Y`. Reaching
`Y` already requires successfully matching `X`, so the effective predicate for
`Y` is the conjunction of every predicate along the path.

The classification call tree for a value with `group = 160` is:

```text
Taxonomy.Classify({group: "160"})
`- test X: 160 in 100..199
   `- match
   `- test Y: 160 in 150..175
      `- match
`- result: Y
```

For other values:

```text
group = 120
    -> X matches
    -> Y does not match
    -> result: X

group = 180
    -> X matches
    -> Y does not match
    -> result: X

group = 220
    -> X does not match
    -> result: component, or unclassified under a closed policy
```

The important property is that the child is not competing with its parent.
The parent establishes a broad meaning, and the child refines it.

### Nested Classification With Multiple Fields

The same refinement model works when the parent and child use different
captured fields:

```text
component
`- resistor                 category = RES
   `- high-power            group in 100..199
      `- precision          tolerance = P
```

The effective meaning of a value is accumulated along the path:

```text
category = RES, group = 137, tolerance = P
    -> component
    -> resistor
    -> high-power
    -> precision
```

The corresponding classification call tree is:

```text
Taxonomy.Classify(parsed)
`- component
   `- resistor: category == RES
      `- high-power: group in 100..199
         `- precision: tolerance == P
`- result: precision
```

This also explains why a taxonomy match should generally describe a condition
local to its node. The path supplies the full condition; each node supplies
the next refinement.

### Nested Versus Overlapping Siblings

Consider two different trees:

```text
Nested refinement:
component
`- X                 group in 100..199
   `- Y              group in 150..175
```

and:

```text
Overlapping siblings:
component
|- X                 group in 100..199
`- Y                 group in 150..175
```

For `group = 160`, the nested tree has one unambiguous result:

```text
component -> X -> Y
```

The sibling tree has two possible classifications:

```text
component -> X
component -> Y
```

That is not refinement. It is ambiguity between competing meanings. The
existing implementation resolves overlapping siblings by declaration order,
but for quality-controlled schemas a stronger ideal policy is preferable:

- Nested predicates express intentional refinement.
- Sibling predicates should be disjoint.
- Validation should reject ambiguous sibling ranges, or require an explicit
  priority when overlap is intentional.

If `Y` is genuinely a special case of `X`, it belongs below `X`. If `X` and `Y`
are alternatives at the same level, their value ranges should not overlap.

### Range Refinement And Allocation

The nested structure provides an allocation context. The accepted allocator
starts with bindings, iterates ordered candidates, and reclassifies them:

```text
Allocate(context = Y)
`- start with the draft's group binding
`- SeqDef.Generate({group: candidate value})
`- Taxonomy.Classify(candidate)
   `- X matches
   `- Y matches
`- if occupied or outside Y, advance SeqDef successor and reclassify
`- Store.CreatePartNumber(candidate)
```

A future implementation could use the proposed `Taxonomy.Constraints(Y)` to
derive the `150..175` intersection, including the inherited `100..199`
condition, before generating candidates. That reverse constraint operation is
deferred and does not replace the current reclassification step.

The inherited `100..199` constraint is redundant for the numeric intersection
in this example, but it is semantically important: a candidate for `Y` must
satisfy every ancestor condition. If a child constraint is outside the
ancestor range, the context has no possible values and allocation should fail
rather than produce a value that cannot reach the requested node.
