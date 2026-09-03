package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	ubom "ubom-v4"
	"ubom-v4/artifact"
	"ubom-v4/store"
)

type Service struct {
	store store.Store
}

func NewService(s store.Store) *Service {
	return &Service{store: s}
}

type SeqDefView struct {
	ID           ubom.SeqDefID `json:"id"`
	DefinitionID string        `json:"definitionId"`
	Version      int           `json:"version"`
	ParentID     ubom.SeqDefID `json:"parentId,omitempty"`
}

type SeqDefDefinitionView struct {
	ID           ubom.SeqDefID  `json:"id"`
	DefinitionID string         `json:"definitionId"`
	Version      int            `json:"version"`
	ParentID     ubom.SeqDefID  `json:"parentId,omitempty"`
	Root         SeqDefNodeView `json:"root"`
}

type SeqDefNodeView struct {
	Kind      string           `json:"kind"`
	Text      string           `json:"text,omitempty"`
	Name      string           `json:"name,omitempty"`
	Min       int              `json:"min,omitempty"`
	Max       int              `json:"max,omitempty"`
	Width     int              `json:"width,omitempty"`
	Pad       string           `json:"pad,omitempty"`
	Alphabets []string         `json:"alphabets,omitempty"`
	Children  []SeqDefNodeView `json:"children,omitempty"`
	Child     *SeqDefNodeView  `json:"child,omitempty"`
}

type SeqDefFieldView struct {
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
	Options []string `json:"options,omitempty"`
	Min     *int     `json:"min,omitempty"`
	Max     *int     `json:"max,omitempty"`
	Width   int      `json:"width,omitempty"`
	Pad     string   `json:"pad,omitempty"`
}
type TaxonomyDefSummary struct {
	ID           ubom.TaxonomyDefID `json:"id"`
	DefinitionID string             `json:"definitionId"`
	Version      int                `json:"version"`
	ParentID     ubom.TaxonomyDefID `json:"parentId,omitempty"`
	SeqDef       ubom.SeqDefID      `json:"seqDef"`
	Current      bool               `json:"current"`
}

type TaxonomyDefDefinitionView struct {
	ID            ubom.TaxonomyDefID  `json:"id"`
	DefinitionID  string              `json:"definitionId"`
	Version       int                 `json:"version"`
	ParentID      ubom.TaxonomyDefID  `json:"parentId,omitempty"`
	Current       bool                `json:"current"`
	SeqDef        ubom.SeqDefID       `json:"seqDef"`
	AttributeDefs []ubom.AttributeDef `json:"attributeDefs"`
	Root          ubom.TaxonomyNode   `json:"root"`
}

