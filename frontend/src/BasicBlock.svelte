<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import { type ComponentNode } from './DndDraggable.svelte'

  export let item: ComponentNode
  export let parentId: number | null
  export let container: string
  export let title = item.label
  export let size: 'normal' | 'small' = 'normal'
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

<BlockShell item={item} dragData={{ ...item, nodeId: item.id, sourceParentId: parentId }} {container} {size} {title} kind={item.blockType} {onDragEnd} {onRemove} />
