<script lang="ts">
  import { onMount } from 'svelte'
  import BomNode from './BomNode.svelte'
  import Breadcrumbs, { type Breadcrumb } from './Breadcrumbs.svelte'
  import IdentityLink from './IdentityLink.svelte'
  import Shell from './Shell.svelte'
  import { getPart, getRevision, getTaxonomyNodeByPath, type PartNumberView, type RevisionView, type TaxonomyNodeView } from './api'
  import { parseRoute, partPath, revisionPath, taxonomyPath, type Route } from './routes'

  let route: Route = parseRoute()
  let taxonomy: TaxonomyNodeView | null = null
  let part: PartNumberView | null = null
  let revision: RevisionView | null = null
  let error = ''
  let loading = true

  $: breadcrumbs = route.kind === 'taxonomy' && taxonomy
    ? taxonomyBreadcrumbs(taxonomy)
    : route.kind === 'part' && part
      ? [{ label: 'Parts' }, { label: part.value }]
      : route.kind === 'revision' && revision
        ? [
            { label: revision.partNumber.value, href: partPath(revision.partNumber.value) },
            { label: `Revision ${revision.id}` },
          ]
        : []

  function taxonomyBreadcrumbs(view: TaxonomyNodeView): Breadcrumb[] {
    return view.path.map((item, index) => ({
      label: item.label,
      href: index < view.path.length - 1 ? taxonomyPath(view.path.slice(0, index + 1).map(pathItem => pathItem.label)) : undefined,
    }))
  }

  async function load() {
    route = parseRoute()
    taxonomy = null
    part = null
    revision = null
    error = ''
    loading = true
    try {
      if (route.kind === 'taxonomy') taxonomy = await getTaxonomyNodeByPath(route.segments)
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
    <p class="eyebrow">Parts</p>
    <h1>{taxonomy.label}</h1>
    {#if taxonomy.children.length > 0}
      <section aria-labelledby="categories-heading">
        <h2 id="categories-heading">Categories</h2>
        <div class="cards">
          {#each taxonomy.children as child}
            <a class="card" href={taxonomyPath([...taxonomy.path.map(item => item.label), child.label])}>
              <span>{child.label}</span>
              <span class="card-arrow" aria-hidden="true">→</span>
            </a>
          {/each}
        </div>
      </section>
    {/if}
    <section aria-labelledby="parts-heading">
      <h2 id="parts-heading">Parts</h2>
      {#if taxonomy.partNumbers.length > 0}
        <div class="part-list">
          {#each taxonomy.partNumbers as item}
            <div class="part-row">
              <IdentityLink partNumber={item.value} />
              <span class="card-arrow" aria-hidden="true">→</span>
            </div>
          {/each}
        </div>
      {:else}
        <p class="empty-state">No parts at this level.</p>
      {/if}
    </section>
  {:else if part}
    <p class="eyebrow">Part number</p>
    <h1>{part.value}</h1>
    <nav class="taxonomy-path" aria-label="Part category">
      {#each part.taxonomyPath as label, index}
        {#if index > 0}<span class="breadcrumb-separator" aria-hidden="true">/</span>{/if}
        <a href={taxonomyPath(part.taxonomyPath.slice(0, index + 1))}>{label}</a>
      {/each}
    </nav>
    <h2>Revisions</h2>
    {#if part.revisions.length > 0}
      <div class="table-frame">
        <table class="revision-table">
          <thead>
            <tr>
              <th scope="col">Rev</th>
              <th scope="col">Status</th>
              <th scope="col">ECO/ECR</th>
              <th scope="col">Effectivity date</th>
            </tr>
          </thead>
          <tbody>
            {#each part.revisions as item}
              <tr>
                <th scope="row"><IdentityLink partNumber={part.value} revision={String(item.id)} /></th>
                <td class="unsupported" aria-label="Status not supported">—</td>
                <td class="unsupported" aria-label="ECO or ECR not supported">—</td>
                <td class="unsupported" aria-label="Effectivity date not supported">—</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}<p class="muted">No revisions.</p>{/if}
  {:else if revision}
    <p class="eyebrow">Revision {revision.id}</p>
    <h1><IdentityLink partNumber={revision.partNumber.value} revision={String(revision.id)} /></h1>
    <h2>Bill of Materials</h2>
    {#if revision.bom.length > 0}<ul>{#each revision.bom as node (node.revisionId)}<BomNode node={node} />{/each}</ul>
    {:else}<p class="muted">No BOM entries.</p>{/if}
  {/if}
</Shell>
