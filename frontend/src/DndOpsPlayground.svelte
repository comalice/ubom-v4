<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import type { DropSlot } from './DndComponentList.svelte'
  import DndDraggable, { type DraggableComponent } from './DndDraggable.svelte'
  import DndField from './DndField.svelte'
  import type { ComponentNode } from './DndDraggable.svelte'
  import StructureTree from './StructureTree.svelte'

  type DragItem = DraggableComponent
  const rootListId = 'component-list:root'
  const component: DraggableComponent = { kind: 'component', label: 'Literal', blockType: 'basic' }
  const nestedComponent: DraggableComponent = { kind: 'component', label: 'Container', blockType: 'nested' }
  let placedComponents: ComponentNode[] = []
  let nextComponentId = 1
  const markerState = { version: 0 }

  function handleDrop(state: DragDropState<DragItem>) {
    if (state.draggedItem?.kind === 'component') placedComponents = [...placedComponents, { ...state.draggedItem, id: nextComponentId++, children: [] }]
  }
  function handleGapDragOver(state: DragDropState<DragItem>) {
    if (state.draggedItem?.kind !== 'component') return
    markerState.version += 1
    const target = (state.targetElement as HTMLElement | null)?.closest('.dnd-insert-zone')
    const markerClass = state.sourceContainer.startsWith('component-list:') ? 'dnd-insert-zone-move' : 'dnd-insert-zone-new'
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
      if (state.sourceContainer.startsWith('component-list:') && state.draggedItem.nodeId) {
        const sourceParentId = state.draggedItem.sourceParentId ?? null
        const sourceList = findChildren(placedComponents, sourceParentId)
        const sourceIndex = sourceList.findIndex((item) => item.id === state.draggedItem.nodeId)
        if (sourceIndex >= 0) {
          const movingItem = sourceList[sourceIndex]
          if (containsNode(movingItem, slot.parentId)) {
            clearDropMarker()
            return
          }
          const remaining = removeNode(placedComponents, movingItem.id)
          const insertIndex = sourceParentId === slot.parentId && sourceIndex < slot.index ? slot.index - 1 : slot.index
          placedComponents = insertNode(remaining, slot.parentId, insertIndex, movingItem)
        }
      } else {
        placedComponents = insertNode(placedComponents, slot.parentId, slot.index, { ...state.draggedItem, id: nextComponentId++, children: [] })
      }
    }
    clearDropMarker()
  }
  function clearDropMarker() {
    document.querySelectorAll('.dnd-ops-playground .dnd-insert-zone-active').forEach((element) => element.classList.remove('dnd-insert-zone-active', 'dnd-insert-zone-new', 'dnd-insert-zone-move'))
  }
  function handleDragEnd() {
    clearDropMarker()
  }
  function findChildren(nodes: ComponentNode[], parentId: number | null): ComponentNode[] {
    if (parentId === null) return nodes
    for (const node of nodes) {
      if (node.id === parentId) return node.children
      const children = findChildren(node.children, parentId)
      if (children.length) return children
    }
    return []
  }
  function removeNode(nodes: ComponentNode[], nodeId: number): ComponentNode[] {
    return nodes.filter((node) => node.id !== nodeId).map((node) => ({ ...node, children: removeNode(node.children, nodeId) }))
  }
  function insertNode(nodes: ComponentNode[], parentId: number | null, index: number, node: ComponentNode): ComponentNode[] {
    if (parentId === null) return [...nodes.slice(0, index), node, ...nodes.slice(index)]
    return nodes.map((item) => item.id === parentId
      ? { ...item, children: [...item.children.slice(0, index), node, ...item.children.slice(index)] }
      : { ...item, children: insertNode(item.children, parentId, index, node) })
  }
  function containsNode(node: ComponentNode, nodeId: number | null): boolean {
    return nodeId !== null && (node.id === nodeId || node.children.some((child) => containsNode(child, nodeId)))
  }
</script>

<div class="dnd-ops-playground">
  <section class="dnd-structure-preview" aria-labelledby="structure-preview-heading">
    <span class="mock-section-label" id="structure-preview-heading">Current structure</span>
    <StructureTree nodes={placedComponents} />
  </section>
  <div class="dnd-ops-heading"><div><span class="mock-section-label">Experiment</span><h3>Drop components into a field</h3></div>{#if placedComponents.length}<button class="dnd-reset" type="button" on:click={() => placedComponents = []}>Reset</button>{/if}</div>
  <div class="dnd-ops-stage">
    <section class="dnd-component-palette" aria-labelledby="component-palette-heading">
      <span class="mock-section-label" id="component-palette-heading">Component palette</span>
      <DndDraggable item={component} container="component-palette" variant="palette" description="Leaf component" onDragEnd={handleDragEnd} />
      <DndDraggable item={nestedComponent} container="component-palette" variant="palette" description="Component with nested field" onDragEnd={handleDragEnd} />
    </section>

    <section class="dnd-field-area" aria-labelledby="field-heading">
      <span class="mock-section-label" id="field-heading">Field</span>
      <DndField nodes={placedComponents} container={rootListId} onDrop={handleDrop} onDragOver={handleGapDragOver} onGapDragLeave={handleGapDragLeave} onGapDrop={handleGapDrop} onDragEnd={handleDragEnd} onRemove={(id) => placedComponents = removeNode(placedComponents, id)} />
    </section>
  </div>
</div>
