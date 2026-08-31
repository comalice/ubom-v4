<script lang="ts">
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import SlideOut from './SlideOut.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyAttributeTableRow from './TaxonomyAttributeTableRow.svelte'
  import { addTaxonomyAttribute, addTaxonomyDraftChild, addTaxonomyDraftRule, effectiveTaxonomyAttributes, findTaxonomyDraftNode, findTaxonomyDraftPath, flattenTaxonomyDraft, removeTaxonomyAttribute, removeTaxonomyDraftRule, reorderTaxonomyDraftRule, replaceTaxonomyDraftPlacements, taxonomyAssignmentErrorKey, updateTaxonomyAttribute, updateTaxonomyDraftNode, updateTaxonomyDraftRule, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyAttributeValueType, type TaxonomyDraftNode, type TaxonomyDragPayload, type TaxonomyGrammar, type TaxonomyPlacement, type TaxonomyEffectiveAttribute, type TaxonomyRulePredicate, type TaxonomyValidationError } from './TaxonomyEditorModel'
  import type { DragDropState } from '@thisux/sveltednd'

  export let grammar: TaxonomyGrammar
  export let initialAttributeDefinitions: TaxonomyAttributeDefinition[] = []
  export let initialDraft: TaxonomyDraftNode
  export let readOnly = false
  export let onChange: (draft: TaxonomyDraftNode, definitions: TaxonomyAttributeDefinition[]) => void = () => {}

  let placementNumber = 1
  let attributeNumber = 1
  let attributesNodeID: string | undefined
  let attributeDefinitions: TaxonomyAttributeDefinition[] = initialAttributeDefinitions
  let taxonomyDraft: TaxonomyDraftNode = initialDraft
  $: placements = flattenTaxonomyDraft(taxonomyDraft)

  function updatePlacements(next: TaxonomyPlacement[]) {
    taxonomyDraft = replaceTaxonomyDraftPlacements(taxonomyDraft, next)
  }

  function newPlacement(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, targetId: string): TaxonomyPlacement {
    const id = `taxonomy-${payload.taxonomyType}-${placementNumber++}`
    if (payload.taxonomyType === 'category') return { id, targetId, type: 'category', title: 'New category', summary: `Attached to ${targetId}`, label: 'New category', nodeId: id, predicate: `${targetId} = A` }
    return { id, targetId, type: 'rule', title: 'New rule', summary: `${targetId} = A`, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
  }

  function handleDrop(state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) {
    const item = state.draggedItem
    if (!item) return
    const targetPlacements = placements.filter((placement) => placement.targetId === targetId)
    if (item.kind === 'palette') {
      const placement = newPlacement(item, targetId)
      if (item.taxonomyType === 'category') {
        const targetChildren = taxonomyDraft.children.filter((child) => child.targetId === targetId)
        const nextChildren = [...taxonomyDraft.children]
        const childIndex = nextChildren.findIndex((child) => child === targetChildren[targetIndex])
        nextChildren.splice(childIndex < 0 ? nextChildren.length : childIndex, 0, { id: placement.id, label: placement.label ?? placement.title, targetId, placements: [placement], rules: [], children: [] })
        taxonomyDraft = { ...taxonomyDraft, children: nextChildren }
        return
      }
      const next = [...targetPlacements]
      next.splice(targetIndex, 0, placement)
      updatePlacements([...placements.filter((placement) => placement.targetId !== targetId), ...next])
      return
    }
    const existing = placements.find((placement) => placement.id === item.placementId)
    if (!existing) return
    const sourceIndex = targetPlacements.findIndex((placement) => placement.id === item.placementId)
    const next = targetPlacements.filter((placement) => placement.id !== item.placementId)
    const insertIndex = existing.targetId === targetId && sourceIndex >= 0 && sourceIndex < targetIndex ? targetIndex - 1 : targetIndex
    next.splice(insertIndex, 0, { ...existing, targetId })
    updatePlacements([...placements.filter((placement) => placement.targetId !== existing.targetId && placement.targetId !== targetId), ...next])
  }

  function handleChildDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string) {
    const item = state.draggedItem
    if (!item || item.kind !== 'palette' || item.taxonomyType !== 'category') return
    const placement = newPlacement(item, parentID)
    taxonomyDraft = addTaxonomyDraftChild(taxonomyDraft, parentID, { id: placement.id, label: placement.label ?? placement.title, targetId: placement.targetId, placements: [placement], rules: [], children: [] })
  }

  function handleRuleDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) {
    const item = state.draggedItem
    if (!item) return
    const parent = findTaxonomyDraftNode(taxonomyDraft, parentID)
    if (!parent) return
    if (item.kind === 'palette' && item.taxonomyType === 'rule') {
      taxonomyDraft = addTaxonomyDraftRule(taxonomyDraft, parentID, newPlacement(item, parent.targetId), targetIndex)
    } else if (item.kind === 'taxonomy' && item.taxonomyType === 'rule' && item.sourceNodeId === parentID) {
      taxonomyDraft = reorderTaxonomyDraftRule(taxonomyDraft, parentID, item.placementId, targetIndex)
    }
  }

  function removePlacement(placementId: string) {
    taxonomyDraft = removeTaxonomyDraftRule(taxonomyDraft, placementId)
    updatePlacements(placements.filter((placement) => placement.id !== placementId))
  }

  function handleDragEnd() {}

  function updateNodeLabel(nodeID: string, label: string) {
    if (!readOnly) taxonomyDraft = updateTaxonomyDraftNode(taxonomyDraft, nodeID, { label })
  }

  function updateNodeID(nodeID: string, newNodeID: string) {
    if (!readOnly) taxonomyDraft = updateTaxonomyDraftNode(taxonomyDraft, nodeID, { id: newNodeID })
  }

  function updateRulePredicate(ruleID: string, predicate: TaxonomyRulePredicate) {
    if (!readOnly) taxonomyDraft = updateTaxonomyDraftRule(taxonomyDraft, ruleID, predicate)
  }

  function closeAttributes() {
    attributesNodeID = undefined
  }

  $: selectedAttributesNode = attributesNodeID ? findTaxonomyDraftNode(taxonomyDraft, attributesNodeID) : undefined
  $: selectedAttributes = attributesNodeID ? effectiveTaxonomyAttributes(taxonomyDraft, attributesNodeID, attributeDefinitions) : []
  $: validation = validateTaxonomyAttributes(attributeDefinitions, taxonomyDraft)
  $: onChange(taxonomyDraft, attributeDefinitions)
  $: selectedAttributeRows = selectedAttributes.map((attribute: TaxonomyEffectiveAttribute) => ({ attribute, errors: attributeValidationErrors(attribute) }))

  function attributeValidationErrors(attribute: TaxonomyEffectiveAttribute): TaxonomyValidationError[] {
    const errors = [...(validation.definitionErrors[attribute.id] ?? [])]
    if (attributesNodeID) for (const node of findTaxonomyDraftPath(taxonomyDraft, attributesNodeID)) {
      for (const [index, assignment] of (node.attributes ?? []).entries()) if (assignment.attributeDefId === attribute.id) errors.push(...(validation.assignmentErrors[taxonomyAssignmentErrorKey(node.id, index)] ?? []))
    }
    return errors
  }

  function updateAttribute(attributeDefID: string, required: boolean) {
    if (attributesNodeID) taxonomyDraft = updateTaxonomyAttribute(taxonomyDraft, attributesNodeID, attributeDefID, { attributeDefId: attributeDefID, required })
  }

  function updateAttributeDefinition(attributeDefID: string, label: string, valueType: TaxonomyAttributeValueType) {
    attributeDefinitions = attributeDefinitions.map((definition) => definition.id === attributeDefID ? { ...definition, label, valueType } : definition)
  }

  function overrideAttribute(attributeDefID: string, required: boolean) {
    if (!attributesNodeID) return
    taxonomyDraft = addTaxonomyAttribute(taxonomyDraft, attributesNodeID, { attributeDefId: attributeDefID, required })
  }

  function addAttribute() {
    if (!attributesNodeID) return
    const available = attributeDefinitions.find((definition) => !selectedAttributes.some((attribute) => attribute.id === definition.id))
    if (available) {
      taxonomyDraft = addTaxonomyAttribute(taxonomyDraft, attributesNodeID, { attributeDefId: available.id, required: false })
      return
    }
    const id = `attribute-${attributeNumber++}`
    attributeDefinitions = [...attributeDefinitions, { id, label: 'New attribute', valueType: 'string' }]
    taxonomyDraft = addTaxonomyAttribute(taxonomyDraft, attributesNodeID, { attributeDefId: id, required: false })
  }

  function removeAttribute(attributeDefID: string) {
    if (attributesNodeID) taxonomyDraft = removeTaxonomyAttribute(taxonomyDraft, attributesNodeID, attributeDefID)
  }
