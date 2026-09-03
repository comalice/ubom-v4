package ubom

import (
	"errors"
	"testing"
)

func materializationSchema() Schema {
	seq, taxonomy := schemaFixture()
	taxonomy.AttributeDefs = []AttributeDef{
		{ID: "finish", Label: "Finish", ValueType: AttributeValueEnum, EnumValues: []string{"matte", "gloss"}},
		{ID: "resistance", Label: "Resistance", ValueType: AttributeValueDecimal},
		{ID: "note", Label: "Note", ValueType: AttributeValueString},
	}
	taxonomy.Taxonomy.Root.Attributes = []AttributeAssignment{
		{AttributeDefID: "finish", Required: true},
		{AttributeDefID: "note"},
	}
	taxonomy.Taxonomy.Root.Children[0].Attributes = []AttributeAssignment{
		{AttributeDefID: "finish", Required: false},
		{AttributeDefID: "resistance", Required: true},
	}
	return NewSchema(seq, taxonomy, SchemaPolicy{})
}

func TestSchemaNewPartNumberMaterializesEffectiveAttributes(t *testing.T) {
	schema := materializationSchema()
	attributes := []PartNumberAttribute{
		{AttributeDefID: "finish", Value: "matte"},
		{AttributeDefID: "resistance", Value: "1000"},
	}
	part, err := schema.NewPartNumber("PN-R-4", attributes)
	if err != nil {
		t.Fatalf("Schema.NewPartNumber() error = %v", err)
	}
	if part.Value != "PN-R-4" || part.SeqDefID != "seq-v1" || part.TaxonomyDefID != "taxonomy-v1" || part.TaxonomyNodeID != "resistor" {
		t.Fatalf("part = %#v, want canonical durable identity", part)
	}
	if len(part.Attributes) != len(attributes) {
		t.Fatalf("attributes = %#v, want %#v", part.Attributes, attributes)
	}
	attributes[0].Value = "gloss"
	if part.Attributes[0].Value != "matte" {
		t.Fatal("materialized attributes alias caller input")
	}
}

func TestSchemaNewPartNumberHonorsOverriddenOptionalAssignment(t *testing.T) {
	schema := materializationSchema()
	if _, err := schema.NewPartNumber("PN-R-4", []PartNumberAttribute{
		{AttributeDefID: "resistance", Value: "1000"},
	}); err != nil {
		t.Fatalf("overridden optional attribute was treated as required: %v", err)
	}
}

func TestSchemaNewPartNumberRejectsAttributeFailures(t *testing.T) {
	schema := materializationSchema()
	tests := []struct {
		name       string
		attributes []PartNumberAttribute
	}{
		{"missing inherited required", []PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}}},
		{"invalid enum", []PartNumberAttribute{{AttributeDefID: "finish", Value: "satin"}, {AttributeDefID: "resistance", Value: "1000"}}},
		{"invalid decimal", []PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}, {AttributeDefID: "resistance", Value: "one thousand"}}},
		{"unknown attribute", []PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}, {AttributeDefID: "resistance", Value: "1000"}, {AttributeDefID: "unknown", Value: "x"}}},
		{"duplicate attribute", []PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}, {AttributeDefID: "finish", Value: "gloss"}, {AttributeDefID: "resistance", Value: "1000"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := schema.NewPartNumber("PN-R-4", test.attributes); err == nil || !errors.Is(err, ErrInvalidValue) {
				t.Fatalf("error = %v, want value error", err)
			}
		})
	}
}

func TestSchemaNewPartNumberRejectsUnclassifiedOpenValue(t *testing.T) {
	seq, taxonomy := schemaFixture()
	taxonomy.Taxonomy.Root.Predicates = []TaxonomyPredicate{ExactPredicate("category", "X")}
	taxonomy.Taxonomy.Root.Children = nil
	part, err := NewSchema(seq, taxonomy, SchemaPolicy{}).NewPartNumber("PN-R-4", nil)
	if err == nil || !errors.Is(err, ErrUnclassified) {
		t.Fatalf("part = %#v, error = %v, want unclassified error", part, err)
	}
}

