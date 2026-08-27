<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import DropZone, { type DropZoneMode } from './DropZone.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import type { TaxonomyDragPayload } from './TaxonomyEditorModel'

  export type TaxonomyCategoryBlockLayout = 'catalogue' | 'tree'

  export let title: string
  export let summary: string
  export let label: string
  export let nodeId: string
  export let predicate: string
  export let predicateLabel: string = 'Rule'
  export let addChildLabel: string = 'Drop child category'
  export let showChildren = false
  export let childContainer = ''
  export let childDropMode: DropZoneMode = 'inactive'
  export let onChildDragOver: ((state: DragDropState<TaxonomyDragPayload>) => void) | undefined
  export let onChildDragLeave: (() => void) | undefined
  export let onChildDrop: ((state: DragDropState<TaxonomyDragPayload>) => void) | undefined
  export let layout: TaxonomyCategoryBlockLayout = 'catalogue'
  export let container = ''
  export let dragData: TaxonomyDragPayload | undefined = undefined
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

  <BlockShell title={title} summary={summary} kind="taxonomy" collapsible initiallyOpen={true} {container} {dragData} dragHandle=".block-shell-header" interactive={['input', 'select', 'button']} {onDragEnd} {onRemove}>
  {#if layout === 'tree'}
    <div class="taxonomy-block-fields">
      <label>Label<input value={label} /></label>
      <label>Node ID<input value={nodeId} /></label>
    </div>
    <div class="taxonomy-block-meta"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
    {#if showChildren}<div class="taxonomy-category-children"><span class="taxonomy-category-children-label">Child categories</span><DropZone variant="gap" mode={childDropMode} className="taxonomy-child-drop-zone" container={childContainer} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDragOver={onChildDragOver} onDragLeave={onChildDragLeave} onDrop={onChildDrop} ariaLabel={addChildLabel} /><slot name="children" /></div>{/if}
  {:else}
    <div class="taxonomy-block-body">
      <label>Label<input value={label} aria-label="Category label" /></label>
      <label>Node ID<input value={nodeId} aria-label="Category node ID" /></label>
      <div class="taxonomy-block-rule"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
      {#if showChildren}<div class="taxonomy-category-children"><span class="taxonomy-category-children-label">Child categories</span><DropZone variant="gap" mode={childDropMode} className="taxonomy-child-drop-zone" container={childContainer} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDragOver={onChildDragOver} onDragLeave={onChildDragLeave} onDrop={onChildDrop} ariaLabel={addChildLabel} /><slot name="children" /></div>{/if}
    </div>
  {/if}
</BlockShell>
