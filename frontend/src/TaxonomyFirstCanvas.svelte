<script lang="ts">
  import type { DragDropState } from '@thisux/sveltednd'
  import TaxonomyCategoryChildren from './TaxonomyCategoryChildren.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyChoiceBranches from './TaxonomyChoiceBranches.svelte'
  import TaxonomyBranchSection from './TaxonomyBranchSection.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import { addTaxonomyDraftBranchChild, addTaxonomyDraftBranchFieldChild, addTaxonomyDraftChild, addTaxonomyDraftRule, findTaxonomyDraftNode, findTaxonomyDraftPath, moveTaxonomyDraftCategory, moveTaxonomyDraftRule, removeTaxonomyDraftCategorySubtree, removeTaxonomyDraftRule, taxonomyContextFields, taxonomyDraftBranchPath, taxonomyGrammarFieldHasChoiceAlternatives, taxonomyTargetContainer, taxonomyTargetField, taxonomyTargetIsUsable, updateTaxonomyDraftNode, updateTaxonomyDraftRule, type TaxonomyDraftNode, type TaxonomyDragPayload, type TaxonomyGrammar, type TaxonomyPlacement, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let grammar: TaxonomyGrammar
  export let taxonomyDraft: TaxonomyDraftNode
  export let readOnly = false
  export let onDraftChange: (draft: TaxonomyDraftNode) => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let onNodeLabelChange: (nodeID: string, label: string) => void = () => {}
  export let onNodeIdChange: (nodeID: string, newNodeID: string) => void = () => {}
  export let onRulePredicateChange: (ruleID: string, predicate: TaxonomyRulePredicate) => void = () => {}

  let placementNumber = 1
  let selectedNodeID = ''
  $: selectedNode = findTaxonomyDraftNode(taxonomyDraft, selectedNodeID) ?? taxonomyDraft
  $: selectedPath = findTaxonomyDraftPath(taxonomyDraft, selectedNode.id)
  $: effectivePredicate = selectedPath.slice(1).map((node) => node.placements.find((placement) => placement.type === 'category')?.predicate).filter(Boolean).join(' AND ')
  $: rootFields = taxonomyContextFields(grammar, taxonomyDraft.targetId)
  $: firstField = rootFields[0]
  $: firstTargetField = firstField ? (taxonomyTargetField(grammar, firstField) ?? firstField) : undefined
  $: suppressRootFieldDropZones = firstField ? taxonomyGrammarFieldHasChoiceAlternatives(grammar, taxonomyDraft.targetId, firstField) : false

  function change(next: TaxonomyDraftNode) { onDraftChange(next) }

  function newPlacement(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, targetId: string): TaxonomyPlacement {
    const id = `taxonomy-${payload.taxonomyType}-${placementNumber++}`
    const predicate = `${targetId} = A`
    return payload.taxonomyType === 'category'
      ? { id, targetId, type: 'category', title: 'New category', summary: `Attached to ${targetId}`, label: 'New category', nodeId: id, predicate, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
      : { id, targetId, type: 'rule', title: 'New rule', summary: predicate, predicate, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
  }

  function addCategory(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, parentID: string, targetId: string, targetIndex?: number) {
    const placement = newPlacement(payload, targetId)
    change(addTaxonomyDraftChild(taxonomyDraft, parentID, { id: placement.id, label: placement.label ?? placement.title, targetId, placements: [placement], rules: [], children: [] }, targetIndex))
    selectedNodeID = placement.id
  }

  function addBranchCategory(state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, targetIndex?: number) {
    if (state.draggedItem?.kind !== 'palette' || state.draggedItem.taxonomyType !== 'category') return
    const id = `taxonomy-category-${placementNumber++}`
    change(addTaxonomyDraftBranchChild(taxonomyDraft, parentID, id, branchPath, 'New category', targetIndex))
    selectedNodeID = id
  }

  function addBranchFieldCategory(state: DragDropState<TaxonomyDragPayload>, parentID: string, branchPath: string, field: string, targetIndex?: number) {
    if (state.draggedItem?.kind !== 'palette' || state.draggedItem.taxonomyType !== 'category') return
    const id = `taxonomy-category-${placementNumber++}`
    change(addTaxonomyDraftBranchFieldChild(taxonomyDraft, parentID, id, branchPath, field, 'New category', targetIndex))
    selectedNodeID = id
  }

  function handleFieldDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string, field: string, targetIndex?: number) {
    const item = state.draggedItem
    if (item?.kind === 'palette' && item.taxonomyType === 'category') addCategory(item, parentID, field, targetIndex)
    else if (item?.kind === 'taxonomy' && item.taxonomyType === 'category' && item.sourceNodeId) change(moveTaxonomyDraftCategory(taxonomyDraft, item.sourceNodeId, parentID, field, targetIndex))
  }

  function handleChildDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex?: number) {
    const item = state.draggedItem
    const parent = findTaxonomyDraftNode(taxonomyDraft, parentID)
    if (!item || !parent || !taxonomyTargetIsUsable(parent.targetId)) return
    const targetField = taxonomyTargetField(grammar, parent.targetId) ?? parent.targetId
    if (item.kind === 'palette' && item.taxonomyType === 'category') addCategory(item, parentID, targetField, targetIndex)
    else if (item.kind === 'taxonomy' && item.taxonomyType === 'category' && item.sourceNodeId) change(moveTaxonomyDraftCategory(taxonomyDraft, item.sourceNodeId, parentID, targetField, targetIndex))
  }

  function handleRuleDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) {
    const item = state.draggedItem
    const parent = findTaxonomyDraftNode(taxonomyDraft, parentID)
    if (!item || !parent || !taxonomyTargetIsUsable(parent.targetId)) return
    const targetField = taxonomyTargetField(grammar, parent.targetId) ?? parent.targetId
    if (item.kind === 'palette' && item.taxonomyType === 'rule') change(addTaxonomyDraftRule(taxonomyDraft, parentID, newPlacement(item, targetField), targetIndex))
    else if (item.kind === 'taxonomy' && item.taxonomyType === 'rule' && item.sourceNodeId) change(moveTaxonomyDraftRule(taxonomyDraft, item.sourceNodeId, parentID, item.placementId, targetIndex))
  }

  function removePlacement(placementID: string) {
    const categoryNode = findTaxonomyDraftNode(taxonomyDraft, placementID)
    const categoryPlacement = categoryNode?.placements.find((placement) => placement.id === placementID)
    if (categoryPlacement?.type === 'category' && categoryNode) {
      const selectedCategory = findTaxonomyDraftPath(taxonomyDraft, selectedNodeID).some((node) => node === categoryNode)
      const path = findTaxonomyDraftPath(taxonomyDraft, categoryNode.id)
      change(removeTaxonomyDraftCategorySubtree(taxonomyDraft, placementID))
      if (selectedCategory) selectedNodeID = path.length > 1 ? path[path.length - 2].id : taxonomyDraft.id
    } else change(removeTaxonomyDraftRule(taxonomyDraft, placementID))
  }

  function openAttributes(event: MouseEvent, nodeID: string) { event.stopPropagation(); onAttributes(nodeID) }
  function selectCategory(event: MouseEvent, fallbackNodeID: string) { if (!(event.target as HTMLElement).closest('.block-shell-remove')) selectedNodeID = fallbackNodeID }
