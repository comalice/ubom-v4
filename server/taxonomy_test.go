package ubom

import (
	"errors"
	"reflect"
	"testing"
)

func TestLegacyTaxonomyPredicateMatching(t *testing.T) {
	node := TaxonomyNode{Predicates: []TaxonomyPredicate{
		ExactPredicate("category", "RES"),
		NumericRangePredicate("group", 100, 199),
	}}
	if !node.predicatesMatch(map[string]string{"category": "RES", "group": "100"}) {
		t.Fatal("predicate did not match lower inclusive bounds")
	}
	if !node.predicatesMatch(map[string]string{"category": "RES", "group": "199"}) {
		t.Fatal("predicate did not match upper inclusive bounds")
	}
	if node.predicatesMatch(map[string]string{"category": "res", "group": "160"}) {
		t.Fatal("exact predicate matched a different value")
	}
	if node.predicatesMatch(map[string]string{"category": "RES", "group": "200"}) {
		t.Fatal("range predicate matched an out-of-range value")
	}
	taxonomy := Taxonomy{Root: TaxonomyNode{ID: "root", Children: []TaxonomyNode{{
		ID: "high-power", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)},
	}}}}
	if got, err := taxonomy.ProjectNode(ParseResult{Bindings: map[string]string{"group": "160"}}); err != nil || got != "high-power" {
		t.Fatalf("ProjectNode() = %q, %v; want high-power", got, err)
	}
}

func TestTaxonomyClassifyUsesStructuredCapturesAndEffectiveMeaning(t *testing.T) {
	seqDef := NewSeqDef(Concat(
		Literal("RES-"),
		Bind("category", PlaceSequence("R", "E", "S")),
		Literal("-"),
		Bind("group", Range(0, 999).Width(3)),
	)).WithID("parts")
	parsed, err := seqDef.ParseValues("RES-RES-160")
	if err != nil {
		t.Fatal(err)
	}
	taxonomy := Taxonomy{Root: TaxonomyNode{
		ID: "root", Label: "component",
		Attributes: []AttributeAssignment{{AttributeDefID: "family"}},
		Children: []TaxonomyNode{{
			ID: "resistor", Label: "resistor",
			Predicates: []TaxonomyPredicate{ExactPredicate("category", "RES")},
			Attributes: []AttributeAssignment{{AttributeDefID: "family", Required: true}, {AttributeDefID: "kind"}},
			Children: []TaxonomyNode{{
				ID: "high-power", Label: "high power",
				Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)},
				Attributes: []AttributeAssignment{{AttributeDefID: "kind", Required: true}},
			}},
		}},
	}}
	classification, err := taxonomy.Classify(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if classification.NodeID != "high-power" || classification.Generic {
		t.Fatalf("classification = %#v, want specific high-power", classification)
	}
	if !reflect.DeepEqual(classification.Path, []TaxonomyNodeID{"root", "resistor", "high-power"}) {
		t.Fatalf("path = %#v", classification.Path)
	}
	if !reflect.DeepEqual(classification.Labels, []string{"component", "resistor", "high power"}) {
		t.Fatalf("labels = %#v", classification.Labels)
	}
	if !reflect.DeepEqual(classification.Attributes, []AttributeAssignment{{AttributeDefID: "family", Required: true}, {AttributeDefID: "kind", Required: true}}) {
		t.Fatalf("attributes = %#v", classification.Attributes)
	}
	if len(classification.Predicates) != 2 {
		t.Fatalf("effective predicates = %#v, want ancestor and child", classification.Predicates)
	}
}

