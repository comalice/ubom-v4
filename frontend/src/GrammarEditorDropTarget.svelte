<script lang="ts">
  import { onMount } from 'svelte'
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'
  import type { GrammarDragPayload, GrammarDropTarget } from './GrammarEditorModel'

  export let container: string
  export let target: GrammarDropTarget
  export let label: string
  export let showLabel = false
  export let onDrop: (state: DragDropState<GrammarDragPayload>, target: GrammarDropTarget) => void

  let palette = false
  let over = false

  function handleDragOver(state: DragDropState<GrammarDragPayload>) {
    over = true
    palette = state.draggedItem?.kind === 'palette'
  }
  function handleDragLeave() {
    over = false
    palette = false
  }
  function handleDrop(state: DragDropState<GrammarDragPayload>) {
    onDrop(state, target)
    handleDragLeave()
  }

  onMount(() => {
    const clearDragState = () => handleDragLeave()
    document.addEventListener('dragend', clearDragState)
    document.addEventListener('pointercancel', clearDragState)
    return () => {
      document.removeEventListener('dragend', clearDragState)
      document.removeEventListener('pointercancel', clearDragState)
    }
  })
</script>

<div class:grammar-editor-drop-target-over={over} class:grammar-editor-drop-target-palette={palette} class:grammar-editor-drop-target-labeled={showLabel} class="grammar-editor-drop-target" role="listitem" aria-label={label} use:droppable={{ container, callbacks: { onDragEnter: handleDragOver, onDragOver: handleDragOver, onDragLeave: handleDragLeave, onDrop: handleDrop } }}>
  <NewDropMarker />
  <MoveDropMarker />
  {#if showLabel}<span class="grammar-editor-drop-target-label">{label}</span>{/if}
</div>
