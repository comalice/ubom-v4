package app

import (
	"testing"

	ubom "ubom-v4"
	"ubom-v4/store"
)

func TestPartNumberViewUsesTypedClassificationAndEffectiveAttributes(t *testing.T) {
	metadata := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Concat(
		ubom.Literal("PN-"),
		ubom.Bind("number", ubom.Range(0, 99).Width(2)),
	)).WithID("view-seq")
	taxonomy := ubom.TaxonomyDef{
		ID: "view-taxonomy", SeqDef: seq.ID,
		AttributeDefs: []ubom.AttributeDef{
			{ID: "finish", Label: "Finish", ValueType: ubom.AttributeValueString},
		},
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "components", Label: "Components",
			Attributes: []ubom.AttributeAssignment{{AttributeDefID: "finish", Required: true}},
			Children: []ubom.TaxonomyNode{{
				ID: "precision", Label: "Precision",
				Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("number", 10, 20)},
				Attributes: []ubom.AttributeAssignment{{AttributeDefID: "finish", Required: false}},
			}},
		}},
	}
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	part, err := ubom.NewSchema(seq, taxonomy, ubom.SchemaPolicy{}).NewPartNumber("PN-12", []ubom.PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}})
	if err != nil {
		t.Fatal(err)
	}
	part, err = metadata.CreatePartNumber(part)
	if err != nil {
		t.Fatal(err)
	}

	view, err := NewService(metadata).GetPartNumberView(part.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.TaxonomyPath) != 2 || view.TaxonomyPath[0] != "Components" || view.TaxonomyPath[1] != "Precision" {
		t.Fatalf("taxonomy path = %#v, want [Components Precision]", view.TaxonomyPath)
	}
	if len(view.Attributes) != 1 || view.Attributes[0].ID != "finish" || view.Attributes[0].Value != "matte" {
		t.Fatalf("attributes = %#v, want effective overridden finish", view.Attributes)
	}
}

func TestTaxonomyNodeViewUsesExactNodeMembership(t *testing.T) {
	metadata := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Bind("kind", ubom.Choice(ubom.Literal("A"), ubom.Literal("B")))).WithID("exact-seq")
	taxonomy := ubom.TaxonomyDef{ID: "exact-taxonomy", SeqDef: seq.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
		ID: "root", Label: "Root", Children: []ubom.TaxonomyNode{{ID: "a", Label: "A", Matches: map[string]string{"kind": "A"}}},
	}}}
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	part, err := ubom.NewSchema(seq, taxonomy, ubom.SchemaPolicy{}).NewPartNumber("A", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := metadata.CreatePartNumber(part); err != nil {
		t.Fatal(err)
	}

	service := NewService(metadata)
	root, err := service.GetTaxonomyNodeView(taxonomy.ID, "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(root.PartNumbers) != 0 {
		t.Fatalf("root parts = %#v, want exact-node query to exclude descendant", root.PartNumbers)
	}
	child, err := service.GetTaxonomyNodeView(taxonomy.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(child.PartNumbers) != 1 || child.PartNumbers[0].Value != "A" {
		t.Fatalf("child parts = %#v, want A", child.PartNumbers)
	}
}

func TestRevisionViewRecursesBOM(t *testing.T) {
	metadata := store.NewMemoryStore()
	parent, err := LoadSampleData(metadata)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := metadata.GetPartNumber(parent.Value)
	if err != nil {
		t.Fatal(err)
	}
	view, err := NewService(metadata).GetRevisionView(stored.PartRevisionID[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(view.BOM) != 1 || view.BOM[0].PartNumber.Value != "PN-B" || view.BOM[0].RevisionID == "" {
		t.Fatalf("BOM = %#v, want recursive child revision", view.BOM)
	}
}
