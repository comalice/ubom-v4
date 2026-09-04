package api

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	ubom "ubom-v4"
	"ubom-v4/app"
	"ubom-v4/artifact"
	"ubom-v4/store"
)

var catalogDecimalPattern = regexp.MustCompile(`^[-+]?(?:\d+(?:\.\d*)?|\.\d+)$`)

type Server struct {
	service         *app.Service
	snapshotService *app.SnapshotService
	readiness       func() error
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

func NewServerWithSnapshotService(service *app.Service, snapshotService *app.SnapshotService) *Server {
	return &Server{service: service, snapshotService: snapshotService}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/api/parts", s.handlePartNumbers)
	mux.HandleFunc("/api/catalog/facets", s.handleCatalogFacets)
	mux.HandleFunc("/api/seq-defs", s.handleSeqDefs)
	mux.HandleFunc("/api/seq-defs/preview", s.handleUnsavedSeqDefPreview)
	mux.HandleFunc("/api/seq-defs/preview/candidate", s.handleSeqDefCandidatePreview)
	mux.HandleFunc("/api/taxonomy-defs", s.handleTaxonomyDefs)
	mux.HandleFunc("/api/taxonomy-defs/", s.handleTaxonomyDefPath)
	mux.HandleFunc("/api/seq-defs/", s.handleSeqDefPath)
	mux.HandleFunc("/api/revisions", s.handleCreateRevision)
	mux.HandleFunc("/api/part-drafts", s.handlePartDrafts)
	mux.HandleFunc("/api/part-drafts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/release") {
			s.handleReleasePartDraft(w, r)
			return
		}
		s.handlePartDraft(w, r)
	})
	mux.HandleFunc("/api/parts/by-value/", s.handlePartNumberByValue)
	mux.HandleFunc("/api/parts/", s.handlePartNumber)
	mux.HandleFunc("/api/taxonomies/", s.handleTaxonomyNode)
	mux.HandleFunc("/api/revisions/", s.handleRevision)
	mux.HandleFunc("/api/artifact-sources", s.handleSources)
	mux.HandleFunc("/api/artifacts", s.handleArtifacts)
	mux.HandleFunc("/api/artifacts/", s.handleArtifact)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	if s.readiness != nil {
		if err := s.readiness(); err != nil {
			writeJSONStatus(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready", "error": "database is not ready"})
			return
		}
	}
	writeJSON(w, map[string]string{"status": "ready"})
}

// SetReadiness adds an optional dependency check without coupling the API to a
// particular persistence implementation.
func (s *Server) SetReadiness(check func() error) { s.readiness = check }

func (s *Server) handleSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	views, err := s.service.ListSourceViews()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load artifact sources")
		return
	}
	writeJSON(w, views)
}

func (s *Server) handleArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	views, err := s.service.ListArtifactViews()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load artifacts")
		return
	}
	writeJSON(w, views)
}

func (s *Server) handleArtifact(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/artifacts/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" || (len(parts) > 1 && parts[1] != "snapshots") || len(parts) > 2 {
		writeError(w, http.StatusBadRequest, "invalid artifact path")
		return
	}
	id := artifact.ID(parts[0])
	if len(parts) == 1 && r.Method == http.MethodGet {
		view, err := s.service.GetArtifactView(id)
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "artifact not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load artifact")
			return
		}
		writeJSON(w, view)
		return
	}
	if len(parts) == 2 && parts[1] == "snapshots" {
		if r.Method == http.MethodGet {
			views, err := s.service.ListSnapshotViews(id)
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "artifact not found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "could not load snapshots")
				return
			}
			writeJSON(w, views)
			return
		}
		if r.Method == http.MethodPost {
			if s.snapshotService == nil {
				writeError(w, http.StatusNotImplemented, "snapshot capture is not configured")
				return
			}
			var request struct {
				Reference string `json:"reference"`
			}
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&request); err != nil {
				writeError(w, http.StatusBadRequest, "invalid snapshot request")
				return
			}
			if request.Reference == "" {
				writeError(w, http.StatusBadRequest, "snapshot reference is required")
				return
			}
			var extra any
			if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "invalid snapshot request")
				return
			}
			snapshot, err := s.snapshotService.Capture(r.Context(), id, request.Reference)
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "artifact not found")
				return
			}
			if err != nil {
				writeError(w, http.StatusBadRequest, "could not capture snapshot")
				return
			}
			writeJSON(w, app.SnapshotView{ID: snapshot.ID, ArtifactID: snapshot.ArtifactID, Reference: snapshot.Reference, Commit: snapshot.Commit, SnapshotKey: snapshot.SnapshotKey, Digest: snapshot.Digest, MediaType: snapshot.MediaType, Size: snapshot.Size})
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSONStatus(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) handlePartNumbers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		s.handleCreatePartNumber(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query, err := parseCatalogQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.service.SearchPartNumberViews(query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load part numbers")
		return
	}
	// Keep the original unpaginated list response for existing callers. Query
	// parameters opt into the part-log envelope and pagination metadata.
	if r.URL.RawQuery == "" {
		writeJSON(w, result.Items)
		return
	}
	writeJSON(w, result)
}

