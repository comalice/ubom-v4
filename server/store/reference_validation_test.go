package store

import (
	"errors"
	"testing"

	ubom "ubom-v4"
)

type revisionReferenceStore struct {
	revisionGraphStore
	seqDef ubom.SeqDef
	parts  map[ubom.PartNumberID]ubom.PartNumber
}

func (s revisionReferenceStore) GetSeqDef(id ubom.SeqDefID) (ubom.SeqDef, error) {
	if s.seqDef.ID != id {
		return ubom.SeqDef{}, ErrNotFound
	}
	return s.seqDef, nil
}

func (s revisionReferenceStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	part, ok := s.parts[id]
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return part, nil
}

type revisionGraphStore struct {
	Store
	revisions map[ubom.PartRevisionID]ubom.PartRevision
}

func (s revisionGraphStore) GetPartRevision(id ubom.PartRevisionID) (ubom.PartRevision, error) {
	revision, ok := s.revisions[id]
	if !ok {
		return ubom.PartRevision{}, ErrNotFound
	}
	return revision, nil
}

func TestValidatePartRevisionCyclesRejectsIndirectCycle(t *testing.T) {
	store := revisionGraphStore{revisions: map[ubom.PartRevisionID]ubom.PartRevision{
		"1": {BOM: []ubom.LineItem{{PartRevisionID: "2"}}},
		"2": {BOM: []ubom.LineItem{{PartRevisionID: "3"}}},
	}}
	revision := ubom.PartRevision{BOM: []ubom.LineItem{{PartRevisionID: "1"}}}
	if err := validatePartRevisionCycles(store, revision, "3"); !errors.Is(err, ubom.ErrConflict) {
		t.Fatal("validatePartRevisionCycles() accepted an indirect cycle")
	}
}

func TestValidatePartRevisionReferencesClassifiesSyntaxAndOwnership(t *testing.T) {
	seqDef := ubom.NewSeqDef(ubom.Range(1, 9).Width(1)).WithID("revision-v1")
	store := revisionReferenceStore{
		revisionGraphStore: revisionGraphStore{revisions: map[ubom.PartRevisionID]ubom.PartRevision{
			"child": {PartNumberID: "other-part"},
		}},
		seqDef: seqDef,
		parts: map[ubom.PartNumberID]ubom.PartNumber{
			"parent":     {ID: "parent"},
			"child-part": {ID: "child-part"},
		},
	}

	if err := validatePartRevisionReferences(store, ubom.PartRevision{
		PartNumberID: "parent", Revision: "x", RevisionSeqDefID: seqDef.ID,
	}); !errors.Is(err, ubom.ErrInvalidSyntax) {
		t.Fatalf("invalid revision error = %v, want ErrInvalidSyntax", err)
	}
	if err := validatePartRevisionReferences(store, ubom.PartRevision{
		PartNumberID: "parent", Revision: "1", RevisionSeqDefID: seqDef.ID,
		BOM: []ubom.LineItem{{PartNumberID: "child-part", PartRevisionID: "child"}},
	}); !errors.Is(err, ubom.ErrConflict) {
		t.Fatalf("ownership error = %v, want ErrConflict", err)
	}
}
