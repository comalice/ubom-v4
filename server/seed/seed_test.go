package seed

import (
	"testing"

	ubom "ubom-v4"
	"ubom-v4/store"
)

func TestPopulateAddsResistorAttributes(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := Populate(persistence, Options{Parts: 2}); err != nil {
		t.Fatalf("Populate() error = %v", err)
	}

	taxonomy, err := persistence.GetTaxonomyDef("sample-taxonomy-v1")
	if err != nil {
		t.Fatalf("GetTaxonomyDef() error = %v", err)
	}
	attributes, err := taxonomy.Taxonomy.EffectiveAttributes("resistors")
	if err != nil {
		t.Fatalf("EffectiveAttributes() error = %v", err)
	}
	if len(attributes) != 4 || attributes[0].AttributeDefID != "footprint" || attributes[1].AttributeDefID != "resistance" || attributes[2].AttributeDefID != "tolerance" || attributes[3].AttributeDefID != "power_rating" {
		t.Fatalf("resistor attributes = %#v, want footprint, resistance, tolerance, power_rating", attributes)
	}

	part, err := persistence.GetPartNumber("PN-A0000")
	if err != nil {
		t.Fatalf("GetPartNumber() error = %v", err)
	}
	if len(part.Attributes) != 4 || part.Attributes[0] != (ubom.PartNumberAttribute{AttributeDefID: "footprint", Value: "0603"}) || part.Attributes[1] != (ubom.PartNumberAttribute{AttributeDefID: "resistance", Value: "1000"}) {
		t.Fatalf("resistor part attributes = %#v, want sample resistor metadata", part.Attributes)
	}
}

func TestPopulateAddsGTDPartNumberSequence(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := Populate(persistence, Options{}); err != nil {
		t.Fatalf("Populate() error = %v", err)
	}

	definition, err := persistence.GetSeqDef("gtd-pn-v1")
	if err != nil {
		t.Fatalf("GetSeqDef() error = %v", err)
	}
	if got, err := definition.ParseValues("302-00103-001"); err != nil || got.Bindings["family"] != "302" || got.Bindings["group"] != "00103" || got.Bindings["uid"] != "001" {
		t.Fatalf("long GTD PN parse = %#v, %v", got, err)
	}
	if got, err := definition.ParseValues("050-00042"); err != nil || got.Bindings["family"] != "050" || got.Bindings["uid"] != "00042" || len(got.Bindings) != 2 {
		t.Fatalf("raw-material GTD PN parse = %#v, %v", got, err)
	}

	taxonomy, err := persistence.GetTaxonomyDef("gtd-pn-taxonomy-v1")
	if err != nil {
		t.Fatalf("GetTaxonomyDef() error = %v", err)
	}
	if len(taxonomy.Taxonomy.Root.Children) != 15 || taxonomy.Taxonomy.Root.Children[0].Label != "Raw Material" || taxonomy.Taxonomy.Root.Children[1].Label != "100-199" || len(taxonomy.Taxonomy.Root.Children[1].Children) != 6 {
		t.Fatalf("GTD PN taxonomy root = %#v", taxonomy.Taxonomy.Root)
	}
}
