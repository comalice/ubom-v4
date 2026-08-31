<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'
  import DropZone from './DropZone.svelte'
  import { acceptsTaxonomyRuleDrop, taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyDraftNode, type TaxonomyPlacement, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let placement: TaxonomyPlacement
  export let draftNode: TaxonomyDraftNode | undefined
  export let parentContainer = ''
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}
  export let readOnly = false

  $: nodeID = draftNode?.id ?? placement.nodeId ?? placement.id
  $: nodeContainer = draftNode ? taxonomyTargetContainer(draftNode.id) : ''
</script>

<TaxonomyCategoryBlock
  title={placement.title}
  summary={placement.summary}
  label={placement.label ?? placement.title}
  nodeId={nodeID}
  predicate={placement.predicate ?? ''}
  layout="tree"
  {readOnly}
  container={parentContainer}
  dragData={{ kind: 'taxonomy', taxonomyType: 'category', placementId: placement.id }}
  showRuleDropZone
  suppressLegacyRuleDropZone
  showChildren
  childContainer={nodeContainer}
  onChildDrop={draftNode ? (state) => onChildDrop(state, draftNode.id) : undefined}
  {onDragEnd}
  onRemove={() => onRemove(placement.id)}
  onAttributes={() => { if (draftNode) onAttributes(draftNode.id) }}
  onLabelChange={draftNode ? (label) => onNodeLabelChange(draftNode.id, label) : undefined}
  onNodeIdChange={draftNode ? (nodeId) => onNodeIdChange(draftNode.id, nodeId) : undefined}
>
  <svelte:fragment slot="rules">
    {#if draftNode}
      {#each draftNode.rules as rule, index (rule.id)}
        {#if !readOnly}<DropZone variant="gap" mode="inactive" container={nodeContainer} accepts={(state) => acceptsTaxonomyRuleDrop(state.draggedItem, draftNode.id)} onDrop={(state) => onRuleDrop(state, draftNode.id, index)} ariaLabel={`Insert rule ${index + 1} for ${draftNode.label}`} />{/if}
        <TaxonomyRuleBlock
          title={rule.title}
          field={draftNode.targetId}
          predicate={rule.predicateData ?? { kind: 'exact', field: draftNode.targetId, value: 'A' }}
          readOnly={readOnly}
          container={nodeContainer}
          dragData={{ kind: 'taxonomy', taxonomyType: 'rule', placementId: rule.id, sourceNodeId: draftNode.id }}
          {onDragEnd}
          onRemove={() => onRemove(rule.id)}
          onPredicateChange={(predicate) => onRulePredicateChange(rule.id, predicate)}
        />
      {/each}
      {#if !readOnly}<DropZone variant="gap" mode="inactive" container={nodeContainer} accepts={(state) => acceptsTaxonomyRuleDrop(state.draggedItem, draftNode.id)} onDrop={(state) => onRuleDrop(state, draftNode.id, draftNode.rules.length)} ariaLabel={`Insert rule after ${draftNode.label}`} />{/if}
    {/if}
  </svelte:fragment>
  <svelte:fragment slot="children">
    {#if draftNode}
      {#each draftNode.children as child (child.id)}
        {#if child.placements[0]}
          <svelte:self
            placement={child.placements[0]}
            draftNode={child}
            parentContainer={nodeContainer}
            {readOnly}
            {onChildDrop}
            {onRuleDrop}
            {onRemove}
            {onDragEnd}
            {onAttributes}
            {onNodeLabelChange}
            {onNodeIdChange}
            {onRulePredicateChange}
          />
        {/if}
      {/each}
    {/if}
  </svelte:fragment>
</TaxonomyCategoryBlock>
