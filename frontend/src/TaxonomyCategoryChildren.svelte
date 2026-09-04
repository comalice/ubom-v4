<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import { acceptsTaxonomyCategoryChildDrop, taxonomyTargetContainer, taxonomyTargetIsUsable, type TaxonomyDragPayload, type TaxonomyDraftNode, type TaxonomyGrammar, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let parent: TaxonomyDraftNode
  export let field: string | undefined = undefined
  export let excludedFields: string[] = []
  export let suppressDropZones = false
  export let suppressLabel = false
  export let readOnly = false
  export let onDrop: ((state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) => void) | undefined
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}
  export let onContextDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, field: string, targetIndex: number) => void = () => {}
  export let contextFieldsForNode: (targetId: string) => string[] = () => []
  export let targetFieldForNode: (targetId: string) => string | undefined = (targetId) => targetId
  export let grammar: TaxonomyGrammar = { id: '', label: '', fields: 0, nodes: [] }
  export let onBranchDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string) => void = () => {}
  export let onBranchFieldDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, field: string) => void = () => {}

  $: sourceChildren = parent.children
  $: normalizedExcludedFields = excludedFields.map((targetId) => targetFieldForNode(targetId) ?? targetId)
  $: visibleChildren = sourceChildren.filter((child) => field === undefined
    ? !normalizedExcludedFields.includes(targetFieldForNode(child.targetId) ?? child.targetId)
    : targetFieldForNode(child.targetId) === field)
  $: container = taxonomyTargetContainer(field === undefined ? parent.id : `${parent.id}:field:${field}`)

  function childIndex(child: TaxonomyDraftNode) {
    return sourceChildren.indexOf(child)
  }

  function accepts(state: DragDropState<TaxonomyDragPayload>) {
    return (field !== undefined || taxonomyTargetIsUsable(parent.targetId)) && acceptsTaxonomyCategoryChildDrop(state.draggedItem, parent.id)
  }

  function drop(index: number) {
    return (state: DragDropState<TaxonomyDragPayload>) => field === undefined
      ? onDrop?.(state, parent.id, index)
      : onContextDrop(state, parent.id, field, index)
  }

</script>

<div class="taxonomy-category-children">
  {#if !suppressLabel}<span class="taxonomy-category-children-label">{field ?? 'Child categories'}</span>{/if}
  {#if !suppressDropZones && !readOnly && (field !== undefined || taxonomyTargetIsUsable(parent.targetId))}
    <DropZone variant="gap" mode="inactive" className="taxonomy-child-drop-zone" container={container} accepts={accepts} onDrop={drop(visibleChildren.length ? childIndex(visibleChildren[0]) : sourceChildren.length)} ariaLabel={field ? `Drop a category on ${field} for ${parent.label}` : 'Drop a child category'} />
  {/if}
  {#each visibleChildren as child (child.id)}
    {#if child.placements[0]}
      <TaxonomyGrammarCategory
        placement={child.placements[0]}
        draftNode={child}
        parentContainer={container}
        {readOnly}
        onChildDrop={onDrop}
        {onRuleDrop}
        {onRemove}
        {onDragEnd}
        {onAttributes}
        {onNodeLabelChange}
        {onNodeIdChange}
        {onRulePredicateChange}
        {grammar}
        {onBranchDrop}
        {onBranchFieldDrop}
        contextFields={field === undefined ? [] : contextFieldsForNode(child.targetId)}
        {onContextDrop}
        {contextFieldsForNode}
        {targetFieldForNode}
      />
    {/if}
    {#if !suppressDropZones && !readOnly && (field !== undefined || taxonomyTargetIsUsable(parent.targetId))}
      <DropZone variant="gap" mode="inactive" className="taxonomy-child-drop-zone" container={container} accepts={accepts} onDrop={drop(childIndex(child) + 1)} ariaLabel={`Insert a category after ${child.label}`} />
    {/if}
  {/each}
</div>
