<script lang="ts">
  import { onMount } from 'svelte'
  import { getSeqDefs, type SeqDefView } from './api'

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
    <p class="eyebrow">Sequence definitions</p>
    <h1>Sequence definitions</h1>
    <p class="muted settings-intro">Immutable languages used to describe part numbers.</p>
  </div>
  <a class="seq-def-new-link" href="/settings/sequence-definitions/new">New</a>
</div>

{#if loading}
  <p class="muted">Loading sequence definitions...</p>
{:else if error}
  <p class="error" role="alert">{error}</p>
{:else if definitions.length === 0}
  <p class="empty-state">No sequence definitions.</p>
{:else}
  <div class="settings-list">
    {#each definitions as definition}
      <a class="settings-item" href={`/settings/sequence-definitions/${encodeURIComponent(definition.id)}`}>
        <strong>{definition.id}</strong>
        <span class="card-arrow" aria-hidden="true">→</span>
      </a>
    {/each}
  </div>
{/if}
