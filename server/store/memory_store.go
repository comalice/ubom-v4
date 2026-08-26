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
	partDrafts     map[ubom.PartNumberDraftID]ubom.PartNumberDraft
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
		partDrafts:     map[ubom.PartNumberDraftID]ubom.PartNumberDraft{},
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

func (s *MemoryStore) ListSeqDefs() ([]ubom.SeqDef, error) {
	items := make([]ubom.SeqDef, 0, len(s.seqDefs))
	for _, def := range s.seqDefs {
		items = append(items, def)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *MemoryStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	if err := def.Validate(); err != nil {
		return err
	}
	seqDef, err := s.GetSeqDef(def.SeqDef)
	if err != nil {
		return err
	}
	if taxonomyHasTypedPredicates(def.Taxonomy.Root) {
		if err := def.ValidateAgainst(seqDef); err != nil {
			return err
		}
	}
	if _, ok := s.taxonomyDefs[def.ID]; ok {
		return ErrAlreadyExists
	}
	s.taxonomyDefs[def.ID] = cloneTaxonomyDef(def)
	return nil
}

func (s *MemoryStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	def, ok := s.taxonomyDefs[id]
	if !ok {
		return ubom.TaxonomyDef{}, ErrNotFound
	}
	return cloneTaxonomyDef(def), nil
}

func (s *MemoryStore) ListTaxonomyDefs() ([]ubom.TaxonomyDef, error) {
	items := make([]ubom.TaxonomyDef, 0, len(s.taxonomyDefs))
	for _, def := range s.taxonomyDefs {
		items = append(items, cloneTaxonomyDef(def))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
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
	s.partNumbers[part.Value] = clonePartNumber(part)
	return clonePartNumber(part), nil
}

func (s *MemoryStore) GetPartNumber(value string) (ubom.PartNumber, error) {
	part, ok := s.partNumbers[value]
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return clonePartNumber(part), nil
}

func (s *MemoryStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	part, ok := s.partNumbersByID(id)
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return clonePartNumber(part), nil
}

func (s *MemoryStore) ListPartNumbers() ([]ubom.PartNumber, error) {
	parts := make([]ubom.PartNumber, 0, len(s.partNumbers))
	for _, part := range s.partNumbers {
		parts = append(parts, clonePartNumber(part))
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
			parts = append(parts, clonePartNumber(part))
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Value < parts[j].Value })
	return parts, nil
}

func (s *MemoryStore) CreatePartNumberDraft(draft ubom.PartNumberDraft) error {
	if err := draft.Validate(); err != nil {
		return err
	}
	if _, err := s.GetSeqDef(draft.SeqDefID); err != nil {
		return err
	}
	taxonomy, err := s.GetTaxonomyDef(draft.TaxonomyDefID)
	if err != nil {
		return err
	}
	if taxonomy.SeqDef != draft.SeqDefID || !taxonomyNodeExists(taxonomy.Taxonomy.Root, draft.ContextNodeID) {
		return ErrNotFound
	}
	if _, ok := s.partDrafts[draft.ID]; ok {
		return ErrAlreadyExists
	}
	s.partDrafts[draft.ID] = clonePartNumberDraft(draft)
	return nil
}

func (s *MemoryStore) GetPartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumberDraft, error) {
	draft, ok := s.partDrafts[id]
	if !ok {
		return ubom.PartNumberDraft{}, ErrNotFound
	}
	return clonePartNumberDraft(draft), nil
}

func (s *MemoryStore) UpdatePartNumberDraft(draft ubom.PartNumberDraft) error {
	if err := draft.Validate(); err != nil {
		return err
	}
	current, err := s.GetPartNumberDraft(draft.ID)
	if err != nil {
		return err
	}
	if current.Status != ubom.DraftStatusDraft {
		return errors.New("part number draft is not editable")
	}
	s.partDrafts[draft.ID] = clonePartNumberDraft(draft)
	return nil
}