func TestTaxonomyClassifyIsCumulativeAndPolicyAware(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{ID: "root", Label: "component", Children: []TaxonomyNode{{
		ID: "resistor", Predicates: []TaxonomyPredicate{ExactPredicate("category", "RES")}, Children: []TaxonomyNode{{
			ID: "high-power", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)},
		}},
	}}}}
	parsed := ParsedSequence{Fields: []CapturedField{{Name: "category", Source: "CAP"}, {Name: "group", Source: "160", Normalized: 160}}}
	classification, err := taxonomy.Classify(parsed)
	if err != nil || classification.NodeID != "root" || !classification.Generic {
		t.Fatalf("classification = %#v, %v; want generic root", classification, err)
	}
	if _, err := taxonomy.ClassifyWithPolicy(parsed, ClassificationPolicy{RequireClassification: true}); !IsCategory(err, ErrorCategoryClassification) {
		t.Fatalf("closed classification error = %v, want classification category", err)
	}
	missing := ParsedSequence{Fields: []CapturedField{{Name: "category", Source: "RES"}}}
	classification, err = taxonomy.Classify(missing)
	if err != nil || classification.NodeID != "resistor" {
		t.Fatalf("missing child field classification = %#v, %v; want resistor", classification, err)
	}
}

func TestTaxonomyClassifyNoRootAndLabelsDoNotChangeIDs(t *testing.T) {
	if _, err := (Taxonomy{}).Classify(ParsedSequence{}); !IsCategory(err, ErrorCategoryClassification) {
		t.Fatalf("no-root error = %v, want classification category", err)
	}
	taxonomy := Taxonomy{Root: TaxonomyNode{ID: "r", Label: "Root label", Children: []TaxonomyNode{{ID: "n", Label: "Node label"}}}}
	classification, err := taxonomy.Classify(ParsedSequence{})
	if err != nil || classification.NodeID != "n" || !reflect.DeepEqual(classification.Labels, []string{"Root label", "Node label"}) {
		t.Fatalf("classification = %#v, %v; want node ID and labels", classification, err)
	}
}

func TestTaxonomyClassifyRejectsInvalidPredicate(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		ID: "root",
		Children: []TaxonomyNode{{
			ID:         "invalid",
			Predicates: []TaxonomyPredicate{{Field: "code", Kind: PredicateKind("unsupported")}},
		}},
	}}
	if _, err := taxonomy.Classify(ParsedSequence{}); !IsCategory(err, ErrorCategoryDefinition) {
		t.Fatalf("Classify() error = %v, want definition error", err)
	}
}

func TestTaxonomyConstraintsReturnsCumulativePredicates(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		ID:         "root",
		Predicates: []TaxonomyPredicate{ExactPredicate("category", "RES")},
		Children: []TaxonomyNode{{
			ID:         "resistor",
			Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 0, 999)},
			Children: []TaxonomyNode{{
				ID:         "high-power",
				Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199), ExactPredicate("finish", "THICK")},
			}},
		}},
	}}

	got, err := taxonomy.Constraints("high-power")
	if err != nil {
		t.Fatal(err)
	}
	want := []TaxonomyPredicate{
		ExactPredicate("category", "RES"),
		NumericRangePredicate("group", 100, 199),
		ExactPredicate("finish", "THICK"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Constraints() = %#v, want %#v", got, want)
	}
}

func TestTaxonomyConstraintsRootUnknownAndOwnership(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{ID: "root", Predicates: []TaxonomyPredicate{ExactPredicate("category", "RES")}}}
	got, err := taxonomy.Constraints("root")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []TaxonomyPredicate{ExactPredicate("category", "RES")}) {
		t.Fatalf("root Constraints() = %#v", got)
	}
	got[0].Value = "CAP"
	again, err := taxonomy.Constraints("root")
	if err != nil || again[0].Value != "RES" {
		t.Fatalf("Constraints() exposed taxonomy storage: %#v, %v", again, err)
	}
	if _, err := taxonomy.Constraints("missing"); err == nil {
		t.Fatal("Constraints() accepted an unknown node")
	}
}

func TestTaxonomyConstraintsRejectsConflictsAndEmptyIntersections(t *testing.T) {
	tests := []TaxonomyNode{
		{ID: "root", Predicates: []TaxonomyPredicate{ExactPredicate("code", "A")}, Children: []TaxonomyNode{{ID: "child", Predicates: []TaxonomyPredicate{ExactPredicate("code", "B")}}}},
		{ID: "root", Predicates: []TaxonomyPredicate{NumericRangePredicate("number", 1, 5)}, Children: []TaxonomyNode{{ID: "child", Predicates: []TaxonomyPredicate{NumericRangePredicate("number", 6, 9)}}}},
	}
	for _, root := range tests {
		if _, err := (Taxonomy{Root: root}).Constraints("child"); err == nil {
			t.Fatalf("Constraints() accepted conflicting path: %#v", root)
		}
	}
}

