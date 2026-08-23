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
