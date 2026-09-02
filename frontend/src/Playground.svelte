<script lang="ts">
  export let navigate: (path: string) => void
  export let page: 'core-components' | 'html-primitives' | 'dnd-ops' | 'block-examples' | 'grammar-blocks' | 'taxonomy-editor-components' | 'grammar-editor' | 'taxonomy-editor' | 'taxonomy-first'
  import CoreComponentsPlayground from './CoreComponentsPlayground.svelte'
  import HtmlPrimitivesPlayground from './HtmlPrimitivesPlayground.svelte'
  import BlockExamplesPlayground from './BlockExamplesPlayground.svelte'
  import DndOpsPlayground from './DndOpsPlayground.svelte'
  import GrammarBlocksPlayground from './GrammarBlocksPlayground.svelte'
  import GrammarEditorPlayground from './GrammarEditorPlayground.svelte'
  import TaxonomyEditorPlayground from './TaxonomyEditorPlayground.svelte'
  import TaxonomyEditorComponentsPlayground from './TaxonomyEditorComponentsPlayground.svelte'
  import TaxonomyFirstPlayground from './TaxonomyFirstPlayground.svelte'

  const links = [
    { label: 'Core components', path: '/playground/core-components', page: 'core-components' as const },
    { label: 'HTML primitives', path: '/playground/html-primitives', page: 'html-primitives' as const },
    { label: 'DnD operations', path: '/playground/dnd-ops', page: 'dnd-ops' as const },
    { label: 'Block examples', path: '/playground/block-examples', page: 'block-examples' as const },
    { label: 'Grammar blocks', path: '/playground/grammar-blocks', page: 'grammar-blocks' as const },
    { label: 'Taxonomy editor components', path: '/playground/taxonomy-editor-components', page: 'taxonomy-editor-components' as const },
    { label: 'Grammar editor', path: '/playground/grammar-editor', page: 'grammar-editor' as const },
    { label: 'Taxonomy editor', path: '/playground/taxonomy-editor', page: 'taxonomy-editor' as const },
    { label: 'Taxonomy-first', path: '/playground/taxonomy-first', page: 'taxonomy-first' as const },
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
    {#if page === 'core-components'}
      <CoreComponentsPlayground />
    {:else if page === 'html-primitives'}
      <HtmlPrimitivesPlayground />
    {:else if page === 'dnd-ops'}
      <DndOpsPlayground />
    {:else if page === 'block-examples'}
      <BlockExamplesPlayground />
    {:else if page === 'grammar-blocks'}
      <GrammarBlocksPlayground />
    {:else if page === 'taxonomy-editor-components'}
      <TaxonomyEditorComponentsPlayground />
    {:else if page === 'grammar-editor'}
      <GrammarEditorPlayground />
    {:else if page === 'taxonomy-editor'}
      <TaxonomyEditorPlayground />
    {:else if page === 'taxonomy-first'}
      <TaxonomyFirstPlayground />
    {/if}
  </main>
</div>
