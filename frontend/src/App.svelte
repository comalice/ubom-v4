<script lang="ts">
  import { onMount } from 'svelte'
  import PartTable from './PartTable.svelte'
  import BomTable from './BomTable.svelte'
  import ArtifactTable from './ArtifactTable.svelte'
  import Breadcrumbs, { type Breadcrumb } from './Breadcrumbs.svelte'
  import IdentityLink from './IdentityLink.svelte'
  import PartIdentity from './PartIdentity.svelte'
  import Shell from './Shell.svelte'
  import Authoring from './Authoring.svelte'
  import Settings from './Settings.svelte'
  import GrammarEditor from './GrammarEditor.svelte'
  import SeqDefEditorPage from './SeqDefEditorPage.svelte'
  import Playground from './Playground.svelte'
  import { getArtifact, getArtifacts, getPart, getParts, getRevision, getSnapshots, getSources, getTaxonomyNodeByPath, type ArtifactView, type PartListItem, type PartNumberView, type RevisionView, type SnapshotView, type SourceView, type TaxonomyNodeView } from './api'
  import { parseRoute, partPath, revisionPath, taxonomyPath, type Route } from './routes'

  let route: Route = parseRoute()
  let taxonomy: TaxonomyNodeView | null = null
  let partsList: PartListItem[] = []
  let part: PartNumberView | null = null
  let revision: RevisionView | null = null
  let artifacts: ArtifactView[] = []
  let sources: SourceView[] = []
  let artifact: ArtifactView | null = null
  let snapshots: SnapshotView[] = []
  let error = ''
  let loading = true

  $: breadcrumbs = route.kind === 'parts'
    ? [{ label: 'Parts' }]
    : route.kind === 'artifacts'
      ? [{ label: 'Artifacts' }]
      : route.kind === 'artifact' && artifact
        ? [{ label: 'Artifacts', href: '/artifacts' }, { label: artifact.name }]
    : route.kind === 'taxonomy' && taxonomy
    ? taxonomyBreadcrumbs(taxonomy)
    : route.kind === 'part' && part
      ? [{ label: 'Parts' }, { label: part.value }]
      : route.kind === 'revision' && revision
        ? [
            { label: revision.partNumber.value, href: partPath(revision.partNumber.value) },
            { label: `Revision ${revision.revision}` },
          ]
      : route.kind === 'grammar-editor'
        ? [{ label: 'Settings', href: '/settings' }, { label: 'Grammar editor' }]
      : (route.kind === 'seq-def-new' || route.kind === 'seq-def')
        ? [{ label: 'Settings', href: '/settings' }, { label: 'Sequence definitions' }]
      : route.kind === 'playground'
        ? [{ label: 'Playground' }]
        : route.kind === 'settings'
          ? [{ label: 'Settings' }]
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
    partsList = []
    part = null
    revision = null
    artifacts = []
    sources = []
    artifact = null
    snapshots = []
    error = ''
    loading = true
    try {
      if (route.kind === 'parts') partsList = await getParts()
      else if (route.kind === 'artifacts') [artifacts, sources] = await Promise.all([getArtifacts(), getSources()])
      else if (route.kind === 'artifact') [artifact, snapshots] = await Promise.all([getArtifact(route.id), getSnapshots(route.id)])
      else if (route.kind === 'taxonomy') taxonomy = await getTaxonomyNodeByPath(route.segments)
      else if (route.kind === 'part') part = await getPart(route.id)
      else if (route.kind === 'revision') revision = await getRevision(route.id)
        else if (route.kind === 'settings' || route.kind === 'grammar-editor' || route.kind === 'seq-def-new' || route.kind === 'seq-def' || route.kind === 'playground') return
      else error = 'Page not found'
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to load page'
    } finally {
      loading = false
    }
  }

  function navigate(path: string) {
    history.pushState({}, '', path)
    load()
  }

  onMount(() => {
    load()
    const handlePopState = () => load()
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  })
</script>

<Shell fullWidth={route.kind === 'parts' || route.kind === 'taxonomy' || route.kind === 'artifacts'}>
  <Breadcrumbs items={breadcrumbs} />
  {#if loading}
    <p class="muted">Loading...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if route.kind === 'parts'}
    <p class="eyebrow">Parts</p>
    <h1>All parts</h1>
    {#if partsList.length > 0}<PartTable parts={partsList} />{:else}<p class="empty-state">No parts.</p>{/if}
  {:else if route.kind === 'artifacts'}
    <p class="eyebrow">Artifact register</p>
    <h1>Sources and artifacts</h1>
    {#if sources.length > 0}<div class="artifact-sources">{#each sources as source}<span><strong>{source.name}</strong> <span class="muted">{source.kind}</span></span>{/each}</div>{/if}
    {#if artifacts.length > 0}<ArtifactTable artifacts={artifacts} />{:else}<p class="empty-state">No artifacts.</p>{/if}
  {:else if artifact}
    <p class="eyebrow">Artifact provenance</p>
    <h1>{artifact.name}</h1>
    <p class="artifact-meta">{artifact.kind} from <span class="artifact-code">{artifact.sourceId}</span> at <span class="artifact-code">{artifact.path}</span></p>
    <h2>Snapshots</h2>
    {#if snapshots.length > 0}
      <div class="table-frame"><table class="revision-table"><thead><tr><th>Reference</th><th>Digest</th><th>Size</th><th>Media type</th></tr></thead><tbody>{#each snapshots as snapshot}<tr><th scope="row">{snapshot.reference || 'Unspecified'}</th><td class="artifact-code">{snapshot.digest}</td><td>{snapshot.size} bytes</td><td>{snapshot.mediaType}</td></tr>{/each}</tbody></table></div>
    {:else}<p class="empty-state">No snapshots captured.</p>{/if}
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
      <Authoring taxonomy={taxonomy} />
      <h2 id="parts-heading">Parts</h2>
      {#if taxonomy.partNumbers.length > 0}
        <PartTable parts={taxonomy.partNumbers} columns={taxonomy.attributes} />
      {:else}
        <p class="empty-state">No parts at this level.</p>
      {/if}
    </section>
  {:else if part}
    <p class="eyebrow">Part number</p>
    <PartIdentity partNumber={part.value} taxonomy={part.taxonomyPath} attributes={part.attributes} />
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
                <th scope="row"><IdentityLink partNumber={part.value} revision={item.revision} revisionID={item.id} /></th>
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
    <p class="eyebrow">Revision {revision.revision}</p>
    <PartIdentity partNumber={revision.partNumber.value} revision={revision.revision} revisionID={revision.id} taxonomy={revision.taxonomyPath} />
    <h2>Bill of Materials</h2>
    {#if revision.bom.length > 0}<BomTable nodes={revision.bom} />
    {:else}<p class="muted">No BOM entries.</p>{/if}
  {:else if route.kind === 'grammar-editor'}
    <GrammarEditor />
  {:else if route.kind === 'seq-def-new'}
    <SeqDefEditorPage />
  {:else if route.kind === 'seq-def'}
    <SeqDefEditorPage id={route.id} />
  {:else if route.kind === 'playground'}
    <Playground navigate={navigate} page={route.page} />
  {:else if route.kind === 'settings'}
    <Settings />
  {/if}
</Shell>
