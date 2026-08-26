<script lang="ts">
  export let navigate: (path: string) => void
  export let page: 'dnd-ops' | 'block-examples' | 'grammar-blocks' | 'grammar-editor'
  import BlockExamplesPlayground from './BlockExamplesPlayground.svelte'
  import DndOpsPlayground from './DndOpsPlayground.svelte'
  import GrammarBlocksPlayground from './GrammarBlocksPlayground.svelte'
  import GrammarEditorPlayground from './GrammarEditorPlayground.svelte'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'

  const links = [
    { label: 'DnD operations', path: '/playground/dnd-ops', page: 'dnd-ops' as const },
    { label: 'Block examples', path: '/playground/block-examples', page: 'block-examples' as const },
    { label: 'Grammar blocks', path: '/playground/grammar-blocks', page: 'grammar-blocks' as const },
    { label: 'Grammar editor', path: '/playground/grammar-editor', page: 'grammar-editor' as const },
  ]
</script>

<div class="playground">
  <aside class="playground-nav" aria-label="Playground navigation">
    <p class="eyebrow">Playground</p>
    <h1>Experiments</h1>
    <p class="muted playground-intro">Small isolated spaces for testing interaction ideas.</p>
    <nav>
      {#each links as link}
        <a class:playground-nav-active={page === link.page} href={link.path} on:click|preventDefault={() => navigate(link.path)}>{link.label}<span aria-hidden="true">→</span></a>
      {/each}
    </nav>
  </aside>

  <main class="playground-content">
    {#if page === 'dnd-ops'}
      <p class="eyebrow">DnD operations</p>
      <h2>DnD operations</h2>
      <section class="playground-components" aria-labelledby="playground-components-heading">
        <p class="mock-section-label" id="playground-components-heading">Components</p>
        <div class="playground-component-list">
          <div class="playground-component-item"><NewDropMarker /><span><strong>New item</strong><small>Creates a new component at the target</small></span></div>
          <div class="playground-component-item"><MoveDropMarker /><span><strong>Moving item</strong><small>Moves an existing component to the target</small></span></div>
        </div>
      </section>
      <DndOpsPlayground />
    {:else if page === 'block-examples'}
      <BlockExamplesPlayground />
    {:else if page === 'grammar-blocks'}
      <GrammarBlocksPlayground />
    {:else if page === 'grammar-editor'}
      <GrammarEditorPlayground />
    {/if}
  </main>
</div>
