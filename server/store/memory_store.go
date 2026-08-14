package store

import (
	"errors"
	"fmt"
	"strconv"

	ubom "ubom-v4"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type MemoryStore struct {
	seqDefs        map[ubom.SeqDefID]ubom.SeqDef
	taxonomyDefs   map[ubom.TaxonomyDefID]ubom.TaxonomyDef
	partNumbers    map[string]ubom.PartNumber
	partRevisions  map[ubom.PartRevisionID]ubom.PartRevision
	nextPartID     int64
	nextRevisionID int64
}

// compile time interface check
var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		seqDefs:        map[ubom.SeqDefID]ubom.SeqDef{},
		taxonomyDefs:   map[ubom.TaxonomyDefID]ubom.TaxonomyDef{},
		partNumbers:    map[string]ubom.PartNumber{},
		partRevisions:  map[ubom.PartRevisionID]ubom.PartRevision{},
		nextPartID:     1,
		nextRevisionID: 1,
	}
}

func (s *MemoryStore) CreateSeqDef(def ubom.SeqDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
	if _, ok := s.seqDefs[def.ID]; ok {
		return ErrAlreadyExists
	}
	s.seqDefs[def.ID] = def
	return nil
}

func (s *MemoryStore) GetSeqDef(id ubom.SeqDefID) (ubom.SeqDef, error) {
	def, ok := s.seqDefs[id]
	if !ok {
		return ubom.SeqDef{}, ErrNotFound
	}
	return def, nil
}

func (s *MemoryStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSeqDef(def.SeqDef); err != nil {
		return err
	}
	if _, ok := s.taxonomyDefs[def.ID]; ok {
		return ErrAlreadyExists
	}
	s.taxonomyDefs[def.ID] = def
	return nil
}

func (s *MemoryStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	def, ok := s.taxonomyDefs[id]
	if !ok {
		return ubom.TaxonomyDef{}, ErrNotFound
	}
	return def, nil
}

func (s *MemoryStore) CreatePartNumber(part ubom.PartNumber) (ubom.PartNumber, error) {
	if err := part.Validate(); err != nil {
		return ubom.PartNumber{}, err
	}
	if _, err := s.GetSeqDef(part.SeqDefID); err != nil {
		return ubom.PartNumber{}, err
	}
	taxonomyDef, err := s.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if taxonomyDef.SeqDef != part.SeqDefID {
		return ubom.PartNumber{}, fmt.Errorf("taxonomy definition does not belong to sequence definition")
	}
	if !taxonomyNodeExists(taxonomyDef.Taxonomy.Root, part.TaxonomyNodeID) {
		return ubom.PartNumber{}, ErrNotFound
	}
	if _, ok := s.partNumbers[part.Value]; ok {
		return ubom.PartNumber{}, ErrAlreadyExists
	}
	part.ID = ubom.PartNumberID(strconv.FormatInt(s.nextPartID, 10))
	s.nextPartID++
	s.partNumbers[part.Value] = part
	return part, nil
}

func (s *MemoryStore) GetPartNumber(value string) (ubom.PartNumber, error) {
	part, ok := s.partNumbers[value]
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return part, nil
}

func (s *MemoryStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	part, ok := s.partNumbersByID(id)
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return part, nil
}

func (s *MemoryStore) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	if err := revision.Validate(); err != nil {
		return ubom.PartRevision{}, err
	}
	part, ok := s.partNumbersByID(revision.PartNumberID)
	if !ok {
		return ubom.PartRevision{}, ErrNotFound
	}
	for _, item := range revision.BOM {
		if _, ok := s.partNumbersByID(item.PartNumberID); !ok {
			return ubom.PartRevision{}, ErrNotFound
		}
		if _, ok := s.partRevisions[item.PartRevisionID]; !ok {
			return ubom.PartRevision{}, ErrNotFound
		}
	}

	revision.ID = ubom.PartRevisionID(strconv.FormatInt(s.nextRevisionID, 10))
	s.nextRevisionID++
	s.partRevisions[revision.ID] = revision
	part.PartRevisionID = append(part.PartRevisionID, revision.ID)
	s.partNumbers[part.Value] = part
	return revision, nil
}

func (s *MemoryStore) GetPartRevision(id ubom.PartRevisionID) (ubom.PartRevision, error) {
	revision, ok := s.partRevisions[id]
	if !ok {
		return ubom.PartRevision{}, ErrNotFound
	}
	return revision, nil
}

func (s *MemoryStore) partNumbersByID(id ubom.PartNumberID) (ubom.PartNumber, bool) {
	for _, part := range s.partNumbers {
		if part.ID == id {
			return part, true
		}
	}
	return ubom.PartNumber{}, false
}

func taxonomyNodeExists(node ubom.TaxonomyNode, id ubom.TaxonomyNodeID) bool {
	if node.ID == id {
		return true
	}
	for _, child := range node.Children {
		if taxonomyNodeExists(child, id) {
			return true
		}
	}
	return false
}
