package store

import (
	"errors"
	"reflect"
	"testing"

	ubom "ubom-v4"
)

func TestMemoryStoreRoundTrip(t *testing.T) {
	store := NewMemoryStore()
	seqDef := ubom.NewSeqDef(ubom.Literal("PN-1")).WithID("pn-v1")
	taxonomyDef := ubom.TaxonomyDef{ID: "taxonomy-v1", SeqDef: seqDef.ID}
	part := ubom.PartNumber{
		Value:          "PN-1",
		SeqDefID:       seqDef.ID,
		TaxonomyDefID:  taxonomyDef.ID,
		TaxonomyNodeID: "root",
	}

	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatalf("CreateSeqDef() error = %v", err)
	}
	if err := store.CreateTaxonomyDef(taxonomyDef); err != nil {
		t.Fatalf("CreateTaxonomyDef() error = %v", err)
	}
	if err := store.CreatePartNumber(part); err != nil {
		t.Fatalf("CreatePartNumber() error = %v", err)
	}

	got, err := store.GetPartNumber(part.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() error = %v", err)
	}
	if !reflect.DeepEqual(got, part) {
		t.Fatalf("GetPartNumber() = %#v, want %#v", got, part)
	}
}

func TestMemoryStoreRejectsDuplicatePartNumber(t *testing.T) {
	store := NewMemoryStore()
	part := ubom.PartNumber{Value: "PN-1"}

	if err := store.CreatePartNumber(part); err != nil {
		t.Fatalf("first CreatePartNumber() error = %v", err)
	}
	if err := store.CreatePartNumber(part); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("second CreatePartNumber() error = %v, want ErrAlreadyExists", err)
	}
}
