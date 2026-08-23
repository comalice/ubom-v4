package store

import (
	"testing"

	ubom "ubom-v4"
)

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
	if err := validatePartRevisionCycles(store, revision, "3"); err == nil {
		t.Fatal("validatePartRevisionCycles() accepted an indirect cycle")
	}
}
