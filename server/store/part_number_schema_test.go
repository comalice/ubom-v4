package store

import (
	"database/sql"
	"testing"

	ubom "ubom-v4"
)

func schemaStoreFixture(t *testing.T, s Store) ubom.TaxonomyDefID {
	t.Helper()
	seq := ubom.NewSeqDef(ubom.Bind("kind", ubom.Literal("R"))).WithID("schema-seq")
	if err := s.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	taxonomy := ubom.TaxonomyDef{ID: "schema-tax-v1", DefinitionID: "schema-tax", Version: 1, SeqDef: seq.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{ID: "root", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("kind", "R")}}}}
	if err := s.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	return taxonomy.ID
}

func TestMemoryPartNumberSchemasAreImmutableAndAllowDuplicateTaxonomyReferences(t *testing.T) {
	s := NewMemoryStore()
	taxonomyID := schemaStoreFixture(t, s)
	first := ubom.PartNumberSchema{ID: "schema-a", Name: "A", TaxonomyDef: taxonomyID, Policy: ubom.SchemaPolicy{RequireClassification: true}}
	second := ubom.PartNumberSchema{ID: "schema-b", Name: "B", TaxonomyDef: taxonomyID}
	if err := s.CreatePartNumberSchema(first); err != nil {
		t.Fatal(err)
	}
	if err := s.CreatePartNumberSchema(second); err != nil {
		t.Fatal(err)
	}
	first.Name = "changed"
	got, err := s.GetPartNumberSchema(first.ID)
	if err != nil || got.Name != "A" || !got.Policy.RequireClassification {
		t.Fatalf("stored schema changed: %#v, %v", got, err)
	}
	if err := s.CreatePartNumberSchema(ubom.PartNumberSchema{ID: "schema-a", Name: "again", TaxonomyDef: taxonomyID}); err != ErrAlreadyExists {
		t.Fatalf("duplicate schema error = %v, want ErrAlreadyExists", err)
	}
	nextTaxonomy := taxonomyWithVersion(t, s, taxonomyID)
	if got, err := s.GetPartNumberSchema(first.ID); err != nil || got.TaxonomyDef != taxonomyID {
		t.Fatalf("schema taxonomy changed after creating %s: %#v, %v", nextTaxonomy, got, err)
	}
	if schemas, err := s.ListPartNumberSchemas(); err != nil || len(schemas) != 2 {
		t.Fatalf("ListPartNumberSchemas() = %#v, %v", schemas, err)
	}
}

func taxonomyWithVersion(t *testing.T, s Store, parentID ubom.TaxonomyDefID) ubom.TaxonomyDefID {
	t.Helper()
	parent, err := s.GetTaxonomyDef(parentID)
	if err != nil {
		t.Fatal(err)
	}
	next := parent
	next.ID = "schema-tax-v2"
	next.Version = 2
	next.ParentID = parent.ID
	if _, err := s.CreateTaxonomyRevision(parent.SeqDef, parent.ID, next); err != nil {
		t.Fatal(err)
	}
	return next.ID
}

func TestSQLitePartNumberSchemaRoundTripAndForeignKey(t *testing.T) {
	db, err := sql.Open("sqlite", "file:schema-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSQLiteStore(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	taxonomyID := schemaStoreFixture(t, s)
	want := ubom.PartNumberSchema{ID: "sqlite-schema", Name: "SQLite", TaxonomyDef: taxonomyID, Policy: ubom.SchemaPolicy{RequireClassification: true}}
	if err := s.CreatePartNumberSchema(want); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetPartNumberSchema(want.ID)
	if err != nil || got != want {
		t.Fatalf("GetPartNumberSchema() = %#v, %v; want %#v", got, err, want)
	}
	if err := s.CreatePartNumberSchema(want); err != ErrAlreadyExists {
		t.Fatalf("duplicate schema error = %v, want ErrAlreadyExists", err)
	}
	if err := s.CreatePartNumberSchema(ubom.PartNumberSchema{ID: "missing-tax", Name: "Missing", TaxonomyDef: "no-such-tax"}); err != ErrNotFound {
		t.Fatalf("missing taxonomy error = %v, want ErrNotFound", err)
	}
}
