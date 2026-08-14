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

func TestTaxonomyNodeView(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomy-definitions/sample-taxonomy-v1/nodes/components", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.TaxonomyNodeView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.ID != "components" || view.Label != "Components" {
		t.Fatalf("node identity = %#v, want components/Components", view)
	}
	if len(view.PartNumbers) != 2 || view.PartNumbers[0].Value != "PN-1" || view.PartNumbers[1].Value != "PN-2" {
		t.Fatalf("part numbers = %#v, want PN-1 then PN-2", view.PartNumbers)
	}
}

func TestTaxonomyNodeViewNotFound(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomy-definitions/missing/nodes/missing", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recording.Code, http.StatusNotFound)
	}
}
