<script context="module" lang="ts">
  export type Breadcrumb = {
    label: string
    href?: string
  }
</script>

<script lang="ts">
  export let items: Breadcrumb[] = []
</script>

<nav class="breadcrumbs" aria-label="Breadcrumb">
  <a href="/parts/components">Home</a>
  {#each items as item, index}
    <span class="breadcrumb-separator" aria-hidden="true">/</span>
    {#if item.href && index < items.length - 1}
      <a href={item.href}>{item.label}</a>
    {:else}
      <span aria-current="page">{item.label}</span>
    {/if}
  {/each}
</nav>
