<script lang="ts">
  import TaxonomyExampleTile, { taxonomyTileModels } from './TaxonomyExampleTile.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'

  const paletteItems = Object.values(taxonomyTileModels)
</script>

<section class="taxonomy-editor-playground" aria-labelledby="taxonomy-editor-heading">
  <p class="eyebrow">Taxonomy editor</p>
  <h2 id="taxonomy-editor-heading">Taxonomy editor</h2>
  <p class="taxonomy-editor-intro">Build a classification overlay against an existing grammar. The grammar is fixed; taxonomy nodes are added at the drop zones.</p>

  <section class="taxonomy-editor" aria-label="Taxonomy editor layout sketch">
    <div class="taxonomy-editor-palette">
        <span class="mock-section-label">Taxonomy palette</span>
      <div class="taxonomy-editor-palette-grid">
        {#each paletteItems as item}
          <TaxonomyExampleTile tile={item} variant="palette" />
        {/each}
      </div>
      <p class="taxonomy-palette-help">Drag taxonomy blocks onto the grammar. The grammar remains an immutable map of available fields.</p>
    </div>

    <section class="taxonomy-editor-canvas" aria-labelledby="taxonomy-canvas-heading">
      <div class="taxonomy-editor-heading">
        <div><span class="mock-section-label" id="taxonomy-canvas-heading">Grammar and taxonomy overlay</span><h3>Part number grammar</h3></div>
        <span class="taxonomy-editor-count">3 fields · local draft</span>
      </div>
      <p class="taxonomy-grammar-note">Read-only grammar <span aria-hidden="true">·</span> taxonomy drop zones are shown below each grammar node</p>

      <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree /></div>
    <div class="taxonomy-tree-drop-zone">Drop a category here to add a sibling taxonomy branch</div>
    </section>

    <section class="taxonomy-editor-status" aria-labelledby="taxonomy-status-heading">
      <div class="taxonomy-editor-heading"><span class="mock-section-label" id="taxonomy-status-heading">Definition status</span><span class="taxonomy-editor-count">local draft</span></div>
      <div class="taxonomy-status-grid">
        <span class="taxonomy-status-ok">✓</span><span>Grammar is immutable</span>
        <span class="taxonomy-status-ok">✓</span><span>Drop zones expose known grammar fields</span>
        <span class="taxonomy-status-pending">•</span><span>Taxonomy validation will come later</span>
      </div>
    </section>
  </section>
</section>