func TestSchemaNewPartNumberRejectsSeqDefTaxonomyMismatch(t *testing.T) {
	seq, taxonomy := schemaFixture()
	taxonomy.SeqDef = "other-seq-v1"
	if _, err := NewSchema(seq, taxonomy, SchemaPolicy{}).NewPartNumber("PN-R-4", nil); err == nil || !errors.Is(err, ErrInvalidDefinition) {
		t.Fatalf("error = %v, want definition error", err)
	}
}

func TestNewPartNumber(t *testing.T) {
	seqDef := NewSeqDef(Concat(
		Bind("category", Range(0, 999).Width(3)),
		Literal("-"),
		Bind("family", Range(0, 99999).Width(5)),
		Literal("-"),
		Bind("id", Range(0, 999).Width(3)),
	)).WithID("pn-v1")
	taxonomy := TaxonomyDef{
		ID:     "components-v1",
		SeqDef: seqDef.ID,
		Taxonomy: Taxonomy{Root: TaxonomyNode{
			ID: "components",
			Children: []TaxonomyNode{{
				ID:         "resistors",
				Predicates: []TaxonomyPredicate{ExactPredicate("category", "001")},
				Children: []TaxonomyNode{{
					ID:         "thick-film",
					Predicates: []TaxonomyPredicate{ExactPredicate("family", "00042")},
				}},
			}},
		}},
	}

	part, err := NewPartNumber("001-00042-001", seqDef, taxonomy)
	if err != nil {
		t.Fatalf("NewPartNumber() error = %v", err)
	}
	if part.Value != "001-00042-001" {
		t.Fatalf("Value = %q, want %q", part.Value, "001-00042-001")
	}
	if part.SeqDefID != seqDef.ID {
		t.Fatalf("SeqDefID = %q, want %q", part.SeqDefID, seqDef.ID)
	}
	if part.TaxonomyDefID != taxonomy.ID {
		t.Fatalf("TaxonomyDefID = %q, want %q", part.TaxonomyDefID, taxonomy.ID)
	}
	if part.TaxonomyNodeID != "thick-film" {
		t.Fatalf("TaxonomyNodeID = %q, want %q", part.TaxonomyNodeID, "thick-film")
	}
}

func TestPartNumberIdentityFlow(t *testing.T) {
	seqDef := NewSeqDef(Concat(
		Bind("category", Range(0, 999).Width(3)),
		Literal("-"),
		Bind("family", Range(0, 99999).Width(5)),
		Literal("-"),
		Bind("id", Range(0, 999).Width(3)),
	)).WithID("component-pn-v1")

	taxonomyDef := TaxonomyDef{
		ID:     "component-taxonomy-v1",
		SeqDef: seqDef.ID,
		Taxonomy: Taxonomy{Root: TaxonomyNode{
			ID: "components",
			Children: []TaxonomyNode{{
				ID:         "resistors",
				Predicates: []TaxonomyPredicate{ExactPredicate("category", "001")},
				Children: []TaxonomyNode{{
					ID:         "thick-film",
					Predicates: []TaxonomyPredicate{ExactPredicate("family", "00042")},
				}},
			}},
		}},
	}

	part, err := NewPartNumber("001-00042-001", seqDef, taxonomyDef)
	if err != nil {
		t.Fatalf("NewPartNumber() error = %v", err)
	}

	want := PartNumber{
		Value:          "001-00042-001",
		SeqDefID:       "component-pn-v1",
		TaxonomyDefID:  "component-taxonomy-v1",
		TaxonomyNodeID: "thick-film",
	}
	if part.Value != want.Value ||
		part.SeqDefID != want.SeqDefID ||
		part.TaxonomyDefID != want.TaxonomyDefID ||
		part.TaxonomyNodeID != want.TaxonomyNodeID {
		t.Fatalf("PartNumber = %#v, want %#v", part, want)
	}
}

