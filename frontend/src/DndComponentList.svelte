<script lang="ts">
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'
  import BasicBlock from './BasicBlock.svelte'
  import NestedBlock from './NestedBlock.svelte'
  import { type ComponentNode, type DraggableComponent } from './DndDraggable.svelte'

  export type DropSlot = { parentId: number | null; index: number }
  export let nodes: ComponentNode[] = []
  export let parentId: number | null = null
  export let container: string
  export let onDragOver: (state: DragDropState<DraggableComponent>) => void
  export let onDragLeave: () => void
  export let onDrop: (slot: DropSlot, state: DragDropState<DraggableComponent>) => void
  export let onDragEnd: () => void
  export let onRemove: (id: number) => void
</script>

<div class="dnd-placed-list">
  {#each nodes as item, index (item.id)}
    <div data-parent-id={parentId ?? 'root'} data-slot-index={index} class:dnd-insert-zone-edge={index === 0} class="dnd-insert-zone" use:droppable={{ container, callbacks: { onDragOver, onDragLeave, onDrop: (state) => onDrop({ parentId, index }, state) } }}><i aria-hidden="true"></i><NewDropMarker /><MoveDropMarker /></div>
    {#if item.blockType === 'nested'}
      <NestedBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)}>
        <svelte:self nodes={item.children} parentId={item.id} container={`component-list:${item.id}`} {onDragOver} {onDragLeave} {onDrop} {onDragEnd} {onRemove} />
      </NestedBlock>
    {:else}
      <BasicBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)} />
    {/if}
  {/each}
  <div data-parent-id={parentId ?? 'root'} data-slot-index={nodes.length} class="dnd-insert-zone dnd-insert-zone-edge" use:droppable={{ container, callbacks: { onDragOver, onDragLeave, onDrop: (state) => onDrop({ parentId, index: nodes.length }, state) } }}><i aria-hidden="true"></i><NewDropMarker /><MoveDropMarker /></div>
</div>
