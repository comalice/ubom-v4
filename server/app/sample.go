package app

import (
	ubom "ubom-v4"
	"ubom-v4/store"
)

func LoadSampleData(s store.Store) (ubom.PartNumber, error) {
	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Literal("PN-"),
		ubom.Bind("category", ubom.Choice(
			ubom.Literal("A"),
			ubom.Literal("B"),
		)),
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
			Children: []ubom.TaxonomyNode{
				{ID: "resistors", Label: "Resistors", Matches: map[string]string{"category": "A"}},
				{ID: "capacitors", Label: "Capacitors", Matches: map[string]string{"category": "B"}},
			},
		}},
	}
	if err := s.CreateTaxonomyDef(taxonomyDef); err != nil {
		return ubom.PartNumber{}, err
	}

	child, err := ubom.NewPartNumber("PN-B", seqDef, taxonomyDef)
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

	parent, err := ubom.NewPartNumber("PN-A", seqDef, taxonomyDef)
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
