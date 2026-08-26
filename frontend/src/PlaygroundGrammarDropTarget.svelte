<script lang="ts">
  import { droppable, dndState, type DragDropState } from '@thisux/sveltednd'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'
  import type { PlaygroundDragPayload, PlaygroundDropTarget } from './PlaygroundGrammar'

  export let container: string
  export let target: PlaygroundDropTarget
  export let label: string
  export let showLabel = false
  export let onDrop: (state: DragDropState<PlaygroundDragPayload>, target: PlaygroundDropTarget) => void

  $: palette = dndState.draggedItem?.kind === 'palette'
  $: over = dndState.targetContainer === container
</script>

<div class:grammar-editor-drop-target-over={over} class:grammar-editor-drop-target-palette={palette} class:grammar-editor-drop-target-labeled={showLabel} class="grammar-editor-drop-target" role="listitem" aria-label={label} use:droppable={{ container, callbacks: { onDrop: (state) => onDrop(state, target) } }}>
  <NewDropMarker />
  <MoveDropMarker />
  {#if showLabel}<span class="grammar-editor-drop-target-label">{label}</span>{/if}
</div>
