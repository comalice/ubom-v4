package store

import (
	"fmt"

	ubom "ubom-v4"
)

func validatePartNumberReferences(s Store, part ubom.PartNumber) error {
	if _, err := s.GetSeqDef(part.SeqDefID); err != nil {
		return err
	}
	taxonomyDef, err := s.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return err
	}
	if taxonomyDef.SeqDef != part.SeqDefID {
		return fmt.Errorf("taxonomy definition does not belong to sequence definition")
	}
	if !taxonomyNodeExists(taxonomyDef.Taxonomy.Root, part.TaxonomyNodeID) {
		return ErrNotFound
	}
	assignments, err := taxonomyDef.Taxonomy.EffectiveAttributes(part.TaxonomyNodeID)
	if err != nil {
		return err
	}
	if err := ubom.ValidatePartNumberAttributes(part.Attributes, assignments, taxonomyDef.AttributeDefs); err != nil {
		return err
	}
	return nil
}

func validatePartRevisionReferences(s Store, revision ubom.PartRevision) error {
	if _, err := s.GetPartNumberByID(revision.PartNumberID); err != nil {
		return err
	}
	revisionSeqDef, err := s.GetSeqDef(revision.RevisionSeqDefID)
	if err != nil {
		return err
	}
	if _, err := revisionSeqDef.Parse(revision.Revision); err != nil {
		return fmt.Errorf("invalid revision value: %w", err)
	}
	for _, item := range revision.BOM {
		if _, err := s.GetPartNumberByID(item.PartNumberID); err != nil {
			return err
		}
		childRevision, err := s.GetPartRevision(item.PartRevisionID)
		if err != nil {
			return err
		}
		if childRevision.PartNumberID != item.PartNumberID {
			return fmt.Errorf("line item revision does not belong to line item part number")
		}
	}
	return nil
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
