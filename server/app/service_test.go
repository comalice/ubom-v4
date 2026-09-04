package app

import (
	"errors"
	"math/big"
	"testing"

	ubom "ubom-v4"
	"ubom-v4/store"
)

func TestPartNumberViewUsesTypedClassificationAndEffectiveAttributes(t *testing.T) {
	metadata := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Concat(
		ubom.Literal("PN-"),
		ubom.Bind("number", ubom.Range(0, 99).Width(2)),
	)).WithID("view-seq")
	taxonomy := ubom.TaxonomyDef{
		ID: "view-taxonomy", SeqDef: seq.ID,
		AttributeDefs: []ubom.AttributeDef{
			{ID: "finish", Label: "Finish", ValueType: ubom.AttributeValueString},
		},
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "components", Label: "Components",
			Attributes: []ubom.AttributeAssignment{{AttributeDefID: "finish", Required: true}},
			Children: []ubom.TaxonomyNode{{
				ID: "precision", Label: "Precision",
				Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("number", 10, 20)},
				Attributes: []ubom.AttributeAssignment{{AttributeDefID: "finish", Required: false}},
			}},
		}},
	}
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	part, err := ubom.NewSchema(seq, taxonomy, ubom.SchemaPolicy{}).NewPartNumber("PN-12", []ubom.PartNumberAttribute{{AttributeDefID: "finish", Value: "matte"}})
	if err != nil {
		t.Fatal(err)
	}
	part, err = metadata.CreatePartNumber(part)
	if err != nil {
		t.Fatal(err)
	}

	view, err := NewService(metadata).GetPartNumberView(part.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.TaxonomyPath) != 2 || view.TaxonomyPath[0] != "Components" || view.TaxonomyPath[1] != "Precision" {
		t.Fatalf("taxonomy path = %#v, want [Components Precision]", view.TaxonomyPath)
	}
	if len(view.Attributes) != 1 || view.Attributes[0].ID != "finish" || view.Attributes[0].Value != "matte" {
		t.Fatalf("attributes = %#v, want effective overridden finish", view.Attributes)
	}
}

