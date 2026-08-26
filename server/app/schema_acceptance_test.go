package app

import (
	"errors"
	"reflect"
	"testing"

	ubom "ubom-v4"
	"ubom-v4/store"
)

func TestSchemaAcceptancePipelineAcrossStores(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T) store.Store
	}{
		{name: "memory", build: func(t *testing.T) store.Store { return store.NewMemoryStore() }},
		{name: "sqlite", build: func(t *testing.T) store.Store {
			persistence, err := store.OpenSQLiteStore("file:schema-acceptance?mode=memory&cache=shared")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { persistence.Close() })
			return persistence
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			persistence := test.build(t)
			seqDef, taxonomy := acceptanceSchema()
			if err := persistence.CreateSeqDef(seqDef); err != nil {
				t.Fatal(err)
			}
			if err := persistence.CreateTaxonomyDef(taxonomy); err != nil {
				t.Fatal(err)
			}
			storedSeqDef, err := persistence.GetSeqDef(seqDef.ID)
			if err != nil {
				t.Fatal(err)
			}
			storedTaxonomy, err := persistence.GetTaxonomyDef(taxonomy.ID)
			if err != nil {
				t.Fatal(err)
			}

			schema := ubom.NewSchema(storedSeqDef, storedTaxonomy, ubom.SchemaPolicy{})
			interpreted, err := schema.Interpret("RES-150-042")
			if err != nil {
				t.Fatal(err)
			}
			if got := interpreted.Parsed.Bindings; !reflect.DeepEqual(got, map[string]string{"category": "RES", "group": "150", "number": "042"}) {
				t.Fatalf("bindings = %#v, want structured captures", got)
			}
			if interpreted.Classification == nil || interpreted.Classification.NodeID != "high-power" || !reflect.DeepEqual(interpreted.Classification.Path, []ubom.TaxonomyNodeID{"component", "resistor", "high-power"}) {
				t.Fatalf("classification = %#v, want high-power path", interpreted.Classification)
			}
			if !reflect.DeepEqual(interpreted.Classification.Attributes, []ubom.AttributeAssignment{
				{AttributeDefID: "family", Required: true},
				{AttributeDefID: "resistance", Required: false},
			}) {
				t.Fatalf("effective attributes = %#v", interpreted.Classification.Attributes)
			}

			part, err := schema.NewPartNumber("RES-150-042", []ubom.PartNumberAttribute{{AttributeDefID: "family", Value: "electronics"}})
			if err != nil {
				t.Fatalf("NewPartNumber() error = %v", err)
			}
			if _, err := persistence.CreatePartNumber(part); err != nil {
				t.Fatalf("CreatePartNumber() error = %v", err)
			}

			service := NewService(persistence)
			if err := persistence.CreatePartNumberDraft(ubom.PartNumberDraft{
				ID: "high-power-draft", SeqDefID: seqDef.ID, TaxonomyDefID: taxonomy.ID,
				ContextNodeID: "high-power", Bindings: map[string]string{"category": "CAP", "group": "000", "number": "001"},
				Attributes: []ubom.PartNumberAttribute{{AttributeDefID: "family", Value: "electronics"}}, Status: ubom.DraftStatusDraft,
			}); err != nil {
				t.Fatal(err)
			}
			allocation, err := service.AllocatePartNumberDraft("high-power-draft")
			if err != nil || allocation.Value != "RES-100-001" {
				t.Fatalf("allocation = %#v, %v; want first available constrained candidate", allocation, err)
			}
			draft, err := persistence.GetPartNumberDraft("high-power-draft")
			if err != nil {
				t.Fatal(err)
			}
			draft.Bindings = allocation.Bindings
			if err := persistence.UpdatePartNumberDraft(draft); err != nil {
				t.Fatal(err)
			}
			released, err := service.ReleasePartNumberDraft("high-power-draft")
			if err != nil || released.Value != allocation.Value {
				t.Fatalf("release = %#v, %v; want allocated value", released, err)
			}
			if _, err := service.ReleasePartNumberDraft("high-power-draft"); err == nil {
				t.Fatal("duplicate release was accepted")
			}

			reloaded, err := persistence.GetPartNumber(allocation.Value)
			if err != nil {
				t.Fatal(err)
			}
			reinterpreted, err := ubom.NewSchema(storedSeqDef, storedTaxonomy, ubom.SchemaPolicy{}).Interpret(reloaded.Value)
			if err != nil || reinterpreted.Classification == nil || reinterpreted.Classification.NodeID != reloaded.TaxonomyNodeID {
				t.Fatalf("reloaded part = %#v, interpretation = %#v, error = %v", reloaded, reinterpreted, err)
			}
			if _, err := persistence.CreatePartNumber(released); !errors.Is(err, store.ErrAlreadyExists) {
				t.Fatalf("duplicate persistence error = %v, want ErrAlreadyExists", err)
			}
		})
	}
}

func acceptanceSchema() (ubom.SeqDef, ubom.TaxonomyDef) {
	seqDef := ubom.NewSeqDef(ubom.Concat(
		ubom.Bind("category", ubom.Choice(ubom.Literal("RES"), ubom.Literal("CAP"))),
		ubom.Literal("-"), ubom.Bind("group", ubom.Range(0, 999).Width(3)),
		ubom.Literal("-"), ubom.Bind("number", ubom.Range(0, 999).Width(3)),
	)).WithID("acceptance-seq-v1")
	taxonomy := ubom.TaxonomyDef{
		ID: "acceptance-taxonomy-v1", SeqDef: seqDef.ID,
		AttributeDefs: []ubom.AttributeDef{
			{ID: "family", Label: "Family", ValueType: ubom.AttributeValueString},
			{ID: "resistance", Label: "Resistance", ValueType: ubom.AttributeValueDecimal, Unit: "ohm"},
		},
		Taxonomy: ubom.Taxonomy{Root: ubom.TaxonomyNode{
			ID: "component", Label: "Component", Attributes: []ubom.AttributeAssignment{{AttributeDefID: "family", Required: true}},
			Children: []ubom.TaxonomyNode{
				{ID: "resistor", Label: "Resistor", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("category", "RES")}, Attributes: []ubom.AttributeAssignment{{AttributeDefID: "resistance", Required: true}}, Children: []ubom.TaxonomyNode{
					{ID: "high-power", Label: "High Power", Predicates: []ubom.TaxonomyPredicate{ubom.NumericRangePredicate("group", 100, 199)}, Attributes: []ubom.AttributeAssignment{{AttributeDefID: "resistance", Required: false}}},
				}},
				{ID: "capacitor", Label: "Capacitor", Predicates: []ubom.TaxonomyPredicate{ubom.ExactPredicate("category", "CAP")}},
			},
		}},
	}
	return seqDef, taxonomy
}
