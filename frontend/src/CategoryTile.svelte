<script lang="ts">
  export type CategoryTileSize = 'compact' | 'standard' | 'feature'

  export let title: string
  export let href: string
  export let size: CategoryTileSize = 'standard'
  export let description = ''
  export let partCount: number | undefined = undefined
  export let image = ''
</script>

<a class:category-tile-compact={size === 'compact'} class:category-tile-feature={size === 'feature'} class="category-tile" {href}>
  {#if image}<span class="category-tile-image" aria-hidden="true">{image}</span>{/if}
  <span class="category-tile-content">
    <strong>{title}</strong>
    {#if description}<span class="category-tile-description">{description}</span>{/if}
    {#if partCount !== undefined}<small>{partCount} {partCount === 1 ? 'part' : 'parts'}</small>{/if}
  </span>
  <span class="category-tile-arrow" aria-hidden="true">›</span>
</a>

<style>
  .category-tile { display: flex; align-items: flex-start; gap: .65rem; min-width: 0; min-height: 8rem; padding: .8rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .3rem; }
  .category-tile:hover { background: var(--accent-light); border-color: var(--accent); text-decoration: none; }
  .category-tile-content { display: flex; flex: 1; min-width: 0; min-height: 6.4rem; flex-direction: column; }
  .category-tile strong { color: var(--accent); font: 600 .78rem ui-monospace, monospace; }
  .category-tile-description { margin-top: .35rem; color: var(--muted); font-size: .7rem; line-height: 1.4; }
  .category-tile-content small { margin-top: auto; padding-top: .65rem; color: var(--muted); font: .62rem ui-monospace, monospace; }
  .category-tile-image { display: grid; flex: none; width: 3.5rem; height: 3.5rem; place-items: center; color: var(--accent); background: var(--accent-light); border: 1px solid var(--hairline); font: .58rem ui-monospace, monospace; }
  .category-tile-arrow { align-self: center; color: var(--accent); font-size: 1.1rem; line-height: 1; }
  .category-tile-compact { min-height: 5.5rem; padding: .6rem; }
  .category-tile-compact .category-tile-content { min-height: 4.3rem; }
  .category-tile-compact .category-tile-image { width: 2.5rem; height: 2.5rem; }
  .category-tile-compact strong { font-size: .7rem; }
  .category-tile-feature { min-height: 11rem; padding: 1rem; }
  .category-tile-feature .category-tile-content { min-height: 9.2rem; }
  .category-tile-feature .category-tile-image { width: 5rem; height: 5rem; }
  .category-tile-feature strong { font-size: .9rem; }
</style>
