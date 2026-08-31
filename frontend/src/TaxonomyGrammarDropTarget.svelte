<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import { findTaxonomyDraftNode, taxonomyEditorContainer, taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyPlacement, type TaxonomyDraftNode, type TaxonomyRulePredicate } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let placements: TaxonomyPlacement[] = []
  export let taxonomyDraft: TaxonomyDraftNode | undefined
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}
  export let readOnly = false

  function dropLabel(node: StructureNode) {
    if (node.id === 'grammar-root') return 'Drop taxonomy at the root'
    return `Drop taxonomy for ${node.label}`
  }

  $: label = dropLabel(node)
  $: filled = placements.length > 0
  $: container = taxonomyTargetContainer(String(node.id))
  function categoryNode(placement: TaxonomyPlacement) { return taxonomyDraft ? findTaxonomyDraftNode(taxonomyDraft, placement.id) : undefined }
  function accepts(state: DragDropState<TaxonomyDragPayload>) {
    const item = state.draggedItem
    const taxonomySource = state.sourceContainer === taxonomyEditorContainer || state.sourceContainer.startsWith(`${taxonomyEditorContainer}:target:`)
    return node.blockType === 'binding' && taxonomySource && Boolean(item && item.taxonomyType === 'category')
  }
  function getMode(state: DragDropState<TaxonomyDragPayload>) {
    if (state.draggedItem?.kind === 'palette') return 'new' as const
    return state.sourceContainer.startsWith('taxonomy-editor:target:') ? 'move' as const : 'inactive' as const
  }
</script>

{#if filled}
  {#each placements as placement, index (placement.id)}
      {#if !readOnly}<DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy ${index + 1} for ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), index)} />{/if}
    {#if placement.type === 'category'}
      {@const draftNode = categoryNode(placement)}
       <TaxonomyGrammarCategory placement={placement} {draftNode} parentContainer={container} {readOnly} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} />
    {:else if placement.type === 'rule'}
      <TaxonomyRuleBlock title={placement.title} field={node.label} predicate={placement.predicateData ?? { kind: 'exact', field: node.label, value: 'A' }} readOnly={readOnly} container={container} dragData={{ kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }} {onDragEnd} onRemove={() => onRemove(placement.id)} onPredicateChange={(predicate) => onRulePredicateChange(placement.id, predicate)} />
    {/if}
  {/each}
   {#if !readOnly}<DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy after ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), placements.length)} />{/if}
{:else}
  {#if !readOnly}<DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone" ariaLabel={`${label} drop zone`} onDrop={(state) => onDrop(state, String(node.id), 0)} />{/if}
{/if}
