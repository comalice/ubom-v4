# Server Performance Benchmarks

## Scope

The catalog service benchmarks use a deterministic in-memory fixture populated
with 100 parts, one revision per part, and seed `1`. They are visibility and
regression baselines, not pass/fail performance tests.

## Run

From the repository root, run:

```sh
cd server && go test ./app -run '^$' -bench 'Benchmark(GetTaxonomyNodeView|SearchPartNumberViewsWithTaxonomyNode|CatalogFacetsForScoped|CatalogRootFor)$' -benchmem
```

Use `-count N` when comparing changes across repeated runs, for example
`-count 5`. Keep the Go version, machine, benchmark flags, and fixture
configuration consistent when comparing results.

## Record

Record the benchmark name, `ns/op`, `B/op`, `allocs/op`, Go version, platform,
and benchmark flags. Note machine load or other environmental conditions that
could affect a comparison. Do not add timing thresholds to these benchmarks;
investigate meaningful changes in the recorded baseline instead.
