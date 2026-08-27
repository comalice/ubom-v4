
<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import type { DropSlot } from './DndComponentList.svelte'
  import DndDraggable, { type DraggableComponent } from './DndDraggable.svelte'
  import DndField from './DndField.svelte'
  import DropZone from './DropZone.svelte'
  import type { ComponentNode } from './DndDraggable.svelte'
  import StructureTree from './StructureTree.svelte'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import NewDropMarker from './NewDropMarker.svelte'
  import InactiveDropMarker from './InactiveDropMarker.svelte'


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
<p class="eyebrow">DnD operations</p>
<h2>DnD operations</h2>
<section class="playground-components" aria-labelledby="playground-components-heading">
  <p class="mock-section-label" id="playground-components-heading">Components</p>
  <div class="playground-component-list">
    <div class="playground-component-item"><NewDropMarker /><span><strong>New item</strong><small>Creates a new component at the target</small></span></div>
    <div class="playground-component-item"><MoveDropMarker /><span><strong>Moving item</strong><small>Moves an existing component to the target</small></span></div>
    <div class="playground-component-item"><InactiveDropMarker /><span><strong>Inactive item</strong><small>Inactive drop zone.</small></span>
  </div>
  </div>
</section>

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
  <section class="dnd-drop-zone-demo" aria-labelledby="drop-zone-demo-heading">
    <span class="mock-section-label" id="drop-zone-demo-heading">Drop-zone states</span>
    <div class="dnd-drop-zone-demo-grid">
      <DropZone variant="slot" mode="inactive" className="dnd-static-drop-zone" label="Inactive target" showHint showLabel />
      <DropZone variant="slot" mode="new" className="dnd-static-drop-zone" label="New item target" showLabel />
      <DropZone variant="slot" mode="move" className="dnd-static-drop-zone" label="Move item target" showLabel />
      <DropZone variant="filled" mode="move" className="dnd-static-drop-zone" label="Filled target" showLabel>
        <span class="dnd-static-drop-zone-content">Existing component</span>
      </DropZone>
    </div>
  </section>
</div>
