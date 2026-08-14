<script context="module" lang="ts">
  export type Breadcrumb = {
    label: string
    href?: string
  }
</script>

<script lang="ts">
  export let items: Breadcrumb[] = []
</script>

{#if items.length > 0}
  <nav class="breadcrumbs" aria-label="Breadcrumb">
    <a href="/taxonomies/sample-taxonomy-v1/nodes/components">Home</a>
    {#each items as item, index}
      <span class="breadcrumb-separator" aria-hidden="true">/</span>
      {#if item.href && index < items.length - 1}
        <a href={item.href}>{item.label}</a>
      {:else}
        <span aria-current="page">{item.label}</span>
      {/if}
    {/each}
  </nav>
{/if}
