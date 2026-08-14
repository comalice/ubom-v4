package app

import (
	ubom "ubom-v4"
	"ubom-v4/store"
)

func LoadSampleData(s store.Store) (ubom.PartNumber, error) {
	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Literal("PN-"),
		ubom.Range(1, 2),
	)).WithID("sample-pn-v1")
	if err := s.CreateSeqDef(seqDef); err != nil {
		return ubom.PartNumber{}, err
	}

	taxonomyDef := ubom.TaxonomyDef{
		ID:     "sample-taxonomy-v1",
		SeqDef: seqDef.ID,
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID:    "components",
			Label: "Components",
		}},
	}
	if err := s.CreateTaxonomyDef(taxonomyDef); err != nil {
		return ubom.PartNumber{}, err
	}

	child, err := ubom.NewPartNumber("PN-2", seqDef, taxonomyDef)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	child, err = s.CreatePartNumber(child)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	childRevision, err := s.CreatePartRevision(ubom.PartRevision{PartNumberID: child.ID})
	if err != nil {
		return ubom.PartNumber{}, err
	}

	parent, err := ubom.NewPartNumber("PN-1", seqDef, taxonomyDef)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	parent, err = s.CreatePartNumber(parent)
	if err != nil {
		return ubom.PartNumber{}, err
	}
	if _, err := s.CreatePartRevision(ubom.PartRevision{
		PartNumberID: parent.ID,
		BOM: []ubom.LineItem{{
			PartNumberID:   child.ID,
			PartRevisionID: childRevision.ID,
		}},
	}); err != nil {
		return ubom.PartNumber{}, err
	}
	return parent, nil
}
