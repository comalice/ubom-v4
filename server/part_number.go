package ubom

import "fmt"

// Part intentionally left bare for now. Leaf node in a BOM tree.
// TODO sort out _what_ goes into Part. Eventually the concept of Artifact, ArtifactSource,
// mfr mpn, etc?
type Part struct {
}

// PartNumberID identifies a durable part number record.
type PartNumberID string

type PartNumberStatus string

const (
	PartNumberStatusActive   PartNumberStatus = "active"
	PartNumberStatusArchived PartNumberStatus = "archived"
)

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
	ID                   PartNumberDraftID     `json:"id"`
	SeqDefID             SeqDefID              `json:"seqDef"`
	TaxonomyDefID        TaxonomyDefID         `json:"taxonomyDef"`
	ContextNodeID        TaxonomyNodeID        `json:"contextNodeId"`
	Bindings             map[string]string     `json:"bindings"`
	Attributes           []PartNumberAttribute `json:"attributes"`
	Status               DraftStatus           `json:"status"`
	ReleasedPartNumberID PartNumberID          `json:"releasedPartNumberId,omitempty"`
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
	PartNumberID   PartNumberID   `json:"partNumberId"`
	PartRevisionID PartRevisionID `json:"partRevisionId"`
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
	ID               PartRevisionID `json:"id"`
	PartNumberID     PartNumberID   `json:"partNumberId"`   // Link to owning part number.
	Revision         string         `json:"revision"`       // Human-facing revision value.
	RevisionSeqDefID SeqDefID       `json:"revisionSeqDef"` // Sequence definition used by Revision.
	BOM              []LineItem     `json:"bom"`
	Part             Part           `json:"part"`
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
	ID             PartNumberID          `json:"id"`              // unique part number ID, created by durable store
	Value          string                `json:"value"`           // SeqDef parseable canonical part number, globally unique.
	PartRevisionID []PartRevisionID      `json:"partRevisionIds"` // Part revisions.
	SeqDefID       SeqDefID              `json:"seqDef"`          // SeqDef version used to generate this part number.
	TaxonomyDefID  TaxonomyDefID         `json:"taxonomyDef"`     // Taxonomy applied to this part number.
	TaxonomyNodeID TaxonomyNodeID        `json:"taxonomyNodeId"`  // Taxonomy node containing this part number.
	Attributes     []PartNumberAttribute `json:"attributes"`
	Status         PartNumberStatus      `json:"status,omitempty"`
}

func (p PartNumber) EffectiveStatus() PartNumberStatus {
	if p.Status == "" {
		return PartNumberStatusActive
	}
	return p.Status
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
	if p.Status != "" && p.Status != PartNumberStatusActive && p.Status != PartNumberStatusArchived {
		return fmt.Errorf("invalid part number status %q", p.Status)
	}
	return nil
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
