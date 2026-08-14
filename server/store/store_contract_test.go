package store

import (
	"errors"
	"reflect"
	"testing"

	ubom "ubom-v4"
)

func TestStoreContract(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) Store
	}{
		{
			name: "memory",
			build: func(t *testing.T) Store {
				return NewMemoryStore()
			},
		},
		{
			name: "sqlite",
			build: func(t *testing.T) Store {
				store, err := OpenSQLiteStore("file:store-contract?mode=memory&cache=shared")
				if err != nil {
					t.Fatalf("OpenSQLiteStore() error = %v", err)
				}
				t.Cleanup(func() { store.Close() })
				return store
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runStoreContract(t, test.build(t))
		})
	}
}

func runStoreContract(t *testing.T, store Store) {
	seqDef := ubom.NewSeqDef(ubom.Literal("PN-1")).WithID("pn-v1")
	taxonomyDef := ubom.TaxonomyDef{
		ID:     "taxonomy-v1",
		SeqDef: seqDef.ID,
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "root",
		}},
	}
	part := ubom.PartNumber{
		Value:          "PN-1",
		SeqDefID:       seqDef.ID,
		TaxonomyDefID:  taxonomyDef.ID,
		TaxonomyNodeID: "root",
	}

	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatalf("CreateSeqDef() error = %v", err)
	}
	gotSeqDef, err := store.GetSeqDef(seqDef.ID)
	if err != nil {
		t.Fatalf("GetSeqDef() error = %v", err)
	}
	if _, err := gotSeqDef.Parse("PN-1"); err != nil {
		t.Fatalf("round-tripped SeqDef.Parse() error = %v", err)
	}

	if err := store.CreateTaxonomyDef(taxonomyDef); err != nil {
		t.Fatalf("CreateTaxonomyDef() error = %v", err)
	}
	if err := store.CreateTaxonomyDef(ubom.TaxonomyDef{
		ID:     "missing-seq-taxonomy",
		SeqDef: "missing-seq-def",
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "root",
		}},
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("CreateTaxonomyDef() with missing sequence definition error = %v, want ErrNotFound", err)
	}
	gotTaxonomy, err := store.GetTaxonomyDef(taxonomyDef.ID)
	if err != nil {
		t.Fatalf("GetTaxonomyDef() error = %v", err)
	}
	if !reflect.DeepEqual(gotTaxonomy, taxonomyDef) {
		t.Fatalf("GetTaxonomyDef() = %#v, want %#v", gotTaxonomy, taxonomyDef)
	}

	createdPart, err := store.CreatePartNumber(part)
	if err != nil {
		t.Fatalf("CreatePartNumber() error = %v", err)
	}
	if _, err := store.CreatePartNumber(ubom.PartNumber{
		Value:          "PN-MISSING-SEQ",
		SeqDefID:       "missing-seq-def",
		TaxonomyDefID:  taxonomyDef.ID,
		TaxonomyNodeID: "root",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("CreatePartNumber() with missing sequence definition error = %v, want ErrNotFound", err)
	}
	if _, err := store.CreatePartNumber(ubom.PartNumber{
		Value:          "PN-MISSING-TAXONOMY",
		SeqDefID:       seqDef.ID,
		TaxonomyDefID:  "missing-taxonomy",
		TaxonomyNodeID: "root",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("CreatePartNumber() with missing taxonomy definition error = %v, want ErrNotFound", err)
	}
	if createdPart.ID == "" {
		t.Fatal("CreatePartNumber() returned an empty ID")
	}
	part = createdPart

	for _, lookup := range []struct {
		name string
		call func() (ubom.PartNumber, error)
	}{
		{name: "value", call: func() (ubom.PartNumber, error) {
			return store.GetPartNumber(part.Value)
		}},
		{name: "id", call: func() (ubom.PartNumber, error) {
			return store.GetPartNumberByID(part.ID)
		}},
	} {
		t.Run("part number by "+lookup.name, func(t *testing.T) {
			got, err := lookup.call()
			if err != nil {
				t.Fatalf("lookup error = %v", err)
			}
			if !reflect.DeepEqual(got, part) {
				t.Fatalf("lookup = %#v, want %#v", got, part)
			}
		})
	}

	if _, err := store.CreatePartNumber(part); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("duplicate CreatePartNumber() error = %v, want ErrAlreadyExists", err)
	}
	if _, err := store.GetPartNumber("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing value lookup error = %v, want ErrNotFound", err)
	}
	if _, err := store.GetPartNumberByID("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing ID lookup error = %v, want ErrNotFound", err)
	}

	child, err := store.CreatePartRevision(ubom.PartRevision{PartNumberID: part.ID})
	if err != nil {
		t.Fatalf("CreatePartRevision() error = %v", err)
	}
	parent, err := store.CreatePartRevision(ubom.PartRevision{
		PartNumberID: part.ID,
		BOM: []ubom.LineItem{{
			PartNumberID:   part.ID,
			PartRevisionID: child.ID,
		}},
	})
	if err != nil {
		t.Fatalf("CreatePartRevision() with BOM error = %v", err)
	}
	gotParent, err := store.GetPartRevision(parent.ID)
	if err != nil {
		t.Fatalf("GetPartRevision() error = %v", err)
	}
	if !reflect.DeepEqual(gotParent, parent) {
		t.Fatalf("GetPartRevision() = %#v, want %#v", gotParent, parent)
	}

	part, err = store.GetPartNumberByID(part.ID)
	if err != nil {
		t.Fatalf("GetPartNumberByID() after revisions error = %v", err)
	}
	if !reflect.DeepEqual(part.PartRevisionID, []ubom.PartRevisionID{child.ID, parent.ID}) {
		t.Fatalf("PartRevisionID = %#v, want %#v", part.PartRevisionID, []ubom.PartRevisionID{child.ID, parent.ID})
	}
}
