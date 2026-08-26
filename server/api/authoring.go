package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

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
	Pad       string                `json:"pad,omitempty"`
	Alphabets []string              `json:"alphabets,omitempty"`
	Children  []sequenceNodeRequest `json:"children,omitempty"`
	Child     *sequenceNodeRequest  `json:"child,omitempty"`
}

type sequenceDefinitionRequest struct {
	ID   ubom.SeqDefID       `json:"id"`
	Root sequenceNodeRequest `json:"root"`
}

func (n sequenceNodeRequest) rangePad() (byte, error) {
	if n.Pad == "" {
		return 0, nil
	}
	if len(n.Pad) != 1 {
		return 0, errors.New("range padding must be one character")
	}
	return n.Pad[0], nil
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
		pad, err := n.rangePad()
		if err != nil {
			return nil, err
		}
		return ubom.RangeNode{Min: n.Min, Max: n.Max, WidthValue: n.Width, Pad: pad}, nil
	default:
		return nil, errors.New("unknown sequence node kind")
	}
}

type taxonomyNodeRequest struct {
	ID         ubom.TaxonomyNodeID        `json:"id"`
	Label      string                     `json:"label"`
	Matches    map[string]string          `json:"matches,omitempty"`
	Predicates []ubom.TaxonomyPredicate   `json:"predicates,omitempty"`
	Attributes []ubom.AttributeAssignment `json:"attributes,omitempty"`
	Children   []taxonomyNodeRequest      `json:"children,omitempty"`
}

func (n taxonomyNodeRequest) node() ubom.TaxonomyNode {
	children := make([]ubom.TaxonomyNode, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, child.node())
	}
	return ubom.TaxonomyNode{ID: n.ID, Label: n.Label, Matches: n.Matches, Predicates: n.Predicates, Attributes: n.Attributes, Children: children}
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

type partNumberDraftRequest struct {
	SeqDef        ubom.SeqDefID       `json:"seqDef"`
	TaxonomyDef   ubom.TaxonomyDefID  `json:"taxonomyDef"`
	ContextNodeID ubom.TaxonomyNodeID `json:"contextNodeId"`
}

type partNumberDraftUpdate struct {
	Bindings   map[string]string          `json:"bindings"`
	Attributes []ubom.PartNumberAttribute `json:"attributes,omitempty"`
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
	writeJSONStatus(w, http.StatusCreated, map[string]any{"id": request.ID})
}

func (s *Server) handleUnsavedSeqDefPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Root     sequenceNodeRequest `json:"root"`
		Bindings map[string]string   `json:"bindings"`
	}
	if !decodeBody(w, r, &request) {
		return
	}
	root, err := request.Root.node()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	preview, err := s.service.PreviewSeqDefRoot(root, request.Bindings)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, preview)
}

func (s *Server) handleSeqDefCandidatePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Root  sequenceNodeRequest `json:"root"`
		Value string              `json:"value"`
	}
	if !decodeBody(w, r, &request) {
		return
	}
	root, err := request.Root.node()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	preview, err := s.service.PreviewSeqDefCandidate(root, request.Value)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, preview)
}

func (s *Server) handleSeqDefFields(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/seq-defs/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "fields" {
		writeError(w, 400, "invalid sequence definition ID")
		return
	}
	fields, err := s.service.SeqDefFieldViews(ubom.SeqDefID(parts[0]))
	if err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSON(w, fields)
}

func (s *Server) handleSeqDefPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/seq-defs/")
	parts := strings.Split(path, "/")
	if len(parts) == 1 && parts[0] != "" && r.Method == http.MethodGet {
		definition, err := s.service.GetSeqDefDefinitionView(ubom.SeqDefID(parts[0]))
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "sequence definition not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load sequence definition")
			return
		}
		writeJSON(w, definition)
		return
	}
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "invalid sequence definition path")
		return
	}
	if parts[1] == "fields" {
		s.handleSeqDefFields(w, r)
		return
	}
	if parts[1] == "definition" {
		writeError(w, http.StatusBadRequest, "invalid sequence definition path")
		return
	}
	if parts[1] != "preview" || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		Bindings map[string]string `json:"bindings"`
	}
	if !decodeBody(w, r, &request) {
		return
	}
	preview, err := s.service.PreviewSeqDef(ubom.SeqDefID(parts[0]), request.Bindings)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSON(w, preview)
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
	writeJSONStatus(w, http.StatusCreated, map[string]any{"id": request.ID})
}

