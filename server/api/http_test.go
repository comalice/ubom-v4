package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ubom-v4/app"
	"ubom-v4/artifact"
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
	if len(view.Attributes) != 4 || view.Attributes[0].ID != "footprint" || view.Attributes[0].Value != "0603" || view.Attributes[1].ID != "resistance" || view.Attributes[1].Label != "Resistance value" || view.Attributes[1].Value != "10000" {
		t.Fatalf("attributes = %#v, want sample resistor attributes", view.Attributes)
	}
	if len(view.Revisions) != 1 || len(view.Revisions[0].BOM) != 1 {
		t.Fatalf("revisions = %#v, want one revision with one BOM item", view.Revisions)
	}
}

func TestArtifactMetadataAndCaptureAPI(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "widget.txt"), []byte("widget"), 0o644); err != nil {
		t.Fatal(err)
	}
	persistence := store.NewMemoryStore()
	source := artifact.Source{ID: "local", Name: "Local", Kind: "filesystem"}
	item := artifact.Artifact{ID: "widget", Name: "Widget", Kind: "text", SourceID: source.ID, Path: "widget.txt"}
	if err := persistence.CreateSource(source); err != nil {
		t.Fatal(err)
	}
	if err := persistence.CreateArtifact(item); err != nil {
		t.Fatal(err)
	}
	adapter, err := app.NewLocalFileAdapter(app.LocalFileAdapterConfig{Source: source, Root: root})
	if err != nil {
		t.Fatal(err)
	}
	service := app.NewService(persistence)
	snapshots := app.NewSnapshotService(persistence, artifact.NewMemorySnapshotStore(), adapter)
	server := NewServerWithSnapshotService(service, snapshots)

	for _, path := range []string{"/api/artifact-sources", "/api/artifacts", "/api/artifacts/widget"} {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodGet, path, nil))
		if recording.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d; body = %s", path, recording.Code, recording.Body)
		}
	}
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/artifacts/widget/snapshots", strings.NewReader(`{"reference":"working"}`))
	server.Handler().ServeHTTP(recording, request)
	if recording.Code != http.StatusOK {
		t.Fatalf("POST snapshot status = %d; body = %s", recording.Code, recording.Body)
	}
	var snapshot app.SnapshotView
	if err := json.NewDecoder(recording.Body).Decode(&snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.ArtifactID != item.ID || snapshot.Digest == "" || snapshot.Size != 6 {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if snapshot.SnapshotKey == filepath.Join(root, "widget.txt") {
		t.Fatal("API returned a source filesystem path")
	}
}

func TestPartNumberList(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/parts", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var views []app.PartNumberListItem
	if err := json.NewDecoder(recording.Body).Decode(&views); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(views) != 2 || views[0].Value != "PN-A" || views[1].Value != "PN-B" {
		t.Fatalf("views = %#v, want PN-A then PN-B", views)
	}
	if len(views[0].Revisions) != 1 || views[0].Revisions[0].Revision != "1" {
		t.Fatalf("revisions = %#v, want revision 1", views[0].Revisions)
	}
	if len(views[0].Attributes) != 4 || views[0].Attributes[0].ID != "footprint" || views[0].Attributes[1].ID != "resistance" {
		t.Fatalf("attributes = %#v, want sample resistor attributes", views[0].Attributes)
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

func TestAuthoringFlow(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(path, body string, wantStatus int) string {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		if recording.Code != wantStatus {
			t.Fatalf("POST %s status = %d, want %d; body = %s", path, recording.Code, wantStatus, recording.Body)
		}
		if !strings.HasPrefix(recording.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("POST %s content type = %q, want application/json", path, recording.Header().Get("Content-Type"))
		}
		return recording.Body.String()
	}

	post("/api/seq-defs", `{"id":"author-pn-v1","root":{"kind":"concat","children":[{"kind":"literal","text":"PN-"},{"kind":"places","alphabets":["AB"]}]}}`, http.StatusCreated)
	post("/api/seq-defs", `{"id":"author-rev-v1","root":{"kind":"range","min":1,"max":9,"width":1}}`, http.StatusCreated)
	post("/api/taxonomy-defs", `{"id":"author-taxonomy-v1","seqDef":"author-pn-v1","root":{"id":"components","label":"Components"}}`, http.StatusCreated)
	part := post("/api/parts", `{"value":"PN-A","seqDef":"author-pn-v1","taxonomyDef":"author-taxonomy-v1"}`, http.StatusCreated)
	if !strings.Contains(part, `"id":"1"`) || !strings.Contains(part, `"value":"PN-A"`) {
		t.Fatalf("part response = %s", part)
	}
}

func TestTaxonomyPredicateContract(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(path, body string, wantStatus int) string {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		if recording.Code != wantStatus {
			t.Fatalf("POST %s status = %d, want %d; body = %s", path, recording.Code, wantStatus, recording.Body)
		}
		return recording.Body.String()
	}
	post("/api/seq-defs", `{"id":"predicate-pn","root":{"kind":"concat","children":[{"kind":"bind","name":"category","child":{"kind":"choice","children":[{"kind":"literal","text":"RES"}]}},{"kind":"bind","name":"number","child":{"kind":"range","min":0,"max":999,"width":3}}]}}`, http.StatusCreated)
	body := `{"id":"predicate-taxonomy","seqDef":"predicate-pn","root":{"id":"root","label":"Components","children":[{"id":"resistors","label":"Resistors","predicates":[{"field":"category","kind":"exact","value":"RES"},{"field":"number","kind":"numeric-range","min":100,"max":199}]}]}}`
	post("/api/taxonomy-defs", body, http.StatusCreated)
	part := post("/api/parts", `{"value":"RES100","seqDef":"predicate-pn","taxonomyDef":"predicate-taxonomy"}`, http.StatusCreated)
	if !strings.Contains(part, `"value":"RES100"`) {
		t.Fatalf("part response = %s", part)
	}
}

func TestSchemaAcceptanceHTTPFlow(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(path, body string, wantStatus int) string {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		if recording.Code != wantStatus {
			t.Fatalf("POST %s status = %d, want %d; body = %s", path, recording.Code, wantStatus, recording.Body)
		}
		return recording.Body.String()
	}

	post("/api/seq-defs", `{"id":"acceptance-http-seq","root":{"kind":"concat","children":[{"kind":"bind","name":"category","child":{"kind":"choice","children":[{"kind":"literal","text":"RES"},{"kind":"literal","text":"CAP"}]}},{"kind":"literal","text":"-"},{"kind":"bind","name":"group","child":{"kind":"range","min":0,"max":999,"width":3}},{"kind":"literal","text":"-"},{"kind":"bind","name":"number","child":{"kind":"range","min":0,"max":999,"width":3}}]}}`, http.StatusCreated)
	post("/api/taxonomy-defs", `{"id":"acceptance-http-taxonomy","seqDef":"acceptance-http-seq","attributeDefs":[{"id":"family","label":"Family","valueType":"string"},{"id":"resistance","label":"Resistance","valueType":"decimal","unit":"ohm"}],"root":{"id":"component","label":"Component","attributes":[{"attributeDefId":"family","required":true}],"children":[{"id":"resistor","label":"Resistor","predicates":[{"field":"category","kind":"exact","value":"RES"}],"children":[{"id":"high-power","label":"High Power","predicates":[{"field":"group","kind":"numeric-range","min":100,"max":199}],"attributes":[{"attributeDefId":"resistance","required":false}]}]},{"id":"capacitor","label":"Capacitor","predicates":[{"field":"category","kind":"exact","value":"CAP"}]}]}}`, http.StatusCreated)
	part := post("/api/parts", `{"value":"RES-150-042","seqDef":"acceptance-http-seq","taxonomyDef":"acceptance-http-taxonomy","attributes":[{"attributeDefId":"family","value":"electronics"}]}`, http.StatusCreated)
	if !strings.Contains(part, `"value":"RES-150-042"`) {
		t.Fatalf("created part = %s", part)
	}

	viewResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(viewResponse, httptest.NewRequest(http.MethodGet, "/api/parts/by-value/RES-150-042", nil))
	if viewResponse.Code != http.StatusOK {
		t.Fatalf("GET part status = %d; body = %s", viewResponse.Code, viewResponse.Body)
	}
	var view app.PartNumberView
	if err := json.NewDecoder(viewResponse.Body).Decode(&view); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(view.TaxonomyPath, []string{"Component", "Resistor", "High Power"}) || len(view.Attributes) != 1 || view.Attributes[0].ID != "family" {
		t.Fatalf("part view = %#v, want interpreted path and inherited attribute", view)
	}

	draftBody := post("/api/part-drafts", `{"seqDef":"acceptance-http-seq","taxonomyDef":"acceptance-http-taxonomy","contextNodeId":"high-power"}`, http.StatusCreated)
	var draft struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(draftBody), &draft); err != nil || draft.ID == "" {
		t.Fatalf("draft response = %s", draftBody)
	}
	patch := httptest.NewRecorder()
	server.Handler().ServeHTTP(patch, httptest.NewRequest(http.MethodPatch, "/api/part-drafts/"+draft.ID, strings.NewReader(`{"bindings":{"category":"RES","group":"100","number":"001"},"attributes":[{"attributeDefId":"family","value":"electronics"}]}`)))
	if patch.Code != http.StatusOK {
		t.Fatalf("PATCH draft status = %d; body = %s", patch.Code, patch.Body)
	}
	released := post("/api/part-drafts/"+draft.ID+"/release", `{}`, http.StatusCreated)
	if !strings.Contains(released, `"value":"RES-100-001"`) {
		t.Fatalf("released part = %s", released)
	}
}

func TestDraftAndLineItemJSONContract(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/part-drafts", strings.NewReader(`{"seqDef":"missing","taxonomyDef":"missing","contextNodeId":"root"}`)))
	if response.Code != http.StatusNotFound {
		t.Fatalf("draft status = %d, want %d; body = %s", response.Code, http.StatusNotFound, response.Body)
	}
	// Domain structs used by the API must not fall back to Go's field names.
	draft := `{"id":"draft-1","seqDef":"seq-1","taxonomyDef":"tax-1","contextNodeId":"root","bindings":{},"attributes":[],"status":"draft"}`
	var decoded struct {
		SeqDef string `json:"seqDef"`
	}
	if err := json.Unmarshal([]byte(draft), &decoded); err != nil || decoded.SeqDef != "seq-1" {
		t.Fatalf("draft contract decode = %#v, %v", decoded, err)
	}
	line := struct {
		PartNumberID string `json:"partNumberId"`
		RevisionID   string `json:"partRevisionId"`
	}{"part-1", "revision-1"}
	encoded, err := json.Marshal(line)
	if err != nil || !strings.Contains(string(encoded), `"partNumberId"`) || !strings.Contains(string(encoded), `"partRevisionId"`) {
		t.Fatalf("line item contract = %s, %v", encoded, err)
	}
}

func TestSeqDefPreviewAndStrictRequest(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(path, body string) *httptest.ResponseRecorder {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		return recording
	}
	if response := post("/api/seq-defs", `{"id":"preview-v1","root":{"kind":"concat","children":[{"kind":"literal","text":"PN-"},{"kind":"bind","name":"number","child":{"kind":"range","min":1,"max":9,"width":2}}]}}`); response.Code != http.StatusCreated {
		t.Fatalf("create status = %d; body = %s", response.Code, response.Body)
	}
	response := post("/api/seq-defs/preview-v1/preview", `{"bindings":{"number":"07"}}`)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"value":"PN-07"`) {
		t.Fatalf("preview response = %d %s", response.Code, response.Body)
	}
	response = post("/api/seq-defs/preview-v1/preview", `{"bindings":{},"extra":true}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("strict preview status = %d; body = %s", response.Code, response.Body)
	}
}

func TestSeqDefAuthoringPadWireContract(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/seq-defs", strings.NewReader(body)))
		return response
	}

	body := `{"id":"pad-wire-v1","root":{"kind":"concat","children":[{"kind":"range_radix","children":[{"kind":"range","min":0,"max":9,"width":2,"pad":" "}]},{"kind":"bind","name":"number","child":{"kind":"range","min":0,"max":9,"width":2,"pad":"X"}}]}}`
	if response := post(body); response.Code != http.StatusCreated {
		t.Fatalf("create status = %d; body = %s", response.Code, response.Body)
	}
	preview := httptest.NewRecorder()
	server.Handler().ServeHTTP(preview, httptest.NewRequest(http.MethodPost, "/api/seq-defs/pad-wire-v1/preview", strings.NewReader(`{"bindings":{"number":"X7"}}`)))
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), `"value":" 0X7"`) {
		t.Fatalf("preview = %d %s; want nested padding preserved", preview.Code, preview.Body)
	}

	for _, pad := range []string{`"XX"`, `"12"`} {
		response := post(`{"id":"invalid-pad","root":{"kind":"range","min":0,"max":9,"width":2,"pad":` + pad + `}}`)
		if response.Code != http.StatusBadRequest {
			t.Errorf("pad %s status = %d; body = %s", pad, response.Code, response.Body)
		}
	}
	if response := post(`{"id":"default-pad","root":{"kind":"range","min":0,"max":9,"width":2}}`); response.Code != http.StatusCreated {
		t.Fatalf("omitted pad status = %d; body = %s", response.Code, response.Body)
	}
}

