<script lang="ts">
  export interface NavigationCardItem {
    label: string
    description: string
    href?: string
    status?: string
  }

  export let items: NavigationCardItem[]
  export let arrowAriaHidden = false
</script>

<div class="settings-list">
  {#each items as item}
    {#if item.href}
      <a class="settings-item" href={item.href}>
        <span><strong>{item.label}</strong><small>{item.description}</small></span><span class="card-arrow" aria-hidden={arrowAriaHidden}>→</span>
      </a>
    {:else}
      <div class="settings-item settings-item-disabled"><span><strong>{item.label}</strong><small>{item.description}</small></span>{#if item.status}<span class="muted">{item.status}</span>{/if}</div>
    {/if}
  {/each}
</div>
