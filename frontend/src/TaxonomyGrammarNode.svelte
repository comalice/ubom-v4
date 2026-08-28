<script lang="ts">
  import TaxonomyGrammarDropTarget from './TaxonomyGrammarDropTarget.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import type { TaxonomyDraftNode, TaxonomyDragPayload, TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let depth = 0
  export let placements: TaxonomyPlacement[] = []
  export let taxonomyDraft: TaxonomyDraftNode | undefined
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (label: string) => void = () => {}
</script>

<li>
  <span class="dnd-structure-node">{#if node.blockType === 'binding'}<span class="taxonomy-grammar-kind">binding</span>{/if}{node.label}<small>{node.summary}</small></span>
    <TaxonomyGrammarDropTarget {node} {taxonomyDraft} placements={placements.filter((placement) => placement.targetId === String(node.id))} {onDrop} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} />
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
        <svelte:self node={child} depth={depth + 1} {placements} {taxonomyDraft} {onDrop} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} />
      {/each}
    </ul>
  {/if}
</li>
