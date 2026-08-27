<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
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
    <DropZone variant="gap" className={index === 0 ? 'dnd-insert-zone-edge' : ''} {container} getMode={(state) => state.sourceContainer.startsWith('component-list:') ? 'move' : 'new'} onDragOver={onDragOver} onDragLeave={onDragLeave} onDrop={(state) => onDrop({ parentId, index }, state)} />
    {#if item.blockType === 'nested'}
      <NestedBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)}>
        <svelte:self nodes={item.children} parentId={item.id} container={`component-list:${item.id}`} {onDragOver} {onDragLeave} {onDrop} {onDragEnd} {onRemove} />
      </NestedBlock>
    {:else}
      <BasicBlock item={item} parentId={parentId} container={container} onDragEnd={onDragEnd} onRemove={() => onRemove(item.id)} />
    {/if}
  {/each}
  <DropZone variant="gap" className="dnd-insert-zone-edge" {container} getMode={(state) => state.sourceContainer.startsWith('component-list:') ? 'move' : 'new'} onDragOver={onDragOver} onDragLeave={onDragLeave} onDrop={(state) => onDrop({ parentId, index: nodes.length }, state)} />
</div>
