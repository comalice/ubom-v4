package ubom

import "testing"

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
				ID:      "resistors",
				Matches: map[string]string{"category": "001"},
				Children: []TaxonomyNode{{
					ID:      "thick-film",
					Matches: map[string]string{"family": "00042"},
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
				ID:      "resistors",
				Matches: map[string]string{"category": "001"},
				Children: []TaxonomyNode{{
					ID:      "thick-film",
					Matches: map[string]string{"family": "00042"},
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
		PartNumberID: "part-1",
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
