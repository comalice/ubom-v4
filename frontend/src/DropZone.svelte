<script lang="ts">
  import { onMount } from 'svelte'
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import { clearDndSession, dndSession, type DndSession } from './DndSession'
  import InactiveDropMarker from './InactiveDropMarker.svelte'
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
  export let accepts: ((session: DndSession) => boolean) | undefined

  let over = false
  let dragMode: DropZoneMode = 'inactive'
  $: interactive = Boolean(container && (onDragOver || onDragLeave || onDrop))
  $: session = $dndSession
  $: sessionAccepted = Boolean(session && (!accepts || accepts(session)))
  $: staticMode = !interactive && mode !== 'inactive' ? mode : 'inactive'
  $: currentMode = interactive
    ? (over ? dragMode : sessionAccepted ? (session?.kind === 'new' ? 'new' : 'move') : 'inactive')
    : mode
  $: dropLabel = ariaLabel || label

  function handleDragOver(state: DragDropState<any>) {
    over = true
    const nextMode = getMode?.(state) ?? (state.draggedItem?.kind === 'palette' ? 'new' : 'move')
    const dragSession: DndSession = {
      kind: nextMode === 'new' ? 'new' : 'move',
      sourceContainer: state.sourceContainer,
    }
    dragMode = accepts && !accepts(dragSession) ? 'inactive' : nextMode
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
    clearDndSession()
  }

  onMount(() => {
    const handleDragStart = (event: Event) => {
      const source = (event.target as HTMLElement | null)?.closest<HTMLElement>('[data-dnd-kind]')
      if (!source) return
      const kind = source.dataset.dndKind === 'new' ? 'new' : 'move'
      dndSession.set({ kind, sourceContainer: source.dataset.dndContainer ?? '' })
    }
    const clearDragState = () => handleDragLeave()
    document.addEventListener('dragstart', handleDragStart)
    document.addEventListener('dragend', clearDragState)
    document.addEventListener('dragend', clearDndSession)
    document.addEventListener('pointercancel', clearDragState)
    document.addEventListener('pointercancel', clearDndSession)
    return () => {
      document.removeEventListener('dragstart', handleDragStart)
      document.removeEventListener('dragend', clearDragState)
      document.removeEventListener('dragend', clearDndSession)
      document.removeEventListener('pointercancel', clearDragState)
      document.removeEventListener('pointercancel', clearDndSession)
    }
  })

</script>

<div
  class={`drop-zone drop-zone-${variant} ${className}`}
  class:drop-zone-available={currentMode !== 'inactive'}
  class:drop-zone-new-hover={over && currentMode === 'new'}
  class:drop-zone-move-hover={over && currentMode === 'move'}
  class:drop-zone-static-new={staticMode === 'new'}
  class:drop-zone-static-move={staticMode === 'move'}
  class:drop-zone-filled={variant === 'filled'}
  aria-label={dropLabel}
  {role}
  use:droppable={{ container, disabled: !interactive, callbacks: { onDragEnter: handleDragOver, onDragOver: handleDragOver, onDragLeave: handleDragLeave, onDrop: handleDrop } }}
>
  {#if currentMode === 'new'}<i class="drop-zone-bar" aria-hidden="true"></i><NewDropMarker />
  {:else if currentMode === 'move'}<i class="drop-zone-bar" aria-hidden="true"></i><MoveDropMarker />
  {:else if session && accepts && sessionAccepted}<i class="drop-zone-bar" aria-hidden="true"></i><InactiveDropMarker />{/if}
  {#if showHint}<span class="drop-zone-hint" aria-hidden="true">+</span>{/if}
  {#if showLabel}<span class="drop-zone-label">{label}</span>{/if}
  <slot />
</div>
