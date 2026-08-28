<script lang="ts">
  import PrimitiveButton from './PrimitiveButton.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyExampleTile, { taxonomyTileModels } from './TaxonomyExampleTile.svelte'
  import { taxonomyCategoryExample, taxonomyRuleExample } from './TaxonomyEditorFixtures'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'
  import TaxonomyAttributeTableRow from './TaxonomyAttributeTableRow.svelte'
  import type { TaxonomyEffectiveAttribute } from './TaxonomyEditorModel'

  const catalogueTiles = [
    { ...taxonomyTileModels.category, description: 'Classify a grammar branch' },
    taxonomyTileModels.rule,
  ]
  const attributeRows: TaxonomyEffectiveAttribute[] = [
    { id: 'material', label: 'Material', valueType: 'string', required: false, source: 'Components', overridden: false },
    { id: 'footprint', label: 'Footprint', valueType: 'enum', required: false, source: 'Resistors', overridden: true },
    { id: 'resistance', label: 'Resistance value', valueType: 'decimal', unit: 'ohm', required: true, source: 'Resistors', overridden: false },
    { id: 'tolerance', label: 'Tolerance', valueType: 'decimal', unit: 'percent', required: false, source: 'Resistors', overridden: false },
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
        <TaxonomyCategoryBlock title="Resistors" summary="category = RES" label="Resistors" nodeId="resistors" predicate="category = RES" showRuleDropZone showChildren addChildLabel="Drop a category here">
          <svelte:fragment slot="children">
            <TaxonomyCategoryBlock title="High Power" summary="number = 100..199" label="High Power" nodeId="high-power" predicate="number = 100..199" showChildren addChildLabel="Drop a category here" />
          </svelte:fragment>
        </TaxonomyCategoryBlock>
        <div class="grammar-block-meta"><p class="grammar-block-description">{taxonomyCategoryExample.description}</p><div class="grammar-block-example"><span>Example</span><code>{taxonomyCategoryExample.example}</code></div></div>
      </article>
      <article class="grammar-basic-example">
        <TaxonomyExampleTile tile={catalogueTiles[1]} variant="catalogue" />
        <TaxonomyRuleBlock title={taxonomyRuleExample.title} field="category" predicate={taxonomyRuleExample.predicate} />
        <div class="grammar-block-meta"><p class="grammar-block-description">{taxonomyRuleExample.description}</p><div class="grammar-block-example"><span>Example</span><code>{taxonomyRuleExample.example}</code></div></div>
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
        <TaxonomyAttributeTableRow attribute={attributeRows[0]} selectedNodeLabel="Resistors" actionLabel="Override" />
        <TaxonomyAttributeTableRow attribute={attributeRows[1]} selectedNodeLabel="Resistors" actionLabel="Remove" />
        <TaxonomyAttributeTableRow attribute={attributeRows[2]} selectedNodeLabel="Resistors" actionLabel="Remove" />
        <TaxonomyAttributeTableRow attribute={attributeRows[3]} selectedNodeLabel="Resistors" actionLabel="Remove" />
      </tbody>
    </table>
  </TableFrame>
  <div class="taxonomy-attribute-actions"><PrimitiveButton variant="secondary" type="button">+ Add attribute assignment</PrimitiveButton><span class="muted">Adds a new local row for Resistors.</span></div>
</section>
