package store

import (
	"errors"
	"sort"
	"strconv"
	"sync"

	ubom "ubom-v4"
	"ubom-v4/artifact"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type MemoryStore struct {
	mu              sync.Mutex
	sources         map[artifact.SourceID]artifact.Source
	artifacts       map[artifact.ID]artifact.Artifact
	snapshots       map[artifact.SnapshotID]artifact.Snapshot
	seqDefs         map[ubom.SeqDefID]ubom.SeqDef
	taxonomyDefs    map[ubom.TaxonomyDefID]ubom.TaxonomyDef
	currentTaxonomy map[ubom.SeqDefID]ubom.TaxonomyDefID
	partNumbers     map[string]ubom.PartNumber
	partRevisions   map[ubom.PartRevisionID]ubom.PartRevision
	partDrafts      map[ubom.PartNumberDraftID]ubom.PartNumberDraft
	nextPartID      int64
	nextRevisionID  int64
}

// compile time interface check
var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sources:         map[artifact.SourceID]artifact.Source{},
		artifacts:       map[artifact.ID]artifact.Artifact{},
		snapshots:       map[artifact.SnapshotID]artifact.Snapshot{},
		seqDefs:         map[ubom.SeqDefID]ubom.SeqDef{},
		taxonomyDefs:    map[ubom.TaxonomyDefID]ubom.TaxonomyDef{},
		currentTaxonomy: map[ubom.SeqDefID]ubom.TaxonomyDefID{},
		partNumbers:     map[string]ubom.PartNumber{},
		partRevisions:   map[ubom.PartRevisionID]ubom.PartRevision{},
		partDrafts:      map[ubom.PartNumberDraftID]ubom.PartNumberDraft{},
		nextPartID:      1,
		nextRevisionID:  1,
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
	def = def.WithLegacyDefaults()
	if err := def.Validate(); err != nil {
		return err
	}
	if _, ok := s.seqDefs[def.ID]; ok {
		return ErrAlreadyExists
	}
	s.seqDefs[def.ID] = def
	return nil
}

func (s *MemoryStore) CreateSeqDefVersion(parent ubom.SeqDefID, def ubom.SeqDef) error {
	original, err := s.GetSeqDef(parent)
	if err != nil {
		return err
	}
	if def.DefinitionID != original.DefinitionID || def.ParentID != parent || def.Version <= original.Version {
		return errors.New("invalid sequence definition version link")
	}
	for _, existing := range s.seqDefs {
		existing = existing.WithLegacyDefaults()
		if existing.DefinitionID == def.DefinitionID && existing.Version == def.Version {
			return ErrAlreadyExists
		}
	}
	return s.CreateSeqDef(def)
}

func (s *MemoryStore) GetSeqDef(id ubom.SeqDefID) (ubom.SeqDef, error) {
	def, ok := s.seqDefs[id]
	if !ok {
		return ubom.SeqDef{}, ErrNotFound
	}
	return def.WithLegacyDefaults(), nil
}

func (s *MemoryStore) ListSeqDefs() ([]ubom.SeqDef, error) {
	items := make([]ubom.SeqDef, 0, len(s.seqDefs))
	for _, def := range s.seqDefs {
		items = append(items, def.WithLegacyDefaults())
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *MemoryStore) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	// Upgrade legacy exact matches before storing so all later interpretation
	// uses typed predicates, while the original Matches remain available.
	taxonomy, err := ubom.UpgradeLegacyTaxonomy(def.Taxonomy)
	if err != nil {
		return err
	}
	def.Taxonomy = taxonomy
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
	if _, exists := s.currentTaxonomy[def.SeqDef]; !exists {
		s.currentTaxonomy[def.SeqDef] = def.ID
	}
	return nil
}

func (s *MemoryStore) CreateTaxonomyRevision(seqDef ubom.SeqDefID, parent ubom.TaxonomyDefID, def ubom.TaxonomyDef) (ubom.TaxonomyDef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.GetCurrentTaxonomyDef(seqDef)
	if parent != "" {
		if err != nil || current.ID != parent {
			return ubom.TaxonomyDef{}, ErrAlreadyExists
		}
	} else if err == nil {
		return ubom.TaxonomyDef{}, ErrAlreadyExists
	}
	if err := s.CreateTaxonomyDef(def); err != nil {
		return ubom.TaxonomyDef{}, err
	}
	s.currentTaxonomy[seqDef] = def.ID
	return cloneTaxonomyDef(def), nil
}

func (s *MemoryStore) GetTaxonomyDef(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	def, ok := s.taxonomyDefs[id]
	if !ok {
		return ubom.TaxonomyDef{}, ErrNotFound
	}
	return cloneTaxonomyDef(def), nil
}

func (s *MemoryStore) GetCurrentTaxonomyDef(seqDef ubom.SeqDefID) (ubom.TaxonomyDef, error) {
	id, ok := s.currentTaxonomy[seqDef]
	if !ok {
		var latest ubom.TaxonomyDef
		for _, candidate := range s.taxonomyDefs {
			if candidate.SeqDef == seqDef && (latest.ID == "" || candidate.WithLegacyDefaults().Version > latest.WithLegacyDefaults().Version) {
				latest = candidate
			}
		}
		if latest.ID == "" {
			return ubom.TaxonomyDef{}, ErrNotFound
		}
		id = latest.ID
		s.currentTaxonomy[seqDef] = id
	}
	return s.GetTaxonomyDef(id)
}

