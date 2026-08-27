<script lang="ts">
  import TaxonomyExampleTile, { taxonomyTileModels, type TaxonomyTile } from './TaxonomyExampleTile.svelte'
  import { taxonomyEditorContainer, type TaxonomyDragPayload } from './TaxonomyEditorModel'

  const paletteItems: TaxonomyTile[] = Object.values(taxonomyTileModels)
  export let onDragEnd: () => void = () => {}
</script>

<div class="taxonomy-editor-palette">
  <span class="mock-section-label">Taxonomy palette</span>
  <div class="taxonomy-editor-palette-grid">
    {#each paletteItems as item}
      <TaxonomyExampleTile tile={item} variant="palette" container={taxonomyEditorContainer} dragData={{ kind: 'palette', taxonomyType: item.type } satisfies TaxonomyDragPayload} {onDragEnd} />
    {/each}
  </div>
  <p class="taxonomy-palette-help">Drag taxonomy blocks onto the grammar. The grammar remains an immutable map of available fields.</p>
</div>
