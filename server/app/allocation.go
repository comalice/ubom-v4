package app

import (
	"errors"

	ubom "ubom-v4"
	"ubom-v4/store"
)

// Allocation is an unreserved candidate. Allocation never changes the draft
// or reserves the value; the store remains authoritative when it is released.
type Allocation struct {
	Value    string
	Bindings map[string]string
}

// AllocatePartNumberDraft finds the first available value at or below the
// draft's taxonomy context. Every successor is generated and classified
// again, rather than assuming that successor preserves the initial meaning.
func (s *Service) AllocatePartNumberDraft(id ubom.PartNumberDraftID) (Allocation, error) {
	draft, err := s.store.GetPartNumberDraft(id)
	if err != nil {
		return Allocation{}, err
	}
	if draft.Status != ubom.DraftStatusDraft {
		return Allocation{}, ubom.NewDomainError(ubom.ErrorCategoryAllocation, "only draft-status drafts can be allocated")
	}
	seqDef, err := s.store.GetSeqDef(draft.SeqDefID)
	if err != nil {
		return Allocation{}, err
	}
	taxonomy, err := s.store.GetTaxonomyDef(draft.TaxonomyDefID)
	if err != nil {
		return Allocation{}, err
	}
	if taxonomy.SeqDef != seqDef.ID || !nodeExists(taxonomy.Taxonomy.Root, draft.ContextNodeID) {
		return Allocation{}, ubom.NewDomainError(ubom.ErrorCategoryDefinition, "draft references an incompatible taxonomy context")
	}
	schema := ubom.NewSchema(seqDef, taxonomy, ubom.SchemaPolicy{})
	if err := schema.Validate(); err != nil {
		return Allocation{}, err
	}
	bindings := cloneStringMap(draft.Bindings)
	for {
		interpreted, err := schema.Generate(bindings)
		if err != nil {
			return Allocation{}, err
		}
		classification := interpreted.Classification
		if classification == nil || !atOrBelow(taxonomy.Taxonomy.Root, draft.ContextNodeID, classification.NodeID) {
			// An ordered sequence may enter the requested subtree later.
			var next ubom.InterpretedValue
			next, err = schema.Successor(interpreted.Parsed.Value)
			if err != nil {
				return Allocation{}, allocationAdvanceError(err)
			}
			bindings = cloneStringMap(next.Parsed.Bindings)
			continue
		}
		part, err := schema.NewPartNumber(interpreted.Parsed.Value, draft.Attributes)
		if err != nil {
			return Allocation{}, err
		}
		_, lookupErr := s.store.GetPartNumber(part.Value)
		if errors.Is(lookupErr, store.ErrNotFound) {
			return Allocation{Value: part.Value, Bindings: bindings}, nil
		}
		if lookupErr != nil {
			return Allocation{}, lookupErr
		}
		next, err := schema.Successor(interpreted.Parsed.Value)
		if err != nil {
			return Allocation{}, allocationAdvanceError(err)
		}
		bindings = cloneStringMap(next.Parsed.Bindings)
	}
}

func allocationAdvanceError(err error) error {
	if errors.Is(err, ubom.ErrOrderingExhausted) {
		return ubom.WrapDomainError(ubom.ErrorCategoryAllocation, "no unused part number is available", err)
	}
	if errors.Is(err, ubom.ErrNoOrdering) {
		return err
	}
	return ubom.WrapDomainError(ubom.ErrorCategoryAllocation, "no unused part number is available", err)
}

func cloneStringMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func nodeExists(root ubom.TaxonomyNode, id ubom.TaxonomyNodeID) bool {
	if root.ID == id {
		return true
	}
	for _, child := range root.Children {
		if nodeExists(child, id) {
			return true
		}
	}
	return false
}

func atOrBelow(root ubom.TaxonomyNode, ancestor, target ubom.TaxonomyNodeID) bool {
	if root.ID == ancestor {
		return nodeExists(root, target)
	}
	for _, child := range root.Children {
		if atOrBelow(child, ancestor, target) {
			return true
		}
	}
	return false
}
