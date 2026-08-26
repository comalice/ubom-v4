<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import type { ComponentNode } from './DndDraggable.svelte'

  export let item: ComponentNode
  export let parentId: number | null
  export let container: string
  export let collapsible = false
  export let initiallyOpen = true
  export let onDragEnd: () => void
  export let onRemove: () => void
</script>

<BlockShell item={item} dragData={{ ...item, nodeId: item.id, sourceParentId: parentId }} {container} title={item.label} kind={item.blockType} {collapsible} {initiallyOpen} {onDragEnd} {onRemove}>
  <div class="dnd-nested-list">
  <span class="mock-section-label">Nested components</span>
  <slot />
  </div>
</BlockShell>
