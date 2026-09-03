<script lang="ts">
  import CategoryTile, { type CategoryTileSize } from './CategoryTile.svelte'

  export type CategoryTileItem = {
    id: string
    title: string
    href: string
    description?: string
    image?: string
    partCount?: number
  }

  export let tiles: CategoryTileItem[] = []
  export let density: CategoryTileSize | 'auto' = 'auto'

  $: tileSize = density === 'auto'
    ? tiles.length <= 3 ? 'feature' : tiles.length <= 6 ? 'standard' : 'compact'
    : density
</script>

{#if tiles.length > 0}
  <div class:category-tile-grid-compact={tileSize === 'compact'} class:category-tile-grid-feature={tileSize === 'feature'} class="category-tile-grid">
    {#each tiles as tile (tile.id)}
      <CategoryTile title={tile.title} href={tile.href} description={tile.description} image={tile.image} partCount={tile.partCount} size={tileSize} />
    {/each}
  </div>
{:else}
  <p class="category-tile-grid-empty">No categories in this scope.</p>
{/if}

<style>
  .category-tile-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; }
  .category-tile-grid-compact { grid-template-columns: repeat(4, minmax(0, 1fr)); }
  .category-tile-grid-feature { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .category-tile-grid-empty { padding: 1rem; color: var(--muted); border: 1px dashed var(--hairline); font-size: .72rem; }
  @media (max-width: 900px) { .category-tile-grid, .category-tile-grid-compact, .category-tile-grid-feature { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  @media (max-width: 600px) { .category-tile-grid, .category-tile-grid-compact, .category-tile-grid-feature { grid-template-columns: 1fr; } }
</style>
