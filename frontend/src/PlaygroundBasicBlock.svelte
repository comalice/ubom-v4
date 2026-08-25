<script lang="ts">
  import PlaygroundDraggable, { type ComponentNode } from './PlaygroundDraggable.svelte'

  export let item: ComponentNode
  export let parentId: number | null
  export let container: string
  export let title = item.label
  export let size: 'normal' | 'small' = 'normal'
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

<PlaygroundDraggable item={item} dragData={{ ...item, nodeId: item.id, sourceParentId: parentId }} {container} variant="placed" {size} {title} {onDragEnd} {onRemove} />
