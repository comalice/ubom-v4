<script lang="ts">
  import { onMount } from 'svelte'
  import { getSeqDefs, type SeqDefView } from './api'
  import AsyncState from './AsyncState.svelte'
  import SectionHeading from './SectionHeading.svelte'

  let definitions: SeqDefView[] = []
  let loading = true
  let error = ''

  onMount(async () => {
    try {
      definitions = await getSeqDefs()
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
  <a class="seq-def-new-link" href="/settings/sequence-definitions/new">New</a>
</div>

{#if loading}
  <AsyncState state="loading" message="Loading sequence definitions..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if definitions.length === 0}
  <AsyncState state="empty" message="No sequence definitions." />
{:else}
  <div class="settings-list">
    {#each definitions as definition}
      <a class="settings-item" href={`/settings/sequence-definitions/${encodeURIComponent(definition.id)}`}>
         <span><strong>{definition.id}</strong><small>{definition.definitionId} · v{definition.version}</small></span>
        <span class="card-arrow" aria-hidden="true">→</span>
      </a>
    {/each}
  </div>
{/if}
