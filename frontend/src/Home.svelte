<script lang="ts">
  import { onMount } from 'svelte'
  import AsyncState from './AsyncState.svelte'
  import PartLog from './PartLog.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import { getCurrentTaxonomyDef, getTaxonomyDefs, type TaxonomyDefSummary } from './api'

  let taxonomies: TaxonomyDefSummary[] = []
  let taxonomyRoots: Record<string, string> = {}
  let totalParts = 0
  let error = ''
  let loading = true
  onMount(async () => {
    try {
      taxonomies = await getTaxonomyDefs()
      const current = taxonomies.filter(taxonomy => taxonomy.current)
      const definitions = await Promise.all(current.map(taxonomy => getCurrentTaxonomyDef(taxonomy.seqDef)))
      taxonomyRoots = Object.fromEntries(definitions.map((definition, index) => [current[index].seqDef, definition.root.id]))
    }
    catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to load summary' }
    finally { loading = false }
  })
</script>

<section class="home-hero" aria-labelledby="home-heading">
  <p class="eyebrow">Part log</p>
  <h1 id="home-heading">The working language of your parts.</h1>
  <p class="home-lede">Search released part numbers, or start with the taxonomy that gives a new part its context.</p>
</section>

{#if loading}<AsyncState state="loading" message="Loading summary..." />
{:else if error}<AsyncState state="error" message={error} />
{:else}
  <div class="summary-cards" aria-label="Part log summary">
    <div><strong>{totalParts}</strong><span>Released part numbers</span></div>
    <div><strong>{taxonomies.filter(taxonomy => taxonomy.current).length}</strong><span>Current taxonomies</span></div>
    <PrimitiveLink className="summary-action" href="/settings/taxonomies"><strong>Taxonomy</strong><span>Manage classification language</span></PrimitiveLink>
  </div>
{/if}

<PartLog compact bind:totalParts />

{#if taxonomies.some(taxonomy => taxonomy.current)}
  <section class="home-taxonomies" aria-labelledby="home-taxonomies-heading">
    <div class="part-log-heading"><div><p class="eyebrow">Browse by context</p><h2 id="home-taxonomies-heading">Taxonomies</h2></div><PrimitiveLink href="/settings/taxonomies">Manage</PrimitiveLink></div>
    <div class="cards">
      {#each taxonomies.filter(taxonomy => taxonomy.current) as taxonomy}
        <PrimitiveLink className="card" href={`/parts/${encodeURIComponent(taxonomyRoots[taxonomy.seqDef])}?seqDef=${encodeURIComponent(taxonomy.seqDef)}`}><span>{taxonomy.definitionId}<small>v{taxonomy.version}</small></span><span class="card-arrow" aria-hidden="true">→</span></PrimitiveLink>
      {/each}
    </div>
  </section>
{/if}
