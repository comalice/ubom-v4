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
	GetSeqDef(ubom.SeqDefID) (ubom.SeqDef, error)
	ListSeqDefs() ([]ubom.SeqDef, error)

	CreateTaxonomyDef(ubom.TaxonomyDef) error
	GetTaxonomyDef(ubom.TaxonomyDefID) (ubom.TaxonomyDef, error)
	ListTaxonomyDefs() ([]ubom.TaxonomyDef, error)

	CreatePartNumber(ubom.PartNumber) (ubom.PartNumber, error)
	GetPartNumber(string) (ubom.PartNumber, error)
	GetPartNumberByID(ubom.PartNumberID) (ubom.PartNumber, error)
	ListPartNumbers() ([]ubom.PartNumber, error)
	ListPartNumbersByTaxonomyNode(ubom.TaxonomyDefID, ubom.TaxonomyNodeID) ([]ubom.PartNumber, error)

	CreatePartRevision(ubom.PartRevision) (ubom.PartRevision, error)
	GetPartRevision(ubom.PartRevisionID) (ubom.PartRevision, error)
}
