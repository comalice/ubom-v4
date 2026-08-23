package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	ubom "ubom-v4"
	"ubom-v4/store"
)

type sequenceNodeRequest struct {
	Kind      string                `json:"kind"`
	Text      string                `json:"text,omitempty"`
	Name      string                `json:"name,omitempty"`
	Min       int                   `json:"min,omitempty"`
	Max       int                   `json:"max,omitempty"`
	Width     int                   `json:"width,omitempty"`
	Pad       byte                  `json:"pad,omitempty"`
	Alphabets []string              `json:"alphabets,omitempty"`
	Children  []sequenceNodeRequest `json:"children,omitempty"`
	Child     *sequenceNodeRequest  `json:"child,omitempty"`
}

type sequenceDefinitionRequest struct {
	ID   ubom.SeqDefID       `json:"id"`
	Root sequenceNodeRequest `json:"root"`
}

func (n sequenceNodeRequest) node() (ubom.SeqDefNode, error) {
	children := make([]ubom.SeqDefNode, 0, len(n.Children))
	for _, child := range n.Children {
		value, err := child.node()
		if err != nil {
			return nil, err
		}
		children = append(children, value)
	}
	switch n.Kind {
	case "literal":
		return ubom.LiteralNode{Text: n.Text}, nil
	case "concat":
		return ubom.ConcatNode{Children: children}, nil
	case "choice":
		return ubom.ChoiceNode{Children: children}, nil
	case "branch":
		return ubom.BranchNode{Children: children}, nil
	case "range_radix":
		return ubom.RangeRadixNode{Children: children}, nil
	case "bind":
		if n.Child == nil {
			return nil, errors.New("bind node has no child")
		}
		child, err := n.Child.node()
		if err != nil {
			return nil, err
		}
		return ubom.BindNode{Name: n.Name, Child: child}, nil
	case "places":
		return ubom.PlaceSequenceNode{Alphabets: n.Alphabets}, nil
	case "range":
		return ubom.RangeNode{Min: n.Min, Max: n.Max, WidthValue: n.Width, Pad: n.Pad}, nil
	default:
		return nil, errors.New("unknown sequence node kind")
	}
}

type taxonomyNodeRequest struct {
	ID         ubom.TaxonomyNodeID        `json:"id"`
	Label      string                     `json:"label"`
	Matches    map[string]string          `json:"matches,omitempty"`
	Attributes []ubom.AttributeAssignment `json:"attributes,omitempty"`
	Children   []taxonomyNodeRequest      `json:"children,omitempty"`
}

func (n taxonomyNodeRequest) node() ubom.TaxonomyNode {
	children := make([]ubom.TaxonomyNode, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, child.node())
	}
	return ubom.TaxonomyNode{ID: n.ID, Label: n.Label, Matches: n.Matches, Attributes: n.Attributes, Children: children}
}

type taxonomyDefinitionRequest struct {
	ID            ubom.TaxonomyDefID  `json:"id"`
	SeqDef        ubom.SeqDefID       `json:"seqDef"`
	AttributeDefs []ubom.AttributeDef `json:"attributeDefs,omitempty"`
	Root          taxonomyNodeRequest `json:"root"`
}

type partNumberRequest struct {
	Value       string                     `json:"value"`
	SeqDef      ubom.SeqDefID              `json:"seqDef"`
	TaxonomyDef ubom.TaxonomyDefID         `json:"taxonomyDef"`
	Attributes  []ubom.PartNumberAttribute `json:"attributes,omitempty"`
}

type partRevisionRequest struct {
	PartNumberID   ubom.PartNumberID `json:"partNumberId"`
	Revision       string            `json:"revision"`
	RevisionSeqDef ubom.SeqDefID     `json:"revisionSeqDef"`
	BOM            []ubom.LineItem   `json:"bom,omitempty"`
}

func decodeBody(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request")
		return false
	}
	return true
}

func (s *Server) handleSeqDefs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		views, err := s.service.ListSeqDefViews()
		if err != nil {
			writeError(w, 500, "could not load sequence definitions")
			return
		}
		writeJSON(w, views)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request sequenceDefinitionRequest
	if !decodeBody(w, r, &request) {
		return
	}
	root, err := request.Root.node()
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := s.service.CreateSeqDef(ubom.NewSeqDef(root).WithID(request.ID)); err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": request.ID})
}

func (s *Server) handleTaxonomyDefs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		views, err := s.service.ListTaxonomyDefSummaries()
		if err != nil {
			writeError(w, 500, "could not load taxonomy definitions")
			return
		}
		writeJSON(w, views)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request taxonomyDefinitionRequest
	if !decodeBody(w, r, &request) {
		return
	}
	definition := ubom.TaxonomyDef{ID: request.ID, SeqDef: request.SeqDef, AttributeDefs: request.AttributeDefs, Taxonomy: ubom.Taxonomy{Root: request.Root.node()}}
	if err := s.service.CreateTaxonomyDef(definition); err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": request.ID})
}

func (s *Server) handleCreatePartNumber(w http.ResponseWriter, r *http.Request) {
	var request partNumberRequest
	if !decodeBody(w, r, &request) {
		return
	}
	seqDef, err := s.service.GetSeqDefForAuthoring(request.SeqDef)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	taxonomy, err := s.service.GetTaxonomyDefForAuthoring(request.TaxonomyDef)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	part, err := ubom.NewPartNumber(request.Value, seqDef, taxonomy)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	part.Attributes = request.Attributes
	created, err := s.service.CreatePartNumber(part)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": created.ID, "value": created.Value})
}

func (s *Server) handleCreateRevision(w http.ResponseWriter, r *http.Request) {
	var request partRevisionRequest
	if !decodeBody(w, r, &request) {
		return
	}
	created, err := s.service.CreatePartRevision(ubom.PartRevision{PartNumberID: request.PartNumberID, Revision: request.Revision, RevisionSeqDefID: request.RevisionSeqDef, BOM: request.BOM})
	if err != nil {
		writeCreateError(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{"id": created.ID, "revision": created.Revision})
}

func writeCreateError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, store.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, store.ErrAlreadyExists) {
		status = http.StatusConflict
	}
	writeError(w, status, err.Error())
}