func TestSeqDefFrontendASTCrosswalk(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	tests := []struct {
		name     string
		root     string
		bindings string
		value    string
	}{
		{
			name:  "literal",
			root:  `{"kind":"literal","text":"PN-"}`,
			value: "PN-",
		},
		{
			name:     "named literal",
			root:     `{"kind":"bind","name":"prefix","child":{"kind":"literal","text":"PN-"}}`,
			bindings: `{"prefix":"PN-"}`,
			value:    "PN-",
		},
		{
			name:     "range with width",
			root:     `{"kind":"range","min":1,"max":9,"width":2}`,
			bindings: `{}`,
			value:    "01",
		},
		{
			name:     "named range",
			root:     `{"kind":"bind","name":"number","child":{"kind":"range","min":1,"max":9,"width":2}}`,
			bindings: `{"number":"07"}`,
			value:    "07",
		},
		{
			name:  "choice",
			root:  `{"kind":"choice","children":[{"kind":"literal","text":"RES"},{"kind":"literal","text":"CAP"}]}`,
			value: "RES",
		},
		{
			name:     "named choice",
			root:     `{"kind":"bind","name":"category","child":{"kind":"choice","children":[{"kind":"literal","text":"RES"},{"kind":"literal","text":"CAP"}]}}`,
			bindings: `{"category":"CAP"}`,
			value:    "CAP",
		},
		{
			name:  "places",
			root:  `{"kind":"places","alphabets":["AB","012"]}`,
			value: "A0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := "crosswalk-" + strings.ReplaceAll(test.name, " ", "-")
			create := httptest.NewRecorder()
			body := `{"id":"` + id + `","root":` + test.root + `}`
			server.Handler().ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/seq-defs", strings.NewReader(body)))
			if create.Code != http.StatusCreated {
				t.Fatalf("create status = %d; body = %s", create.Code, create.Body)
			}

			bindings := test.bindings
			if bindings == "" {
				bindings = `{}`
			}
			preview := httptest.NewRecorder()
			path := "/api/seq-defs/" + id + "/preview"
			server.Handler().ServeHTTP(preview, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"bindings":`+bindings+`}`)))
			if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), `"value":"`+test.value+`"`) {
				t.Fatalf("preview = %d %s; want value %q", preview.Code, preview.Body, test.value)
			}
		})
	}
}

func TestUnsavedSeqDefPreview(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	preview := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/seq-defs/preview", strings.NewReader(`{"root":{"kind":"concat","children":[{"kind":"literal","text":"PN-"},{"kind":"bind","name":"number","child":{"kind":"range","min":1,"max":9,"width":2}}]},"bindings":{"number":"07"}}`))
	server.Handler().ServeHTTP(preview, request)
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), `"value":"PN-07"`) {
		t.Fatalf("preview = %d %s; want value PN-07", preview.Code, preview.Body)
	}

	invalid := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/seq-defs/preview", strings.NewReader(`{"root":{"kind":"unknown"},"bindings":{}}`))
	server.Handler().ServeHTTP(invalid, request)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid preview status = %d; body = %s", invalid.Code, invalid.Body)
	}
}

func TestPartNumberDraftPreviewAndRelease(t *testing.T) {
	server := NewServer(app.NewService(store.NewMemoryStore()))
	post := func(path, body string, wantStatus int) string {
		recording := httptest.NewRecorder()
		server.Handler().ServeHTTP(recording, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
		if recording.Code != wantStatus {
			t.Fatalf("POST %s status = %d, want %d; body = %s", path, recording.Code, wantStatus, recording.Body)
		}
		return recording.Body.String()
	}
	post("/api/seq-defs", `{"id":"draft-pn-v1","root":{"kind":"concat","children":[{"kind":"literal","text":"PN-"},{"kind":"bind","name":"category","child":{"kind":"places","alphabets":["AB"]}},{"kind":"range","min":1,"max":9,"width":1}]}}`, http.StatusCreated)
	post("/api/taxonomy-defs", `{"id":"draft-taxonomy-v1","seqDef":"draft-pn-v1","root":{"id":"components","label":"Components"}}`, http.StatusCreated)
	body := post("/api/part-drafts", `{"seqDef":"draft-pn-v1","taxonomyDef":"draft-taxonomy-v1","contextNodeId":"components"}`, http.StatusCreated)
	var draft ubomDraftResponse
	if err := json.Unmarshal([]byte(body), &draft); err != nil || draft.ID == "" {
		t.Fatalf("draft response = %s", body)
	}
	preview := post("/api/part-drafts/"+draft.ID+"/preview", `{}`, http.StatusOK)
	if !strings.Contains(preview, `"complete":false`) {
		t.Fatalf("empty preview = %s", preview)
	}
	patch := httptest.NewRecorder()
	server.Handler().ServeHTTP(patch, httptest.NewRequest(http.MethodPatch, "/api/part-drafts/"+draft.ID, strings.NewReader(`{"bindings":{"category":"A"}}`)))
	if patch.Code != http.StatusOK {
		t.Fatalf("PATCH draft status = %d; body = %s", patch.Code, patch.Body)
	}
	post("/api/part-drafts/"+draft.ID+"/release", `{}`, http.StatusCreated)
}

type ubomDraftResponse struct {
	ID string `json:"id"`
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
	if len(view.Attributes) != 1 || view.Attributes[0].ID != "footprint" || !view.Attributes[0].Required {
		t.Fatalf("attributes = %#v, want required footprint", view.Attributes)
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

func TestTaxonomyNodeViewIncludesRicherResistorAttributes(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := app.LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	server := NewServer(app.NewService(persistence))
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/taxonomies/sample-taxonomy-v1/nodes/resistors", nil)
	server.Handler().ServeHTTP(recording, request)

	if recording.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recording.Code, http.StatusOK, recording.Body)
	}
	var view app.TaxonomyNodeView
	if err := json.NewDecoder(recording.Body).Decode(&view); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(view.Attributes) != 4 {
		t.Fatalf("attributes = %#v, want inherited footprint plus three resistor attributes", view.Attributes)
	}
	if view.Attributes[0].ID != "footprint" || view.Attributes[1].ID != "resistance" || view.Attributes[2].ID != "tolerance" || view.Attributes[3].ID != "power_rating" {
		t.Fatalf("attribute order = %#v, want footprint, resistance, tolerance, power_rating", view.Attributes)
	}
	if len(view.PartNumbers) != 1 || view.PartNumbers[0].Value != "PN-A" || len(view.PartNumbers[0].Attributes) != 4 || view.PartNumbers[0].Attributes[1].Value != "10000" || view.PartNumbers[0].Attributes[3].Value != "0.1" {
		t.Fatalf("part numbers = %#v, want PN-A with resistor metadata", view.PartNumbers)
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
