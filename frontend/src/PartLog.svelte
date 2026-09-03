<script lang="ts">
  import { onMount } from 'svelte'
  import AsyncState from './AsyncState.svelte'
  import PartTable from './PartTable.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import type { PartListItem } from './api'
  import { searchParts } from './api'

  export let compact = false
  export let initialQuery = ''
  export let taxonomyNode: string | undefined = undefined
  export let taxonomyDef: string | undefined = undefined
  export let totalParts = 0

  let query = initialQuery
  let parts: PartListItem[] = []
  let total = 0
  let page = 1
  let pageSize = compact ? 6 : 25
  let loading = true
  let error = ''

  async function load(nextPage = 1) {
    loading = true
    error = ''
    try {
      const result = await searchParts({ q: query, taxonomyNode, taxonomyDef, page: nextPage, pageSize })
      parts = result.items
      total = result.total
      totalParts = result.total
      page = result.page
      pageSize = result.pageSize
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to load the part log'
    } finally { loading = false }
  }

  function submit() { load() }
  onMount(() => load())
</script>

<section class="part-log" aria-labelledby="part-log-heading">
  {#if !compact}<div class="part-log-heading"><div><p class="eyebrow">Part log</p><h1 id="part-log-heading">Find a part number</h1></div><span class="part-count">{total} {total === 1 ? 'part' : 'parts'}</span></div>{/if}
  <form class="part-search" on:submit|preventDefault={submit}>
    <label for="part-search-input">Part number</label>
    <div class="part-search-controls">
      <input id="part-search-input" bind:value={query} placeholder="Search complete or partial number" autocomplete="off" />
      <PrimitiveButton type="submit" disabled={loading}>{loading ? 'Searching...' : 'Search'}</PrimitiveButton>
    </div>
  </form>
  {#if loading}<AsyncState state="loading" message="Loading part log..." />
  {:else if error}<AsyncState state="error" message={error} />
  {:else if parts.length === 0}<AsyncState state="empty" message={query ? `No parts match "${query}".` : 'No parts have been released yet.'} />
  {:else}<PartTable parts={parts} />{/if}
  {#if !compact && total > pageSize}
    <nav class="part-pagination" aria-label="Part log pages">
      <PrimitiveButton variant="secondary" size="small" disabled={page <= 1 || loading} on:click={() => load(page - 1)}>Previous</PrimitiveButton>
      <span>Page {page} of {Math.ceil(total / pageSize)}</span>
      <PrimitiveButton variant="secondary" size="small" disabled={page >= Math.ceil(total / pageSize) || loading} on:click={() => load(page + 1)}>Next</PrimitiveButton>
    </nav>
  {/if}
</section>