</script>

<section class="taxonomy-editor" aria-label="Taxonomy editor layout sketch">
  {#if !readOnly}<TaxonomyPalette onDragEnd={handleDragEnd} />{/if}

  <section class="taxonomy-editor-canvas" aria-labelledby="taxonomy-canvas-heading">
    <div class="taxonomy-editor-heading">
        <div><span class="mock-section-label" id="taxonomy-canvas-heading">Grammar and taxonomy overlay</span><h3>{grammar.label}</h3></div>
        <div class="taxonomy-editor-heading-controls"><span class="taxonomy-editor-count">{grammar.fields} fields · {placements.length} taxonomy items</span></div>
    </div>
    <p class="taxonomy-grammar-note">Read-only grammar <span aria-hidden="true">·</span> taxonomy drop zones are shown below each grammar node</p>

    <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree grammarNodes={grammar.nodes} {placements} {taxonomyDraft} {readOnly} onDrop={handleDrop} onChildDrop={handleChildDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} onDragEnd={handleDragEnd} onAttributes={(nodeID) => attributesNodeID = nodeID} onNodeLabelChange={updateNodeLabel} onNodeIdChange={updateNodeID} onRulePredicateChange={updateRulePredicate} /></div>
  </section>

  <section class="taxonomy-editor-status" aria-labelledby="taxonomy-status-heading">
    <div class="taxonomy-editor-heading"><span class="mock-section-label" id="taxonomy-status-heading">Definition status</span><span class="taxonomy-editor-count">{readOnly ? 'historical revision' : 'local draft'}</span></div>
    <div class="taxonomy-status-grid">
      <span class="taxonomy-status-ok">✓</span><span>{readOnly ? 'Historical revision is immutable' : 'Grammar is immutable'}</span>
      <span class="taxonomy-status-ok">✓</span><span>Drop zones expose known grammar fields</span>
       <span class={validation.valid ? 'taxonomy-status-ok' : 'taxonomy-status-pending'}>{validation.valid ? '✓' : '!'}</span><span>{validation.valid ? 'Taxonomy draft is valid' : 'Taxonomy draft has validation errors'}</span>
    </div>
  </section>
</section>

{#if attributesNodeID && selectedAttributesNode}
  <SlideOut open eyebrow={`Attributes for ${selectedAttributesNode.label}`} title="Assigned part properties" onClose={closeAttributes}>
    <p class="slide-out-copy">Inherited assignments remain visible with their source.</p>
    <TableFrame className="taxonomy-attribute-table-frame">
      <table class="taxonomy-attribute-table">
        <caption class="sr-only">Attributes assigned to {selectedAttributesNode.label}</caption>
        <thead><tr><th scope="col">Attribute</th><th scope="col">Value type</th><th scope="col">Required</th><th scope="col">Source</th></tr></thead>
       <tbody>{#each selectedAttributeRows as row}<TaxonomyAttributeTableRow attribute={row.attribute} errors={row.errors} selectedNodeLabel={selectedAttributesNode.label} editable={!readOnly && row.attribute.source === selectedAttributesNode.label} onChange={(required) => updateAttribute(row.attribute.id, required)} onDefinitionChange={(label, valueType) => updateAttributeDefinition(row.attribute.id, label, valueType)} actionLabel={readOnly ? '' : 'Override'} onAction={() => overrideAttribute(row.attribute.id, row.attribute.required)} onRemove={() => removeAttribute(row.attribute.id)} />{/each}</tbody>
      </table>
    </TableFrame>
     {#if !readOnly}<PrimitiveButton variant="secondary" size="small" type="button" on:click={addAttribute}>+ Add attribute assignment</PrimitiveButton>{/if}
  </SlideOut>
{/if}
