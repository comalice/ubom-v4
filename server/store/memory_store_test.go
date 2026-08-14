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
	created, err := store.CreatePartNumber(part)
	if err != nil {
		t.Fatalf("CreatePartNumber() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("CreatePartNumber() returned an empty ID")
	}
	part = created

	got, err := store.GetPartNumber(part.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() error = %v", err)
	}
	if !reflect.DeepEqual(got, part) {
		t.Fatalf("GetPartNumber() = %#v, want %#v", got, part)
	}

	revision, err := store.CreatePartRevision(ubom.PartRevision{PartNumberID: part.ID})
	if err != nil {
		t.Fatalf("CreatePartRevision() error = %v", err)
	}
	gotRevision, err := store.GetPartRevision(revision.ID)
	if err != nil {
		t.Fatalf("GetPartRevision() error = %v", err)
	}
	if !reflect.DeepEqual(gotRevision, revision) {
		t.Fatalf("GetPartRevision() = %#v, want %#v", gotRevision, revision)
	}
	got, err = store.GetPartNumber(part.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() after revision error = %v", err)
	}
	if !reflect.DeepEqual(got.PartRevisionID, []ubom.PartRevisionID{revision.ID}) {
		t.Fatalf("PartRevisionID = %#v, want %#v", got.PartRevisionID, []ubom.PartRevisionID{revision.ID})
	}

	parent, err := store.CreatePartRevision(ubom.PartRevision{
		PartNumberID: part.ID,
		BOM: []ubom.LineItem{{
			PartNumberID:   part.ID,
			PartRevisionID: revision.ID,
		}},
	})
	if err != nil {
		t.Fatalf("CreatePartRevision() with BOM error = %v", err)
	}
	gotParent, err := store.GetPartRevision(parent.ID)
	if err != nil {
		t.Fatalf("GetPartRevision() with BOM error = %v", err)
	}
	if !reflect.DeepEqual(gotParent, parent) {
		t.Fatalf("GetPartRevision() with BOM = %#v, want %#v", gotParent, parent)
	}
}

func TestMemoryStoreRejectsDuplicatePartNumber(t *testing.T) {
	store := NewMemoryStore()
	part := ubom.PartNumber{Value: "PN-1"}

	if _, err := store.CreatePartNumber(part); err != nil {
		t.Fatalf("first CreatePartNumber() error = %v", err)
	}
	if _, err := store.CreatePartNumber(part); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("second CreatePartNumber() error = %v, want ErrAlreadyExists", err)
	}
}