func TestSearchPartNumberViewsMatchesAnyValueWithinEachAttribute(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := LoadSampleData(persistence); err != nil {
		t.Fatalf("LoadSampleData() error = %v", err)
	}

	result, err := NewService(persistence).SearchPartNumberViews(PartSearchQuery{
		Attributes: map[ubom.AttributeDefID][]string{
			"footprint":  {"0805", "0603"},
			"resistance": {"999", "10000"},
		},
	})
	if err != nil {
		t.Fatalf("SearchPartNumberViews() error = %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].Value != "PN-A" {
		t.Fatalf("result = %#v, want only PN-A matching one value from each filter group", result)
	}
}

func TestSearchPartNumberViewsMatchesDecimalRangesExactly(t *testing.T) {
	persistence := store.NewMemoryStore()
	if _, err := LoadSampleData(persistence); err != nil {
		t.Fatal(err)
	}
	minimum, _ := new(big.Rat).SetString("9999.999")
	maximum, _ := new(big.Rat).SetString("10000.000")
	result, err := NewService(persistence).SearchPartNumberViews(PartSearchQuery{
		Ranges: map[ubom.AttributeDefID]PartSearchRange{"resistance": {Minimum: minimum, Maximum: maximum}},
	})
	if err != nil || result.Total != 1 || result.Items[0].Value != "PN-A" {
		t.Fatalf("result = %#v, error = %v", result, err)
	}
}

func TestGetSeqDefDefinitionViewSerializesAuthoringRoot(t *testing.T) {
	metadata := store.NewMemoryStore()
	definition := ubom.NewSeqDef(ubom.Concat(ubom.Literal("PN-"), ubom.Range(0, 9).Width(2, ' '))).WithID("authoring-view")
	if err := metadata.CreateSeqDef(definition); err != nil {
		t.Fatal(err)
	}

	view, err := NewService(metadata).GetSeqDefDefinitionView(definition.ID)
	if err != nil {
		t.Fatal(err)
	}
	if view.ID != definition.ID || view.DefinitionID != string(definition.ID) || view.Version != 1 || view.Root.Kind != "concat" || len(view.Root.Children) != 2 {
		t.Fatalf("view = %#v, want complete root", view)
	}
	if view.Root.Children[1].Pad != " " {
		t.Fatalf("pad = %q, want one-character string", view.Root.Children[1].Pad)
	}
}

func TestCreateSeqDefVersionPreservesParent(t *testing.T) {
	metadata := store.NewMemoryStore()
	parent := ubom.NewSeqDef(ubom.Literal("A")).WithID("widget-v1")
	if err := metadata.CreateSeqDef(parent); err != nil {
		t.Fatal(err)
	}
	service := NewService(metadata)
	view, err := service.CreateSeqDefVersion(parent.ID, ubom.Literal("B"), 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if view.ID != "widget-v1-v2" || view.DefinitionID != "widget-v1" || view.Version != 2 || view.ParentID != parent.ID {
		t.Fatalf("view = %#v", view)
	}
	original, err := metadata.GetSeqDef(parent.ID)
	if err != nil || original.Root().(ubom.LiteralNode).Text != "A" {
		t.Fatalf("parent changed: %#v, %v", original, err)
	}
	if _, err := service.CreateSeqDefVersion(parent.ID, ubom.Literal("C"), 1, ""); err == nil {
		t.Fatal("accepted conflicting version")
	}
}

func TestCreateTaxonomyRevisionsAreImmutableAndCurrent(t *testing.T) {
	metadata := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Literal("PN")).WithID("part-v1")
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	service := NewService(metadata)
	root := ubom.TaxonomyNode{ID: "root", Label: "Root"}
	first, err := service.CreateTaxonomyRevision(seq.ID, "", "part-taxonomy", nil, root)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "part-taxonomy-v1" || first.Version != 1 || first.DefinitionID != "part-taxonomy" || first.Current == false {
		t.Fatalf("first = %#v", first)
	}
	second, err := service.CreateTaxonomyRevision(seq.ID, first.ID, "", nil, ubom.TaxonomyNode{ID: "new-root"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != "part-taxonomy-v2" || second.Version != 2 || second.ParentID != first.ID {
		t.Fatalf("second = %#v", second)
	}
	old, err := service.GetTaxonomyDefDefinitionView(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if old.Root.ID != "root" || old.Current {
		t.Fatalf("old revision changed = %#v", old)
	}
	current, err := service.GetCurrentTaxonomyDefDefinitionView(seq.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.ID != second.ID || !current.Current {
		t.Fatalf("current = %#v", current)
	}
	history, err := service.ListTaxonomyDefHistory("part-taxonomy")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Version != 1 || !history[1].Current {
		t.Fatalf("history = %#v", history)
	}
	if _, err := service.CreateTaxonomyRevision(seq.ID, first.ID, "", nil, root); !errors.Is(err, store.ErrAlreadyExists) {
		t.Fatalf("stale save error = %v, want conflict", err)
	}
}

func TestTaxonomyNodeViewUsesExactNodeMembership(t *testing.T) {
	metadata := store.NewMemoryStore()
	seq := ubom.NewSeqDef(ubom.Bind("kind", ubom.Choice(ubom.Literal("A"), ubom.Literal("B")))).WithID("exact-seq")
	taxonomy := ubom.TaxonomyDef{ID: "exact-taxonomy", SeqDef: seq.ID, Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
		ID: "root", Label: "Root", Children: []ubom.TaxonomyNode{{ID: "a", Label: "A", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("kind", "A")}}},
	}}}
	if err := metadata.CreateSeqDef(seq); err != nil {
		t.Fatal(err)
	}
	if err := metadata.CreateTaxonomyDef(taxonomy); err != nil {
		t.Fatal(err)
	}
	part, err := ubom.NewSchema(seq, taxonomy, ubom.SchemaPolicy{}).NewPartNumber("A", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := metadata.CreatePartNumber(part); err != nil {
		t.Fatal(err)
	}

	service := NewService(metadata)
	root, err := service.GetTaxonomyNodeView(taxonomy.ID, "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(root.PartNumbers) != 0 {
		t.Fatalf("root parts = %#v, want exact-node query to exclude descendant", root.PartNumbers)
	}
	child, err := service.GetTaxonomyNodeView(taxonomy.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(child.PartNumbers) != 1 || child.PartNumbers[0].Value != "A" {
		t.Fatalf("child parts = %#v, want A", child.PartNumbers)
	}
}

func TestPreviewSeqDefCandidateParsesStructuredFields(t *testing.T) {
	root := ubom.Concat(
		ubom.Bind("prefix", ubom.Choice(ubom.Literal("PN"), ubom.Literal("SN"))),
		ubom.Literal("-"),
		ubom.Bind("series", ubom.Choice(ubom.Literal("SN"), ubom.Literal("PN"))),
		ubom.Literal("-"),
		ubom.Bind("revision", ubom.Concat(ubom.Literal("B"), ubom.PlaceSequence("0123456789"))),
	)

	preview, err := NewService(store.NewMemoryStore()).PreviewSeqDefCandidate(root, "PN-SN-B0")
	if err != nil {
		t.Fatalf("PreviewSeqDefCandidate() error = %v", err)
	}
	if preview.Value != "PN-SN-B0" {
		t.Fatalf("value = %q, want PN-SN-B0", preview.Value)
	}
	want := map[string]string{"prefix": "PN", "series": "SN", "revision": "B0"}
	if len(preview.Bindings) != len(want) {
		t.Fatalf("bindings = %#v, want %#v", preview.Bindings, want)
	}
	for name, value := range want {
		if preview.Bindings[name] != value {
			t.Errorf("binding %q = %q, want %q", name, preview.Bindings[name], value)
		}
	}
	if len(preview.Fields) != 3 || preview.Fields[2].Name != "revision" || preview.Fields[2].Source != "B0" {
		t.Fatalf("fields = %#v, want ordered captures", preview.Fields)
	}
}

func TestPreviewSeqDefCandidateRejectsInvalidCandidate(t *testing.T) {
	root := ubom.NewSeqDef(ubom.Bind("code", ubom.PlaceSequence("AB"))).Root()
	if _, err := NewService(store.NewMemoryStore()).PreviewSeqDefCandidate(root, "C"); !ubom.IsCategory(err, ubom.ErrorCategorySyntax) {
		t.Fatalf("PreviewSeqDefCandidate() error = %v, want syntax error", err)
	}
}

func TestRevisionViewRecursesBOM(t *testing.T) {
	metadata := store.NewMemoryStore()
	parent, err := LoadSampleData(metadata)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := metadata.GetPartNumber(parent.Value)
	if err != nil {
		t.Fatal(err)
	}
	view, err := NewService(metadata).GetRevisionView(stored.PartRevisionID[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(view.BOM) != 1 || view.BOM[0].PartNumber.Value != "PN-B" || view.BOM[0].RevisionID == "" {
		t.Fatalf("BOM = %#v, want recursive child revision", view.BOM)
	}
}
