<script lang="ts">
  import DropZone from './DropZone.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
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

</script>

<DropZone variant="slot" className="grammar-editor-drop-target" {container} {label} {showLabel} role="listitem" onDragOver={handleDragOver} onDragLeave={handleDragLeave} onDrop={handleDrop} />