func (s *MemoryStore) ReleasePartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumber, error) {
	draft, err := s.GetPartNumberDraft(id)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if draft.Status != ubom.DraftStatusDraft {
		return ubom.PartNumber{}, errors.New("part number draft is not releasable")
	}
	seqDef, err := s.GetSeqDef(draft.SeqDefID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	taxonomy, err := s.GetTaxonomyDef(draft.TaxonomyDefID)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	value, err := seqDef.Generate(draft.Bindings)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	part, err := ubom.NewPartNumber(value, seqDef, taxonomy)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if !taxonomyNodeIsAtOrBelow(taxonomy.Taxonomy.Root, draft.ContextNodeID, part.TaxonomyNodeID) {
		return ubom.PartNumber{}, errors.New("part number does not match draft taxonomy context")
	}
	part.Attributes = draft.Attributes
	created, err := s.CreatePartNumber(part)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	draft.Status = ubom.DraftStatusReleased
	draft.ReleasedPartNumberID = created.ID
	s.partDrafts[id] = clonePartNumberDraft(draft)
	return created, nil
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
	part = clonePartNumber(part)
	for _, existingID := range part.PartRevisionID {
		existing, _ := s.GetPartRevision(existingID)
		if existing.Revision == revision.Revision {
			return ubom.PartRevision{}, ErrAlreadyExists
		}
	}

	revision.ID = prospectiveID
	s.nextRevisionID++
	s.partRevisions[revision.ID] = clonePartRevision(revision)
	part.PartRevisionID = append(part.PartRevisionID, revision.ID)
	s.partNumbers[part.Value] = part
	return clonePartRevision(revision), nil
}

func (s *MemoryStore) GetPartRevision(id ubom.PartRevisionID) (ubom.PartRevision, error) {
	revision, ok := s.partRevisions[id]
	if !ok {
		return ubom.PartRevision{}, ErrNotFound
	}
	return clonePartRevision(revision), nil
}

func clonePartNumber(part ubom.PartNumber) ubom.PartNumber {
	part.PartRevisionID = append([]ubom.PartRevisionID(nil), part.PartRevisionID...)
	part.Attributes = append([]ubom.PartNumberAttribute(nil), part.Attributes...)
	return part
}

func clonePartNumberDraft(draft ubom.PartNumberDraft) ubom.PartNumberDraft {
	clone := draft
	clone.Bindings = map[string]string{}
	for key, value := range draft.Bindings {
		clone.Bindings[key] = value
	}
	clone.Attributes = append([]ubom.PartNumberAttribute(nil), draft.Attributes...)
	return clone
}

func taxonomyNodeIsAtOrBelow(root ubom.TaxonomyNode, ancestor, target ubom.TaxonomyNodeID) bool {
	if root.ID == ancestor {
		return taxonomyNodeExists(root, target)
	}
	for _, child := range root.Children {
		if taxonomyNodeIsAtOrBelow(child, ancestor, target) {
			return true
		}
	}
	return false
}

func clonePartRevision(revision ubom.PartRevision) ubom.PartRevision {
	revision.BOM = append([]ubom.LineItem(nil), revision.BOM...)
	return revision
}

func cloneTaxonomyDef(def ubom.TaxonomyDef) ubom.TaxonomyDef {
	def.AttributeDefs = append([]ubom.AttributeDef(nil), def.AttributeDefs...)
	for i := range def.AttributeDefs {
		def.AttributeDefs[i].EnumValues = append([]string(nil), def.AttributeDefs[i].EnumValues...)
	}
	def.Taxonomy.Root = cloneTaxonomyNode(def.Taxonomy.Root)
	return def
}

func cloneTaxonomyNode(node ubom.TaxonomyNode) ubom.TaxonomyNode {
	clone := node
	if node.Matches != nil {
		clone.Matches = make(map[string]string, len(node.Matches))
		for key, value := range node.Matches {
			clone.Matches[key] = value
		}
	}
	clone.Predicates = append([]ubom.TaxonomyPredicate(nil), node.Predicates...)
	clone.Attributes = append([]ubom.AttributeAssignment(nil), node.Attributes...)
	if node.Children != nil {
		clone.Children = make([]ubom.TaxonomyNode, len(node.Children))
		for i, child := range node.Children {
			clone.Children[i] = cloneTaxonomyNode(child)
		}
	}
	return clone
}

func (s *MemoryStore) partNumbersByID(id ubom.PartNumberID) (ubom.PartNumber, bool) {
	for _, part := range s.partNumbers {
		if part.ID == id {
			return part, true
		}
	}
	return ubom.PartNumber{}, false
}