func (s *Server) handleCreatePartNumber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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
	part, err := ubom.NewSchema(seqDef, taxonomy, ubom.SchemaPolicy{}).NewPartNumber(request.Value, request.Attributes)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	created, err := s.service.CreatePartNumber(part)
	if err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSONStatus(w, http.StatusCreated, map[string]any{"id": created.ID, "value": created.Value})
}

func (s *Server) handleCreateRevision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request partRevisionRequest
	if !decodeBody(w, r, &request) {
		return
	}
	created, err := s.service.CreatePartRevision(ubom.PartRevision{PartNumberID: request.PartNumberID, Revision: request.Revision, RevisionSeqDefID: request.RevisionSeqDef, BOM: request.BOM})
	if err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSONStatus(w, http.StatusCreated, map[string]any{"id": created.ID, "revision": created.Revision})
}

func (s *Server) handlePartDrafts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request partNumberDraftRequest
	if !decodeBody(w, r, &request) {
		return
	}
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		writeError(w, 500, "could not create draft ID")
		return
	}
	draft := ubom.PartNumberDraft{ID: ubom.PartNumberDraftID("draft-" + hex.EncodeToString(bytes)), SeqDefID: request.SeqDef, TaxonomyDefID: request.TaxonomyDef, ContextNodeID: request.ContextNodeID, Bindings: map[string]string{}, Status: ubom.DraftStatusDraft}
	if err := s.service.CreatePartNumberDraft(draft); err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSONStatus(w, http.StatusCreated, draft)
}

func (s *Server) handlePartDraft(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/part-drafts/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" || len(parts) > 2 {
		writeError(w, 400, "invalid part draft path")
		return
	}
	id := ubom.PartNumberDraftID(parts[0])
	if len(parts) == 2 && parts[1] == "preview" && r.Method == http.MethodPost {
		preview, err := s.service.PreviewPartNumberDraft(id)
		if err != nil {
			writeCreateError(w, err)
			return
		}
		writeJSON(w, preview)
		return
	}
	if len(parts) != 1 {
		writeError(w, 400, "invalid part draft path")
		return
	}
	if r.Method == http.MethodGet {
		draft, err := s.service.GetPartNumberDraft(id)
		if err != nil {
			writeCreateError(w, err)
			return
		}
		writeJSON(w, draft)
		return
	}
	if r.Method == http.MethodPatch {
		var request partNumberDraftUpdate
		if !decodeBody(w, r, &request) {
			return
		}
		draft, err := s.service.GetPartNumberDraft(id)
		if err != nil {
			writeCreateError(w, err)
			return
		}
		draft.Bindings, draft.Attributes = request.Bindings, request.Attributes
		if err := s.service.UpdatePartNumberDraft(draft); err != nil {
			writeCreateError(w, err)
			return
		}
		writeJSON(w, draft)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) handleReleasePartDraft(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/part-drafts/")
	path = strings.TrimSuffix(path, "/release")
	if r.Method != http.MethodPost || path == "" || strings.Contains(path, "/") {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	part, err := s.service.ReleasePartNumberDraft(ubom.PartNumberDraftID(path))
	if err != nil {
		writeCreateError(w, err)
		return
	}
	writeJSONStatus(w, http.StatusCreated, map[string]any{"id": part.ID, "value": part.Value})
}

func writeCreateError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, store.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, store.ErrAlreadyExists) {
		status = http.StatusConflict
	}
	if ubom.IsCategory(err, ubom.ErrorCategoryConflict) || errors.Is(err, ubom.ErrAllocation) {
		status = http.StatusConflict
	}
	if ubom.IsCategory(err, ubom.ErrorCategoryOrdering) {
		status = http.StatusUnprocessableEntity
	}
	writeError(w, status, err.Error())
}
