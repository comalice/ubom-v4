<script lang="ts">
  const items = [
    { label: 'Sequence generator mock', description: 'Explore the editor interaction before backend integration.', href: '/settings/sequence-definitions/mock', available: true },
    { label: 'Grammar workbench mock', description: 'Explore the underlying sequence language one primitive at a time.', href: '/settings/sequence-definitions/grammar-mock', available: true },
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
