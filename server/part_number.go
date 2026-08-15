package ubom

import "fmt"

// Part intentionally left bare for now. Leaf node in a BOM tree.
// TODO sort out _what_ goes into Part. Eventually the concept of Artifact, ArtifactSource,
// mfr mpn, etc?
type Part struct {
}

type PartNumberID string
type PartRevisionID string

type LineItem struct {
	PartNumberID   PartNumberID
	PartRevisionID PartRevisionID
}

// Validate checks the identity fields required for a BOM line item.
func (l LineItem) Validate() error {
	if l.PartNumberID == "" {
		return fmt.Errorf("line item has no part number ID")
	}
	if l.PartRevisionID == "" {
		return fmt.Errorf("line item has no part revision ID")
	}
	return nil
}

type PartRevision struct {
	ID               PartRevisionID
	PartNumberID     PartNumberID // Link to owning part number.
	Revision         string       // Human-facing revision value.
	RevisionSeqDefID SeqDefID     // Sequence definition used by Revision.
	BOM              []LineItem
	Part             Part
}

// Validate checks the structure of a part revision and its BOM.
// It does not verify that referenced records exist in a store.
func (r PartRevision) Validate() error {
	if r.PartNumberID == "" {
		return fmt.Errorf("part revision has no part number ID")
	}
	if r.Revision == "" {
		return fmt.Errorf("part revision has no revision value")
	}
	if r.RevisionSeqDefID == "" {
		return fmt.Errorf("part revision has no revision sequence definition ID")
	}
	for _, item := range r.BOM {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// PartNumber, a unique value, taxonomy, and schema used to identify a given part. May be a
// partial construction or fully qualified -- where a fully qualified part number possesses values
// for every level of the related taxonomy definition.
type PartNumber struct {
	ID             PartNumberID     // unique part number ID, created by durable store
	Value          string           // SeqDef parseable canonical part number, globally unique.
	PartRevisionID []PartRevisionID // Part revisions.
	SeqDefID       SeqDefID         // SeqDef version used to generate this part number.
	TaxonomyDefID  TaxonomyDefID    // Taxonomy applied to this part number.
	TaxonomyNodeID TaxonomyNodeID   // Taxonomy node containing this part number.
	// TODO, post MVP, sort out attributes and their types.
	// AttributeSchemas []AttributeSchemaID // Attribute schema attached to this part number.
}

// Validate checks the fields required to persist a part number.
func (p PartNumber) Validate() error {
	// No ID validation occurs here, durable stores populate this value.
	if p.Value == "" {
		return fmt.Errorf("part number has no value")
	}
	if p.SeqDefID == "" {
		return fmt.Errorf("part number has no sequence definition ID")
	}
	if p.TaxonomyDefID == "" {
		return fmt.Errorf("part number has no taxonomy definition ID")
	}
	if p.TaxonomyNodeID == "" {
		return fmt.Errorf("part number has no taxonomy node ID")
	}
	return nil
}

// NewPartNumber validates value against seqDef and permanently links the
// resulting part number to both definition IDs.
func NewPartNumber(value string, seqDef SeqDef, taxonomy TaxonomyDef) (PartNumber, error) {
	if seqDef.ID == "" {
		return PartNumber{}, fmt.Errorf("sequence definition has no ID")
	}
	if taxonomy.ID == "" {
		return PartNumber{}, fmt.Errorf("taxonomy definition has no ID")
	}
	if taxonomy.SeqDef != seqDef.ID {
		return PartNumber{}, fmt.Errorf("taxonomy does not belong to sequence definition")
	}
	parsed, err := seqDef.ParseValues(value)
	if err != nil {
		return PartNumber{}, err
	}
	nodeID, err := taxonomy.Taxonomy.ProjectNode(parsed)
	if err != nil {
		return PartNumber{}, err
	}

	return PartNumber{
		Value:          parsed.Value,
		SeqDefID:       seqDef.ID,
		TaxonomyDefID:  taxonomy.ID,
		TaxonomyNodeID: nodeID,
	}, nil
}

// Next generates the next part number from the part number's SeqDef.
// TODO MVP+1 wire in attribute fields -- require an attribute object with (key,value) pairs for
// all defined attribues in the taxonomy tree AND/OR on the part number.

// NextFromPartial generates the next part number from the part number's SeqDef AND relevant SeqDef
// field selections.
// TODO MVP+1 wire in attribute fields.

// Validate consumes a part number struct OR part number string + SeqDef and determines if the part
// number could be produced from the SeqDef.

// ProjectTaxonomy produces a PartNumberTaxonomy object with taxonomy nodes attached to each
// segment of the part number.
