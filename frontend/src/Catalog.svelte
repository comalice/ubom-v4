<script lang="ts">
  import { onMount } from 'svelte'
  import AsyncState from './AsyncState.svelte'
  import CatalogBrowser from './CatalogBrowser.svelte'
  import CategoryList, { type CategoryListItem } from './CategoryList.svelte'
  import CategoryTileGrid from './CategoryTileGrid.svelte'
  import FilterOptionList, { type FilterOption } from './FilterOptionList.svelte'
  import FilterWidget from './FilterWidget.svelte'
  import PartTable from './PartTable.svelte'
  import { getCurrentTaxonomyDef, getTaxonomyDef, getTaxonomyDefs, getTaxonomyNode, resolveTaxonomyPath, searchParts, type PartListItem, type TaxonomyDefinitionView, type TaxonomyNodeView } from './api'
  import { catalogPageSizeOptions, catalogPath, catalogServerMaxPageSize, taxonomySlug, type CatalogRouteState } from './routes'

  export let state: CatalogRouteState | undefined = undefined
  export let navigate: (path: string, replace?: boolean) => void

  let definition: TaxonomyDefinitionView | null = null
  let node: TaxonomyNodeView | null = null
  let parts: PartListItem[] = []
  let total = 0
  let loading = true
  let error = ''
  let query = ''
  let filters: Record<string, string[]> = {}
  let currentPage = 1
  let currentPageSize = 50
  let request = 0

  const pageSizeOptions = [...catalogPageSizeOptions, catalogServerMaxPageSize]
  const sortOptions = [
    { value: 'value', label: 'Part number' },
    { value: 'id', label: 'ID' },
  ]

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
      const resolvedPath = routeState.path ? resolveTaxonomyPath(taxonomy.root, routeState.path) : undefined
      if (routeState.path && !resolvedPath) throw new Error('404 Not Found')
      const nodeID = resolvedPath?.at(-1) ?? routeState.node ?? taxonomy.root.id
      const resolvedNode = await getTaxonomyNode(taxonomy.id, nodeID)
      const effectiveAttributeIDs = new Set(resolvedNode.attributes
        .filter(attribute => attribute.valueType === 'enum' && (attribute.enumValues?.length ?? 0) > 0)
        .map(attribute => attribute.id))
      const effectiveFilters = Object.fromEntries(Object.entries(routeState.filters ?? {})
        .filter(([attribute]) => effectiveAttributeIDs.has(attribute)))
      const effectiveState = { ...routeState, taxonomyDef: taxonomy.id, node: nodeID, filters: effectiveFilters }
      const result = await searchParts(effectiveState)
      if (requestID !== request) return
      definition = taxonomy
      node = resolvedNode
      parts = result.items
      total = result.total
      currentPage = result.page
      currentPageSize = result.pageSize
      const canonicalPage = Math.min(Math.max(1, result.page), Math.max(1, Math.ceil(result.total / result.pageSize)))
      const canonicalState = { ...routeState, filters: effectiveFilters, page: canonicalPage === 1 ? undefined : canonicalPage }
      const canonicalPath = catalogPath(canonicalState)
      const currentPath = `${window.location.pathname}${window.location.search}`
      if (canonicalPath !== currentPath) navigate(canonicalPath, true)
    } catch (reason) {
      if (requestID !== request) return
      definition = null
      node = null
      parts = []
      total = 0
      currentPage = 1
      currentPageSize = 50
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
    const path = node?.path.map(item => taxonomySlug(item.label)) ?? []
    const child = node?.children.find(item => item.id === id)
    return catalogPath(scopedState({ path: [...path, taxonomySlug(child?.label ?? id)], node: undefined, page: undefined }))
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
      href: index < node!.path.length - 1
        ? catalogPath(scopedState({ path: node!.path.slice(0, index + 1).map(pathItem => taxonomySlug(pathItem.label)), node: undefined, page: undefined }))
        : undefined,
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

  function updatePage(page: number) {
    navigate(catalogPath(scopedState({ node: node?.id, page })))
  }

  function updatePageSize(event: Event) {
    const pageSize = Number((event.currentTarget as HTMLSelectElement).value)
    navigate(catalogPath(scopedState({ node: node?.id, pageSize, page: undefined })))
  }

  function updateSort(event: Event) {
    const sort = (event.currentTarget as HTMLSelectElement).value
    navigate(catalogPath(scopedState({ node: node?.id, sort, page: undefined })))
  }

  function updateOrder(event: Event) {
    const order = (event.currentTarget as HTMLSelectElement).value as 'asc' | 'desc'
    navigate(catalogPath(scopedState({ node: node?.id, order, page: undefined })))
  }

  $: totalPages = Math.max(1, Math.ceil(total / currentPageSize))
  $: selectedSort = state?.sort ?? 'value'
  $: selectedOrder = state?.order ?? 'asc'
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
        <a href={catalogPath({ ...scopedState({ node: undefined, page: undefined }), path: undefined })}>Catalog</a>
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
        <div class="catalog-controls" aria-label="Catalog result controls">
          <label>Sort by
            <select value={selectedSort} on:change={updateSort}>
              {#each sortOptions as option}<option value={option.value}>{option.label}</option>{/each}
            </select>
          </label>
          <label>Order
            <select value={selectedOrder} on:change={updateOrder}>
              <option value="asc">Ascending</option>
              <option value="desc">Descending</option>
            </select>
          </label>
          <label>Results per page
            <select value={currentPageSize} on:change={updatePageSize}>
              {#if !pageSizeOptions.includes(currentPageSize)}<option value={currentPageSize}>{currentPageSize}</option>{/if}
              {#each pageSizeOptions as option}<option value={option}>{option}</option>{/each}
            </select>
          </label>
        </div>
        {#if parts.length > 0}<PartTable parts={parts} columns={node.attributes} seqDef={node.seqDefId} />{:else}<AsyncState state="empty" message="No parts in this scope." />{/if}
        <nav class="catalog-pagination" aria-label="Catalog pagination">
          <button type="button" on:click={() => updatePage(currentPage - 1)} disabled={currentPage <= 1}>Previous</button>
          <span aria-live="polite">Page {currentPage} of {totalPages}</span>
          <button type="button" on:click={() => updatePage(currentPage + 1)} disabled={currentPage >= totalPages}>Next</button>
        </nav>
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
  .catalog-controls { display: flex; flex-wrap: wrap; gap: .75rem 1rem; margin-bottom: 1rem; color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-controls label { display: grid; gap: .3rem; }
  .catalog-controls select { min-width: 8rem; padding: .45rem .55rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .72rem Inter, sans-serif; }
  .catalog-pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; margin-top: 1rem; color: var(--muted); font: .7rem ui-monospace, monospace; }
  .catalog-pagination button { padding: .45rem .7rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; cursor: pointer; font: inherit; }
  .catalog-pagination button:disabled { color: var(--muted); cursor: not-allowed; opacity: .55; }
</style>
