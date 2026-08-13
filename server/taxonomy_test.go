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
