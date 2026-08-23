package store

import (
	"errors"
	"sort"
	"strconv"

	ubom "ubom-v4"
	"ubom-v4/artifact"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type MemoryStore struct {
	sources        map[artifact.SourceID]artifact.Source
	artifacts      map[artifact.ID]artifact.Artifact
	snapshots      map[artifact.SnapshotID]artifact.Snapshot
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
		sources:        map[artifact.SourceID]artifact.Source{},
		artifacts:      map[artifact.ID]artifact.Artifact{},
		snapshots:      map[artifact.SnapshotID]artifact.Snapshot{},
		seqDefs:        map[ubom.SeqDefID]ubom.SeqDef{},
		taxonomyDefs:   map[ubom.TaxonomyDefID]ubom.TaxonomyDef{},
		partNumbers:    map[string]ubom.PartNumber{},
		partRevisions:  map[ubom.PartRevisionID]ubom.PartRevision{},
		nextPartID:     1,
		nextRevisionID: 1,
	}
}

func (s *MemoryStore) CreateSource(source artifact.Source) error {
	if err := source.Validate(); err != nil {
		return err
	}
	if _, ok := s.sources[source.ID]; ok {
		return ErrAlreadyExists
	}
	s.sources[source.ID] = source
	return nil
}

func (s *MemoryStore) GetSource(id artifact.SourceID) (artifact.Source, error) {
	source, ok := s.sources[id]
	if !ok {
		return artifact.Source{}, ErrNotFound
	}
	return source, nil
}

func (s *MemoryStore) ListSources() ([]artifact.Source, error) {
	items := make([]artifact.Source, 0, len(s.sources))
	for _, source := range s.sources {
		items = append(items, source)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *MemoryStore) CreateArtifact(item artifact.Artifact) error {
	if err := item.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSource(item.SourceID); err != nil {
		return err
	}
	if _, ok := s.artifacts[item.ID]; ok {
		return ErrAlreadyExists
	}
	s.artifacts[item.ID] = item
	return nil
}

func (s *MemoryStore) GetArtifact(id artifact.ID) (artifact.Artifact, error) {
	item, ok := s.artifacts[id]
	if !ok {
		return artifact.Artifact{}, ErrNotFound
	}
	return item, nil
}

func (s *MemoryStore) ListArtifacts() ([]artifact.Artifact, error) {
	items := make([]artifact.Artifact, 0, len(s.artifacts))
	for _, item := range s.artifacts {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *MemoryStore) CreateSnapshot(snapshot artifact.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if _, err := s.GetArtifact(snapshot.ArtifactID); err != nil {
		return err
	}
	if _, ok := s.snapshots[snapshot.ID]; ok {
		return ErrAlreadyExists
	}
	s.snapshots[snapshot.ID] = snapshot
	return nil
}

func (s *MemoryStore) GetSnapshot(id artifact.SnapshotID) (artifact.Snapshot, error) {
	snapshot, ok := s.snapshots[id]
	if !ok {
		return artifact.Snapshot{}, ErrNotFound
	}
	return snapshot, nil
}

func (s *MemoryStore) ListSnapshotsByArtifact(id artifact.ID) ([]artifact.Snapshot, error) {
	if _, err := s.GetArtifact(id); err != nil {
		return nil, err
	}
	items := []artifact.Snapshot{}
	for _, snapshot := range s.snapshots {
		if snapshot.ArtifactID == id {
			items = append(items, snapshot)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
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
	if err := validatePartNumberReferences(s, part); err != nil {
		return ubom.PartNumber{}, err
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

func (s *MemoryStore) ListPartNumbers() ([]ubom.PartNumber, error) {
	parts := make([]ubom.PartNumber, 0, len(s.partNumbers))
	for _, part := range s.partNumbers {
		parts = append(parts, part)
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Value < parts[j].Value })
	return parts, nil
}

func (s *MemoryStore) ListPartNumbersByTaxonomyNode(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	taxonomy, err := s.GetTaxonomyDef(taxonomyID)
	if err != nil {
		return nil, err
	}
	if !taxonomyNodeExists(taxonomy.Taxonomy.Root, nodeID) {
		return nil, ErrNotFound
	}
	parts := []ubom.PartNumber{}
	for _, part := range s.partNumbers {
		if part.TaxonomyDefID == taxonomyID && part.TaxonomyNodeID == nodeID {
			parts = append(parts, part)
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Value < parts[j].Value })
	return parts, nil
}

func (s *MemoryStore) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	prospectiveID := ubom.PartRevisionID(strconv.FormatInt(s.nextRevisionID, 10))
	if err := revision.Validate(); err != nil {
		return ubom.PartRevision{}, err
	}
	if err := validatePartRevisionCycles(s, revision, prospectiveID); err != nil {
		return ubom.PartRevision{}, err
	}
	if err := validatePartRevisionReferences(s, revision); err != nil {
		return ubom.PartRevision{}, err
	}
	part, _ := s.partNumbersByID(revision.PartNumberID)
	for _, existingID := range part.PartRevisionID {
		existing, _ := s.GetPartRevision(existingID)
		if existing.Revision == revision.Revision {
			return ubom.PartRevision{}, ErrAlreadyExists
		}
	}

	revision.ID = prospectiveID
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