func TestTaxonomyPredicateValidation(t *testing.T) {
	tests := []TaxonomyPredicate{
		{Field: "", Kind: PredicateExact, Value: "x"},
		{Field: "group", Kind: PredicateExact},
		{Field: "group", Kind: PredicateNumericRange, Min: 10, Max: 9},
		{Field: "group", Kind: PredicateKind("other"), Value: "x"},
	}
	for _, predicate := range tests {
		if err := predicate.Validate(); err == nil || !errors.Is(err, ErrValidation) {
			t.Errorf("Validate(%+v) error = %v, want validation error", predicate, err)
		}
	}
}

func TestTaxonomyValidateAgainstSeqDefFields(t *testing.T) {
	seqDef := NewSeqDef(Concat(
		Literal("RES-"),
		Bind("group", Range(0, 999).Width(3)),
	)).WithID("pn-v1")
	valid := TaxonomyDef{ID: "taxonomy-v1", SeqDef: "pn-v1", Taxonomy: Taxonomy{Root: TaxonomyNode{
		ID: "root", Children: []TaxonomyNode{{ID: "resistor", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)}}},
	}}}
	if err := valid.ValidateAgainst(seqDef); err != nil {
		t.Fatalf("ValidateAgainst() error = %v", err)
	}
	valid.Taxonomy.Root.Children[0].Predicates[0].Field = "missing"
	if err := valid.ValidateAgainst(seqDef); err == nil {
		t.Fatal("ValidateAgainst() accepted an unknown captured field")
	}
	textual := NewSeqDef(Concat(Literal("PN-"), Bind("code", PlaceSequence("A", "B")))).WithID("text-v1")
	numeric := TaxonomyDef{ID: "taxonomy-v1", SeqDef: textual.ID, Taxonomy: Taxonomy{Root: TaxonomyNode{
		ID: "root", Predicates: []TaxonomyPredicate{NumericRangePredicate("code", 1, 2)},
	}}}
	if err := numeric.ValidateAgainst(textual); err == nil {
		t.Fatal("ValidateAgainst() accepted a numeric predicate on a non-range field")
	}
	// Legacy exact matches remain source-text predicates and are valid for any
	// captured field, including fields whose structural node is numeric.
	exact := numeric
	exact.Taxonomy.Root.Predicates = nil
	exact.Taxonomy.Root.Matches = map[string]string{"code": "A"}
	if err := exact.ValidateAgainst(textual); err != nil {
		t.Fatalf("ValidateAgainst() rejected a textual exact predicate: %v", err)
	}
}

func TestTaxonomyLegacyMatchesCannotConflictWithTypedPredicates(t *testing.T) {
	definition := TaxonomyDef{ID: "taxonomy-v1", SeqDef: "pn-v1", Taxonomy: Taxonomy{Root: TaxonomyNode{
		ID: "root", Matches: map[string]string{"category": "RES"}, Predicates: []TaxonomyPredicate{ExactPredicate("category", "CAP")},
	}}}
	if err := definition.Validate(); err == nil {
		t.Fatal("Validate() accepted conflicting legacy and typed predicates")
	}
}

