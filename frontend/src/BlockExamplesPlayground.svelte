<script lang="ts">
  import PlaygroundBasicBlock from './PlaygroundBasicBlock.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'

  const withoutRemove = { id: 101, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
  const withRemove = { id: 102, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
  const tiles = [
    { kind: 'sequence-tile' as const, type: 'literal', title: 'Literal', description: 'Fixed text value' },
    { kind: 'sequence-tile' as const, type: 'select', title: 'Select', description: 'Choose one of several values' },
    { kind: 'sequence-tile' as const, type: 'range', title: 'Pattern', description: 'Generate a formatted value' },
  ]
</script>

<p class="eyebrow">Block examples</p>
<h2>Block examples</h2>
<section class="block-examples" aria-labelledby="optional-remove-heading">
  <div class="block-examples-heading">
    <span class="mock-section-label" id="optional-remove-heading">Optional remove control</span>
    <p>Compare the same block with and without the optional upper-right control.</p>
  </div>
  <div class="block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Without x</span>
      <PlaygroundBasicBlock item={withoutRemove} parentId={null} container="component-list:examples" title="Literal" />
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">With x</span>
      <PlaygroundBasicBlock item={withRemove} parentId={null} container="component-list:examples" title="Literal" onRemove={() => {}} />
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="custom-copy-heading">
  <div class="block-examples-heading">
    <span class="mock-section-label" id="custom-copy-heading">Custom block copy</span>
    <p>Titles can be supplied independently for each block; descriptions belong to palette tiles.</p>
  </div>
  <div class="block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Short title</span>
      <PlaygroundBasicBlock item={withoutRemove} parentId={null} container="component-list:examples" title="Value" />
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">Long title</span>
      <PlaygroundBasicBlock item={withRemove} parentId={null} container="component-list:examples" title="Reference value" />
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="sequence-tiles-heading">
  <div class="block-examples-heading">
    <span class="mock-section-label" id="sequence-tiles-heading">Sequence tiles</span>
    <p>Descriptions belong to the draggable tile and are not carried into a dropped block.</p>
  </div>
  <div class="block-example-tile-grid">
    {#each tiles as tile}<SequenceExampleTile tile={tile} variant="palette" container="sequence-example:palette" />{/each}
  </div>
</section>
