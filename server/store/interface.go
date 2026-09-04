package store

import ubom "ubom-v4"
import "ubom-v4/artifact"

type Store interface {
	CreateSource(artifact.Source) error
	GetSource(artifact.SourceID) (artifact.Source, error)
	ListSources() ([]artifact.Source, error)
	CreateArtifact(artifact.Artifact) error
	GetArtifact(artifact.ID) (artifact.Artifact, error)
	ListArtifacts() ([]artifact.Artifact, error)
	CreateSnapshot(artifact.Snapshot) error
	GetSnapshot(artifact.SnapshotID) (artifact.Snapshot, error)
	ListSnapshotsByArtifact(artifact.ID) ([]artifact.Snapshot, error)

	CreateSeqDef(ubom.SeqDef) error
	CreateSeqDefVersion(parent ubom.SeqDefID, def ubom.SeqDef) error
	GetSeqDef(ubom.SeqDefID) (ubom.SeqDef, error)
	ListSeqDefs() ([]ubom.SeqDef, error)

	CreateTaxonomyDef(ubom.TaxonomyDef) error
	CreateTaxonomyRevision(seqDef ubom.SeqDefID, parent ubom.TaxonomyDefID, def ubom.TaxonomyDef) (ubom.TaxonomyDef, error)
	GetTaxonomyDef(ubom.TaxonomyDefID) (ubom.TaxonomyDef, error)
	ListTaxonomyDefs() ([]ubom.TaxonomyDef, error)
	GetCurrentTaxonomyDef(ubom.SeqDefID) (ubom.TaxonomyDef, error)
	ListTaxonomyDefHistory(string) ([]ubom.TaxonomyDef, error)

	CreatePartNumber(ubom.PartNumber) (ubom.PartNumber, error)
	ArchivePartNumber(ubom.PartNumberID) error
	GetPartNumber(string) (ubom.PartNumber, error)
	GetPartNumberByID(ubom.PartNumberID) (ubom.PartNumber, error)
	ListPartNumbers() ([]ubom.PartNumber, error)
	// ListPartNumbersByTaxonomyNode has exact-node semantics. It does not
	// include parts classified at descendants of nodeID.
	ListPartNumbersByTaxonomyNode(ubom.TaxonomyDefID, ubom.TaxonomyNodeID) ([]ubom.PartNumber, error)
	CreatePartNumberDraft(ubom.PartNumberDraft) error
	GetPartNumberDraft(ubom.PartNumberDraftID) (ubom.PartNumberDraft, error)
	UpdatePartNumberDraft(ubom.PartNumberDraft) error
	ReleasePartNumberDraft(ubom.PartNumberDraftID) (ubom.PartNumber, error)

	CreatePartRevision(ubom.PartRevision) (ubom.PartRevision, error)
	GetPartRevision(ubom.PartRevisionID) (ubom.PartRevision, error)
}

// TaxonomyNodePartNumberLister is the internal fast path for callers that
// have already validated the taxonomy node.
type TaxonomyNodePartNumberLister interface {
	ListPartNumbersByTaxonomyNodeUnvalidated(ubom.TaxonomyDefID, ubom.TaxonomyNodeID) ([]ubom.PartNumber, error)
}

// TaxonomyNodePartNumberRevisionLister is the optimized node path. It returns
// complete part rows and the revision data needed for list rendering.
type TaxonomyNodePartNumberRevisionLister interface {
	ListPartNumbersByTaxonomyNodeWithRevisionSummaries(ubom.TaxonomyDefID, ubom.TaxonomyNodeID) (TaxonomyNodePartNumbers, error)
}

// AllPartNumberRevisionLister is the optimized general list path. It returns
// all part rows together with the revision data needed for list rendering.
type AllPartNumberRevisionLister interface {
	ListPartNumbersWithRevisionSummaries() (TaxonomyNodePartNumbers, error)
}

type TaxonomyNodePartNumbers struct {
	Parts             []ubom.PartNumber
	RevisionSummaries map[ubom.PartRevisionID]string
}
