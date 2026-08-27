<script lang="ts">
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
</script>

<section class="taxonomy-editor" aria-label="Taxonomy editor layout sketch">
  <TaxonomyPalette />

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
