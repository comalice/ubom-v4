<script lang="ts">
  import { onMount } from 'svelte'
  import BomNode from './BomNode.svelte'
  import Breadcrumbs, { type Breadcrumb } from './Breadcrumbs.svelte'
  import Shell from './Shell.svelte'
  import { getPart, getRevision, getTaxonomyNode, type PartNumberView, type RevisionView, type TaxonomyNodeView } from './api'
  import { parseRoute, partPath, revisionPath, taxonomyPath, type Route } from './routes'

  let route: Route = parseRoute()
  let taxonomy: TaxonomyNodeView | null = null
  let part: PartNumberView | null = null
  let revision: RevisionView | null = null
  let error = ''
  let loading = true

  $: breadcrumbs = getBreadcrumbs()

  function getBreadcrumbs(): Breadcrumb[] {
    if (route.kind === 'taxonomy' && taxonomy) return [{ label: taxonomy.label }]
    if (route.kind === 'part' && part) return [{ label: 'Part numbers' }, { label: part.value }]
    if (route.kind === 'revision' && revision) {
      return [
        { label: revision.partNumber.value, href: partPath(revision.partNumber.id) },
        { label: `Revision ${revision.id}` },
      ]
    }
    return []
  }

  async function load() {
    route = parseRoute()
    taxonomy = null
    part = null
    revision = null
    error = ''
    loading = true
    try {
      if (route.kind === 'taxonomy') taxonomy = await getTaxonomyNode(route.taxonomyID, route.nodeID)
      else if (route.kind === 'part') part = await getPart(route.id)
      else if (route.kind === 'revision') revision = await getRevision(route.id)
      else error = 'Page not found'
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to load page'
    } finally {
      loading = false
    }
  }

  onMount(() => {
    load()
    const handlePopState = () => load()
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  })
</script>

<Shell>
  <Breadcrumbs items={breadcrumbs} />
  {#if loading}
    <p class="muted">Loading...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if taxonomy && route.kind === 'taxonomy'}
    <p class="eyebrow">Taxonomy {route.taxonomyID}</p>
    <h1>{taxonomy.label}</h1>
    {#if taxonomy.children.length > 0}
      <h2>Categories</h2>
      <div class="cards">
        {#each taxonomy.children as child}
          <a class="card" href={taxonomyPath(route.taxonomyID, child.id)}>{child.label}<span>→</span></a>
        {/each}
      </div>
    {/if}
    {#if taxonomy.partNumbers.length > 0}
      <h2>Parts</h2>
      <ul>{#each taxonomy.partNumbers as item}<li><a href={partPath(item.id)}>{item.value}</a></li>{/each}</ul>
    {:else}
      <p class="muted">No parts at this level.</p>
    {/if}
  {:else if part}
    <p class="eyebrow">Part number</p>
    <h1>{part.value}</h1>
    <p class="muted">{part.taxonomyPath.join(' / ')}</p>
    <h2>Revisions</h2>
    {#if part.revisions.length > 0}
      <ul>{#each part.revisions as item}<li><a href={revisionPath(item.id)}>Revision {item.id}</a></li>{/each}</ul>
    {:else}<p class="muted">No revisions.</p>{/if}
  {:else if revision}
    <p class="eyebrow">Revision {revision.id}</p>
    <h1><a href={partPath(revision.partNumber.id)}>{revision.partNumber.value}</a></h1>
    <h2>Bill of Materials</h2>
    {#if revision.bom.length > 0}<ul>{#each revision.bom as node (node.revisionId)}<BomNode node={node} />{/each}</ul>
    {:else}<p class="muted">No BOM entries.</p>{/if}
  {/if}
</Shell>