</script>

<section class="taxonomy-first-workflow taxonomy-editor">
  {#if !readOnly}<section class="taxonomy-first-authoring-row" aria-label="Authoring blocks"><TaxonomyPalette /></section>{/if}
  <main class="taxonomy-editor-canvas">
    <div class="taxonomy-editor-heading"><div><span class="mock-section-label">Taxonomy tree</span><h3>{taxonomyDraft.label}</h3></div><span class="taxonomy-editor-count">{grammar.id} · {grammar.fields} fields</span></div>
    <p class="taxonomy-grammar-note">The taxonomy tree is the work surface. Categories expose later grammar fields as context lanes.</p>
    <div class="taxonomy-grammar-tree">
      <div class:selected={selectedNodeID === taxonomyDraft.id} class="taxonomy-first-root"><span class="taxonomy-grammar-kind">root scope</span><strong>{taxonomyDraft.label}</strong><small>all values matching {grammar.id}</small>{#if !readOnly}<button class="primitive-button primitive-button-secondary button-size-small" type="button" aria-label={`View attributes for ${taxonomyDraft.label}`} on:click={(event) => openAttributes(event, taxonomyDraft.id)}>attributes</button>{/if}<div class="taxonomy-first-root-context"><TaxonomyChoiceBranches {grammar} targetId={taxonomyDraft.targetId} {readOnly} onBranchDrop={(state, path, index) => addBranchCategory(state, taxonomyDraft.id, path, index)} onBranchFieldDrop={(state, path, field, index) => addBranchFieldCategory(state, taxonomyDraft.id, path, field, index)}><svelte:fragment slot="branch" let:path><TaxonomyBranchSection parent={taxonomyDraft} {path} {grammar} {readOnly} contextFieldsForNode={(targetId) => taxonomyContextFields(grammar, targetId)} targetFieldForNode={(targetId) => taxonomyTargetField(grammar, targetId)} onDrop={addBranchCategory} onChildDrop={handleChildDrop} onContextDrop={handleFieldDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} onBranchDrop={addBranchCategory} onBranchFieldDrop={addBranchFieldCategory} /></svelte:fragment></TaxonomyChoiceBranches>{#if firstField && !suppressRootFieldDropZones}<div class="taxonomy-first-context-lane"><TaxonomyCategoryChildren parent={taxonomyDraft} field={firstField} {readOnly} onDrop={handleChildDrop} onContextDrop={handleFieldDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} contextFieldsForNode={(targetId) => taxonomyContextFields(grammar, targetId)} targetFieldForNode={(targetId) => taxonomyTargetField(grammar, targetId)} {grammar} onBranchDrop={addBranchCategory} onBranchFieldDrop={addBranchFieldCategory} /></div>{/if}</div></div>
      <div class="taxonomy-first-branches taxonomy-grammar-structure-tree">{#each taxonomyDraft.children.filter((child) => !taxonomyDraftBranchPath(child) && taxonomyTargetField(grammar, child.targetId) !== firstTargetField) as child (child.id)}{#if child.placements[0]}<div class:selected={selectedNodeID === child.id} class="taxonomy-first-category" role="button" tabindex="0" on:click={(event) => selectCategory(event, child.id)} on:keydown={(event) => event.key === 'Enter' && (selectedNodeID = child.id)}><TaxonomyGrammarCategory placement={child.placements[0]} draftNode={child} parentContainer={taxonomyTargetContainer(taxonomyDraft.id)} contextFields={taxonomyContextFields(grammar, child.targetId)} contextFieldsForNode={(targetId) => taxonomyContextFields(grammar, targetId)} targetFieldForNode={(targetId) => taxonomyTargetField(grammar, targetId)} {grammar} {readOnly} onContextDrop={handleFieldDrop} onChildDrop={handleChildDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} {onAttributes} {onNodeLabelChange} {onNodeIdChange} {onRulePredicateChange} onBranchDrop={addBranchCategory} onBranchFieldDrop={addBranchFieldCategory} /></div>{/if}{/each}</div>
    </div>
    {#if readOnly}<PrimitiveButton variant="secondary" size="small" type="button" on:click={() => onAttributes(taxonomyDraft.id)}>attributes</PrimitiveButton>{/if}
    <div class="taxonomy-first-effective"><span class="mock-section-label">Effective condition for {selectedNode.label}</span><code>{effectivePredicate || 'any valid part number'}</code><small>Inherited from the taxonomy path.</small></div>
  </main>
</section>

<style>
  .taxonomy-first-workflow { display: grid; gap: 1rem; }
  .taxonomy-first-authoring-row { display: grid; gap: 1rem; background: var(--ground); }
  .taxonomy-first-root { display: grid; gap: .15rem; padding: .6rem .55rem; border-bottom: 1px solid var(--hairline); }
  .taxonomy-first-root.selected { color: var(--accent); background: var(--accent-light); }
  .taxonomy-first-root small { color: var(--muted); font: .65rem ui-monospace, monospace; }
  .taxonomy-first-root-context { display: grid; margin-top: .35rem; }
  .taxonomy-first-context-lane { display: grid; gap: .45rem; min-width: 0; }
  .taxonomy-first-branches { display: grid; gap: .5rem; padding: .5rem 0 0 1.15rem; border-left: 1px solid var(--hairline); }
  .taxonomy-first-category { padding: .15rem; border: 1px solid transparent; border-radius: .25rem; cursor: pointer; }
  .taxonomy-first-category:hover, .taxonomy-first-category.selected { border-color: var(--accent); background: var(--accent-light); }
  .taxonomy-first-effective { display: grid; gap: .35rem; margin-top: .7rem; padding: .7rem; color: var(--surface); background: var(--ink); border-radius: .2rem; }
  .taxonomy-first-effective .mock-section-label { color: var(--muted); }
  .taxonomy-first-effective code { color: var(--surface); font: .78rem ui-monospace, monospace; }
  .taxonomy-first-effective small { color: var(--muted); font: .65rem ui-monospace, monospace; }
</style>
