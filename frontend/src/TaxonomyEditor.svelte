<script lang="ts">
  import TaxonomyFirstCanvas from './TaxonomyFirstCanvas.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import SlideOut from './SlideOut.svelte'
  import TableFrame from './TableFrame.svelte'
  import TaxonomyAttributeTableRow from './TaxonomyAttributeTableRow.svelte'
  import { addTaxonomyAttribute, effectiveTaxonomyAttributes, findTaxonomyDraftNode, findTaxonomyDraftPath, flattenTaxonomyDraft, removeTaxonomyAttribute, taxonomyAssignmentErrorKey, updateTaxonomyAttribute, updateTaxonomyDraftNode, updateTaxonomyDraftRule, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyAttributeValueType, type TaxonomyDraftNode, type TaxonomyGrammar, type TaxonomyEffectiveAttribute, type TaxonomyRulePredicate, type TaxonomyValidationError } from './TaxonomyEditorModel'

  export let grammar: TaxonomyGrammar
  export let initialAttributeDefinitions: TaxonomyAttributeDefinition[] = []
  export let initialDraft: TaxonomyDraftNode
  export let readOnly = false
  export let onChange: (draft: TaxonomyDraftNode, definitions: TaxonomyAttributeDefinition[]) => void = () => {}

  let attributeNumber = 1
  let attributesNodeID: string | undefined
  let attributeDefinitions: TaxonomyAttributeDefinition[] = initialAttributeDefinitions
  let taxonomyDraft: TaxonomyDraftNode = initialDraft

  $: placements = flattenTaxonomyDraft(taxonomyDraft)

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
    let id = `attribute-${attributeNumber++}`
    while (attributeDefinitions.some((definition) => definition.id === id)) id = `attribute-${attributeNumber++}`
    attributeDefinitions = [...attributeDefinitions, { id, label: 'New attribute', valueType: 'string' }]
    taxonomyDraft = addTaxonomyAttribute(taxonomyDraft, attributesNodeID, { attributeDefId: id, required: false })
  }

  function removeAttribute(attributeDefID: string) {
    if (attributesNodeID) taxonomyDraft = removeTaxonomyAttribute(taxonomyDraft, attributesNodeID, attributeDefID)
  }
</script>

<section class="taxonomy-editor" aria-label="Taxonomy editor layout sketch">
  <TaxonomyFirstCanvas grammar={grammar} taxonomyDraft={taxonomyDraft} {readOnly} onDraftChange={(next) => taxonomyDraft = next} onNodeLabelChange={updateNodeLabel} onNodeIdChange={updateNodeID} onRulePredicateChange={updateRulePredicate} onAttributes={(nodeID) => attributesNodeID = nodeID} />

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
