package ubom

import (
	"reflect"
	"testing"
)

func TestTaxonomyProject(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		Label: "component",
		Children: []TaxonomyNode{
			{
				Label:   "resistor",
				Matches: map[string]string{"category": "001"},
				Children: []TaxonomyNode{
					{
						Label:   "thick-film",
						Matches: map[string]string{"family": "00042"},
					},
				},
			},
			{
				Label:   "capacitor",
				Matches: map[string]string{"category": "002"},
			},
		},
	}}

	got := taxonomy.Project(ParseResult{Bindings: map[string]string{
		"category": "001",
		"family":   "00042",
		"id":       "001",
	}})
	want := []string{"component", "resistor", "thick-film"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Project() = %#v, want %#v", got, want)
	}
}

func TestTaxonomyProjectStopsAtUnknownValue(t *testing.T) {
	taxonomy := Taxonomy{Root: TaxonomyNode{
		Label: "component",
		Children: []TaxonomyNode{{
			Label:   "resistor",
			Matches: map[string]string{"category": "001"},
		}},
	}}

	got := taxonomy.Project(ParseResult{Bindings: map[string]string{
		"category": "999",
	}})
	want := []string{"component"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Project() = %#v, want %#v", got, want)
	}
}

func TestTaxonomyDefValidate(t *testing.T) {
	tests := []struct {
		name string
		def  TaxonomyDef
		want bool
	}{
		{
			name: "valid",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{ID: "resistors"}},
				}},
			},
			want: true,
		},
		{
			name: "missing ID",
			def: TaxonomyDef{
				SeqDef:   "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root"}},
			},
		},
		{
			name: "missing sequence definition ID",
			def: TaxonomyDef{
				ID:       "taxonomy-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{ID: "root"}},
			},
		},
		{
			name: "missing root ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
			},
		},
		{
			name: "missing child ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{}},
				}},
			},
		},
		{
			name: "duplicate node ID",
			def: TaxonomyDef{
				ID:     "taxonomy-v1",
				SeqDef: "pn-v1",
				Taxonomy: Taxonomy{Root: TaxonomyNode{
					ID:       "root",
					Children: []TaxonomyNode{{ID: "root"}},
				}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.def.Validate()
			if test.want && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !test.want && err == nil {
				t.Fatal("Validate() accepted invalid definition")
			}
		})
	}
}
