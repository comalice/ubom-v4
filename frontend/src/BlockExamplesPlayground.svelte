<script lang="ts">
  import PlaygroundBasicBlock from './PlaygroundBasicBlock.svelte'
  import PlaygroundExpandableBlock from './PlaygroundExpandableBlock.svelte'
  import PlaygroundNestedBlock from './PlaygroundNestedBlock.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'

  const withoutRemove = { id: 101, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
  const withRemove = { id: 102, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
  const nestedExample = { id: 103, kind: 'component' as const, label: 'Container', blockType: 'nested' as const, children: [] }
  const nestedWithChildren = {
    id: 104,
    kind: 'component' as const,
    label: 'Container',
    blockType: 'nested' as const,
    children: [
      { id: 105, kind: 'component' as const, label: 'Literal A', blockType: 'basic' as const, children: [] },
      { id: 106, kind: 'component' as const, label: 'Literal B', blockType: 'basic' as const, children: [] },
      { id: 107, kind: 'component' as const, label: 'Literal C', blockType: 'basic' as const, children: [] },
    ],
  }
  const tiles = [
    { kind: 'sequence-tile' as const, type: 'literal', title: 'Literal', description: 'Fixed text value' },
    { kind: 'sequence-tile' as const, type: 'select', title: 'Select', description: 'Choose one of several values' },
    { kind: 'sequence-tile' as const, type: 'range', title: 'Pattern', description: 'Generate a formatted value' },
  ]
</script>

<p class="eyebrow">Block examples</p>
<h2>Block examples</h2>
<section class="block-example-part" aria-labelledby="basic-block-heading">
  <h3 id="basic-block-heading">Basic block</h3>
<section class="block-examples" aria-labelledby="draggable-tile-components-heading">
  <div class="block-examples-heading">
    <span class="mock-section-label" id="draggable-tile-components-heading">Draggable tile components</span>
    <p>Descriptions belong to the draggable tile and are not carried into a dropped block.</p>
  </div>
  <div class="block-example-tile-grid">
    {#each tiles as tile}<SequenceExampleTile tile={tile} variant="palette" container="sequence-example:palette" />{/each}
  </div>
</section>
<section class="block-examples" aria-labelledby="block-view-sizes-heading">
  <span class="mock-section-label" id="block-view-sizes-heading">Block view sizes</span>
  <div class="block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Normal</span>
      <PlaygroundBasicBlock item={withoutRemove} parentId={null} container="component-list:examples" title="Literal" />
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">Small</span>
      <PlaygroundBasicBlock item={withRemove} parentId={null} container="component-list:examples" title="Literal" size="small" />
    </article>
  </div>
</section>
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
<section class="block-examples" aria-labelledby="expand-collapse-heading">
  <div class="block-examples-heading">
    <span class="mock-section-label" id="expand-collapse-heading">Expand / collapse</span>
    <p>The top-right control toggles an empty body while keeping the header distinct.</p>
  </div>
  <div class="block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Expanded</span>
      <PlaygroundExpandableBlock title="Literal" />
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">Collapsed</span>
      <PlaygroundExpandableBlock title="Literal" initiallyOpen={false} />
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
</section>
<section class="block-example-part" aria-labelledby="nested-block-heading">
  <h3 id="nested-block-heading">Nested block</h3>
<section class="block-examples" aria-labelledby="nested-blocks-heading">
  <span class="mock-section-label" id="nested-blocks-heading">Nested blocks</span>
  <div class="block-example-grid nested-block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Container with empty child list</span>
      <PlaygroundNestedBlock item={nestedExample} parentId={null} container="component-list:examples"><span class="nested-block-example-empty">Empty nested component list</span></PlaygroundNestedBlock>
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">Container with three sub blocks</span>
      <PlaygroundNestedBlock item={nestedWithChildren} parentId={null} container="component-list:examples">
        <div class="nested-block-example-children">
          {#each nestedWithChildren.children as child}<PlaygroundBasicBlock item={child} parentId={nestedWithChildren.id} container="component-list:104" title={child.label} />{/each}
        </div>
      </PlaygroundNestedBlock>
    </article>
  </div>
  <div class="block-examples-heading nested-block-example-heading">
    <span class="mock-section-label">Expand / collapse</span>
    <p>A nested block can hide or reveal its child list.</p>
  </div>
  <div class="block-example-grid nested-block-example-grid">
    <article class="block-example-card">
      <span class="block-example-kind">Expanded</span>
      <PlaygroundNestedBlock item={nestedWithChildren} parentId={null} container="component-list:examples" collapsible>
        <div class="nested-block-example-children">
          {#each nestedWithChildren.children as child}<PlaygroundBasicBlock item={child} parentId={nestedWithChildren.id} container="component-list:104" title={child.label} />{/each}
        </div>
      </PlaygroundNestedBlock>
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">Collapsed</span>
      <PlaygroundNestedBlock item={nestedWithChildren} parentId={null} container="component-list:examples" collapsible initiallyOpen={false}>
        <div class="nested-block-example-children">
          {#each nestedWithChildren.children as child}<PlaygroundBasicBlock item={child} parentId={nestedWithChildren.id} container="component-list:104" title={child.label} />{/each}
        </div>
      </PlaygroundNestedBlock>
    </article>
  </div>
</section>
</section>
