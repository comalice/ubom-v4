<script lang="ts">
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import PlaygroundMoveMarker from './PlaygroundMoveMarker.svelte'
  import PlaygroundNewMarker from './PlaygroundNewMarker.svelte'
  import PlaygroundBasicBlock from './PlaygroundBasicBlock.svelte'
  import PlaygroundNestedBlock from './PlaygroundNestedBlock.svelte'
  import { type ComponentNode, type DraggableComponent } from './PlaygroundDraggable.svelte'

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
    {#if item.blockType === 'nested'}
      <PlaygroundNestedBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)}>
        <svelte:self nodes={item.children} parentId={item.id} container={`component-list:${item.id}`} {onDragOver} {onDragLeave} {onDrop} {onDragEnd} {onRemove} />
      </PlaygroundNestedBlock>
    {:else}
      <PlaygroundBasicBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)} />
    {/if}
  {/each}
  <div data-parent-id={parentId ?? 'root'} data-slot-index={nodes.length} class="dnd-insert-zone dnd-insert-zone-edge" use:droppable={{ container, callbacks: { onDragOver, onDragLeave, onDrop: (state) => onDrop({ parentId, index: nodes.length }, state) } }}><i aria-hidden="true"></i><PlaygroundNewMarker /><PlaygroundMoveMarker /></div>
</div>
