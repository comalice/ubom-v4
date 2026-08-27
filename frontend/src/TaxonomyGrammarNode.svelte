<script lang="ts">
  import TaxonomyGrammarDropTarget from './TaxonomyGrammarDropTarget.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import type { TaxonomyDragPayload, TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let depth = 0
  export let placements: TaxonomyPlacement[] = []
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
</script>

<li>
  <span class="dnd-structure-node">{node.label}<small>{node.summary}</small></span>
  <TaxonomyGrammarDropTarget {node} placements={placements.filter((placement) => placement.targetId === String(node.id))} {onDrop} {onRemove} {onDragEnd} />
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
        <svelte:self node={child} depth={depth + 1} {placements} {onDrop} {onRemove} {onDragEnd} />
      {/each}
    </ul>
  {/if}
</li>
