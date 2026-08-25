<script lang="ts">
  import PlaygroundBasicBlock from './PlaygroundBasicBlock.svelte'

  const withoutRemove = { id: 101, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
  const withRemove = { id: 102, kind: 'component' as const, label: 'Literal', blockType: 'basic' as const, children: [] }
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
      <PlaygroundBasicBlock item={withoutRemove} parentId={null} container="component-list:examples" />
    </article>
    <article class="block-example-card">
      <span class="block-example-kind">With x</span>
      <PlaygroundBasicBlock item={withRemove} parentId={null} container="component-list:examples" onRemove={() => {}} />
    </article>
  </div>
</section>
