<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import { acceptsTaxonomyCategoryChildDrop, taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyDraftNode, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let parent: TaxonomyDraftNode
  export let children: TaxonomyDraftNode[] | undefined = undefined
  export let field: string | undefined = undefined
  export let excludedFields: string[] = []
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

  $: sourceChildren = children ?? parent.children
  $: visibleChildren = sourceChildren.filter((child) => field === undefined
    ? !excludedFields.includes(child.targetId)
    : child.targetId === field)
  $: container = taxonomyTargetContainer(field === undefined ? parent.id : `${parent.id}:field:${field}`)

  function childIndex(child: TaxonomyDraftNode) {
    return sourceChildren.indexOf(child)
  }

  function accepts(state: DragDropState<TaxonomyDragPayload>) {
    return acceptsTaxonomyCategoryChildDrop(state.draggedItem, parent.id)
  }

  function drop(index: number) {
    return (state: DragDropState<TaxonomyDragPayload>) => field === undefined
      ? onDrop?.(state, parent.id, index)
      : onContextDrop(state, parent.id, field, index)
  }

</script>

{#if !readOnly}
  <DropZone variant="gap" mode="inactive" className="taxonomy-child-drop-zone" container={container} accepts={accepts} onDrop={drop(visibleChildren.length ? childIndex(visibleChildren[0]) : sourceChildren.length)} ariaLabel={field ? `Drop a category on ${field} for ${parent.label}` : 'Drop a child category'} />
{/if}
{#each visibleChildren as child, index (child.id)}
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
      contextFields={field === undefined ? [] : (child.targetId === 'family' ? ['group', 'uid'] : ['uid'])}
      {onContextDrop}
    />
  {/if}
  {#if !readOnly}
    <DropZone variant="gap" mode="inactive" className="taxonomy-child-drop-zone" container={container} accepts={accepts} onDrop={drop(childIndex(child) + 1)} ariaLabel={`Insert a category after ${child.label}`} />
  {/if}
{/each}
