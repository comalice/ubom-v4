package ubom

import "fmt"

// Schema is the interpretation boundary for a sequence definition. It binds
// the syntax, taxonomy, classification policy, and (when available) the
// sequence definition's ordering into one immutable composition.
type Schema struct {
	SeqDef   SeqDef
	Taxonomy TaxonomyDef
	Policy   ClassificationPolicy
}

// SchemaPolicy is the schema-level name for the classification policy.
// ClassificationPolicy remains the underlying type used by Taxonomy APIs.
type SchemaPolicy = ClassificationPolicy

// InterpretedValue contains the syntax result and its optional meaning. An
// open schema can return a nil Classification when the value is syntactically
// valid but does not match the taxonomy root. A generic root classification is
// returned non-nil and is marked Generic.
type InterpretedValue struct {
	Parsed         ParsedSequence
	Classification *Classification
}

// NewSchema constructs a composed schema. Definitions are checked by Validate
// or by the operations that require a valid schema.
func NewSchema(seqDef SeqDef, taxonomy TaxonomyDef, policy ClassificationPolicy) Schema {
	return Schema{SeqDef: seqDef, Taxonomy: taxonomy, Policy: policy}
}

// Validate checks both definitions and their cross-definition invariants.
func (s Schema) Validate() error {
	if err := s.SeqDef.Validate(); err != nil {
		return err
	}
	if err := s.Taxonomy.ValidateAgainst(s.SeqDef); err != nil {
		return err
	}
	return nil
}

// Parse parses and classifies a value. Syntax errors and invalid definitions
// are always returned. Under an open policy, a classification mismatch is a
// valid but unclassified interpretation.
func (s Schema) Parse(value string) (InterpretedValue, error) {
	if err := s.Validate(); err != nil {
		return InterpretedValue{}, err
	}
	parsed, err := s.SeqDef.ParseValues(value)
	if err != nil {
		return InterpretedValue{}, err
	}
	classification, err := s.Taxonomy.ClassifyWithPolicy(parsed, s.Policy)
	if err == nil {
		return InterpretedValue{Parsed: parsed, Classification: classification}, nil
	}
	if !s.Policy.RequireClassification && IsCategory(err, ErrorCategoryClassification) {
		return InterpretedValue{Parsed: parsed}, nil
	}
	return InterpretedValue{}, err
}

// Interpret is a descriptive alias for Parse.
func (s Schema) Interpret(value string) (InterpretedValue, error) {
	return s.Parse(value)
}

// Generate renders bindings and interprets the generated canonical value,
// ensuring generation and parsing/classification cannot disagree.
func (s Schema) Generate(bindings map[string]string) (InterpretedValue, error) {
	if err := s.Validate(); err != nil {
		return InterpretedValue{}, err
	}
	value, err := s.SeqDef.Generate(bindings)
	if err != nil {
		return InterpretedValue{}, err
	}
	return s.Parse(value)
}

// Successor advances a value using the SeqDef ordering and interprets the
// resulting candidate. Definitions without meaningful ordering return the
// SeqDef ordering error.
func (s Schema) Successor(value string) (InterpretedValue, error) {
	if err := s.Validate(); err != nil {
		return InterpretedValue{}, err
	}
	next, err := s.SeqDef.Successor(value)
	if err != nil {
		return InterpretedValue{}, err
	}
	result, err := s.Parse(next)
	if err != nil {
		return InterpretedValue{}, fmt.Errorf("successor is not accepted by schema: %w", err)
	}
	return result, nil
}
