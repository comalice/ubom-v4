<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import { acceptsTaxonomyBranchDrop, taxonomyBranchContainer, taxonomyDraftBranchPath, type TaxonomyDragPayload, type TaxonomyDraftNode, type TaxonomyGrammar, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let parent: TaxonomyDraftNode
  export let path: string
  export let readOnly = false
  export let grammar: TaxonomyGrammar
  export let contextFieldsForNode: (targetId: string) => string[] = () => []
  export let targetFieldForNode: (targetId: string) => string | undefined = (targetId) => targetId
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, targetIndex: number) => void = () => {}
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex?: number) => void = () => {}
  export let onContextDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, field: string, targetIndex?: number) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}
  export let onBranchDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, targetIndex?: number) => void = () => {}
  export let onBranchFieldDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, field: string, targetIndex?: number) => void = () => {}

  $: categories = parent.children.filter((child) => taxonomyDraftBranchPath(child) === path)
  $: container = taxonomyBranchContainer(path)

  function drop(index: number) {
    return (state: DragDropState<TaxonomyDragPayload>) => onDrop(state, parent.id, path, index)
  }
</script>

<section class="taxonomy-branch-section" aria-label={`Branch ${path}`}>
  <div class="taxonomy-branch-heading">
    <span class="taxonomy-branch-eyebrow">Branch</span>
    <span class="taxonomy-branch-label">{path}</span>
  </div>
  <div class="taxonomy-branch-rail">
    {#if !readOnly}<DropZone variant="gap" mode="inactive" className="taxonomy-branch-gap" {container} accepts={(state) => acceptsTaxonomyBranchDrop(state.draggedItem)} onDrop={drop(0)} ariaLabel={`Insert category at the start of branch ${path}`} />{/if}
    {#each categories as child, index (child.id)}
      {#if child.placements[0]}
        <TaxonomyGrammarCategory
          placement={child.placements[0]}
          draftNode={child}
          parentContainer={container}
          {readOnly}
          {onChildDrop}
          {onContextDrop}
          {onRuleDrop}
          {onRemove}
          {onDragEnd}
          {onAttributes}
          {onNodeLabelChange}
          {onNodeIdChange}
          {onRulePredicateChange}
          contextFields={contextFieldsForNode(path)}
          {contextFieldsForNode}
          {targetFieldForNode}
          {grammar}
          {onBranchDrop}
          {onBranchFieldDrop}
        />
      {/if}
      {#if !readOnly}<DropZone variant="gap" mode="inactive" className="taxonomy-branch-gap" {container} accepts={(state) => acceptsTaxonomyBranchDrop(state.draggedItem)} onDrop={drop(index + 1)} ariaLabel={`Insert category after ${child.label} in branch ${path}`} />{/if}
    {/each}
  </div>
</section>
