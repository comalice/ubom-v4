<script lang="ts">
  import { draggable, type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'
  import { taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let placements: TaxonomyPlacement[] = []
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}

  function dropLabel(node: StructureNode) {
    if (node.id === 'grammar-root') return 'Drop taxonomy at the root'
    return `Drop taxonomy for ${node.label}`
  }

  $: label = dropLabel(node)
  $: filled = placements.length > 0
  $: container = taxonomyTargetContainer(String(node.id))
  function getMode(state: DragDropState<TaxonomyDragPayload>) {
    if (state.draggedItem?.kind === 'palette') return 'new' as const
    return state.sourceContainer.startsWith('taxonomy-editor:target:') ? 'move' as const : 'inactive' as const
  }
</script>

{#if filled}
  {#each placements as placement, index (placement.id)}
      <DropZone variant="gap" {container} {getMode} accepts={(session) => session.sourceContainer === 'taxonomy-editor' || session.sourceContainer.startsWith('taxonomy-editor:target:')} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy ${index + 1} for ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), index)} />
    {#if placement.type === 'category'}
      <TaxonomyCategoryBlock title={placement.title} summary={placement.summary} label={placement.label ?? placement.title} nodeId={placement.nodeId ?? placement.id} predicate={placement.predicate ?? ''} layout="tree" container={container} dragData={{ kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }} {onDragEnd} onRemove={() => onRemove(placement.id)} />
    {:else if placement.type === 'rule'}
      <TaxonomyRuleBlock title={placement.title} summary={placement.summary} predicate={placement.predicateData ?? { kind: 'exact', field: node.label, value: 'A' }} container={container} dragData={{ kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }} {onDragEnd} onRemove={() => onRemove(placement.id)} />
    {:else}
      <div data-dnd-kind="move" data-dnd-container={container} class="taxonomy-placed-attribute" use:draggable={{ container, dragData: { kind: 'taxonomy', taxonomyType: placement.type, placementId: placement.id }, callbacks: { onDragEnd } }}><span class="taxonomy-tree-mark" aria-hidden="true"></span><span><strong>{placement.title}</strong><small>{placement.attribute ?? 'Part property'}: {placement.value ?? '10'}</small></span><button type="button" aria-label="Remove {placement.title}" on:click={() => onRemove(placement.id)}>x</button></div>
    {/if}
  {/each}
  <DropZone variant="gap" {container} {getMode} accepts={(session) => session.sourceContainer === 'taxonomy-editor' || session.sourceContainer.startsWith('taxonomy-editor:target:')} className="taxonomy-drop-zone taxonomy-drop-gap" ariaLabel={`Insert taxonomy after ${node.label}`} onDrop={(state) => onDrop(state, String(node.id), placements.length)} />
{:else}
  <DropZone variant="gap" {container} {getMode} accepts={(session) => session.sourceContainer === 'taxonomy-editor' || session.sourceContainer.startsWith('taxonomy-editor:target:')} className="taxonomy-drop-zone" ariaLabel={`${label} drop zone`} onDrop={(state) => onDrop(state, String(node.id), 0)} />
{/if}