func (s *Server) handleCatalogFacets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query, err := parseCatalogQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	facets, err := s.service.CatalogFacetsFor(query)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "catalog scope not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load catalog facets")
		return
	}
	writeJSON(w, facets)
}

func parseCatalogQuery(r *http.Request) (app.PartSearchQuery, error) {
	query := app.PartSearchQuery{
		Q: r.URL.Query().Get("q"), TaxonomyDef: ubom.TaxonomyDefID(firstQueryValue(r, "taxonomyDef", "taxonomy")),
		TaxonomyNode: ubom.TaxonomyNodeID(r.URL.Query().Get("taxonomyNode")), Attributes: map[ubom.AttributeDefID][]string{},
		Ranges: map[ubom.AttributeDefID]app.PartSearchRange{}, Sort: r.URL.Query().Get("sort"),
		Descending: r.URL.Query().Get("order") == "desc" || r.URL.Query().Get("direction") == "desc",
	}
	if raw, ok := r.URL.Query()["includeArchived"]; ok {
		if len(raw) != 1 || (raw[0] != "true" && raw[0] != "false") {
			return query, errors.New("invalid includeArchived value")
		}
		includeArchived := raw[0] == "true"
		query.IncludeArchived = &includeArchived
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		query.Page = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil {
		query.PageSize = value
	}
	for _, raw := range r.URL.Query()["attribute"] {
		key, value, ok := strings.Cut(raw, "=")
		if !ok || key == "" || value == "" {
			return query, errors.New("invalid attribute filter")
		}
		query.Attributes[ubom.AttributeDefID(key)] = append(query.Attributes[ubom.AttributeDefID(key)], value)
	}
	for key, values := range r.URL.Query() {
		prefix := "attribute."
		bound := ""
		if strings.HasPrefix(key, "attributeMin.") {
			prefix, bound = "attributeMin.", "min"
		} else if strings.HasPrefix(key, "attributeMax.") {
			prefix, bound = "attributeMax.", "max"
		}
		if bound != "" {
			attributeID := strings.TrimPrefix(key, prefix)
			if attributeID == "" || len(values) != 1 || values[0] == "" {
				return query, errors.New("invalid numeric range filter for " + attributeID)
			}
			parsed, ok := parseCatalogDecimal(values[0])
			if !ok {
				return query, errors.New("invalid numeric range value for " + attributeID)
			}
			rangeValue := query.Ranges[ubom.AttributeDefID(attributeID)]
			if bound == "min" {
				rangeValue.Minimum = parsed
			} else {
				rangeValue.Maximum = parsed
			}
			query.Ranges[ubom.AttributeDefID(attributeID)] = rangeValue
			continue
		}
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		attributeID := strings.TrimPrefix(key, "attribute.")
		if attributeID == "" || len(values) == 0 {
			return query, errors.New("invalid attribute filter")
		}
		for _, value := range values {
			if value == "" {
				return query, errors.New("invalid attribute filter")
			}
			query.Attributes[ubom.AttributeDefID(attributeID)] = append(query.Attributes[ubom.AttributeDefID(attributeID)], value)
		}
	}
	for attributeID, bounds := range query.Ranges {
		if bounds.Minimum != nil && bounds.Maximum != nil && bounds.Minimum.Cmp(bounds.Maximum) > 0 {
			return query, errors.New("numeric range minimum exceeds maximum for " + string(attributeID))
		}
	}
	return query, nil
}

func parseCatalogDecimal(value string) (*big.Rat, bool) {
	if !catalogDecimalPattern.MatchString(value) {
		return nil, false
	}
	return new(big.Rat).SetString(value)
}

func firstQueryValue(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if value := r.URL.Query().Get(key); value != "" {
			return value
		}
	}
	return ""
}

func (s *Server) handlePartNumberByValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	value := strings.TrimPrefix(r.URL.Path, "/api/parts/by-value/")
	if value == "" || strings.Contains(value, "/") {
		writeError(w, http.StatusBadRequest, "invalid part number value")
		return
	}
	view, err := s.service.GetPartNumberViewByValue(value)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "part number not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load part number")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		return
	}
}

func (s *Server) handlePartNumber(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/parts/")
	if strings.HasSuffix(path, "/archive") {
		id := strings.TrimSuffix(path, "/archive")
		if r.Method != http.MethodPost || id == "" || strings.Contains(id, "/") {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := s.service.ArchivePartNumber(ubom.PartNumberID(id)); err != nil {
			writeCreateError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := path
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "invalid part number ID")
		return
	}
	view, err := s.service.GetPartNumberView(ubom.PartNumberID(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "part number not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load part number")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		return
	}
}

func (s *Server) handleTaxonomyNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/taxonomies/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] != "nodes" || parts[2] == "" {
		writeError(w, http.StatusBadRequest, "invalid taxonomy node path")
		return
	}
	view, err := s.service.GetTaxonomyNodeView(ubom.TaxonomyDefID(parts[0]), ubom.TaxonomyNodeID(parts[2]))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "taxonomy node not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load taxonomy node")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		return
	}
}

func (s *Server) handleRevision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/revisions/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, "invalid revision ID")
		return
	}
	view, err := s.service.GetRevisionView(ubom.PartRevisionID(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "revision not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load revision")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(view); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