func (s *MemoryStore) ListTaxonomyDefHistory(definitionID string) ([]ubom.TaxonomyDef, error) {
	items := []ubom.TaxonomyDef{}
	for _, def := range s.taxonomyDefs {
		if def.WithLegacyDefaults().DefinitionID == definitionID {
			items = append(items, cloneTaxonomyDef(def))
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].WithLegacyDefaults().Version < items[j].WithLegacyDefaults().Version
	})
	return items, nil
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
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createPartNumber(part)
}

func (s *MemoryStore) createPartNumber(part ubom.PartNumber) (ubom.PartNumber, error) {
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

func (s *MemoryStore) ArchivePartNumber(id ubom.PartNumberID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for value, part := range s.partNumbers {
		if part.ID == id {
			if part.EffectiveStatus() == ubom.PartNumberStatusArchived {
				return ErrAlreadyExists
			}
			part.Status = ubom.PartNumberStatusArchived
			s.partNumbers[value] = part
			return nil
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) GetPartNumberByID(id ubom.PartNumberID) (ubom.PartNumber, error) {
	part, ok := s.partNumbersByID(id)
	if !ok {
		return ubom.PartNumber{}, ErrNotFound
	}
	return clonePartNumber(part), nil
}

func (s *MemoryStore) ListPartNumbers() ([]ubom.PartNumber, error) {
	result, err := s.ListPartNumbersWithRevisionSummaries()
	if err != nil {
		return nil, err
	}
	return result.Parts, nil
}

func (s *MemoryStore) ListPartNumbersWithRevisionSummaries() (TaxonomyNodePartNumbers, error) {
	parts := make([]ubom.PartNumber, 0, len(s.partNumbers))
	revisionSummaries := map[ubom.PartRevisionID]string{}
	for _, part := range s.partNumbers {
		cloned := clonePartNumber(part)
		parts = append(parts, cloned)
		for _, revisionID := range cloned.PartRevisionID {
			if revision, ok := s.partRevisions[revisionID]; ok {
				revisionSummaries[revisionID] = revision.Revision
			}
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Value < parts[j].Value })
	return TaxonomyNodePartNumbers{Parts: parts, RevisionSummaries: revisionSummaries}, nil
}

func (s *MemoryStore) ListPartNumbersByTaxonomyNode(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	taxonomy, err := s.GetTaxonomyDef(taxonomyID)
	if err != nil {
		return nil, err
	}
	if !taxonomyNodeExists(taxonomy.Taxonomy.Root, nodeID) {
		return nil, ErrNotFound
	}
	return s.ListPartNumbersByTaxonomyNodeUnvalidated(taxonomyID, nodeID)
}

func (s *MemoryStore) ListPartNumbersByTaxonomyNodeUnvalidated(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) ([]ubom.PartNumber, error) {
	result, err := s.ListPartNumbersByTaxonomyNodeWithRevisionSummaries(taxonomyID, nodeID)
	if err != nil {
		return nil, err
	}
	return result.Parts, nil
}

func (s *MemoryStore) ListPartNumbersByTaxonomyNodeWithRevisionSummaries(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) (TaxonomyNodePartNumbers, error) {
	parts := []ubom.PartNumber{}
	revisionSummaries := map[ubom.PartRevisionID]string{}
	for _, part := range s.partNumbers {
		if part.TaxonomyDefID == taxonomyID && part.TaxonomyNodeID == nodeID {
			cloned := clonePartNumber(part)
			parts = append(parts, cloned)
			for _, revisionID := range cloned.PartRevisionID {
				if revision, ok := s.partRevisions[revisionID]; ok {
					revisionSummaries[revisionID] = revision.Revision
				}
			}
		}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].Value < parts[j].Value })
	return TaxonomyNodePartNumbers{Parts: parts, RevisionSummaries: revisionSummaries}, nil
}

func (s *MemoryStore) CreatePartNumberDraft(draft ubom.PartNumberDraft) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
	schema := ubom.NewSchema(seqDef, taxonomy, ubom.SchemaPolicy{RequireClassification: true})
	interpreted, err := schema.Generate(draft.Bindings)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	part, err := schema.NewPartNumber(interpreted.Parsed.Value, draft.Attributes)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if !taxonomyNodeIsAtOrBelow(taxonomy.Taxonomy.Root, draft.ContextNodeID, part.TaxonomyNodeID) || !taxonomyNodeIsLeaf(taxonomy.Taxonomy.Root, part.TaxonomyNodeID) {
		return ubom.PartNumber{}, ubom.NewDomainError(ubom.ErrorCategoryClassification, "part number does not match draft taxonomy context")
	}
	created, err := s.createPartNumber(part)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	draft.Status = ubom.DraftStatusReleased
	draft.ReleasedPartNumberID = created.ID
	s.partDrafts[id] = clonePartNumberDraft(draft)
	return created, nil
}

func taxonomyNodeIsLeaf(root ubom.TaxonomyNode, id ubom.TaxonomyNodeID) bool {
	if root.ID == id {
		return len(root.Children) == 0
	}
	for _, child := range root.Children {
		if taxonomyNodeIsLeaf(child, id) {
			return true
		}
	}
	return false
}

func (s *MemoryStore) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
		existing, ok := s.partRevisions[existingID]
		if !ok {
			return ubom.PartRevision{}, ErrNotFound
		}
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
