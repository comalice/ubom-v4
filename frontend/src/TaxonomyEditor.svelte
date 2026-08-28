<script lang="ts">
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import { addTaxonomyDraftChild, addTaxonomyDraftRule, findTaxonomyDraftNode, flattenTaxonomyDraft, removeTaxonomyDraftRule, replaceTaxonomyDraftPlacements, type TaxonomyDraftNode, type TaxonomyDragPayload, type TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { DragDropState } from '@thisux/sveltednd'

  let placementNumber = 1
  let taxonomyDraft: TaxonomyDraftNode = {
    id: 'taxonomy-root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [], children: [{
      id: 'high-power', label: 'High Power', targetId: 'number', placements: [{
        id: 'high-power', targetId: 'number', type: 'category', title: 'High Power', summary: 'number = 100..199',
        label: 'High Power', nodeId: 'high-power', predicate: 'number = 100..199',
      }], rules: [], children: [],
    }],
  }
  $: placements = flattenTaxonomyDraft(taxonomyDraft)

  function updatePlacements(next: TaxonomyPlacement[]) {
    taxonomyDraft = replaceTaxonomyDraftPlacements(taxonomyDraft, next)
  }

  function newPlacement(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, targetId: string): TaxonomyPlacement {
    const id = `taxonomy-${payload.taxonomyType}-${placementNumber++}`
    if (payload.taxonomyType === 'category') return { id, targetId, type: 'category', title: 'New category', summary: `Attached to ${targetId}`, label: 'New category', nodeId: id, predicate: `${targetId} = A` }
    if (payload.taxonomyType === 'rule') return { id, targetId, type: 'rule', title: 'New rule', summary: `${targetId} = A`, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
    return { id, targetId, type: 'attribute', title: 'Attribute', summary: 'Part property', attribute: 'Resistance value', value: '10' }
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
        nextChildren.splice(childIndex < 0 ? nextChildren.length : childIndex, 0, { id: placement.id, label: placement.label ?? placement.title, targetId, placements: [placement], children: [] })
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
    updatePlacements(placements.filter((placement) => placement.id !== placementId))
    taxonomyDraft = removeTaxonomyDraftRule(taxonomyDraft, placementId)
  }

  function handleDragEnd() {}
</script>

<section class="taxonomy-editor" aria-label="Taxonomy editor layout sketch">
  <TaxonomyPalette onDragEnd={handleDragEnd} />

  <section class="taxonomy-editor-canvas" aria-labelledby="taxonomy-canvas-heading">
    <div class="taxonomy-editor-heading">
      <div><span class="mock-section-label" id="taxonomy-canvas-heading">Grammar and taxonomy overlay</span><h3>Part number grammar</h3></div>
       <span class="taxonomy-editor-count">3 fields · {placements.length} taxonomy items</span>
    </div>
    <p class="taxonomy-grammar-note">Read-only grammar <span aria-hidden="true">·</span> taxonomy drop zones are shown below each grammar node</p>

    <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree {placements} {taxonomyDraft} onDrop={handleDrop} onChildDrop={handleChildDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} onDragEnd={handleDragEnd} /></div>
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
