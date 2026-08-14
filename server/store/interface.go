package store

import ubom "ubom-v4"

type Store interface {
	CreateSeqDef(ubom.SeqDef) error
	GetSeqDef(ubom.SeqDefID) (ubom.SeqDef, error)

	CreateTaxonomyDef(ubom.TaxonomyDef) error
	GetTaxonomyDef(ubom.TaxonomyDefID) (ubom.TaxonomyDef, error)

	CreatePartNumber(ubom.PartNumber) (ubom.PartNumber, error)
	GetPartNumber(string) (ubom.PartNumber, error)

	CreatePartRevision(ubom.PartRevision) (ubom.PartRevision, error)
	GetPartRevision(ubom.PartRevisionID) (ubom.PartRevision, error)
}
