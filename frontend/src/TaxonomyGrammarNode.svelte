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
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: import('./TaxonomyEditorModel').TaxonomyRulePredicate) => void = () => {}
  export let readOnly = false
</script>

<li>
  <span class="dnd-structure-node">{#if node.blockType === 'binding'}<span class="taxonomy-grammar-kind">binding</span>{/if}{node.label}<small>{node.summary}</small></span>
      <TaxonomyGrammarDropTarget {node} {taxonomyDraft} placements={placements.filter((placement) => placement.targetId === String(node.id))} {readOnly} {onDrop} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} />
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
         <svelte:self node={child} depth={depth + 1} {placements} {taxonomyDraft} {readOnly} {onDrop} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} />
      {/each}
    </ul>
  {/if}
</li>
