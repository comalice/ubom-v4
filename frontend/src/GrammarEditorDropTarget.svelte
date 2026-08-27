<script lang="ts">
  import DropZone from './DropZone.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import type { GrammarDragPayload, GrammarDropTarget } from './GrammarEditorModel'

  export let container: string
  export let target: GrammarDropTarget
  export let label: string
  export let showLabel = false
  export let getMode: ((state: DragDropState<GrammarDragPayload>) => 'inactive' | 'new' | 'move') | undefined
  export let onDragOver: ((state: DragDropState<GrammarDragPayload>) => void) | undefined
  export let onDragLeave: (() => void) | undefined
  export let onDrop: (state: DragDropState<GrammarDragPayload>, target: GrammarDropTarget) => void

  function handleDrop(state: DragDropState<GrammarDragPayload>) { onDrop(state, target) }

</script>

<DropZone variant="slot" className="grammar-editor-drop-target" {container} {label} {showLabel} {getMode} {onDragOver} {onDragLeave} accepts={(state) => state.sourceContainer.startsWith('grammar-editor:')} role="listitem" onDrop={handleDrop} />
