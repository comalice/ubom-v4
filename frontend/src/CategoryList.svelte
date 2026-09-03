<script lang="ts">
  export type CategoryListItem = {
    id: string
    label: string
    href: string
    count?: number
    selected?: boolean
  }

  export let categories: CategoryListItem[] = []
</script>

<nav class="category-list" aria-label="Catalog categories">
  {#each categories as category}
    <a class:selected={category.selected} data-category-id={category.id} href={category.href} aria-current={category.selected ? 'page' : undefined}>
      <span>{category.label}</span>
      {#if category.count !== undefined}<small>{category.count}</small>{/if}
      <span class="category-list-arrow" aria-hidden="true">›</span>
    </a>
  {/each}
</nav>

<style>
  .category-list { display: grid; gap: .1rem; }
  .category-list a { display: grid; grid-template-columns: minmax(0, 1fr) auto auto; align-items: center; gap: .5rem; padding: .55rem .6rem; color: var(--muted); border-left: 2px solid transparent; font-size: .75rem; }
  .category-list a:hover { color: var(--ink); background: var(--accent-light); text-decoration: none; }
  .category-list a.selected { color: var(--ink); background: var(--accent-light); border-left-color: var(--accent); }
  .category-list small { color: var(--muted); font: .6rem ui-monospace, monospace; }
  .category-list-arrow { color: var(--accent); font-size: 1rem; line-height: 1; }
</style>