func (s *Service) CreateSeqDef(def ubom.SeqDef) error { return s.store.CreateSeqDef(def) }
func (s *Service) CreateSeqDefVersion(parent ubom.SeqDefID, root ubom.SeqDefNode, requestedVersion int, requestedID ubom.SeqDefID) (SeqDefDefinitionView, error) {
	original, err := s.store.GetSeqDef(parent)
	if err != nil {
		return SeqDefDefinitionView{}, err
	}
	defs, err := s.store.ListSeqDefs()
	if err != nil {
		return SeqDefDefinitionView{}, err
	}
	next := original.Version + 1
	for _, candidate := range defs {
		if candidate.DefinitionID == original.DefinitionID && candidate.Version >= next {
			next = candidate.Version + 1
		}
	}
	if requestedVersion != 0 && requestedVersion != next {
		if requestedVersion < next {
			return SeqDefDefinitionView{}, store.ErrAlreadyExists
		}
		return SeqDefDefinitionView{}, fmt.Errorf("version must be %d", next)
	}
	if requestedID == "" {
		requestedID = ubom.SeqDefID(fmt.Sprintf("%s-v%d", original.DefinitionID, next))
	}
	def := ubom.NewSeqDef(root).WithID(requestedID).WithVersionMetadata(original.DefinitionID, next, parent)
	if err := s.store.CreateSeqDefVersion(parent, def); err != nil {
		return SeqDefDefinitionView{}, err
	}
	return s.GetSeqDefDefinitionView(def.ID)
}
func (s *Service) CreateTaxonomyDef(def ubom.TaxonomyDef) error {
	return s.store.CreateTaxonomyDef(def)
}
func (s *Service) CreateTaxonomyRevision(seqDef ubom.SeqDefID, parent ubom.TaxonomyDefID, definitionID string, attributeDefs []ubom.AttributeDef, root ubom.TaxonomyNode) (TaxonomyDefDefinitionView, error) {
	seq, err := s.store.GetSeqDef(seqDef)
	if err != nil {
		return TaxonomyDefDefinitionView{}, err
	}
	version := 1
	if parent != "" {
		old, err := s.store.GetTaxonomyDef(parent)
		if err != nil {
			return TaxonomyDefDefinitionView{}, err
		}
		old = old.WithLegacyDefaults()
		if old.SeqDef != seqDef {
			return TaxonomyDefDefinitionView{}, ubom.NewDomainError(ubom.ErrorCategoryConflict, "taxonomy parent references a different sequence definition")
		}
		if definitionID == "" {
			definitionID = old.DefinitionID
		}
		if definitionID != old.DefinitionID {
			return TaxonomyDefDefinitionView{}, ubom.NewDomainError(ubom.ErrorCategoryConflict, "taxonomy parent is from a different lineage")
		}
		version = old.Version + 1
	} else if definitionID == "" {
		definitionID = string(seqDef) + "-taxonomy"
	}
	def := ubom.TaxonomyDef{ID: ubom.TaxonomyDefID(fmt.Sprintf("%s-v%d", definitionID, version)), DefinitionID: definitionID, Version: version, ParentID: parent, SeqDef: seqDef, AttributeDefs: attributeDefs, Taxonomy: ubom.Taxonomy{Root: root}}
	if err := def.ValidateAgainst(seq); err != nil {
		return TaxonomyDefDefinitionView{}, err
	}
	created, err := s.store.CreateTaxonomyRevision(seqDef, parent, def)
	if err != nil {
		return TaxonomyDefDefinitionView{}, err
	}
	return s.taxonomyDefinitionView(created, true), nil
}
func (s *Service) CreatePartNumber(part ubom.PartNumber) (ubom.PartNumber, error) {
	return s.store.CreatePartNumber(part)
}
func (s *Service) ArchivePartNumber(id ubom.PartNumberID) error {
	return s.store.ArchivePartNumber(id)
}
func (s *Service) CreatePartRevision(revision ubom.PartRevision) (ubom.PartRevision, error) {
	return s.store.CreatePartRevision(revision)
}
func (s *Service) GetSeqDefForAuthoring(id ubom.SeqDefID) (ubom.SeqDef, error) {
	return s.store.GetSeqDef(id)
}

func (s *Service) GetSeqDefDefinitionView(id ubom.SeqDefID) (SeqDefDefinitionView, error) {
	def, err := s.store.GetSeqDef(id)
	if err != nil {
		return SeqDefDefinitionView{}, err
	}
	root, err := seqDefNodeView(def.Root())
	if err != nil {
		return SeqDefDefinitionView{}, err
	}
	def = def.WithLegacyDefaults()
	return SeqDefDefinitionView{ID: def.ID, DefinitionID: def.DefinitionID, Version: def.Version, ParentID: def.ParentID, Root: root}, nil
}

