package store

import (
	"errors"
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
	revisionSeqDef := ubom.NewSeqDef(ubom.Range(1, 9).Width(1)).WithID("revision-v1")
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
	if err := store.CreateSeqDef(revisionSeqDef); err != nil {
		t.Fatalf("CreateSeqDef(revision) error = %v", err)
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
	gotPart, err = store.GetPartNumberByID(part.ID)
	if err != nil {
		t.Fatalf("GetPartNumberByID() error = %v", err)
	}
	if !reflect.DeepEqual(gotPart, part) {
		t.Fatalf("GetPartNumberByID() = %#v, want %#v", gotPart, part)
	}
	if _, err := store.GetPartNumberByID("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetPartNumberByID(missing) error = %v, want ErrNotFound", err)
	}

	revision, err := store.CreatePartRevision(ubom.PartRevision{PartNumberID: part.ID, Revision: "1", RevisionSeqDefID: revisionSeqDef.ID})
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
	gotPart, err = store.GetPartNumber(part.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() after revision error = %v", err)
	}
	if !reflect.DeepEqual(gotPart.PartRevisionID, []ubom.PartRevisionID{revision.ID}) {
		t.Fatalf("PartRevisionID = %#v, want %#v", gotPart.PartRevisionID, []ubom.PartRevisionID{revision.ID})
	}

	parent, err := store.CreatePartRevision(ubom.PartRevision{
		PartNumberID: part.ID, Revision: "2", RevisionSeqDefID: revisionSeqDef.ID,
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
