package store

import (
	"database/sql"
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
				ID:         "resistors",
				Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("category", "001")},
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

func TestSQLiteStoreRoundTripsTypedTaxonomyPredicates(t *testing.T) {
	store, err := OpenSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLiteStore() error = %v", err)
	}
	defer store.Close()
	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Bind("group", ubom.Range(100, 999).Width(3)), ubom.Literal("-"),
		ubom.Bind("code", ubom.Range(0, 9).Width(1)),
	)).WithID("pn-range-v1")
	definition := ubom.TaxonomyDef{ID: "taxonomy-range-v1", SeqDef: seqDef.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
		ID: "root", Children: []ubom.TaxonomyNode{{ID: "mid", Predicates: []ubom.TaxonomyPredicate{
			ubom.NumericRangePredicate("group", 100, 199), ubom.ExactPredicate("code", "4"),
		}}},
	}}}
	if err := store.CreateSeqDef(seqDef); err != nil {
		t.Fatalf("CreateSeqDef() error = %v", err)
	}
	if err := store.CreateTaxonomyDef(definition); err != nil {
		t.Fatalf("CreateTaxonomyDef() error = %v", err)
	}
	got, err := store.GetTaxonomyDef(definition.ID)
	if err != nil {
		t.Fatalf("GetTaxonomyDef() error = %v", err)
	}
	if !reflect.DeepEqual(got, definition) {
		t.Fatalf("GetTaxonomyDef() = %#v, want %#v", got, definition)
	}
}

func TestSQLiteStoreMigratesLegacyTaxonomyNodes(t *testing.T) {
	// A legacy exact-match database is upgraded in place and keeps Matches.
	// This is deliberately tested separately from typed predicate persistence.
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE seq_defs (id TEXT PRIMARY KEY, ast TEXT NOT NULL);
		CREATE TABLE taxonomy_defs (id TEXT PRIMARY KEY, seq_def_id TEXT NOT NULL, attribute_defs TEXT NOT NULL);
		CREATE TABLE taxonomy_nodes (taxonomy_def_id TEXT NOT NULL, id TEXT NOT NULL, parent_id TEXT, label TEXT NOT NULL, matches TEXT NOT NULL, attributes TEXT NOT NULL, position INTEGER NOT NULL, PRIMARY KEY (taxonomy_def_id, id));`); err != nil {
		t.Fatal(err)
	}
	// NewSQLiteStore creates the rest of the schema and performs the column migration.
	store, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := db.Exec(`INSERT INTO seq_defs(id, ast) VALUES ('legacy-seq', ?)`, `{"id":"legacy-seq","root":{"kind":"literal","text":"X"}}`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO taxonomy_defs(id, seq_def_id, attribute_defs) VALUES ('legacy-tax', 'legacy-seq', '[]')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO taxonomy_nodes(taxonomy_def_id,id,parent_id,label,matches,attributes,position) VALUES ('legacy-tax','root',NULL,'Root','{"field":"value"}','[]',0)`); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetTaxonomyDef("legacy-tax")
	if err != nil || got.Taxonomy.Root.Matches["field"] != "value" {
		t.Fatalf("legacy taxonomy = %#v, error = %v", got, err)
	}
}
