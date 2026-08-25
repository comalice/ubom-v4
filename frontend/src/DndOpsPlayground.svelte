<script lang="ts">
  import { droppable, type DragDropState } from '@thisux/sveltednd'
  import PlaygroundMoveMarker from './PlaygroundMoveMarker.svelte'
  import PlaygroundNewMarker from './PlaygroundNewMarker.svelte'
  import PlaygroundDraggable, { type DraggableComponent } from './PlaygroundDraggable.svelte'

  type ComponentNode = DraggableComponent & { id: number; children: ComponentNode[] }
  type DragItem = DraggableComponent
  type DropSlot = { parentId: number | null; index: number }
  const rootListId = 'component-list:root'
  const component: DraggableComponent = { kind: 'component', label: 'Literal' }
  let placedComponents: ComponentNode[] = []
  let nextComponentId = 1
  let newItemOver = false
  const markerState = { version: 0 }

  function handleDrop(state: DragDropState<DragItem>) {
    if (state.draggedItem?.kind === 'component') placedComponents = [...placedComponents, { ...state.draggedItem, id: nextComponentId++, children: [] }]
    newItemOver = false
  }
  function handleDragEnter(state: DragDropState<DragItem>) { newItemOver = state.draggedItem?.kind === 'component' }
  function handleDragLeave() { newItemOver = false }
  function handleGapDragOver(state: DragDropState<DragItem>) {
    if (state.draggedItem?.kind !== 'component') return
    markerState.version += 1
    const target = (state.targetElement as HTMLElement | null)?.closest('.dnd-insert-zone')
    const markerClass = state.sourceContainer === rootListId ? 'dnd-insert-zone-move' : 'dnd-insert-zone-new'
    if (!target?.classList.contains(markerClass)) {
      clearDropMarker()
      target?.classList.add('dnd-insert-zone-active', markerClass)
    }
  }
  function handleGapDragLeave() {
    const version = markerState.version
    requestAnimationFrame(() => {
      if (version === markerState.version) clearDropMarker()
    })
  }
  function handleGapDrop(slot: DropSlot, state: DragDropState<DragItem>) {
    if (state.draggedItem?.kind === 'component') {
      if (state.sourceContainer === rootListId && state.draggedItem.nodeId) {
        const sourceIndex = placedComponents.findIndex((item) => item.id === state.draggedItem.nodeId)
        if (sourceIndex >= 0) {
          const movingItem = placedComponents[sourceIndex]
          const remaining = placedComponents.filter((item) => item.id !== state.draggedItem?.nodeId)
          const insertIndex = sourceIndex < slot.index ? slot.index - 1 : slot.index
          placedComponents = [...remaining.slice(0, insertIndex), movingItem, ...remaining.slice(insertIndex)]
        }
      } else {
        placedComponents = [...placedComponents.slice(0, slot.index), { ...state.draggedItem, id: nextComponentId++, children: [] }, ...placedComponents.slice(slot.index)]
      }
    }
    clearDropMarker()
  }
  function clearDropMarker() {
    document.querySelectorAll('.dnd-ops-playground .dnd-insert-zone-active').forEach((element) => element.classList.remove('dnd-insert-zone-active', 'dnd-insert-zone-new', 'dnd-insert-zone-move'))
  }
  function handleDragEnd() {
    newItemOver = false
    clearDropMarker()
  }
</script>

<div class="dnd-ops-playground">
  <div class="dnd-ops-heading"><div><span class="mock-section-label">Experiment</span><h3>Drop components into a field</h3></div>{#if placedComponents.length}<button class="dnd-reset" type="button" on:click={() => placedComponents = []}>Reset</button>{/if}</div>
  <div class="dnd-ops-stage">
    <section class="dnd-component-palette" aria-labelledby="component-palette-heading">
      <span class="mock-section-label" id="component-palette-heading">Component palette</span>
      <PlaygroundDraggable item={component} container="component-palette" variant="palette" description="Drag into the field" onDragEnd={handleDragEnd} />
    </section>

    <section class="dnd-field-area" aria-labelledby="field-heading">
      <span class="mock-section-label" id="field-heading">Field</span>
      <div class="dnd-field" class:dnd-field-filled={placedComponents.length > 0} use:droppable={{ container: rootListId, disabled: placedComponents.length > 0, callbacks: { onDragEnter: handleDragEnter, onDragLeave: handleDragLeave, onDrop: handleDrop }, attributes: { dragOverClass: 'dnd-field-over' } }}>
        {#if placedComponents.length}<div class="dnd-placed-list">{#each placedComponents as item, index (item.id)}
          <div data-parent-id="root" data-slot-index={index} class:dnd-insert-zone-edge={index === 0} class="dnd-insert-zone" use:droppable={{ container: rootListId, callbacks: { onDragOver: handleGapDragOver, onDragLeave: handleGapDragLeave, onDrop: (state) => handleGapDrop({ parentId: null, index }, state) } }}><i aria-hidden="true"></i><PlaygroundNewMarker /><PlaygroundMoveMarker /></div>
          <PlaygroundDraggable item={item} dragData={{ ...item, nodeId: item.id, sourceParentId: null }} container={rootListId} variant="placed" onDragEnd={handleDragEnd} onRemove={() => placedComponents = placedComponents.filter((existingItem) => existingItem.id !== item.id)} />
        {/each}
          <div data-parent-id="root" data-slot-index={placedComponents.length} class="dnd-insert-zone dnd-insert-zone-edge" use:droppable={{ container: rootListId, callbacks: { onDragOver: handleGapDragOver, onDragLeave: handleGapDragLeave, onDrop: (state) => handleGapDrop({ parentId: null, index: placedComponents.length }, state) } }}><i aria-hidden="true"></i><PlaygroundNewMarker /><PlaygroundMoveMarker /></div>
        </div>{/if}
        {#if newItemOver}<span class="dnd-new-drop-target" aria-hidden="true">+</span>{:else if !placedComponents.length}<span>Drop a component here</span>{/if}
      </div>
    </section>
  </div>
</div>
