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
	ID       ubom.PartRevisionID `json:"id"`
	Revision string              `json:"revision"`
	BOM      []LineItemView      `json:"bom"`
}

type LineItemView struct {
	PartNumberID   ubom.PartNumberID   `json:"partNumberId"`
	PartRevisionID ubom.PartRevisionID `json:"partRevisionId"`
}

type TaxonomyNodeView struct {
	ID          ubom.TaxonomyNodeID `json:"id"`
	Label       string              `json:"label"`
	Path        []TaxonomyPathItem  `json:"path"`
	Children    []TaxonomyChildView `json:"children"`
	PartNumbers []PartNumberSummary `json:"partNumbers"`
}

type TaxonomyPathItem struct {
	ID    ubom.TaxonomyNodeID `json:"id"`
	Label string              `json:"label"`
}

type TaxonomyChildView struct {
	ID    ubom.TaxonomyNodeID `json:"id"`
	Label string              `json:"label"`
}

type PartNumberSummary struct {
	ID    ubom.PartNumberID `json:"id"`
	Value string            `json:"value"`
}

type RevisionDetailView struct {
	ID           ubom.PartRevisionID `json:"id"`
	Revision     string              `json:"revision"`
	PartNumber   PartNumberSummary   `json:"partNumber"`
	TaxonomyPath []string            `json:"taxonomyPath"`
	BOM          []BOMNodeView       `json:"bom"`
}

type BOMNodeView struct {
	PartNumber PartNumberSummary   `json:"partNumber"`
	RevisionID ubom.PartRevisionID `json:"revisionId"`
	Revision   string              `json:"revision"`
	BOM        []BOMNodeView       `json:"bom"`
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
			ID:       revision.ID,
			Revision: revision.Revision,
			BOM:      make([]LineItemView, 0, len(revision.BOM)),
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

func (s *Service) GetPartNumberViewByValue(value string) (PartNumberView, error) {
	part, err := s.store.GetPartNumber(value)
	if err != nil {
		return PartNumberView{}, err
	}
	return s.GetPartNumberView(part.ID)
}

func (s *Service) GetTaxonomyNodeView(taxonomyID ubom.TaxonomyDefID, nodeID ubom.TaxonomyNodeID) (TaxonomyNodeView, error) {
	taxonomyDef, err := s.store.GetTaxonomyDef(taxonomyID)
	if err != nil {
		return TaxonomyNodeView{}, err
	}
	node, ok := findTaxonomyNode(taxonomyDef.Taxonomy.Root, nodeID)
	if !ok {
		return TaxonomyNodeView{}, store.ErrNotFound
	}
	path, ok := findTaxonomyPath(taxonomyDef.Taxonomy.Root, nodeID)
	if !ok {
		return TaxonomyNodeView{}, store.ErrNotFound
	}
	parts, err := s.store.ListPartNumbersByTaxonomyNode(taxonomyID, nodeID)
	if err != nil {
		return TaxonomyNodeView{}, err
	}

	view := TaxonomyNodeView{
		ID:          node.ID,
		Label:       node.Label,
		Path:        path,
		Children:    make([]TaxonomyChildView, 0, len(node.Children)),
		PartNumbers: make([]PartNumberSummary, 0, len(parts)),
	}
	for _, child := range node.Children {
		view.Children = append(view.Children, TaxonomyChildView{ID: child.ID, Label: child.Label})
	}
	for _, part := range parts {
		view.PartNumbers = append(view.PartNumbers, PartNumberSummary{ID: part.ID, Value: part.Value})
	}
	return view, nil
}

func (s *Service) GetRevisionView(id ubom.PartRevisionID) (RevisionDetailView, error) {
	revision, err := s.store.GetPartRevision(id)
	if err != nil {
		return RevisionDetailView{}, err
	}
	return s.buildRevisionView(revision, map[ubom.PartRevisionID]bool{})
}

func (s *Service) buildRevisionView(revision ubom.PartRevision, active map[ubom.PartRevisionID]bool) (RevisionDetailView, error) {
	if active[revision.ID] {
		return RevisionDetailView{}, errors.New("cycle detected in BOM")
	}
	active[revision.ID] = true
	defer delete(active, revision.ID)

	part, err := s.store.GetPartNumberByID(revision.PartNumberID)
	if err != nil {
		return RevisionDetailView{}, err
	}
	taxonomyPath, err := s.partTaxonomyPath(part)
	if err != nil {
		return RevisionDetailView{}, err
	}
	view := RevisionDetailView{
		ID:       revision.ID,
		Revision: revision.Revision,
		PartNumber: PartNumberSummary{
			ID:    part.ID,
			Value: part.Value,
		},
		TaxonomyPath: taxonomyPath,
		BOM:          make([]BOMNodeView, 0, len(revision.BOM)),
	}
	for _, item := range revision.BOM {
		childRevision, err := s.store.GetPartRevision(item.PartRevisionID)
		if err != nil {
			return RevisionDetailView{}, err
		}
		if childRevision.PartNumberID != item.PartNumberID {
			return RevisionDetailView{}, errors.New("line item revision does not belong to line item part number")
		}
		childView, err := s.buildRevisionView(childRevision, active)
		if err != nil {
			return RevisionDetailView{}, err
		}
		view.BOM = append(view.BOM, BOMNodeView{
			PartNumber: childView.PartNumber,
			RevisionID: childView.ID,
			Revision:   childView.Revision,
			BOM:        childView.BOM,
		})
	}
	return view, nil
}

func (s *Service) partTaxonomyPath(part ubom.PartNumber) ([]string, error) {
	seqDef, err := s.store.GetSeqDef(part.SeqDefID)
	if err != nil {
		return nil, err
	}
	taxonomyDef, err := s.store.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return nil, err
	}
	parsed, err := seqDef.ParseValues(part.Value)
	if err != nil {
		return nil, err
	}
	return taxonomyDef.Taxonomy.Project(parsed), nil
}

func findTaxonomyPath(node ubom.TaxonomyNode, id ubom.TaxonomyNodeID) ([]TaxonomyPathItem, bool) {
	path := []TaxonomyPathItem{}
	if node.Label != "" {
		path = append(path, TaxonomyPathItem{ID: node.ID, Label: node.Label})
	}
	if node.ID == id {
		return path, true
	}
	for _, child := range node.Children {
		childPath, ok := findTaxonomyPath(child, id)
		if ok {
			return append(path, childPath...), true
		}
	}
	return nil, false
}

func findTaxonomyNode(node ubom.TaxonomyNode, id ubom.TaxonomyNodeID) (ubom.TaxonomyNode, bool) {
	if node.ID == id {
		return node, true
	}
	for _, child := range node.Children {
		if found, ok := findTaxonomyNode(child, id); ok {
			return found, true
		}
	}
	return ubom.TaxonomyNode{}, false
}
