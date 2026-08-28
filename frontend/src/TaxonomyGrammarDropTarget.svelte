<script lang="ts">
  import { draggable, type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'
  import { findTaxonomyDraftNode, taxonomyEditorContainer, taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyPlacement, type TaxonomyDraftNode } from './TaxonomyEditorModel'
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
      <DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy ${index + 1} for ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), index)} />
    {#if placement.type === 'category'}
      {@const draftNode = categoryNode(placement)}
      <TaxonomyCategoryBlock title={placement.title} summary={placement.summary} label={placement.label ?? placement.title} nodeId={placement.nodeId ?? placement.id} layout="tree" container={container} dragData={{ kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }} showRuleDropZone ruleContainer={draftNode ? taxonomyTargetContainer(draftNode.id) : ''} onRuleDrop={draftNode ? (state) => onRuleDrop(state, draftNode.id) : undefined} showChildren childContainer={draftNode ? taxonomyTargetContainer(draftNode.id) : ''} onChildDrop={draftNode ? (state) => onChildDrop(state, draftNode.id) : undefined} {onDragEnd} onRemove={() => onRemove(placement.id)} onAttributes={() => { if (draftNode) onAttributes(draftNode.id) }}>
        <svelte:fragment slot="rules">{#if draftNode}{#each draftNode.rules as rule (rule.id)}<TaxonomyRuleBlock title={rule.title} field={draftNode.targetId} predicate={rule.predicateData ?? { kind: 'exact', field: draftNode.targetId, value: 'A' }} container={taxonomyTargetContainer(draftNode.id)} dragData={{ kind: 'taxonomy', taxonomyType: 'rule', placementId: rule.id }} {onDragEnd} onRemove={() => onRemove(rule.id)} />{/each}{/if}</svelte:fragment>
            <svelte:fragment slot="children">{#if draftNode}{#each draftNode.children as child}{#if child.placements[0]}<TaxonomyCategoryBlock title={child.placements[0].title} summary={child.placements[0].summary} label={child.placements[0].label ?? child.label} nodeId={child.placements[0].nodeId ?? child.id} predicate={child.placements[0].predicate ?? ''} layout="tree" showChildren addChildLabel="Drop a category here" onAttributes={() => onAttributes(child.id)} />{/if}{/each}{/if}</svelte:fragment>
      </TaxonomyCategoryBlock>
    {:else if placement.type === 'rule'}
      <TaxonomyRuleBlock title={placement.title} field={node.label} predicate={placement.predicateData ?? { kind: 'exact', field: node.label, value: 'A' }} container={container} dragData={{ kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }} {onDragEnd} onRemove={() => onRemove(placement.id)} />
    {:else}
      <div data-dnd-kind="move" data-dnd-container={container} class="taxonomy-placed-attribute" use:draggable={{ container, dragData: { kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }, callbacks: { onDragEnd } }}><span class="taxonomy-tree-mark" aria-hidden="true"></span><span><strong>{placement.title}</strong><small>{placement.attribute ?? 'Part property'}: {placement.value ?? '10'}</small></span><button type="button" aria-label="Remove {placement.title}" on:click={() => onRemove(placement.id)}>x</button></div>
    {/if}
  {/each}
  <DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy after ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), placements.length)} />
{:else}
  <DropZone variant="gap" {container} {getMode} accepts={accepts} className="taxonomy-drop-zone" ariaLabel={`${label} drop zone`} onDrop={(state) => onDrop(state, String(node.id), 0)} />
{/if}
