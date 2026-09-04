package app

import (
	"testing"

	"ubom-v4/seed"
	"ubom-v4/store"
)

var (
	benchmarkTaxonomyNodeView TaxonomyNodeView
	benchmarkPartSearch       PartSearchResult
	benchmarkCatalogFacets    CatalogFacets
	benchmarkCatalogRoot      CatalogRootView
)

func benchmarkCatalogService(b *testing.B) *Service {
	b.Helper()
	persistence := store.NewMemoryStore()
	if _, err := seed.Populate(persistence, seed.Options{
		Parts:        100,
		MaxRevisions: 1,
		Seed:         1,
	}); err != nil {
		b.Fatalf("seed.Populate() error = %v", err)
	}
	return NewService(persistence)
}

func BenchmarkGetTaxonomyNodeView(b *testing.B) {
	service := benchmarkCatalogService(b)
	b.ResetTimer()
	for range b.N {
		view, err := service.GetTaxonomyNodeView("sample-taxonomy-v1", "resistors")
		if err != nil {
			b.Fatal(err)
		}
		benchmarkTaxonomyNodeView = view
	}
}

func BenchmarkSearchPartNumberViewsWithTaxonomyNode(b *testing.B) {
	service := benchmarkCatalogService(b)
	query := PartSearchQuery{
		TaxonomyDef:  "sample-taxonomy-v1",
		TaxonomyNode: "resistors",
	}
	b.ResetTimer()
	for range b.N {
		result, err := service.SearchPartNumberViews(query)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkPartSearch = result
	}
}

func BenchmarkCatalogFacetsForScoped(b *testing.B) {
	service := benchmarkCatalogService(b)
	query := PartSearchQuery{
		TaxonomyDef:  "sample-taxonomy-v1",
		TaxonomyNode: "components",
	}
	b.ResetTimer()
	for range b.N {
		facets, err := service.CatalogFacetsFor(query)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkCatalogFacets = facets
	}
}

func BenchmarkCatalogRootFor(b *testing.B) {
	service := benchmarkCatalogService(b)
	b.ResetTimer()
	for range b.N {
		root, err := service.CatalogRootFor(PartSearchQuery{})
		if err != nil {
			b.Fatal(err)
		}
		benchmarkCatalogRoot = root
	}
}
