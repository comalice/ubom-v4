<script lang="ts">
  import { onMount } from 'svelte'
  import PartTable from './PartTable.svelte'
  import BomTable from './BomTable.svelte'
  import ArtifactTable from './ArtifactTable.svelte'
  import SnapshotTable from './SnapshotTable.svelte'
  import RevisionTable from './RevisionTable.svelte'
  import Breadcrumbs, { type Breadcrumb } from './Breadcrumbs.svelte'
  import PartIdentity from './PartIdentity.svelte'
  import Shell from './Shell.svelte'
  import Authoring from './Authoring.svelte'
  import Settings from './Settings.svelte'
  import SeqDefList from './SeqDefList.svelte'
  import GrammarEditor from './GrammarEditor.svelte'
  import SeqDefEditorPage from './SeqDefEditorPage.svelte'
  import TaxonomyList from './TaxonomyList.svelte'
  import TaxonomyEditorPage from './TaxonomyEditorPage.svelte'
  import AsyncState from './AsyncState.svelte'
  import Playground from './Playground.svelte'
  import TaxonomyCardGrid from './TaxonomyCardGrid.svelte'
  import Home from './Home.svelte'
  import PartLog from './PartLog.svelte'
  import { getArtifact, getArtifacts, getCurrentTaxonomyDef, getPart, getRevision, getSnapshots, getSources, getTaxonomyDefs, getTaxonomyNodeByPath, searchParts, type ArtifactView, type PartListItem, type PartNumberView, type RevisionView, type SnapshotView, type SourceView, type TaxonomyNodeView } from './api'
  import { parseRoute, partPath, revisionPath, taxonomyPath, type Route } from './routes'

  let route: Route = parseRoute()
  let taxonomy: TaxonomyNodeView | null = null
  let taxonomySeqDef = ''
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
       : (route.kind === 'seq-def-list' || route.kind === 'seq-def-new' || route.kind === 'seq-def')
         ? [{ label: 'Settings', href: '/settings' }, { label: 'Sequence definitions' }]
       : (route.kind === 'taxonomy-list' || route.kind === 'taxonomy-new' || route.kind === 'taxonomy-detail')
         ? [{ label: 'Settings', href: '/settings' }, { label: 'Taxonomies' }]
      : route.kind === 'playground'
        ? [{ label: 'Playground' }]
        : route.kind === 'settings'
          ? [{ label: 'Settings' }]
        : []

  function taxonomyBreadcrumbs(view: TaxonomyNodeView): Breadcrumb[] {
    return view.path.map((item, index) => ({
      label: item.label,
      href: index < view.path.length - 1 ? taxonomyPath(view.path.slice(0, index + 1).map(pathItem => pathItem.label), taxonomySeqDef) : undefined,
    }))
  }

  async function load() {
    route = parseRoute()
    taxonomy = null
    part = null
    revision = null
    artifacts = []
    sources = []
    artifact = null
    snapshots = []
    error = ''
    loading = true
    try {
      if (route.kind === 'parts' || route.kind === 'home') return
      else if (route.kind === 'artifacts') [artifacts, sources] = await Promise.all([getArtifacts(), getSources()])
      else if (route.kind === 'artifact') [artifact, snapshots] = await Promise.all([getArtifact(route.id), getSnapshots(route.id)])
      else if (route.kind === 'taxonomy') {
        // URLs may omit the context for existing bookmarks; choose a current taxonomy explicitly here.
        taxonomySeqDef = route.seqDef ?? (await getTaxonomyDefs()).find(summary => summary.current)?.seqDef ?? ''
        if (!taxonomySeqDef) throw new Error('No current taxonomy is available')
        const currentTaxonomy = await getCurrentTaxonomyDef(taxonomySeqDef)
        taxonomy = await getTaxonomyNodeByPath(currentTaxonomy.id, route.segments)
        const matching = await searchParts({ taxonomyDef: currentTaxonomy.id, taxonomyNode: taxonomy.id, pageSize: 50 })
        taxonomy = { ...taxonomy, partNumbers: matching.items }
      }
      else if (route.kind === 'part') part = await getPart(route.id)
      else if (route.kind === 'revision') revision = await getRevision(route.id)
        else if (route.kind === 'settings' || route.kind === 'grammar-editor' || route.kind === 'seq-def-list' || route.kind === 'seq-def-new' || route.kind === 'seq-def' || route.kind === 'taxonomy-list' || route.kind === 'taxonomy-new' || route.kind === 'taxonomy-detail' || route.kind === 'playground') return
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
    <AsyncState state="loading" message="Loading..." />
  {:else if error}
    <AsyncState state="error" message={error} />
  {:else if route.kind === 'home'}
    <Home />
  {:else if route.kind === 'parts'}
    <PartLog />
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
      <SnapshotTable snapshots={snapshots} />
    {:else}<p class="empty-state">No snapshots captured.</p>{/if}
  {:else if taxonomy && route.kind === 'taxonomy'}
    <p class="eyebrow">Parts</p>
    <h1>{taxonomy.label}</h1>
    <TaxonomyCardGrid taxonomyPath={taxonomy.path} childCategories={taxonomy.children} seqDef={taxonomySeqDef} />
    <section aria-labelledby="parts-heading">
      <Authoring taxonomy={taxonomy} />
      <h2 id="parts-heading">Parts</h2>
      {#if taxonomy.partNumbers.length > 0}
        <PartTable parts={taxonomy.partNumbers} columns={taxonomy.attributes} seqDef={taxonomySeqDef} />
      {:else}
        <p class="empty-state">No parts at this level.</p>
      {/if}
    </section>
  {:else if part}
    <p class="eyebrow">Part number</p>
    <PartIdentity partNumber={part.value} taxonomy={part.taxonomyPath} attributes={part.attributes} />
    <h2>Revisions</h2>
    {#if part.revisions.length > 0}
      <RevisionTable partNumber={part.value} revisions={part.revisions} />
    {:else}<p class="muted">No revisions.</p>{/if}
  {:else if revision}
    <p class="eyebrow">Revision {revision.revision}</p>
    <PartIdentity partNumber={revision.partNumber.value} revision={revision.revision} revisionID={revision.id} taxonomy={revision.taxonomyPath} />
    <h2>Bill of Materials</h2>
    {#if revision.bom.length > 0}<BomTable nodes={revision.bom} />
    {:else}<p class="muted">No BOM entries.</p>{/if}
  {:else if route.kind === 'grammar-editor'}
    <GrammarEditor />
  {:else if route.kind === 'seq-def-list'}
    <SeqDefList />
  {:else if route.kind === 'seq-def-new'}
    <SeqDefEditorPage />
  {:else if route.kind === 'seq-def'}
    <SeqDefEditorPage id={route.id} />
  {:else if route.kind === 'taxonomy-list'}
    <TaxonomyList />
  {:else if route.kind === 'taxonomy-new'}
    <TaxonomyEditorPage seqDef={route.seqDef} />
  {:else if route.kind === 'taxonomy-detail'}
    <TaxonomyEditorPage id={route.id} />
  {:else if route.kind === 'playground'}
    <Playground navigate={navigate} page={route.page} />
  {:else if route.kind === 'settings'}
    <Settings />
  {/if}
</Shell>