func TestNewPartNumberRejectsInvalidDefinitionLinks(t *testing.T) {
	seqDef := NewSeqDef(Literal("PN-1")).WithID("pn-v1")

	tests := []struct {
		name     string
		value    string
		seqDef   SeqDef
		taxonomy TaxonomyDef
	}{
		{
			name:     "missing sequence ID",
			value:    "PN-1",
			seqDef:   NewSeqDef(Literal("PN-1")),
			taxonomy: TaxonomyDef{ID: "taxonomy-v1"},
		},
		{
			name:     "missing taxonomy ID",
			value:    "PN-1",
			seqDef:   seqDef,
			taxonomy: TaxonomyDef{SeqDef: seqDef.ID},
		},
		{
			name:     "taxonomy points elsewhere",
			value:    "PN-1",
			seqDef:   seqDef,
			taxonomy: TaxonomyDef{ID: "taxonomy-v1", SeqDef: "other-pn-v1"},
		},
		{
			name:     "invalid value",
			value:    "PN-2",
			seqDef:   seqDef,
			taxonomy: TaxonomyDef{ID: "taxonomy-v1", SeqDef: seqDef.ID},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewPartNumber(test.value, test.seqDef, test.taxonomy); err == nil {
				t.Fatal("NewPartNumber() accepted invalid input")
			}
		})
	}
}

func TestPartNumberValidate(t *testing.T) {
	valid := PartNumber{
		Value:          "PN-1",
		SeqDefID:       "pn-v1",
		TaxonomyDefID:  "taxonomy-v1",
		TaxonomyNodeID: "root",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name string
		part PartNumber
		want bool
	}{
		{name: "missing ID", part: PartNumber{
			SeqDefID: "v1", TaxonomyDefID: "v1", TaxonomyNodeID: "root",
		}},
		{name: "missing value", part: PartNumber{
			ID: "1", SeqDefID: "pn-v1", TaxonomyDefID: "taxonomy-v1", TaxonomyNodeID: "root",
		}, want: true},
		{name: "missing sequence definition ID", part: PartNumber{
			ID: "1", Value: "PN-1", TaxonomyDefID: "taxonomy-v1", TaxonomyNodeID: "root",
		}, want: true},
		{name: "missing taxonomy definition ID", part: PartNumber{
			ID: "1", Value: "PN-1", SeqDefID: "pn-v1", TaxonomyNodeID: "root",
		}, want: true},
		{name: "missing taxonomy node ID", part: PartNumber{
			ID: "1", Value: "PN-1", SeqDefID: "pn-v1", TaxonomyDefID: "taxonomy-v1",
		}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.part.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid part number")
			}
		})
	}
}

func TestPartNumberStatus(t *testing.T) {
	part := PartNumber{Value: "PN-1", SeqDefID: "pn-v1", TaxonomyDefID: "taxonomy-v1", TaxonomyNodeID: "root"}
	if part.EffectiveStatus() != PartNumberStatusActive {
		t.Fatalf("zero status = %q, want active", part.EffectiveStatus())
	}
	part.Status = PartNumberStatusArchived
	if err := part.Validate(); err != nil {
		t.Fatalf("archived part validation error = %v", err)
	}
	part.Status = "retired"
	if err := part.Validate(); err == nil {
		t.Fatal("invalid part status accepted")
	}
}

func TestLineItemValidate(t *testing.T) {
	valid := LineItem{PartNumberID: "part-1", PartRevisionID: "revision-1"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name string
		item LineItem
	}{
		{name: "missing part number ID", item: LineItem{PartRevisionID: "revision-1"}},
		{name: "missing part revision ID", item: LineItem{PartNumberID: "part-1"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.item.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid line item")
			}
		})
	}
}

func TestPartRevisionValidate(t *testing.T) {
	valid := PartRevision{
		PartNumberID:     "part-1",
		Revision:         "1",
		RevisionSeqDefID: "revision-v1",
		BOM: []LineItem{{
			PartNumberID:   "child-part-1",
			PartRevisionID: "child-revision-1",
		}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name     string
		revision PartRevision
	}{
		{name: "missing part number ID", revision: PartRevision{}},
		{name: "invalid BOM line item", revision: PartRevision{
			PartNumberID: "part-1",
			BOM:          []LineItem{{PartNumberID: "child-part-1"}},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.revision.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid part revision")
			}
		})
	}
}
