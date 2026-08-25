<script lang="ts">
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import PlaygroundComponentList, { type DropSlot } from './PlaygroundComponentList.svelte'
  import type { ComponentNode, DraggableComponent } from './PlaygroundDraggable.svelte'

  export let nodes: ComponentNode[] = []
  export let filled = nodes.length > 0
  export let container = 'component-list:root'
  export let onDrop: (state: DragDropState<DraggableComponent>) => void
  export let onDragOver: (state: DragDropState<DraggableComponent>) => void
  export let onGapDragLeave: () => void
  export let onGapDrop: (slot: DropSlot, state: DragDropState<DraggableComponent>) => void
  export let onDragEnd: () => void
  export let onRemove: (id: number) => void

  let newItemOver = false

  function handleDragEnter(state: DragDropState<DraggableComponent>) {
    newItemOver = state.draggedItem?.kind === 'component'
  }

  function handleDragLeave() {
    newItemOver = false
  }

  function handleDrop(state: DragDropState<DraggableComponent>) {
    onDrop(state)
    newItemOver = false
  }

  function handleDragEnd() {
    newItemOver = false
    onDragEnd()
  }
</script>

<div class:dnd-field-filled={filled || nodes.length > 0} class="dnd-field" use:droppable={{ container, disabled: filled || nodes.length > 0, callbacks: { onDragEnter: handleDragEnter, onDragLeave: handleDragLeave, onDrop: handleDrop }, attributes: { dragOverClass: 'dnd-field-over' } }}>
  {#if filled || nodes.length > 0}
    <slot>
      <PlaygroundComponentList {nodes} parentId={null} {container} onDragOver={onDragOver} onDragLeave={onGapDragLeave} onDrop={onGapDrop} onDragEnd={handleDragEnd} {onRemove} />
    </slot>
  {:else if newItemOver}<span class="dnd-new-drop-target" aria-hidden="true">+</span>{:else}<span>Drop a component here</span>{/if}
</div>
