  <script lang="ts">
  const items = [
    { label: 'Sequence definitions', description: 'Create and inspect immutable sequence languages.', href: '/settings/sequence-definitions', available: true },
    { label: 'Taxonomies', description: 'Organize parts and assign their attributes.', href: '', available: false },
    { label: 'Part-number schemas', description: 'Connect sequence definitions to taxonomy rules.', href: '', available: false },
    { label: 'Artifact sources', description: 'Configure external sources and captured files.', href: '/artifacts', available: true },
  ]
</script>

<p class="eyebrow">Settings</p>
<h1>Settings</h1>
<p class="muted settings-intro">Rules and connections that shape the UBOM.</p>
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
