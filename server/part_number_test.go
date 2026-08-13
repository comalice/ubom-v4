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
	taxonomy := TaxonomyDef{ID: "components-v1", SeqDef: seqDef.ID}

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
