<script context="module" lang="ts">
  export type Breadcrumb = {
    label: string
    href?: string
  }
</script>

<script lang="ts">
  import PrimitiveLink from './PrimitiveLink.svelte'

  export let items: Breadcrumb[] = []
</script>

<nav class="breadcrumbs" aria-label="Breadcrumb">
  <PrimitiveLink variant="muted" href="/">Home</PrimitiveLink>
  {#each items as item, index}
    <span class="breadcrumb-separator" aria-hidden="true">/</span>
    {#if item.href && index < items.length - 1}
      <PrimitiveLink variant="muted" href={item.href}>{item.label}</PrimitiveLink>
    {:else}
      <span aria-current="page">{item.label}</span>
    {/if}
  {/each}
</nav>
