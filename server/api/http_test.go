package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ubom-v4/app"
	"ubom-v4/store"
)

func TestPartNumberView(t *testing.T) {
	persistence := store.NewMemoryStore()
	parent, err := app.LoadSampleData(persistence)
	if err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/part-numbers/"+string(parent.ID), nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.PartNumberView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.ID != parent.ID || view.Value != "PN-1" {
		t.Fatalf("view identity = %#v, want ID %q and value %q", view, parent.ID, "PN-1")
	}
	if len(view.TaxonomyPath) != 1 || view.TaxonomyPath[0] != "Components" {
		t.Fatalf("taxonomy path = %#v, want [Components]", view.TaxonomyPath)
	}
	if len(view.Revisions) != 1 || len(view.Revisions[0].BOM) != 1 {
		t.Fatalf("revisions = %#v, want one revision with one BOM item", view.Revisions)
	}
}

func TestPartNumberViewNotFound(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/part-numbers/missing", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recording.Code, http.StatusNotFound)
	}
}
