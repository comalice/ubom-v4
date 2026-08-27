<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
</script>

<p class="eyebrow">Taxonomy editor components</p>
<h2>Taxonomy editor components</h2>
<p class="taxonomy-blocks-intro">Hard-coded sketches for the blocks and drop targets that will make up the taxonomy editor.</p>

<section class="block-examples taxonomy-grammar-example" aria-labelledby="taxonomy-grammar-heading">
  <span class="mock-section-label" id="taxonomy-grammar-heading">Grammar tree</span>
  <p class="taxonomy-component-description">The immutable grammar tree provides the structure and fields against which taxonomy nodes are authored.</p>
  <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree /></div>
</section>

<section class="block-examples taxonomy-block-examples" aria-labelledby="taxonomy-category-heading">
  <span class="mock-section-label" id="taxonomy-category-heading">Category block</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <div class="taxonomy-example-tile taxonomy-example-tile-category"><span class="taxonomy-example-dot" aria-hidden="true"></span><span><strong>Category</strong><small>Classify a grammar branch</small></span></div>
      <BlockShell title="category" summary="Resistors" kind="taxonomy" collapsible initiallyOpen={true} onRemove={() => {}}>
        <div class="taxonomy-block-body"><label>Label<input value="Resistors" aria-label="Category label" /></label><label>Node ID<input value="resistors" aria-label="Category node ID" /></label><div class="taxonomy-block-rule"><span>Rule</span><strong>category = A</strong></div><button class="taxonomy-block-add" type="button">+ add child category</button></div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">A named classification node attached to a grammar location.</p><div class="grammar-block-example"><span>Example</span><code>Resistors</code></div></div>
    </article>
  </div>
</section>

<section class="block-examples taxonomy-attribute-inspector" aria-labelledby="taxonomy-attributes-heading">
  <div class="taxonomy-inspector-title"><div><span class="mock-section-label" id="taxonomy-attributes-heading">Attributes for Resistors</span><h3>Assigned part properties</h3></div><span class="taxonomy-editor-count">selected grammar node: category</span></div>
  <p class="taxonomy-attribute-intro">Inherited assignments remain visible with their source. Add local assignments for the selected taxonomy node as new table rows.</p>
  <TableFrame className="taxonomy-attribute-table-frame">
    <table class="taxonomy-attribute-table">
      <caption class="sr-only">Attributes assigned to Resistors</caption>
      <thead><tr><th scope="col">Attribute</th><th scope="col">Value type</th><th scope="col">Required</th><th scope="col">Source</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead>
      <tbody>
        <tr><th scope="row">Footprint</th><td>enum</td><td>Required</td><td><span class="taxonomy-inherited">Components</span></td><td><PrimitiveButton variant="quiet" type="button">Override</PrimitiveButton></td></tr>
        <tr><th scope="row">Resistance value</th><td>decimal · ohm</td><td>Required</td><td>This node</td><td><PrimitiveButton variant="quiet" type="button">Remove</PrimitiveButton></td></tr>
        <tr><th scope="row">Tolerance</th><td>decimal · percent</td><td>Optional</td><td>This node</td><td><PrimitiveButton variant="quiet" type="button">Remove</PrimitiveButton></td></tr>
      </tbody>
    </table>
  </TableFrame>
  <div class="taxonomy-attribute-actions"><PrimitiveButton variant="secondary" type="button">+ Add attribute assignment</PrimitiveButton><span class="muted">Adds a new local row for Resistors.</span></div>
</section>
