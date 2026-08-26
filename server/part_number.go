package ubom

import "fmt"

// Part intentionally left bare for now. Leaf node in a BOM tree.
// TODO sort out _what_ goes into Part. Eventually the concept of Artifact, ArtifactSource,
// mfr mpn, etc?
type Part struct {
}

// PartNumberID identifies a durable part number record.
type PartNumberID string

// PartRevisionID identifies a revision owned by a part number.
type PartRevisionID string

// PartNumberDraftID identifies an in-progress part-number draft.
type PartNumberDraftID string

type DraftStatus string

const (
	DraftStatusDraft     DraftStatus = "draft"
	DraftStatusReleased  DraftStatus = "released"
	DraftStatusAbandoned DraftStatus = "abandoned"
)

type PartNumberDraft struct {
	ID                   PartNumberDraftID
	SeqDefID             SeqDefID
	TaxonomyDefID        TaxonomyDefID
	ContextNodeID        TaxonomyNodeID
	Bindings             map[string]string
	Attributes           []PartNumberAttribute
	Status               DraftStatus
	ReleasedPartNumberID PartNumberID
}

func (d PartNumberDraft) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("part number draft has no ID")
	}
	if d.SeqDefID == "" {
		return fmt.Errorf("part number draft has no sequence definition ID")
	}
	if d.TaxonomyDefID == "" {
		return fmt.Errorf("part number draft has no taxonomy definition ID")
	}
	if d.ContextNodeID == "" {
		return fmt.Errorf("part number draft has no context node ID")
	}
	if d.Status != DraftStatusDraft && d.Status != DraftStatusReleased && d.Status != DraftStatusAbandoned {
		return fmt.Errorf("invalid part number draft status %q", d.Status)
	}
	return nil
}

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
	Attributes     []PartNumberAttribute
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

// NewPartNumber is the legacy adapter for Schema.NewPartNumber. Durable part
// numbers require a taxonomy classification node, even though callers of this
// adapter do not provide an explicit open/closed policy.
func NewPartNumber(value string, seqDef SeqDef, taxonomy TaxonomyDef) (PartNumber, error) {
	// The historical API has no attribute parameter and callers commonly fill
	// attributes immediately afterward. Keep that construction behavior while
	// sharing all schema parsing and classification logic.
	return NewSchema(seqDef, taxonomy, SchemaPolicy{}).materializePartNumber(value, nil, false)
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
