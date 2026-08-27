<script lang="ts">
  import BlockShell from './BlockShell.svelte'

  export type TaxonomyCategoryBlockLayout = 'catalogue' | 'tree'

  export let title: string
  export let summary: string
  export let label: string
  export let nodeId: string
  export let predicate: string
  export let predicateLabel: string = 'Rule'
  export let addChildLabel: string = '+ add child category'
  export let layout: TaxonomyCategoryBlockLayout = 'catalogue'
</script>

<BlockShell title={title} summary={summary} kind="taxonomy" collapsible initiallyOpen={true} onRemove={() => {}}>
  {#if layout === 'tree'}
    <div class="taxonomy-block-fields">
      <label>Label<input value={label} /></label>
      <label>Node ID<input value={nodeId} /></label>
    </div>
    <div class="taxonomy-block-meta"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
    <button class="taxonomy-block-add" type="button">{addChildLabel}</button>
  {:else}
    <div class="taxonomy-block-body">
      <label>Label<input value={label} aria-label="Category label" /></label>
      <label>Node ID<input value={nodeId} aria-label="Category node ID" /></label>
      <div class="taxonomy-block-rule"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
      <button class="taxonomy-block-add" type="button">{addChildLabel}</button>
    </div>
  {/if}
</BlockShell>
