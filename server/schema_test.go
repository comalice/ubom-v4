package ubom

import (
	"errors"
	"testing"
)

func schemaFixture() (SeqDef, TaxonomyDef) {
	seq := NewSeqDef(Concat(
		Literal("PN-"),
		Bind("category", PlaceSequence("RC")),
		Literal("-"),
		Bind("number", Range(0, 9)),
	)).WithID("seq-v1")
	taxonomy := TaxonomyDef{
		ID: "taxonomy-v1", SeqDef: seq.ID,
		Taxonomy: Taxonomy{Root: TaxonomyNode{
			ID: "root",
			Children: []TaxonomyNode{
				{ID: "resistor", Predicates: []TaxonomyPredicate{ExactPredicate("category", "R")}},
				{ID: "capacitor", Predicates: []TaxonomyPredicate{ExactPredicate("category", "C")}},
			},
		}},
	}
	return seq, taxonomy
}

func TestSchemaValidateComposition(t *testing.T) {
	seq, taxonomy := schemaFixture()
	if err := NewSchema(seq, taxonomy, SchemaPolicy{}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	taxonomy.SeqDef = "other-seq-v1"
	if err := NewSchema(seq, taxonomy, SchemaPolicy{}).Validate(); err == nil || !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("mismatched definitions error = %v, want definition error", err)
	}

	taxonomy.SeqDef = seq.ID
	taxonomy.Taxonomy.Root.Children[0].Predicates[0].Field = "missing"
	if err := NewSchema(seq, taxonomy, SchemaPolicy{}).Validate(); err == nil || !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("unknown field error = %v, want definition error", err)
	}
}

func TestSchemaParseClassifiesAndHandlesPolicy(t *testing.T) {
	seq, taxonomy := schemaFixture()
	open := NewSchema(seq, taxonomy, SchemaPolicy{})
	interpreted, err := open.Parse("PN-R-4")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if interpreted.Parsed.Value != "PN-R-4" || interpreted.Classification == nil || interpreted.Classification.NodeID != "resistor" {
		t.Fatalf("interpretation = %#v, want resistor classification", interpreted)
	}

	genericTaxonomy := taxonomy
	genericTaxonomy.Taxonomy.Root.Children = nil
	interpreted, err = NewSchema(seq, genericTaxonomy, SchemaPolicy{}).Parse("PN-R-4")
	if err != nil || interpreted.Classification == nil || !interpreted.Classification.Generic {
		t.Fatalf("open generic interpretation = %#v, %v", interpreted, err)
	}

	unknownTaxonomy := taxonomy
	unknownTaxonomy.Taxonomy.Root.Predicates = []TaxonomyPredicate{ExactPredicate("category", "X")}
	unknownTaxonomy.Taxonomy.Root.Children = nil
	interpreted, err = NewSchema(seq, unknownTaxonomy, SchemaPolicy{}).Parse("PN-R-4")
	if err != nil || interpreted.Classification != nil || interpreted.Parsed.Value != "PN-R-4" {
		t.Fatalf("open unclassified interpretation = %#v, %v", interpreted, err)
	}
	if _, err := NewSchema(seq, unknownTaxonomy, SchemaPolicy{RequireClassification: true}).Parse("PN-R-4"); !IsCategory(err, ErrorCategoryClassification) {
		t.Fatalf("closed unclassified error = %v, want classification error", err)
	}
	if _, err := open.Parse("not-a-part-number"); !IsCategory(err, ErrorCategorySyntax) {
		t.Fatalf("invalid syntax error = %v, want syntax error", err)
	}
}

func TestSchemaGenerateAndSuccessorRoundTrip(t *testing.T) {
	seq := NewSeqDef(Concat(
		Literal("PN-"), Bind("number", Range(0, 9)),
	)).WithID("ordered-seq-v1")
	taxonomy := TaxonomyDef{
		ID: "ordered-taxonomy-v1", SeqDef: seq.ID,
		Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root", Children: []TaxonomyNode{{
			ID: "numbered", Predicates: []TaxonomyPredicate{NumericRangePredicate("number", 0, 9)},
		}}}},
	}
	schema := NewSchema(seq, taxonomy, SchemaPolicy{})
	generated, err := schema.Generate(map[string]string{"number": "4"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if generated.Parsed.Value != "PN-4" || generated.Classification == nil || generated.Classification.NodeID != "numbered" {
		t.Fatalf("generated = %#v, want parsed numbered value", generated)
	}
	next, err := schema.Successor(generated.Parsed.Value)
	if err != nil {
		t.Fatalf("Successor() error = %v", err)
	}
	if next.Parsed.Value != "PN-5" || next.Classification == nil || next.Classification.NodeID != "numbered" {
		t.Fatalf("successor = %#v, want PN-5 numbered value", next)
	}
}
