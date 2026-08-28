<script lang="ts">
  import { onMount } from 'svelte'
  import { getSeqDefs, getTaxonomyDefs, type SeqDefView, type TaxonomyDefSummary } from './api'
  import AsyncState from './AsyncState.svelte'
  import NavigationCardList, { type NavigationCardItem } from './NavigationCardList.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import SectionHeading from './SectionHeading.svelte'

  let definitions: SeqDefView[] = []
  let taxonomies: TaxonomyDefSummary[] = []
  let loading = true
  let error = ''

  onMount(async () => {
    try {
      ;[definitions, taxonomies] = await Promise.all([getSeqDefs(), getTaxonomyDefs()])
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to load sequence definitions'
    } finally {
      loading = false
    }
  })
</script>

<div class="seq-def-list-header">
  <div>
    <SectionHeading eyebrow="Sequence definitions" title="Sequence definitions" description="Immutable languages used to describe part numbers." />
  </div>
  <PrimitiveLink className="seq-def-new-link" href="/settings/sequence-definitions/new">New</PrimitiveLink>
</div>

{#if loading}
  <AsyncState state="loading" message="Loading sequence definitions..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if definitions.length === 0}
  <AsyncState state="empty" message="No sequence definitions." />
{:else}
  {@const items: NavigationCardItem[] = definitions.map((definition) => ({
    label: definition.id,
      description: `${definition.definitionId} · v${definition.version}`,
      href: `/settings/sequence-definitions/${encodeURIComponent(definition.id)}`,
      action: taxonomies.some(taxonomy => taxonomy.seqDef === definition.id)
        ? { label: 'Edit taxonomy', href: `/settings/taxonomies/${encodeURIComponent(taxonomies.find(taxonomy => taxonomy.seqDef === definition.id)!.id)}` }
        : { label: 'New taxonomy', href: `/settings/taxonomies/new?seqDef=${encodeURIComponent(definition.id)}` },
  }))}
  <NavigationCardList {items} arrowAriaHidden />
{/if}