func TestTaxonomyValidateRejectsInvalidTreeRelationships(t *testing.T) {
	base := TaxonomyDef{ID: "taxonomy-v1", SeqDef: "pn-v1"}
	tests := []struct {
		name string
		root TaxonomyNode
	}{
		{
			name: "overlapping sibling ranges",
			root: TaxonomyNode{ID: "root", Children: []TaxonomyNode{
				{ID: "x", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)}},
				{ID: "y", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 150, 175)}},
			}},
		},
		{
			name: "conflicting nested refinement",
			root: TaxonomyNode{ID: "root", Children: []TaxonomyNode{{
				ID: "x", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)}, Children: []TaxonomyNode{{
					ID: "y", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 200, 300)},
				}},
			}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base.Taxonomy = Taxonomy{Root: test.root}
			if err := base.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid taxonomy")
			}
		})
	}

	// The same ranges are valid when the narrower meaning is a child: it is a
	// refinement, not an ambiguous sibling alternative.
	base.Taxonomy = Taxonomy{Root: TaxonomyNode{ID: "root", Children: []TaxonomyNode{{
		ID: "x", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 100, 199)}, Children: []TaxonomyNode{{
			ID: "y", Predicates: []TaxonomyPredicate{NumericRangePredicate("group", 150, 175)},
		}},
	}}}}
	if err := base.Validate(); err != nil {
		t.Fatalf("Validate() rejected nested refinement: %v", err)
	}
}

func TestLegacyTaxonomyProject(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		Label: "component",
		Children: []TaxonomyNode{
			{
				Label:   "resistor",
				Matches: map[string]string{"category": "001"},
				Children: []TaxonomyNode{
					{
						Label:   "thick-film",
						Matches: map[string]string{"family": "00042"},
					},
				},
			},
			{
				Label:   "capacitor",
				Matches: map[string]string{"category": "002"},
			},
		},
	}}

	got := taxonomy.Project(ParseResult{Bindings: map[string]string{
		"category": "001",
		"family":   "00042",
		"id":       "001",
	}})
	want := []string{"component", "resistor", "thick-film"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Project() = %#v, want %#v", got, want)
	}
}

func TestLegacyTaxonomyProjectStopsAtUnknownValue(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		Label: "component",
		Children: []TaxonomyNode{{
			Label:   "resistor",
			Matches: map[string]string{"category": "001"},
		}},
	}}

	got := taxonomy.Project(ParseResult{Bindings: map[string]string{
		"category": "999",
	}})
	want := []string{"component"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Project() = %#v, want %#v", got, want)
	}
}

func TestTaxonomyDefValidate(t *testing.T) {
	tests := []struct {
		name string
		def  TaxonomyDef
		want bool
	}{
		{
			name: "valid",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{ID: "resistors"}},
				}},
			},
			want: true,
		},
		{
			name: "missing ID",
			def: TaxonomyDef{
				SeqDef:   "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root"}},
			},
		},
		{
			name: "missing sequence definition ID",
			def: TaxonomyDef{
				ID:       "taxonomy-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root"}},
			},
		},
		{
			name: "missing root ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
			},
		},
		{
			name: "missing child ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{}},
				}},
			},
		},
		{
			name: "duplicate node ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{ID: "root"}},
				}},
			},
		},
		{
			name: "duplicate node attribute assignment",
			def:  TaxonomyDef{ID: "taxonomy-v1", SeqDef: "pn-v1", AttributeDefs: []AttributeDef{{ID: "size", Label: "Size", ValueType: AttributeValueString}}, Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root", Attributes: []AttributeAssignment{{AttributeDefID: "size"}, {AttributeDefID: "size"}}}}},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.def.Validate()
			if test.want && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !test.want && err == nil {
				t.Fatal("Validate() accepted invalid definition")
			}
		})
	}
}

func TestTaxonomyEffectiveAttributesChildOverridesAncestor(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		ID: "root",
		Attributes: []AttributeAssignment{
			{AttributeDefID: "footprint", Required: false},
			{AttributeDefID: "material", Required: true},
		},
		Children: []TaxonomyNode{{
			ID: "resistors",
			Attributes: []AttributeAssignment{
				{AttributeDefID: "footprint", Required: true},
				{AttributeDefID: "resistance", Required: true},
			},
		}},
	}}

	got, err := taxonomy.EffectiveAttributes("resistors")
	if err != nil {
		t.Fatalf("EffectiveAttributes() error = %v", err)
	}
	want := []AttributeAssignment{
		{AttributeDefID: "footprint", Required: true},
		{AttributeDefID: "material", Required: true},
		{AttributeDefID: "resistance", Required: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EffectiveAttributes() = %#v, want %#v", got, want)
	}
}
