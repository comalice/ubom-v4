<script lang="ts">
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import PlaygroundMoveMarker from './PlaygroundMoveMarker.svelte'
  import PlaygroundNewMarker from './PlaygroundNewMarker.svelte'
  import PlaygroundDraggable, { type DraggableComponent } from './PlaygroundDraggable.svelte'

  export type ComponentNode = DraggableComponent & { id: number; children: ComponentNode[] }
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
    <div data-parent-id={parentId ?? 'root'} data-slot-index={index} class:dnd-insert-zone-edge={index === 0} class="dnd-insert-zone" use:droppable={{ container, callbacks: { onDragOver, onDragLeave, onDrop: (state) => onDrop({ parentId, index }, state) } }}><i aria-hidden="true"></i><PlaygroundNewMarker /><PlaygroundMoveMarker /></div>
    <PlaygroundDraggable item={item} dragData={{ ...item, nodeId: item.id, sourceParentId: parentId }} container={container} variant="placed" onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)} />
  {/each}
  <div data-parent-id={parentId ?? 'root'} data-slot-index={nodes.length} class="dnd-insert-zone dnd-insert-zone-edge" use:droppable={{ container, callbacks: { onDragOver, onDragLeave, onDrop: (state) => onDrop({ parentId, index: nodes.length }, state) } }}><i aria-hidden="true"></i><PlaygroundNewMarker /><PlaygroundMoveMarker /></div>
</div>
