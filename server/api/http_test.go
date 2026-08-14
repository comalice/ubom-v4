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
	request := httptest.NewRequest(http.MethodGet, "/api/parts/"+string(parent.ID), nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.PartNumberView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.ID != parent.ID || view.Value != "PN-A" {
		t.Fatalf("view identity = %#v, want ID %q and value %q", view, parent.ID, "PN-A")
	}
	if len(view.TaxonomyPath) != 2 || view.TaxonomyPath[0] != "Components" || view.TaxonomyPath[1] != "Resistors" {
		t.Fatalf("taxonomy path = %#v, want [Components Resistors]", view.TaxonomyPath)
	}
	if len(view.Revisions) != 1 || len(view.Revisions[0].BOM) != 1 {
		t.Fatalf("revisions = %#v, want one revision with one BOM item", view.Revisions)
	}
}

func TestPartNumberViewNotFound(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/parts/missing", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recording.Code, http.StatusNotFound)
	}
}

func TestPartNumberViewByValue(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/parts/by-value/PN-A", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.PartNumberView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.Value != "PN-A" {
		t.Fatalf("value = %q, want PN-A", view.Value)
	}
}

func TestTaxonomyNodeView(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomies/sample-taxonomy-v1/nodes/components", nil)
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
	if len(view.Path) != 1 || view.Path[0].ID != "components" || view.Path[0].Label != "Components" {
		t.Fatalf("path = %#v, want Components", view.Path)
	}
	if len(view.Children) != 2 || view.Children[0].ID != "resistors" || view.Children[1].ID != "capacitors" {
		t.Fatalf("children = %#v, want resistors then capacitors", view.Children)
	}
	if len(view.PartNumbers) != 0 {
		t.Fatalf("part numbers = %#v, want no parts at taxonomy root", view.PartNumbers)
	}
}

func TestTaxonomyNodeViewIncludesAncestorPath(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomies/sample-taxonomy-v1/nodes/capacitors", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.TaxonomyNodeView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(view.Path) != 2 || view.Path[0].ID != "components" || view.Path[0].Label != "Components" || view.Path[1].ID != "capacitors" || view.Path[1].Label != "Capacitors" {
		t.Fatalf("path = %#v, want Components / Capacitors", view.Path)
	}
}

func TestTaxonomyNodeViewNotFound(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomies/missing/nodes/missing", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recording.Code, http.StatusNotFound)
	}
}

func TestRevisionView(t *testing.T) {
	persistence := store.NewMemoryStore()
	parent, err := app.LoadSampleData(persistence)
	if err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}
	parent, err = persistence.GetPartNumber(parent.Value)
	if err != nil {
		t.Fatalf("GetPartNumber() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/revisions/"+string(parent.PartRevisionID[0]), nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.RevisionDetailView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if view.PartNumber.Value != "PN-A" || len(view.BOM) != 1 {
		t.Fatalf("revision view = %#v, want PN-A with one BOM node", view)
	}
	if len(view.TaxonomyPath) != 2 || view.TaxonomyPath[0] != "Components" || view.TaxonomyPath[1] != "Resistors" {
		t.Fatalf("taxonomy path = %#v, want [Components Resistors]", view.TaxonomyPath)
	}
	if view.BOM[0].PartNumber.Value != "PN-B" || view.BOM[0].RevisionID == "" {
		t.Fatalf("BOM node = %#v, want PN-B with revision ID", view.BOM[0])
	}
}

func TestRevisionViewNotFound(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/revisions/missing", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recording.Code, http.StatusNotFound)
	}
}
