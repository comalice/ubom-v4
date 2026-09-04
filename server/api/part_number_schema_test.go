package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ubom "ubom-v4"
	"ubom-v4/app"
	"ubom-v4/store"
)

func TestPartNumberSchemaEndpoints(t *testing.T) {
	persistence := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Bind("kind", ubom.Literal("R"))).WithID("http-schema-seq")
	if err := persistence.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	taxonomy := ubom.TaxonomyDef{ID: "http-schema-tax-v1", DefinitionID: "http-schema-tax", Version: 1, SeqDef: seq.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{ID: "root", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("kind", "R")}}}}
	if err := persistence.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	server := NewServer(app.NewService(persistence))

	create := func(id, name string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{"id": id, "name": name, "taxonomyDef": taxonomy.ID, "policy": map[string]bool{"requireClassification": true}})
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, "/api/part-number-schemas", bytes.NewReader(body)))
		return recording
	}
	if response := create("http-schema-a", "Schema A"); response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", response.Code, response.Body)
	}
	if response := create("http-schema-b", "Schema B"); response.Code != http.StatusCreated {
		t.Fatalf("second create status = %d, body = %s", response.Code, response.Body)
	}
	list := httptest.NewRecorder()
	server.Handler().ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/part-number-schemas", nil))
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d", list.Code)
	}
	var views []app.PartNumberSchemaSummary
	if err := json.NewDecoder(list.Body).Decode(&views); err != nil || len(views) != 2 || views[0].SeqDef != seq.ID || views[0].TaxonomyDefinitionID != taxonomy.DefinitionID {
		t.Fatalf("list views = %#v, %v", views, err)
	}
	detail := httptest.NewRecorder()
	server.Handler().ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/part-number-schemas/http-schema-a", nil))
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d", detail.Code)
	}
	missing := httptest.NewRecorder()
	server.Handler().ServeHTTP(missing, httptest.NewRequest(http.MethodPost, "/api/part-number-schemas", bytes.NewBufferString(`{"id":"missing","name":"Missing","taxonomyDef":"no-such-tax"}`)))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing taxonomy status = %d, body = %s", missing.Code, missing.Body)
	}
	if schemas, err := persistence.ListPartNumberSchemas(); err != nil || len(schemas) != 2 {
		t.Fatalf("taxonomy creation/schema side effect check = %#v, %v", schemas, err)
	}
}
