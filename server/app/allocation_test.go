package app

import (
	"errors"
	"testing"

	ubom "ubom-v4"
	"ubom-v4/store"
)

func allocationFixture(t *testing.T, root ubom.TaxonomyNode, seed string) (*Service, ubom.SeqDef, *store.MemoryStore) {
	t.Helper()
	seq := ubom.NewSeqDef(ubom.Concat(ubom.Literal("PN-"), ubom.Bind("number", ubom.Range(0, 3).Width(1)))).WithID("allocation-seq")
	taxonomy := ubom.TaxonomyDef{ID: "allocation-taxonomy", SeqDef: seq.ID, Taxonomy: ubom.Taxonomy{Root: root}}
	metadata := store.NewMemoryStore()
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreatePartNumberDraft(ubom.PartNumberDraft{ID: "draft-1", SeqDefID: seq.ID, TaxonomyDefID: taxonomy.ID, ContextNodeID: "context", Bindings: map[string]string{"number": seed}, Status: ubom.DraftStatusDraft}); err != nil {
		t.Fatal(err)
	}
	return NewService(metadata), seq, metadata
}

func TestAllocateDraftAcceptsContextAndDescendant(t *testing.T) {
	service, _, _ := allocationFixture(t, ubom.TaxonomyNode{ID: "root", Children: []ubom.TaxonomyNode{{ID: "context", Children: []ubom.TaxonomyNode{{ID: "descendant", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("number", 2, 3)}}}}}}, "2")
	got, err := service.AllocatePartNumberDraft("draft-1")
	if err != nil || got.Value != "PN-2" {
		t.Fatalf("allocation = %#v, %v; want PN-2", got, err)
	}
}

func TestAllocateDraftRevalidatesAfterCollisions(t *testing.T) {
	service, seq, metadata := allocationFixture(t, ubom.TaxonomyNode{ID: "root", Children: []ubom.TaxonomyNode{{ID: "context", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("number", 2, 3)}}}}, "2")
	for _, value := range []string{"PN-2"} {
		part, err := ubom.NewSchema(seq, mustTaxonomy(t, metadata), ubom.SchemaPolicy{}).NewPartNumber(value, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := metadata.CreatePartNumber(part); err != nil {
			t.Fatal(err)
		}
	}
	got, err := service.AllocatePartNumberDraft("draft-1")
	if err != nil || got.Value != "PN-3" {
		t.Fatalf("allocation = %#v, %v; want PN-3", got, err)
	}
}

func TestAllocateDraftReportsExhaustionAndNoOrdering(t *testing.T) {
	service, seq, metadata := allocationFixture(t, ubom.TaxonomyNode{ID: "root", Children: []ubom.TaxonomyNode{{ID: "context", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("number", 2, 3)}}}}, "2")
	taxonomy := mustTaxonomy(t, metadata)
	for _, value := range []string{"PN-2", "PN-3"} {
		part, err := ubom.NewSchema(seq, taxonomy, ubom.SchemaPolicy{}).NewPartNumber(value, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := metadata.CreatePartNumber(part); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.AllocatePartNumberDraft("draft-1"); err == nil || !errors.Is(err, ubom.ErrAllocation) {
		t.Fatalf("exhaustion error = %v, want allocation category", err)
	}

	choice := ubom.NewSeqDef(ubom.Bind("kind", ubom.Choice(ubom.Literal("A"), ubom.Literal("B")))).WithID("choice-seq")
	if err := metadata.CreateSeqDef(choice); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(ubom.TaxonomyDef{ID: "choice-taxonomy", SeqDef: choice.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{ID: "root", Children: []ubom.TaxonomyNode{{ID: "a", Matches: map[string]string{"kind": "A"}}, {ID: "b", Matches: map[string]string{"kind": "B"}}}}}}); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreatePartNumberDraft(ubom.PartNumberDraft{ID: "choice-draft", SeqDefID: choice.ID, TaxonomyDefID: "choice-taxonomy", ContextNodeID: "b", Bindings: map[string]string{"kind": "A"}, Status: ubom.DraftStatusDraft}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AllocatePartNumberDraft("choice-draft"); !errors.Is(err, ubom.ErrNoOrdering) {
		t.Fatalf("no-ordering error = %v", err)
	}
}

func mustTaxonomy(t *testing.T, metadata *store.MemoryStore) ubom.TaxonomyDef {
	t.Helper()
	result, err := metadata.GetTaxonomyDef("allocation-taxonomy")
	if err != nil {
		t.Fatal(err)
	}
	return result
}
