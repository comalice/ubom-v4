<script lang="ts">
  import { onMount } from 'svelte'
  import { getSeqDefs, getTaxonomyDefs, type SeqDefView, type TaxonomyDefSummary } from './api'
  import AsyncState from './AsyncState.svelte'
  import NavigationCardList, { type NavigationCardItem } from './NavigationCardList.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import SectionHeading from './SectionHeading.svelte'
  import { currentTaxonomySummaries } from './taxonomySelection'

  let taxonomies: TaxonomyDefSummary[] = []
  let definitions: SeqDefView[] = []
  let loading = true
  let error = ''
  $: currentTaxonomies = currentTaxonomySummaries(taxonomies)

  onMount(async () => {
    try {
      ;[taxonomies, definitions] = await Promise.all([getTaxonomyDefs(), getSeqDefs()])
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to load taxonomies'
    } finally {
      loading = false
    }
  })

  function definitionFor(taxonomy: TaxonomyDefSummary) {
    return definitions.find(definition => definition.id === taxonomy.seqDef)
  }
</script>

<div class="seq-def-list-header">
  <SectionHeading eyebrow="Taxonomies" title="Taxonomies" description="Organize parts and assign their attributes." />
  <PrimitiveLink className="seq-def-new-link" href="/settings/taxonomies/new">New taxonomy</PrimitiveLink>
</div>

{#if loading}
  <AsyncState state="loading" message="Loading taxonomies..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if currentTaxonomies.length === 0}
  <AsyncState state="empty" message="No taxonomies." />
{:else}
  {@const items: NavigationCardItem[] = currentTaxonomies.map((taxonomy) => {
    const definition = definitionFor(taxonomy)
    return {
      label: taxonomy.id,
      description: definition
        ? `${definition.definitionId} · sequence v${definition.version} · taxonomy v${taxonomy.version}`
        : `Sequence definition: ${taxonomy.seqDef} · taxonomy v${taxonomy.version}`,
      href: `/settings/taxonomies/${encodeURIComponent(taxonomy.id)}`,
      action: { label: 'Edit taxonomy', href: `/settings/taxonomies/${encodeURIComponent(taxonomy.id)}` },
    }
  })}
  <NavigationCardList {items} arrowAriaHidden />
{/if}
