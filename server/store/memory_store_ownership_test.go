package store

import (
	"testing"

	ubom "ubom-v4"
)

func TestMemoryStoreCopiesMutableValues(t *testing.T) {
	store := NewMemoryStore()
	seqDef := ubom.NewSeqDef(ubom.Literal("PN-1")).WithID("pn")
	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatal(err)
	}
	taxonomy := ubom.TaxonomyDef{ID: "tax", SeqDef: seqDef.ID, AttributeDefs: []ubom.AttributeDef{{ID: "x", Label: "X", ValueType: ubom.AttributeValueString}}, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{ID: "root", Matches: map[string]string{"kind": "one"}, Attributes: []ubom.AttributeAssignment{{AttributeDefID: "x"}}}}}
	if err := store.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	taxonomy.Taxonomy.Root.Matches["kind"] = "mutated"
	gotTaxonomy, err := store.GetTaxonomyDef(taxonomy.ID)
	if err != nil || gotTaxonomy.Taxonomy.Root.Matches["kind"] != "one" {
		t.Fatalf("taxonomy was not isolated: %#v, %v", gotTaxonomy, err)
	}
	part, err := store.CreatePartNumber(ubom.PartNumber{Value: "PN-1", SeqDefID: seqDef.ID, TaxonomyDefID: taxonomy.ID, TaxonomyNodeID: "root", Attributes: []ubom.PartNumberAttribute{{AttributeDefID: "x", Value: "y"}}})
	if err != nil {
		t.Fatal(err)
	}
	part.Attributes[0].Value = "mutated"
	part.PartRevisionID = append(part.PartRevisionID, "mutated")
	gotPart, err := store.GetPartNumber(part.Value)
	if err != nil || gotPart.Attributes[0].Value != "y" || len(gotPart.PartRevisionID) != 0 {
		t.Fatalf("part number was not isolated: %#v, %v", gotPart, err)
	}
}

func TestMemoryStoreCopiesTypedPredicates(t *testing.T) {
	store := NewMemoryStore()
	seqDef := ubom.NewSeqDef(ubom.Bind("kind", ubom.Literal("one"))).WithID("predicate-seq")
	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatal(err)
	}
	taxonomy := ubom.TaxonomyDef{ID: "predicate-tax", SeqDef: seqDef.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
		ID: "root", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("kind", "one")},
	}}}
	if err := store.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	taxonomy.Taxonomy.Root.Predicates[0].Value = "mutated"
	got, err := store.GetTaxonomyDef(taxonomy.ID)
	if err != nil || got.Taxonomy.Root.Predicates[0].Value != "one" {
		t.Fatalf("typed predicates were not isolated: %#v, %v", got, err)
	}
}
