package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	ubom "ubom-v4"
	"ubom-v4/app"
	"ubom-v4/artifact"
	"ubom-v4/store"
)

type Server struct {
	service         *app.Service
	snapshotService *app.SnapshotService
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

func NewServerWithSnapshotService(service *app.Service, snapshotService *app.SnapshotService) *Server {
	return &Server{service: service, snapshotService: snapshotService}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/parts", s.handlePartNumbers)
	mux.HandleFunc("/api/seq-defs", s.handleSeqDefs)
	mux.HandleFunc("/api/seq-defs/", s.handleSeqDefFields)
	mux.HandleFunc("/api/taxonomy-defs", s.handleTaxonomyDefs)
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

func (s *Server) handlePartNumbers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	views, err := s.service.ListPartNumberViews()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load part numbers")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(views); err != nil {
		return
	}
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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/parts/")
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
