<script lang="ts">
  import TaxonomyEditor from './TaxonomyEditor.svelte'
  import { grammarNodes } from './TaxonomyEditorFixtures'
  import type { TaxonomyAttributeDefinition, TaxonomyDraftNode, TaxonomyGrammar } from './TaxonomyEditorModel'

  const grammarFixtures: TaxonomyGrammar[] = [
    { id: 'sample-pn-v1', label: 'Sample part number', fields: 3, nodes: grammarNodes },
    { id: 'sample-pn-v2', label: 'Extended part number', fields: 4, nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v2', children: [
      { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
      { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [{ id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] }] },
      { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
      { id: 'revision', label: 'revision', blockType: 'binding', summary: 'choice: A · B', children: [] },
    ] }] },
    { id: 'sample-pn-v3', label: 'Regional part number', fields: 4, nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v3', children: [
      { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"UB-"', children: [] },
      { id: 'region', label: 'region', blockType: 'binding', summary: 'choice: EU · US', children: [] },
      { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [{ id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] }] },
      { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
    ] }] },
  ]

  let selectedGrammarID = grammarFixtures[0].id
  $: selectedGrammar = grammarFixtures.find((grammar) => grammar.id === selectedGrammarID) ?? grammarFixtures[0]

  const fixtureAttributeDefinitions: TaxonomyAttributeDefinition[] = [
    { id: 'footprint', label: 'Footprint', valueType: 'enum', enumValues: ['Axial', 'Radial', 'SMD'] },
    { id: 'material', label: 'Material', valueType: 'string' },
    { id: 'resistance', label: 'Resistance value', valueType: 'decimal', unit: 'ohm' },
    { id: 'tolerance', label: 'Tolerance', valueType: 'decimal', unit: 'percent' },
    { id: 'power-rating', label: 'Power rating', valueType: 'decimal', unit: 'watt' },
  ]

  export function createTaxonomyFixtureDraft(grammar: TaxonomyGrammar): TaxonomyDraftNode {
    return { id: 'taxonomy-root', label: 'Parts', targetId: String(grammar.nodes[0]?.id ?? 'grammar-root'), placements: [], rules: [], attributes: [], children: [] }
  }
</script>

<section class="taxonomy-editor-playground" aria-labelledby="taxonomy-editor-heading">
  <p class="eyebrow">Taxonomy editor</p>
  <h2 id="taxonomy-editor-heading">Taxonomy editor</h2>
  <p class="taxonomy-editor-intro">Build a classification overlay against an existing grammar. The grammar is fixed; taxonomy nodes are added at the drop zones.</p>

  <label for="taxonomy-grammar-select">Grammar<select id="taxonomy-grammar-select" bind:value={selectedGrammarID}>{#each grammarFixtures as grammar}<option value={grammar.id}>{grammar.label}</option>{/each}</select></label>
  {#key selectedGrammar.id}
    <TaxonomyEditor grammar={selectedGrammar} initialAttributeDefinitions={fixtureAttributeDefinitions} initialDraft={createTaxonomyFixtureDraft(selectedGrammar)} />
  {/key}
</section>
