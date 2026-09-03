<script lang="ts">
  import { onMount } from 'svelte'
  import AsyncState from './AsyncState.svelte'
  import CatalogBrowser from './CatalogBrowser.svelte'
  import CategoryList, { type CategoryListItem } from './CategoryList.svelte'
  import CategoryTileGrid from './CategoryTileGrid.svelte'
  import FilterOptionList, { type FilterOption } from './FilterOptionList.svelte'
  import FilterWidget from './FilterWidget.svelte'
  import PartTable from './PartTable.svelte'
  import { getCurrentTaxonomyDef, getTaxonomyDef, getTaxonomyDefs, getTaxonomyNode, searchParts, type PartListItem, type TaxonomyDefinitionView, type TaxonomyNodeView } from './api'
  import { catalogPath, type CatalogRouteState } from './routes'

  export let state: CatalogRouteState | undefined = undefined
  export let navigate: (path: string) => void

  let definition: TaxonomyDefinitionView | null = null
  let node: TaxonomyNodeView | null = null
  let parts: PartListItem[] = []
  let total = 0
  let loading = true
  let error = ''
  let query = ''
  let filters: Record<string, string[]> = {}
  let request = 0

  $: stateKey = JSON.stringify(state ?? {})
  $: if (stateKey !== loadedStateKey) {
    loadedStateKey = stateKey
    query = state?.q ?? ''
    filters = state?.filters ?? {}
    load(state ?? {})
  }
  let loadedStateKey = ''

  async function load(routeState: CatalogRouteState) {
    const requestID = ++request
    loading = true
    error = ''
    try {
      const taxonomy = routeState.taxonomyDef
        ? await getTaxonomyDef(routeState.taxonomyDef)
        : await getCurrentTaxonomy()
      const nodeID = routeState.node ?? taxonomy.root.id
      const [resolvedNode, result] = await Promise.all([
        getTaxonomyNode(taxonomy.id, nodeID),
        searchParts({ ...routeState, taxonomyDef: taxonomy.id, node: nodeID }),
      ])
      if (requestID !== request) return
      definition = taxonomy
      node = resolvedNode
      parts = result.items
      total = result.total
    } catch (reason) {
      if (requestID !== request) return
      definition = null
      node = null
      parts = []
      total = 0
      error = reason instanceof Error ? reason.message : 'Unable to load catalog'
    } finally {
      if (requestID === request) loading = false
    }
  }

  async function getCurrentTaxonomy() {
    const current = (await getTaxonomyDefs()).find(item => item.current)
    if (!current) throw new Error('No current taxonomy is available')
    return getCurrentTaxonomyDef(current.seqDef)
  }

  function scopedState(overrides: Partial<CatalogRouteState> = {}): CatalogRouteState {
    return { ...state, ...overrides, taxonomyDef: definition?.id ?? state?.taxonomyDef, filters: overrides.filters ?? filters }
  }

  function categoryHref(id: string) {
    return catalogPath(scopedState({ node: id, page: undefined }))
  }

  function handleCatalogLink(event: MouseEvent) {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return
    const link = (event.target as Element | null)?.closest('a')
    const href = link?.getAttribute('href')
    if (!href || !(/^\/catalog(?:[?#]|$)/.test(href))) return
    event.preventDefault()
    navigate(href)
  }

  onMount(() => {
    document.addEventListener('click', handleCatalogLink)
    return () => document.removeEventListener('click', handleCatalogLink)
  })

  function breadcrumbs() {
    if (!node || !definition) return []
    return node.path.map((item, index) => ({
      label: item.label,
      href: index < node!.path.length - 1 ? categoryHref(item.id) : undefined,
    }))
  }

  function categoryItems(): CategoryListItem[] {
    return node?.children.map(child => ({ id: child.id, label: child.label, href: categoryHref(child.id), selected: false })) ?? []
  }

  function enumFilters(): { attribute: string; options: FilterOption[] }[] {
    return node?.attributes
      .filter(attribute => attribute.valueType === 'enum' && (attribute.enumValues?.length ?? 0) > 0)
      .map(attribute => ({
        attribute: attribute.id,
        options: (attribute.enumValues ?? []).map(value => ({ label: value, value })),
      })) ?? []
  }

  function updateFilter(attribute: string, values: string[]) {
    const nextFilters = { ...filters }
    if (values.length) nextFilters[attribute] = values
    else delete nextFilters[attribute]
    filters = nextFilters
    navigate(catalogPath(scopedState({ node: node?.id, q: query || undefined, filters: nextFilters, page: undefined })))
  }

  function clearFilters() {
    filters = {}
    navigate(catalogPath(scopedState({ node: node?.id, q: query || undefined, filters: {}, page: undefined })))
  }

  function submitSearch() {
    navigate(catalogPath(scopedState({ node: node?.id, q: query || undefined, page: undefined })))
  }
</script>

{#if loading}
  <AsyncState state="loading" message="Loading catalog..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if node && definition}
  <CatalogBrowser title={node.label} subtitle="Browse parts by taxonomy and filter the current scope.">
    <form class="catalog-search" slot="search" on:submit|preventDefault={submitSearch}>
      <label for="catalog-search-input">Search parts</label>
      <input id="catalog-search-input" bind:value={query} placeholder="Part number or description" />
    </form>
    <svelte:fragment slot="sidebar">
      <div class="catalog-panel-content">
        <section class="catalog-panel-section" aria-labelledby="catalog-categories-heading">
          <p class="catalog-panel-kicker">Browse catalog</p>
          <h2 id="catalog-categories-heading">Categories</h2>
          <CategoryList categories={categoryItems()} />
        </section>
        <section class="catalog-panel-section" aria-labelledby="catalog-filters-heading">
          <div class="catalog-panel-heading"><h2 id="catalog-filters-heading">Filter by</h2><button type="button" on:click={clearFilters}>Clear</button></div>
          {#each enumFilters() as filter}
            <FilterWidget title={node.attributes.find(attribute => attribute.id === filter.attribute)?.label ?? filter.attribute}>
              <FilterOptionList options={filter.options} selectedValues={filters[filter.attribute] ?? []} onChange={values => updateFilter(filter.attribute, values)} />
            </FilterWidget>
          {:else}
            <p class="muted">No supported filters in this scope.</p>
          {/each}
        </section>
      </div>
    </svelte:fragment>
    <div class="catalog-main">
      <nav class="catalog-breadcrumbs" aria-label="Catalog scope">
        <a href={catalogPath(scopedState({ node: undefined, page: undefined }))}>Catalog</a>
        {#each breadcrumbs() as crumb, index}<span aria-hidden="true">/</span>{#if crumb.href}<a href={crumb.href}>{crumb.label}</a>{:else}<strong aria-current="page">{crumb.label}</strong>{/if}{/each}
      </nav>
      <div class="catalog-main-heading"><div><p class="eyebrow">Category</p><h2>{node.label}</h2></div><span class="catalog-result-count">{total} {total === 1 ? 'part' : 'parts'}</span></div>
      {#if node.children.length > 0}
        <section class="catalog-category-section" aria-labelledby="catalog-categories-main-heading">
          <div class="catalog-section-heading"><h3 id="catalog-categories-main-heading">Categories</h3><span>{node.children.length} categories</span></div>
          <CategoryTileGrid tiles={node.children.map(child => ({ id: child.id, title: child.label, href: categoryHref(child.id) }))} />
        </section>
      {/if}
      <section class="catalog-parts-section" aria-labelledby="catalog-parts-heading">
        <div class="catalog-section-heading"><h3 id="catalog-parts-heading">Parts in this scope</h3><span>{parts.length} shown</span></div>
        {#if parts.length > 0}<PartTable parts={parts} columns={node.attributes} seqDef={node.seqDefId} />{:else}<AsyncState state="empty" message="No parts in this scope." />{/if}
      </section>
    </div>
  </CatalogBrowser>
{/if}

<style>
  .catalog-search { display: grid; flex: 0 1 20rem; gap: .3rem; color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-search input { min-width: 0; padding: .6rem .7rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .75rem Inter, sans-serif; }
  .catalog-search input:focus { border-color: var(--accent); outline: none; }
  .catalog-panel-content { min-width: 0; }
  .catalog-panel-section { padding: 1rem; }
  .catalog-panel-section + .catalog-panel-section { border-top: 1px solid var(--hairline); }
  .catalog-panel-kicker { margin: 0 0 .35rem; color: var(--accent); font: 600 .65rem ui-monospace, monospace; letter-spacing: .1em; text-transform: uppercase; }
  .catalog-panel-content h2 { margin: 0 0 .7rem; font: 400 1.15rem Georgia, serif; }
  .catalog-panel-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .5rem; }
  .catalog-panel-heading button { padding: 0; color: var(--accent); background: none; border: 0; cursor: pointer; font: .62rem ui-monospace, monospace; }
  .catalog-breadcrumbs { display: flex; flex-wrap: wrap; gap: .4rem; margin-bottom: 1rem; color: var(--muted); font-size: .7rem; }
  .catalog-main-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; }
  .catalog-main-heading h2 { margin: .2rem 0 0; font: 400 1.6rem Georgia, serif; }
  .catalog-result-count, .catalog-section-heading span { color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-category-section, .catalog-parts-section { margin-top: 1.75rem; }
  .catalog-section-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: .7rem; padding-bottom: .4rem; border-bottom: 1px solid var(--hairline); }
  .catalog-section-heading h3 { margin: 0; font: 400 1.1rem Georgia, serif; }
</style>