func seqDefNodeView(node ubom.SeqDefNode) (SeqDefNodeView, error) {
	view := SeqDefNodeView{}
	children := func(nodes []ubom.SeqDefNode) ([]SeqDefNodeView, error) {
		result := make([]SeqDefNodeView, 0, len(nodes))
		for _, child := range nodes {
			converted, err := seqDefNodeView(child)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		return result, nil
	}
	switch n := node.(type) {
	case ubom.LiteralNode:
		view.Kind, view.Text = "literal", n.Text
	case ubom.ConcatNode:
		converted, err := children(n.Children)
		if err != nil {
			return SeqDefNodeView{}, err
		}
		view.Kind, view.Children = "concat", converted
	case ubom.ChoiceNode:
		converted, err := children(n.Children)
		if err != nil {
			return SeqDefNodeView{}, err
		}
		view.Kind, view.Children = "choice", converted
	case ubom.RangeRadixNode:
		converted, err := children(n.Children)
		if err != nil {
			return SeqDefNodeView{}, err
		}
		view.Kind, view.Children = "range_radix", converted
	case ubom.BindNode:
		child, err := seqDefNodeView(n.Child)
		if err != nil {
			return SeqDefNodeView{}, err
		}
		view.Kind, view.Name, view.Child = "bind", n.Name, &child
	case ubom.PlaceSequenceNode:
		view.Kind, view.Alphabets = "places", n.Alphabets
	case ubom.RangeNode:
		view.Kind, view.Min, view.Max, view.Width = "range", n.Min, n.Max, n.WidthValue
		if n.Pad != 0 {
			view.Pad = string(n.Pad)
		}
	default:
		return SeqDefNodeView{}, errors.New("cannot serialize unknown sequence node")
	}
	return view, nil
}

type SeqDefPreview struct {
	ID       ubom.SeqDefID     `json:"id"`
	Bindings map[string]string `json:"bindings"`
	Value    string            `json:"value,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type SeqDefCandidateField struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	Normalized any    `json:"normalized"`
}

type SeqDefCandidatePreview struct {
	Value    string                 `json:"value"`
	Bindings map[string]string      `json:"bindings"`
	Fields   []SeqDefCandidateField `json:"fields"`
}

func (s *Service) PreviewSeqDef(id ubom.SeqDefID, bindings map[string]string) (SeqDefPreview, error) {
	def, err := s.store.GetSeqDef(id)
	if err != nil {
		return SeqDefPreview{}, err
	}
	return previewSeqDef(def, id, bindings)
}

func (s *Service) PreviewSeqDefRoot(root ubom.SeqDefNode, bindings map[string]string) (SeqDefPreview, error) {
	return previewSeqDef(ubom.NewSeqDef(root).WithID("preview"), "", bindings)
}

func (s *Service) PreviewSeqDefCandidate(root ubom.SeqDefNode, value string) (SeqDefCandidatePreview, error) {
	def := ubom.NewSeqDef(root).WithID("preview")
	if err := def.Validate(); err != nil {
		return SeqDefCandidatePreview{}, err
	}
	parsed, err := def.ParseValues(value)
	if err != nil {
		return SeqDefCandidatePreview{}, err
	}
	fields := make([]SeqDefCandidateField, 0, len(parsed.Fields))
	for _, field := range parsed.Fields {
		fields = append(fields, SeqDefCandidateField{
			Name: field.Name, Source: field.Source, Normalized: field.Normalized,
		})
	}
	return SeqDefCandidatePreview{Value: parsed.Value, Bindings: parsed.Bindings, Fields: fields}, nil
}

func previewSeqDef(def ubom.SeqDef, id ubom.SeqDefID, bindings map[string]string) (SeqDefPreview, error) {
	if err := def.Validate(); err != nil {
		return SeqDefPreview{}, err
	}
	preview := SeqDefPreview{ID: id, Bindings: bindings}
	value, err := def.Generate(bindings)
	if err != nil {
		preview.Error = err.Error()
	} else {
		preview.Value = value
	}
	return preview, nil
}
func (s *Service) GetTaxonomyDefForAuthoring(id ubom.TaxonomyDefID) (ubom.TaxonomyDef, error) {
	return s.store.GetTaxonomyDef(id)
}

func (s *Service) GetTaxonomyDefDefinitionView(id ubom.TaxonomyDefID) (TaxonomyDefDefinitionView, error) {
	definition, err := s.GetTaxonomyDefForAuthoring(id)
	if err != nil {
		return TaxonomyDefDefinitionView{}, err
	}
	current, _ := s.store.GetCurrentTaxonomyDef(definition.SeqDef)
	return s.taxonomyDefinitionView(definition, current.ID == definition.ID), nil
}
func (s *Service) taxonomyDefinitionView(definition ubom.TaxonomyDef, current bool) TaxonomyDefDefinitionView {
	definition = definition.WithLegacyDefaults()
	return TaxonomyDefDefinitionView{ID: definition.ID, DefinitionID: definition.DefinitionID, Version: definition.Version, ParentID: definition.ParentID, Current: current, SeqDef: definition.SeqDef, AttributeDefs: definition.AttributeDefs, Root: definition.Taxonomy.Root}
}
func (s *Service) GetCurrentTaxonomyDefDefinitionView(seqDef ubom.SeqDefID) (TaxonomyDefDefinitionView, error) {
	def, err := s.store.GetCurrentTaxonomyDef(seqDef)
	if err != nil {
		return TaxonomyDefDefinitionView{}, err
	}
	return s.taxonomyDefinitionView(def, true), nil
}
func (s *Service) ListTaxonomyDefHistory(definitionID string) ([]TaxonomyDefSummary, error) {
	defs, err := s.store.ListTaxonomyDefHistory(definitionID)
	if err != nil {
		return nil, err
	}
	result := make([]TaxonomyDefSummary, 0, len(defs))
	for _, def := range defs {
		def = def.WithLegacyDefaults()
		current, _ := s.store.GetCurrentTaxonomyDef(def.SeqDef)
		result = append(result, TaxonomyDefSummary{ID: def.ID, DefinitionID: def.DefinitionID, Version: def.Version, ParentID: def.ParentID, SeqDef: def.SeqDef, Current: current.ID == def.ID})
	}
	return result, nil
}
func (s *Service) CreatePartNumberDraft(draft ubom.PartNumberDraft) error {
	return s.store.CreatePartNumberDraft(draft)
}
func (s *Service) GetPartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumberDraft, error) {
	return s.store.GetPartNumberDraft(id)
}
func (s *Service) UpdatePartNumberDraft(draft ubom.PartNumberDraft) error {
	return s.store.UpdatePartNumberDraft(draft)
}
func (s *Service) ReleasePartNumberDraft(id ubom.PartNumberDraftID) (ubom.PartNumber, error) {
	return s.store.ReleasePartNumberDraft(id)
}

type PartNumberDraftPreview struct {
	ID                ubom.PartNumberDraftID `json:"id"`
	Status            ubom.DraftStatus       `json:"status"`
	Complete          bool                   `json:"complete"`
	Value             string                 `json:"value,omitempty"`
	Bindings          map[string]string      `json:"bindings"`
	SuggestedBindings map[string]string      `json:"suggestedBindings,omitempty"`
	Error             string                 `json:"error,omitempty"`
}

func (s *Service) PreviewPartNumberDraft(id ubom.PartNumberDraftID) (PartNumberDraftPreview, error) {
	draft, err := s.store.GetPartNumberDraft(id)
	if err != nil {
		return PartNumberDraftPreview{}, err
	}
	preview := PartNumberDraftPreview{ID: draft.ID, Status: draft.Status, Bindings: draft.Bindings}
	allocation, err := s.AllocatePartNumberDraft(id)
	if err != nil {
		preview.Error = err.Error()
		return preview, nil
	}
	preview.Value, preview.Bindings, preview.SuggestedBindings = allocation.Value, allocation.Bindings, allocation.Bindings
	preview.Complete = true
	return preview, nil
}
func (s *Service) ListSeqDefViews() ([]SeqDefView, error) {
	defs, err := s.store.ListSeqDefs()
	if err != nil {
		return nil, err
	}
	views := make([]SeqDefView, 0, len(defs))
	for _, def := range defs {
		def = def.WithLegacyDefaults()
		views = append(views, SeqDefView{ID: def.ID, DefinitionID: def.DefinitionID, Version: def.Version, ParentID: def.ParentID})
	}
	return views, nil
}
func (s *Service) SeqDefFieldViews(id ubom.SeqDefID) ([]SeqDefFieldView, error) {
	def, err := s.store.GetSeqDef(id)
	if err != nil {
		return nil, err
	}
	fields := make([]SeqDefFieldView, 0)
	for _, field := range def.Fields() {
		view := SeqDefFieldView{Name: field.Name}
		switch node := field.Node.(type) {
		case ubom.ChoiceNode:
			view.Kind = "choice"
			for _, child := range node.Children {
				if literal, ok := child.(ubom.LiteralNode); ok {
					view.Options = append(view.Options, literal.Text)
				}
			}
		case ubom.PlaceSequenceNode:
			view.Kind = "text"
		case ubom.RangeNode:
			view.Kind, view.Min, view.Max, view.Width = "range", &node.Min, &node.Max, node.WidthValue
			if node.Pad != 0 {
				view.Pad = string(node.Pad)
			}
		default:
			view.Kind = "text"
		}
		fields = append(fields, view)
	}
	return fields, nil
}
func (s *Service) ListTaxonomyDefSummaries() ([]TaxonomyDefSummary, error) {
	defs, err := s.store.ListTaxonomyDefs()
	if err != nil {
		return nil, err
	}
	views := make([]TaxonomyDefSummary, 0, len(defs))
	for _, def := range defs {
		def = def.WithLegacyDefaults()
		current, _ := s.store.GetCurrentTaxonomyDef(def.SeqDef)
		views = append(views, TaxonomyDefSummary{ID: def.ID, DefinitionID: def.DefinitionID, Version: def.Version, ParentID: def.ParentID, SeqDef: def.SeqDef, Current: current.ID == def.ID})
	}
	return views, nil
}

type PartNumberView struct {
	ID           ubom.PartNumberID     `json:"id"`
	Value        string                `json:"value"`
	TaxonomyPath []string              `json:"taxonomyPath"`
	Attributes   []PartAttributeView   `json:"attributes"`
	Revisions    []RevisionView        `json:"revisions"`
	Status       ubom.PartNumberStatus `json:"status"`
}

type PartAttributeView struct {
	ID        ubom.AttributeDefID     `json:"id"`
	Label     string                  `json:"label"`
	Value     string                  `json:"value"`
	ValueType ubom.AttributeValueType `json:"valueType"`
	Unit      string                  `json:"unit,omitempty"`
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
	ID            ubom.TaxonomyNodeID  `json:"id"`
	TaxonomyDefID ubom.TaxonomyDefID   `json:"taxonomyDefId"`
	SeqDefID      ubom.SeqDefID        `json:"seqDefId"`
	Label         string               `json:"label"`
	Path          []TaxonomyPathItem   `json:"path"`
	Attributes    []AttributeView      `json:"attributes"`
	Children      []TaxonomyChildView  `json:"children"`
	PartNumbers   []PartNumberListItem `json:"partNumbers"`
}

type AttributeView struct {
	ID         ubom.AttributeDefID     `json:"id"`
	Label      string                  `json:"label"`
	ValueType  ubom.AttributeValueType `json:"valueType"`
	Unit       string                  `json:"unit,omitempty"`
	EnumValues []string                `json:"enumValues,omitempty"`
	Required   bool                    `json:"required"`
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

type RevisionSummary struct {
	ID       ubom.PartRevisionID `json:"id"`
	Revision string              `json:"revision"`
}

type PartNumberListItem struct {
	ID           ubom.PartNumberID     `json:"id"`
	Value        string                `json:"value"`
	TaxonomyPath []string              `json:"taxonomyPath"`
	Attributes   []PartAttributeView   `json:"attributes"`
	Revisions    []RevisionSummary     `json:"revisions"`
	Status       ubom.PartNumberStatus `json:"status"`
}

// PartSearchQuery is deliberately store-independent for the initial part log.
// Attribute values are ORed within an attribute and attribute IDs are ANDed.
// Stores remain responsible for durable writes; search can use the same domain
// view for SQLite, memory, and Postgres deployments.
type PartSearchQuery struct {
	Q            string
	TaxonomyDef  ubom.TaxonomyDefID
	TaxonomyNode ubom.TaxonomyNodeID
	Attributes   map[ubom.AttributeDefID][]string
	Page         int
	PageSize     int
	Sort         string
	Descending   bool
}

type PartSearchResult struct {
	Items    []PartNumberListItem `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

func (s *Service) SearchPartNumberViews(query PartSearchQuery) (PartSearchResult, error) {
	parts, err := s.store.ListPartNumbers()
	if err != nil {
		return PartSearchResult{}, err
	}
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 50
	}
	if query.PageSize > 200 {
		query.PageSize = 200
	}
	q := strings.ToLower(strings.TrimSpace(query.Q))
	filtered := make([]ubom.PartNumber, 0, len(parts))
	for _, part := range parts {
		if q != "" && !strings.Contains(strings.ToLower(part.Value), q) {
			continue
		}
		if query.TaxonomyDef != "" && part.TaxonomyDefID != query.TaxonomyDef {
			continue
		}
		if query.TaxonomyNode != "" {
			taxonomy, taxonomyErr := s.store.GetTaxonomyDef(part.TaxonomyDefID)
			if taxonomyErr != nil {
				return PartSearchResult{}, taxonomyErr
			}
			path, ok := findTaxonomyPath(taxonomy.Taxonomy.Root, part.TaxonomyNodeID)
			found := false
			for _, item := range path {
				if item.ID == query.TaxonomyNode {
					found = ok
					break
				}
			}
			if !found {
				continue
			}
		}
		matched := true
		for id, values := range query.Attributes {
			attributeMatched := false
			for _, attribute := range part.Attributes {
				if attribute.AttributeDefID != id {
					continue
				}
				for _, value := range values {
					if attribute.Value == value {
						attributeMatched = true
						break
					}
				}
				if attributeMatched {
					break
				}
			}
			if !attributeMatched {
				matched = false
				break
			}
		}
		if matched {
			filtered = append(filtered, part)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left, right := filtered[i], filtered[j]
		if query.Sort == "id" {
			if query.Descending {
				return left.ID > right.ID
			}
			return left.ID < right.ID
		}
		if query.Descending {
			return left.Value > right.Value
		}
		return left.Value < right.Value
	})
	total := len(filtered)
	start := (query.Page - 1) * query.PageSize
	if start > total {
		start = total
	}
	end := start + query.PageSize
	if end > total {
		end = total
	}
	items := make([]PartNumberListItem, 0, end-start)
	for _, part := range filtered[start:end] {
		item, itemErr := s.partNumberListItem(part)
		if itemErr != nil {
			return PartSearchResult{}, itemErr
		}
		items = append(items, item)
	}
	return PartSearchResult{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, nil
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

type SourceView struct {
	ID   artifact.SourceID `json:"id"`
	Name string            `json:"name"`
	Kind string            `json:"kind"`
}

type ArtifactView struct {
	ID       artifact.ID       `json:"id"`
	Name     string            `json:"name"`
	Kind     string            `json:"kind"`
	SourceID artifact.SourceID `json:"sourceId"`
	Path     string            `json:"path"`
}

type SnapshotView struct {
	ID          artifact.SnapshotID `json:"id"`
	ArtifactID  artifact.ID         `json:"artifactId"`
	Reference   string              `json:"reference"`
	Commit      string              `json:"commit,omitempty"`
	SnapshotKey string              `json:"snapshotKey"`
	Digest      string              `json:"digest"`
	MediaType   string              `json:"mediaType"`
	Size        int64               `json:"size"`
}

func (s *Service) ListSourceViews() ([]SourceView, error) {
	sources, err := s.store.ListSources()
	if err != nil {
		return nil, err
	}
	views := make([]SourceView, 0, len(sources))
	for _, source := range sources {
		views = append(views, SourceView{ID: source.ID, Name: source.Name, Kind: source.Kind})
	}
	return views, nil
}

func (s *Service) ListArtifactViews() ([]ArtifactView, error) {
	items, err := s.store.ListArtifacts()
	if err != nil {
		return nil, err
	}
	views := make([]ArtifactView, 0, len(items))
	for _, item := range items {
		views = append(views, ArtifactView{ID: item.ID, Name: item.Name, Kind: item.Kind, SourceID: item.SourceID, Path: item.Path})
	}
	return views, nil
}

func (s *Service) GetArtifactView(id artifact.ID) (ArtifactView, error) {
	item, err := s.store.GetArtifact(id)
	if err != nil {
		return ArtifactView{}, err
	}
	return ArtifactView{ID: item.ID, Name: item.Name, Kind: item.Kind, SourceID: item.SourceID, Path: item.Path}, nil
}

func (s *Service) ListSnapshotViews(id artifact.ID) ([]SnapshotView, error) {
	snapshots, err := s.store.ListSnapshotsByArtifact(id)
	if err != nil {
		return nil, err
	}
	views := make([]SnapshotView, 0, len(snapshots))
	for _, snapshot := range snapshots {
		views = append(views, snapshotView(snapshot))
	}
	return views, nil
}

func snapshotView(snapshot artifact.Snapshot) SnapshotView {
	return SnapshotView{ID: snapshot.ID, ArtifactID: snapshot.ArtifactID, Reference: snapshot.Reference, Commit: snapshot.Commit, SnapshotKey: snapshot.SnapshotKey, Digest: snapshot.Digest, MediaType: snapshot.MediaType, Size: snapshot.Size}
}

func (s *Service) GetPartNumberView(id ubom.PartNumberID) (PartNumberView, error) {
	part, err := s.store.GetPartNumberByID(id)
	if err != nil {
		return PartNumberView{}, err
	}
	interpreted, taxonomyDef, err := s.interpretPart(part)
	if err != nil {
		return PartNumberView{}, err
	}
	attributes, err := partAttributeViews(part, taxonomyDef)
	if err != nil {
		return PartNumberView{}, err
	}

	view := PartNumberView{
		ID:           part.ID,
		Value:        part.Value,
		TaxonomyPath: interpreted.Classification.Labels,
		Attributes:   attributes,
		Revisions:    make([]RevisionView, 0, len(part.PartRevisionID)),
		Status:       part.EffectiveStatus(),
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

func partAttributeViews(part ubom.PartNumber, taxonomyDef ubom.TaxonomyDef) ([]PartAttributeView, error) {
	assignments, err := taxonomyDef.Taxonomy.EffectiveAttributes(part.TaxonomyNodeID)
	if err != nil {
		return nil, err
	}
	definitions := make(map[ubom.AttributeDefID]ubom.AttributeDef, len(taxonomyDef.AttributeDefs))
	for _, definition := range taxonomyDef.AttributeDefs {
		definitions[definition.ID] = definition
	}
	values := make(map[ubom.AttributeDefID]string, len(part.Attributes))
	for _, attribute := range part.Attributes {
		values[attribute.AttributeDefID] = attribute.Value
	}
	result := make([]PartAttributeView, 0, len(part.Attributes))
	for _, assignment := range assignments {
		definition, ok := definitions[assignment.AttributeDefID]
		if !ok {
			return nil, errors.New("taxonomy attribute assignment references unknown definition")
		}
		value, ok := values[definition.ID]
		if !ok {
			continue
		}
		result = append(result, PartAttributeView{
			ID: definition.ID, Label: definition.Label, Value: value,
			ValueType: definition.ValueType, Unit: definition.Unit,
		})
	}
	return result, nil
}

func (s *Service) ListPartNumberViews() ([]PartNumberListItem, error) {
	parts, err := s.store.ListPartNumbers()
	if err != nil {
		return nil, err
	}
	items := make([]PartNumberListItem, 0, len(parts))
	for _, part := range parts {
		item, err := s.partNumberListItem(part)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
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
	effectiveAttributes, err := taxonomyDef.Taxonomy.EffectiveAttributes(nodeID)
	if err != nil {
		return TaxonomyNodeView{}, err
	}
	attributeDefs := make(map[ubom.AttributeDefID]ubom.AttributeDef, len(taxonomyDef.AttributeDefs))
	for _, definition := range taxonomyDef.AttributeDefs {
		attributeDefs[definition.ID] = definition
	}
	attributes := make([]AttributeView, 0, len(effectiveAttributes))
	for _, assignment := range effectiveAttributes {
		definition, ok := attributeDefs[assignment.AttributeDefID]
		if !ok {
			return TaxonomyNodeView{}, errors.New("taxonomy attribute assignment references unknown definition")
		}
		attributes = append(attributes, AttributeView{
			ID: definition.ID, Label: definition.Label, ValueType: definition.ValueType,
			Unit: definition.Unit, EnumValues: definition.EnumValues, Required: assignment.Required,
		})
	}

	view := TaxonomyNodeView{
		ID:            node.ID,
		TaxonomyDefID: taxonomyID,
		SeqDefID:      taxonomyDef.SeqDef,
		Label:         node.Label,
		Path:          path,
		Attributes:    attributes,
		Children:      make([]TaxonomyChildView, 0, len(node.Children)),
		PartNumbers:   make([]PartNumberListItem, 0, len(parts)),
	}
	for _, child := range node.Children {
		view.Children = append(view.Children, TaxonomyChildView{ID: child.ID, Label: child.Label})
	}
	for _, part := range parts {
		item, err := s.partNumberListItem(part)
		if err != nil {
			return TaxonomyNodeView{}, err
		}
		view.PartNumbers = append(view.PartNumbers, item)
	}
	return view, nil
}

func (s *Service) partNumberListItem(part ubom.PartNumber) (PartNumberListItem, error) {
	taxonomyPath, err := s.partTaxonomyPath(part)
	if err != nil {
		return PartNumberListItem{}, err
	}
	item := PartNumberListItem{
		ID:           part.ID,
		Value:        part.Value,
		TaxonomyPath: taxonomyPath,
		Attributes:   []PartAttributeView{},
		Revisions:    make([]RevisionSummary, 0, len(part.PartRevisionID)),
		Status:       part.EffectiveStatus(),
	}
	taxonomyDef, err := s.store.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return PartNumberListItem{}, err
	}
	attributes, err := partAttributeViews(part, taxonomyDef)
	if err != nil {
		return PartNumberListItem{}, err
	}
	item.Attributes = attributes
	for _, revisionID := range part.PartRevisionID {
		revision, err := s.store.GetPartRevision(revisionID)
		if err != nil {
			return PartNumberListItem{}, err
		}
		item.Revisions = append(item.Revisions, RevisionSummary{ID: revision.ID, Revision: revision.Revision})
	}
	return item, nil
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
	interpreted, _, err := s.interpretPart(part)
	if err != nil {
		return nil, err
	}
	return interpreted.Classification.Labels, nil
}

// interpretPart is the application boundary for durable part-number meaning.
// Views must consume Schema's parsed and classified result rather than
// reimplementing SeqDef parsing or taxonomy projection.
func (s *Service) interpretPart(part ubom.PartNumber) (ubom.InterpretedValue, ubom.TaxonomyDef, error) {
	seqDef, err := s.store.GetSeqDef(part.SeqDefID)
	if err != nil {
		return ubom.InterpretedValue{}, ubom.TaxonomyDef{}, err
	}
	taxonomyDef, err := s.store.GetTaxonomyDef(part.TaxonomyDefID)
	if err != nil {
		return ubom.InterpretedValue{}, ubom.TaxonomyDef{}, err
	}
	// Root-only classifications remain valid for the open policy used by the
	// durable part-number contract.
	interpreted, err := ubom.NewSchema(seqDef, taxonomyDef, ubom.SchemaPolicy{}).Interpret(part.Value)
	if err != nil {
		return ubom.InterpretedValue{}, ubom.TaxonomyDef{}, err
	}
	if interpreted.Classification == nil || interpreted.Classification.NodeID != part.TaxonomyNodeID {
		return ubom.InterpretedValue{}, ubom.TaxonomyDef{}, ubom.NewDomainError(ubom.ErrorCategoryConflict, "part number classification does not match its stored taxonomy node")
	}
	return interpreted, taxonomyDef, nil
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
