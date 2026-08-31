<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import ControlButton from './ControlButton.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import DropZone, { type DropZoneMode } from './DropZone.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import { acceptsTaxonomyRuleDrop, type TaxonomyDragPayload } from './TaxonomyEditorModel'

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
  export let showRuleDropZone = false
  export let ruleDropLabel = 'Drop a rule here'
  export let ruleContainer = ''
  export let ruleDropMode: DropZoneMode = 'inactive'
  export let onRuleDragOver: ((state: DragDropState<TaxonomyDragPayload>) => void) | undefined
  export let onRuleDragLeave: (() => void) | undefined
  export let onRuleDrop: ((state: DragDropState<TaxonomyDragPayload>) => void) | undefined
  export let onAttributes: () => void = () => {}
  export let onLabelChange: (label: string) => void = () => {}
  export let onNodeIdChange: (nodeId: string) => void = () => {}
  export let layout: TaxonomyCategoryBlockLayout = 'catalogue'
  export let container = ''
  export let dragData: TaxonomyDragPayload | undefined = undefined
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
  export let readOnly = false
</script>

  <BlockShell title={title} summary={summary} kind="taxonomy" collapsible initiallyOpen={true} container={readOnly ? '' : container} dragData={readOnly ? undefined : dragData} dragHandle=".block-shell-drag" interactive={['input', 'select', 'button']} {onDragEnd} onRemove={readOnly ? undefined : onRemove}>
  <svelte:fragment slot="header-actions">{#if !readOnly}<PrimitiveButton variant="secondary" size="small" type="button" aria-label={`View attributes for ${label}`} on:click={onAttributes}>attributes</PrimitiveButton>{/if}</svelte:fragment>
  {#if layout === 'tree'}
    <div class="taxonomy-block-fields">
       <label>Label<input value={label} readonly={readOnly} on:input={(event) => onLabelChange(event.currentTarget.value)} /></label>
       <label>Node ID<input value={nodeId} readonly={readOnly} on:input={(event) => onNodeIdChange(event.currentTarget.value)} /></label>
    </div>
    <div class="taxonomy-block-meta"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
    {#if showRuleDropZone}<div class="taxonomy-category-rules"><span class="taxonomy-category-rules-label">Rules</span>{#if !readOnly}<DropZone variant="slot" mode={ruleDropMode} className="taxonomy-rule-drop-zone" container={ruleContainer} accepts={(state) => acceptsTaxonomyRuleDrop(state.draggedItem)} onDragOver={onRuleDragOver} onDragLeave={onRuleDragLeave} onDrop={onRuleDrop} ariaLabel={ruleDropLabel} />{/if}<slot name="rules" /></div>{/if}
    {#if showChildren}<div class="taxonomy-category-children"><span class="taxonomy-category-children-label">Child categories</span>{#if !readOnly}<DropZone variant="gap" mode={childDropMode} className="taxonomy-child-drop-zone" container={childContainer} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDragOver={onChildDragOver} onDragLeave={onChildDragLeave} onDrop={onChildDrop} ariaLabel={addChildLabel} />{/if}<slot name="children" /></div>{/if}
  {:else}
    <div class="taxonomy-block-body">
       <label>Label<input value={label} readonly={readOnly} aria-label="Category label" on:input={(event) => onLabelChange(event.currentTarget.value)} /></label>
       <label>Node ID<input value={nodeId} readonly={readOnly} aria-label="Category node ID" on:input={(event) => onNodeIdChange(event.currentTarget.value)} /></label>
      <div class="taxonomy-block-rule"><span>{predicateLabel}</span><strong>{predicate}</strong></div>
       {#if showRuleDropZone}<div class="taxonomy-category-rules"><span class="taxonomy-category-rules-label">Rules</span>{#if !readOnly}<DropZone variant="slot" mode={ruleDropMode} className="taxonomy-rule-drop-zone" container={ruleContainer} accepts={(state) => acceptsTaxonomyRuleDrop(state.draggedItem)} onDragOver={onRuleDragOver} onDragLeave={onRuleDragLeave} onDrop={onRuleDrop} ariaLabel={ruleDropLabel} />{/if}<slot name="rules" /></div>{/if}
      {#if showChildren}<div class="taxonomy-category-children"><span class="taxonomy-category-children-label">Child categories</span>{#if !readOnly}<DropZone variant="gap" mode={childDropMode} className="taxonomy-child-drop-zone" container={childContainer} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDragOver={onChildDragOver} onDragLeave={onChildDragLeave} onDrop={onChildDrop} ariaLabel={addChildLabel} />{/if}<slot name="children" /></div>{/if}
    </div>
  {/if}
</BlockShell>
