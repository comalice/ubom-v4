<script lang="ts">
  import { onMount } from 'svelte'
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'

  export type DropZoneMode = 'inactive' | 'new' | 'move'
  export type DropZoneVariant = 'gap' | 'slot' | 'filled'
  export type DropZoneModeResolver = (state: DragDropState<any>) => DropZoneMode

  export let container = ''
  export let variant: DropZoneVariant = 'gap'
  export let mode: DropZoneMode = 'inactive'
  export let label = ''
  export let showLabel = false
  export let showHint = false
  export let className = ''
  export let ariaLabel = label
  export let role = ''
  export let onDragOver: ((state: DragDropState<any>) => void) | undefined
  export let onDragLeave: (() => void) | undefined
  export let onDrop: ((state: DragDropState<any>) => void) | undefined
  export let getMode: DropZoneModeResolver | undefined

  let over = false
  let dragMode: DropZoneMode = 'inactive'
  $: interactive = Boolean(container && (onDragOver || onDragLeave || onDrop))
  $: currentMode = interactive ? (over ? dragMode : 'inactive') : mode
  $: dropLabel = ariaLabel || label

  function handleDragOver(state: DragDropState<any>) {
    over = true
    dragMode = getMode?.(state) ?? (state.draggedItem?.kind === 'palette' ? 'new' : 'move')
    onDragOver?.(state)
  }

  function handleDragLeave() {
    over = false
    dragMode = 'inactive'
    onDragLeave?.()
  }

  function handleDrop(state: DragDropState<any>) {
    onDrop?.(state)
    handleDragLeave()
  }

  onMount(() => {
    if (!interactive) return
    const clearDragState = () => handleDragLeave()
    document.addEventListener('dragend', clearDragState)
    document.addEventListener('pointercancel', clearDragState)
    return () => {
      document.removeEventListener('dragend', clearDragState)
      document.removeEventListener('pointercancel', clearDragState)
    }
  })
</script>

<div
  class={`drop-zone drop-zone-${variant} ${className}`}
  class:drop-zone-active={currentMode !== 'inactive'}
  class:drop-zone-new={currentMode === 'new'}
  class:drop-zone-move={currentMode === 'move'}
  class:dnd-insert-zone-active={currentMode !== 'inactive'}
  class:dnd-insert-zone-new={currentMode === 'new'}
  class:dnd-insert-zone-move={currentMode === 'move'}
  class:drop-zone-filled={variant === 'filled'}
  class:dnd-insert-zone={variant === 'gap'}
  aria-label={dropLabel}
  {role}
  use:droppable={{ container, disabled: !interactive, callbacks: { onDragEnter: handleDragOver, onDragOver: handleDragOver, onDragLeave: handleDragLeave, onDrop: handleDrop } }}
>
  <i aria-hidden="true"></i>
  <NewDropMarker />
  <MoveDropMarker />
  {#if showHint}<span class="drop-zone-hint" aria-hidden="true">+</span>{/if}
  {#if showLabel}<span class="drop-zone-label">{label}</span>{/if}
  <slot />
</div>
