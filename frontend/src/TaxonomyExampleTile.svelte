<script context="module" lang="ts">
  export type TaxonomyTileType = 'category' | 'rule' | 'attribute'
  export type TaxonomyTile = {
    kind: 'taxonomy-tile'
    type: TaxonomyTileType
    title: string
    description: string
    color: string
  }

  export const taxonomyTileModels = {
    category: { kind: 'taxonomy-tile', type: 'category', title: 'Category', description: 'Classify this branch', color: 'var(--accent)' },
    rule: { kind: 'taxonomy-tile', type: 'rule', title: 'Rule', description: 'Match a grammar field', color: '#7655a8' },
    attribute: { kind: 'taxonomy-tile', type: 'attribute', title: 'Attribute', description: 'Assign a part property', color: '#337d91' },
  } satisfies Record<TaxonomyTileType, TaxonomyTile>

</script>

<script lang="ts">
  import { draggable } from '@thisux/sveltednd'
  import type { TaxonomyDragPayload } from './TaxonomyEditorModel'
  export let tile: TaxonomyTile
  export let variant: 'palette' | 'catalogue'
  export let container = ''
  export let dragData: TaxonomyDragPayload | undefined = undefined
  export let onDragEnd: () => void = () => {}
</script>

{#if variant === 'palette'}
  <div data-dnd-kind="new" data-dnd-container={container} class="taxonomy-palette-item" style={`--taxonomy-color: ${tile.color}`} use:draggable={{ container, dragData, callbacks: { onDragEnd } }}>
    <i aria-hidden="true"></i>
    <span><strong>{tile.title}</strong><small>{tile.description}</small></span>
  </div>
{:else}
  <div class={`taxonomy-example-tile taxonomy-example-tile-${tile.type}`} style={`--taxonomy-color: ${tile.color}`}>
    <span class="taxonomy-example-dot" aria-hidden="true"></span>
    <span><strong>{tile.title}</strong><small>{tile.description}</small></span>
  </div>
{/if}
