package store

import (
	"errors"
	"strconv"

	ubom "ubom-v4"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type MemoryStore struct {
	seqDefs      map[ubom.SeqDefID]ubom.SeqDef
	taxonomyDefs map[ubom.TaxonomyDefID]ubom.TaxonomyDef
	partNumbers  map[string]ubom.PartNumber
	nextPartID   int64
}

// compile time interface check
var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		seqDefs:      map[ubom.SeqDefID]ubom.SeqDef{},
		taxonomyDefs: map[ubom.TaxonomyDefID]ubom.TaxonomyDef{},
		partNumbers:  map[string]ubom.PartNumber{},
		nextPartID:   1,
	}
}

func (s *MemoryStore) CreateSeqDef(def ubom.SeqDef) error {
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
