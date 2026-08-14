package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	ubom "ubom-v4"
	"ubom-v4/app"
	"ubom-v4/store"
)

type Server struct {
	service *app.Service
}

func NewServer(service *app.Service) *Server {
	return &Server{service: service}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/part-numbers/", s.handlePartNumber)
	return mux
}

func (s *Server) handlePartNumber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/part-numbers/")
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

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
