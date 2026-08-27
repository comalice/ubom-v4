<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyExampleTile, { taxonomyTileModels } from './TaxonomyExampleTile.svelte'
  import { taxonomyCategoryExample, taxonomyRuleExample } from './TaxonomyEditorFixtures'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'

  const catalogueTiles = [
    { ...taxonomyTileModels.category, description: 'Classify a grammar branch' },
    taxonomyTileModels.rule,
    taxonomyTileModels.attribute,
  ]
</script>

<p class="eyebrow">Taxonomy editor components</p>
<h2>Taxonomy editor components</h2>
<p class="taxonomy-blocks-intro">Hard-coded sketches for the taxonomy nodes, predicates, and attribute assignments that will make up the editor.</p>

<section class="block-examples taxonomy-grammar-example" aria-labelledby="taxonomy-grammar-heading">
  <span class="mock-section-label" id="taxonomy-grammar-heading">Grammar tree</span>
  <p class="taxonomy-component-description">The immutable grammar tree provides the structure and fields against which taxonomy nodes are authored.</p>
  <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree /></div>
</section>

<section class="block-examples taxonomy-block-examples" aria-labelledby="taxonomy-category-heading">
  <span class="mock-section-label" id="taxonomy-category-heading">Category block</span>
  <div class="grammar-basic-example-list">
      <article class="grammar-basic-example">
        <TaxonomyExampleTile tile={catalogueTiles[0]} variant="catalogue" />
        <TaxonomyCategoryBlock title="Resistors" summary="category = RES" label="Resistors" nodeId="resistors" predicate="category = RES" showChildren addChildLabel="Drop a category here">
          <svelte:fragment slot="children">
            <TaxonomyCategoryBlock title="High Power" summary="number = 100..199" label="High Power" nodeId="high-power" predicate="number = 100..199" showChildren addChildLabel="Drop a category here" />
          </svelte:fragment>
        </TaxonomyCategoryBlock>
        <div class="grammar-block-meta"><p class="grammar-block-description">{taxonomyCategoryExample.description}</p><div class="grammar-block-example"><span>Example</span><code>{taxonomyCategoryExample.example}</code></div></div>
      </article>
      <article class="grammar-basic-example">
        <TaxonomyExampleTile tile={catalogueTiles[1]} variant="catalogue" />
        <TaxonomyRuleBlock title={taxonomyRuleExample.title} summary={taxonomyRuleExample.summary} predicate={taxonomyRuleExample.predicate} />
        <div class="grammar-block-meta"><p class="grammar-block-description">{taxonomyRuleExample.description}</p><div class="grammar-block-example"><span>Example</span><code>{taxonomyRuleExample.example}</code></div></div>
      </article>
      <article class="grammar-basic-example">
        <TaxonomyExampleTile tile={catalogueTiles[2]} variant="catalogue" />
        <BlockShell title="attribute" summary="Resistance value" kind="taxonomy" collapsible initiallyOpen={true} onRemove={() => {}}>
          <div class="taxonomy-block-body"><label>Attribute<select aria-label="Assigned attribute"><option>Resistance value</option><option>Tolerance</option><option>Footprint</option></select></label><label>Value<input value="10" aria-label="Attribute value" /></label><label class="taxonomy-checkbox"><input type="checkbox" checked aria-label="Attribute required" /> Required</label><div class="taxonomy-block-rule"><span>Type</span><strong>decimal · ohm</strong></div></div>
        </BlockShell>
        <div class="grammar-block-meta"><p class="grammar-block-description">A domain property assigned to every part classified by this taxonomy node.</p><div class="grammar-block-example"><span>Example</span><code>10 ohm</code></div></div>
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
