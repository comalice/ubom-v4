package app

import (
	"errors"

	ubom "ubom-v4"
	"ubom-v4/store"
)

type Service struct {
	store store.Store
}

func NewService(s store.Store) *Service {
	return &Service{store: s}
}

type PartNumberView struct {
	ID           ubom.PartNumberID `json:"id"`
	Value        string            `json:"value"`
	TaxonomyPath []string          `json:"taxonomyPath"`
	Revisions    []RevisionView    `json:"revisions"`
}

type RevisionView struct {
	ID  ubom.PartRevisionID `json:"id"`
	BOM []LineItemView      `json:"bom"`
}

type LineItemView struct {
	PartNumberID   ubom.PartNumberID   `json:"partNumberId"`
	PartRevisionID ubom.PartRevisionID `json:"partRevisionId"`
}

func (s *Service) GetPartNumberView(id ubom.PartNumberID) (PartNumberView, error) {
	part, err := s.store.GetPartNumberByID(id)
	if err != nil {
		return PartNumberView{}, err
	}
	seqDef, err := s.store.GetSeqDef(part.SeqDefID)
	if err != nil {
		return PartNumberView{}, err
	}
	taxonomyDef, err := s.store.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return PartNumberView{}, err
	}
	parsed, err := seqDef.ParseValues(part.Value)
	if err != nil {
		return PartNumberView{}, err
	}

	view := PartNumberView{
		ID:           part.ID,
		Value:        part.Value,
		TaxonomyPath: taxonomyDef.Taxonomy.Project(parsed),
		Revisions:    make([]RevisionView, 0, len(part.PartRevisionID)),
	}
	for _, revisionID := range part.PartRevisionID {
		revision, err := s.store.GetPartRevision(revisionID)
		if err != nil {
			return PartNumberView{}, err
		}
		if revision.PartNumberID != part.ID {
			return PartNumberView{}, errors.New("part revision does not belong to part number")
		}
		revisionView := RevisionView{
			ID:  revision.ID,
			BOM: make([]LineItemView, 0, len(revision.BOM)),
		}
		for _, item := range revision.BOM {
			revisionView.BOM = append(revisionView.BOM, LineItemView{
				PartNumberID:   item.PartNumberID,
				PartRevisionID: item.PartRevisionID,
			})
		}
		view.Revisions = append(view.Revisions, revisionView)
	}
	return view, nil
}
