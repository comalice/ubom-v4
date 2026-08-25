<script lang="ts">
  import { droppable, dndState, type DragDropState } from '@thisux/sveltednd'
  import { type DropTarget, type DragPayload } from './grammarMock'

  export let container: string
  export let target: DropTarget
  export let label: string
  export let showLabel = false
  export let onDrop: (state: DragDropState<DragPayload>, target: DropTarget) => void

  $: palette = dndState.draggedItem?.kind === 'palette'
  $: over = dndState.targetContainer === container
</script>

<div
  class="grammar-drop-target"
  class:grammar-drop-target-dragging={dndState.isDragging}
  class:grammar-drop-target-over={over}
  class:grammar-drop-target-palette={palette}
  class:grammar-drop-target-labeled={showLabel}
  role="listitem"
  aria-label={label}
  use:droppable={{ container, callbacks: { onDrop: (state) => onDrop(state, target) }, attributes: { dragOverClass: 'grammar-drop-target-over' } }}
>
  <span class="grammar-drop-target-visual" aria-hidden="true">{palette ? '+' : ''}</span>
  {#if showLabel}<span class="grammar-drop-target-label">{label}</span>{/if}
</div>
