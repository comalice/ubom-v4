<script lang="ts">
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import SlideOut from './SlideOut.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyAttributeTableRow from './TaxonomyAttributeTableRow.svelte'
  import { grammarFixtures } from './TaxonomyEditorFixtures'
  import { addTaxonomyAttribute, addTaxonomyDraftChild, addTaxonomyDraftRule, effectiveTaxonomyAttributes, findTaxonomyDraftNode, flattenTaxonomyDraft, removeTaxonomyAttribute, removeTaxonomyDraftRule, replaceTaxonomyDraftPlacements, updateTaxonomyAttribute, type TaxonomyAttributeDefinition, type TaxonomyDraftNode, type TaxonomyDragPayload, type TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { DragDropState } from '@thisux/sveltednd'

  let placementNumber = 1
  let attributeNumber = 1
  let selectedGrammarID = grammarFixtures[0].id
  let attributesNodeID: string | undefined
  let attributeDefinitions: TaxonomyAttributeDefinition[] = [
    { id: 'footprint', label: 'Footprint', valueType: 'enum', enumValues: ['Axial', 'Radial', 'SMD'] },
    { id: 'material', label: 'Material', valueType: 'string' },
    { id: 'resistance', label: 'Resistance value', valueType: 'decimal', unit: 'ohm' },
    { id: 'tolerance', label: 'Tolerance', valueType: 'decimal', unit: 'percent' },
    { id: 'power-rating', label: 'Power rating', valueType: 'decimal', unit: 'watt' },
  ]
  let taxonomyDraft: TaxonomyDraftNode = {
    id: 'taxonomy-root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [], attributes: [{ attributeDefId: 'footprint', required: true }, { attributeDefId: 'material', required: false }], children: [{
      id: 'high-power', label: 'High Power', targetId: 'number', placements: [{
        id: 'high-power', targetId: 'number', type: 'category', title: 'High Power', summary: 'number = 100..199',
        label: 'High Power', nodeId: 'high-power', predicate: 'number = 100..199',
      }], rules: [], attributes: [{ attributeDefId: 'footprint', required: false }, { attributeDefId: 'resistance', required: true }, { attributeDefId: 'tolerance', required: false }], children: [],
    }],
  }
  $: placements = flattenTaxonomyDraft(taxonomyDraft)
  $: selectedGrammar = grammarFixtures.find((grammar) => grammar.id === selectedGrammarID) ?? grammarFixtures[0]

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

  function handleRuleDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string) {
    const item = state.draggedItem
    if (!item || item.kind !== 'palette' || item.taxonomyType !== 'rule') return
    const parent = findTaxonomyDraftNode(taxonomyDraft, parentID)
    if (!parent) return
    const placement = newPlacement(item, parent.targetId)
    taxonomyDraft = addTaxonomyDraftRule(taxonomyDraft, parentID, placement)
  }

  function removePlacement(placementId: string) {
    taxonomyDraft = removeTaxonomyDraftRule(taxonomyDraft, placementId)
    updatePlacements(placements.filter((placement) => placement.id !== placementId))
  }

  function handleDragEnd() {}

  function closeAttributes() {
    attributesNodeID = undefined
  }

  $: selectedAttributesNode = attributesNodeID ? findTaxonomyDraftNode(taxonomyDraft, attributesNodeID) : undefined
  $: selectedAttributes = attributesNodeID ? effectiveTaxonomyAttributes(taxonomyDraft, attributesNodeID, attributeDefinitions) : []

  function updateAttribute(attributeDefID: string, required: boolean) {
    if (attributesNodeID) taxonomyDraft = updateTaxonomyAttribute(taxonomyDraft, attributesNodeID, attributeDefID, { attributeDefId: attributeDefID, required })
  }

  function updateAttributeDefinition(attributeDefID: string, label: string, valueType: string) {
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
  <TaxonomyPalette onDragEnd={handleDragEnd} />

  <section class="taxonomy-editor-canvas" aria-labelledby="taxonomy-canvas-heading">
    <div class="taxonomy-editor-heading">
        <div><span class="mock-section-label" id="taxonomy-canvas-heading">Grammar and taxonomy overlay</span><h3>Part number grammar</h3></div>
        <div class="taxonomy-editor-heading-controls"><label for="taxonomy-grammar-select">Grammar<select id="taxonomy-grammar-select" bind:value={selectedGrammarID}>{#each grammarFixtures as grammar}<option value={grammar.id}>{grammar.label}</option>{/each}</select></label><span class="taxonomy-editor-count">{selectedGrammar.fields} fields · {placements.length} taxonomy items</span></div>
    </div>
    <p class="taxonomy-grammar-note">Read-only grammar <span aria-hidden="true">·</span> taxonomy drop zones are shown below each grammar node</p>

    <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree grammarNodes={selectedGrammar.nodes} {placements} {taxonomyDraft} onDrop={handleDrop} onChildDrop={handleChildDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} onDragEnd={handleDragEnd} onAttributes={(nodeID) => attributesNodeID = nodeID} /></div>
  </section>

  <section class="taxonomy-editor-status" aria-labelledby="taxonomy-status-heading">
    <div class="taxonomy-editor-heading"><span class="mock-section-label" id="taxonomy-status-heading">Definition status</span><span class="taxonomy-editor-count">local draft</span></div>
    <div class="taxonomy-status-grid">
      <span class="taxonomy-status-ok">✓</span><span>Grammar is immutable</span>
      <span class="taxonomy-status-ok">✓</span><span>Drop zones expose known grammar fields</span>
      <span class="taxonomy-status-pending">•</span><span>Taxonomy validation will come later</span>
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
        <tbody>{#each selectedAttributes as attribute}<TaxonomyAttributeTableRow {attribute} selectedNodeLabel={selectedAttributesNode.label} editable={attribute.source === selectedAttributesNode.label} onChange={(required) => updateAttribute(attribute.id, required)} onDefinitionChange={(label, valueType) => updateAttributeDefinition(attribute.id, label, valueType)} actionLabel="Override" onAction={() => overrideAttribute(attribute.id, attribute.required)} onRemove={() => removeAttribute(attribute.id)} />{/each}</tbody>
      </table>
    </TableFrame>
    <PrimitiveButton variant="secondary" size="small" type="button" on:click={addAttribute}>+ Add attribute assignment</PrimitiveButton>
  </SlideOut>
{/if}
