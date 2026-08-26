  <script lang="ts">
  import { onMount } from 'svelte'
  import { getSeqDefs, type SeqDefView } from './api'
  let definitions: SeqDefView[] = []
  let error = ''
  onMount(async () => { try { definitions = await getSeqDefs() } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to load sequence definitions' } })
  const items = [
    { label: 'Sequence definitions', description: 'Create and inspect immutable sequence languages.', href: '/settings/sequence-definitions/new', available: true },
    { label: 'Taxonomies', description: 'Organize parts and assign their attributes.', href: '', available: false },
    { label: 'Part-number schemas', description: 'Connect sequence definitions to taxonomy rules.', href: '', available: false },
    { label: 'Artifact sources', description: 'Configure external sources and captured files.', href: '/artifacts', available: true },
  ]
</script>

<p class="eyebrow">Settings</p>
<h1>Settings</h1>
<p class="muted settings-intro">Rules and connections that shape the UBOM.</p>
{#if error}<p class="error" role="alert">{error}</p>{/if}
<div class="settings-list">
  {#each items as item}
    {#if item.available}
      <a class="settings-item" href={item.href}>
        <span><strong>{item.label}</strong><small>{item.description}</small></span><span class="card-arrow">→</span>
      </a>
    {:else}
      <div class="settings-item settings-item-disabled"><span><strong>{item.label}</strong><small>{item.description}</small></span><span class="muted">Soon</span></div>
    {/if}
  {/each}
</div>
{#if definitions.length}<section class="settings-definitions" aria-labelledby="sequence-definitions-heading"><h2 id="sequence-definitions-heading">Existing sequence definitions</h2>{#each definitions as definition}<a class="settings-item" href={`/settings/sequence-definitions/${encodeURIComponent(definition.id)}`}><strong>{definition.id}</strong><span class="card-arrow">→</span></a>{/each}</section>{/if}
