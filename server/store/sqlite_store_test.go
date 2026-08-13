package store

import (
	"reflect"
	"testing"

	ubom "ubom-v4"
)

func TestSQLiteStoreRoundTrip(t *testing.T) {
	store, err := OpenSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	defer store.Close()

	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Bind("category", ubom.Range(0, 999).Width(3)),
		ubom.Literal("-"),
		ubom.Bind("id", ubom.Range(0, 99).Width(2)),
	)).WithID("pn-v1")
	taxonomyDef := ubom.TaxonomyDef{
		ID:     "taxonomy-v1",
		SeqDef: seqDef.ID,
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "components",
			Children: []ubom.TaxonomyNode{{
				ID:      "resistors",
				Matches: map[string]string{"category": "001"},
			}},
		}},
	}
	part := ubom.PartNumber{
		Value:          "001-42",
		SeqDefID:       seqDef.ID,
		TaxonomyDefID:  taxonomyDef.ID,
		TaxonomyNodeID: "resistors",
	}

	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatalf("CreateSeqDef() error = %v", err)
	}
	if err := store.CreateTaxonomyDef(taxonomyDef); err != nil {
		t.Fatalf("CreateTaxonomyDef() error = %v", err)
	}
	createdPart, err := store.CreatePartNumber(part)
	if err != nil {
		t.Fatalf("CreatePartNumber() error = %v", err)
	}
	if createdPart.ID == "" {
		t.Fatal("CreatePartNumber() returned an empty ID")
	}
	part = createdPart

	gotSeqDef, err := store.GetSeqDef(seqDef.ID)
	if err != nil {
		t.Fatalf("GetSeqDef() error = %v", err)
	}
	if got, err := gotSeqDef.ParseValues("001-42"); err != nil || got.Bindings["category"] != "001" {
		t.Fatalf("round-tripped SeqDef did not parse correctly: %#v, %v", got, err)
	}

	gotTaxonomy, err := store.GetTaxonomyDef(taxonomyDef.ID)
	if err != nil {
		t.Fatalf("GetTaxonomyDef() error = %v", err)
	}
	if !reflect.DeepEqual(gotTaxonomy, taxonomyDef) {
		t.Fatalf("GetTaxonomyDef() = %#v, want %#v", gotTaxonomy, taxonomyDef)
	}

	gotPart, err := store.GetPartNumber(part.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() error = %v", err)
	}
	if !reflect.DeepEqual(gotPart, part) {
		t.Fatalf("GetPartNumber() = %#v, want %#v", gotPart, part)
	}
}
