<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import TaxonomyCategoryChildren from './TaxonomyCategoryChildren.svelte'
  import TaxonomyRuleBlock from './TaxonomyRuleBlock.svelte'
  import DropZone from './DropZone.svelte'
  import TaxonomyChoiceBranches from './TaxonomyChoiceBranches.svelte'
  import { acceptsTaxonomyRuleDrop, taxonomyBranchContainer, taxonomyGrammarNode, taxonomyTargetContainer, type TaxonomyDragPayload, type TaxonomyDraftNode, type TaxonomyGrammar, type TaxonomyPlacement, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let placement: TaxonomyPlacement
  export let draftNode: TaxonomyDraftNode | undefined
  export let parentContainer = ''
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex?: number) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}
  export let contextFields: string[] = []
  export let onContextDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, field: string, targetIndex?: number) => void = () => {}
  export let readOnly = false
  export let contextFieldsForNode: (targetId: string) => string[] = () => []
  export let targetFieldForNode: (targetId: string) => string | undefined = (targetId) => targetId
  export let grammar: TaxonomyGrammar = { id: '', label: '', fields: 0, nodes: [] }
  export let onBranchDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string) => void = () => {}
  export let onBranchFieldDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, field: string) => void = () => {}

  $: nodeID = draftNode?.id ?? placement.nodeId ?? placement.id
  $: nodeContainer = draftNode ? taxonomyTargetContainer(draftNode.id) : ''
  $: targetMessage = draftNode?.targetResolution?.status === 'ambiguous'
    ? `Ambiguous ${draftNode.targetResolution.field} target: choose a branch occurrence`
    : draftNode?.targetResolution?.status === 'unresolved'
      ? `No grammar occurrence found for ${draftNode.targetResolution.field}`
      : ''
  $: progressiveContextFields = draftNode?.branch && placement.predicateData?.field
    ? (taxonomyGrammarNode(grammar, draftNode.targetId)?.capturedFields ?? []).slice(Math.max(0, (taxonomyGrammarNode(grammar, draftNode.targetId)?.capturedFields ?? []).indexOf(placement.predicateData.field) + 1))
    : contextFields
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
  dragData={{ kind: 'taxonomy', taxonomyType: 'category', placementId: placement.id, sourceNodeId: nodeID }}
  showRuleDropZone
  suppressLegacyRuleDropZone
  showChildren
  suppressChildDropZone
  childContainer={nodeContainer}
  onChildDrop={draftNode ? (state, _parentID, targetIndex) => onChildDrop(state, draftNode.id, targetIndex) : undefined}
  {onDragEnd}
  onRemove={() => onRemove(placement.id)}
  onAttributes={() => { if (draftNode) onAttributes(draftNode.id) }}
  onLabelChange={draftNode ? (label) => onNodeLabelChange(draftNode.id, label) : undefined}
  onNodeIdChange={draftNode ? (nodeId) => onNodeIdChange(draftNode.id, nodeId) : undefined}
>
  {#if targetMessage}<p class="taxonomy-target-warning" role="status">{targetMessage}</p>{/if}
  <svelte:fragment slot="context">
    {#if draftNode}
      <TaxonomyChoiceBranches {grammar} targetId={draftNode.targetId} {readOnly} onBranchDrop={(state, branchPath) => onBranchDrop(state, draftNode!.id, branchPath)} onBranchFieldDrop={(state, branchPath, field) => onBranchFieldDrop(state, draftNode!.id, branchPath, field)}><svelte:fragment slot="branch" let:path>{#each draftNode.children.filter((child) => child.branch?.path[child.branch.path.length - 1] === path) as child (child.id)}{#if child.placements[0]}<TaxonomyGrammarCategory placement={child.placements[0]} draftNode={child} parentContainer={taxonomyBranchContainer(path)} contextFields={contextFieldsForNode(path)} {readOnly} onContextDrop={onContextDrop} onChildDrop={onChildDrop} onRuleDrop={onRuleDrop} onRemove={onRemove} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} {grammar} {onBranchDrop} {onBranchFieldDrop} />{/if}{/each}</svelte:fragment></TaxonomyChoiceBranches>
    {/if}
    {#if draftNode && progressiveContextFields.length}
      <div class="taxonomy-category-context">
        <span class="taxonomy-category-context-label">Grammar context</span>
        <div class="taxonomy-category-context-lanes">
          {#each progressiveContextFields as field}
            <div class="taxonomy-category-context-lane">
              <TaxonomyCategoryChildren parent={draftNode} field={field} {readOnly} onDrop={onChildDrop} onContextDrop={onContextDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} {contextFieldsForNode} {targetFieldForNode} {grammar} {onBranchDrop} {onBranchFieldDrop} />
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </svelte:fragment>
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
      <TaxonomyCategoryChildren parent={draftNode} excludedFields={contextFields} {readOnly} onDrop={onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} {onContextDrop} {contextFieldsForNode} {targetFieldForNode} {grammar} {onBranchDrop} {onBranchFieldDrop} />
    {/if}
  </svelte:fragment>
</TaxonomyCategoryBlock>
