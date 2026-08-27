<script lang="ts">
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import TaxonomyGrammarTree from './TaxonomyGrammarTree.svelte'
  import type { TaxonomyDragPayload, TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { DragDropState } from '@thisux/sveltednd'

  let placementNumber = 1
  let placements: TaxonomyPlacement[] = [{
    id: 'high-power', targetId: 'number', type: 'category', title: 'High Power', summary: 'number = 100..199',
    label: 'High Power', nodeId: 'high-power', predicate: 'number = 100..199',
  }]

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
      const next = [...targetPlacements]
      next.splice(targetIndex, 0, newPlacement(item, targetId))
      placements = [...placements.filter((placement) => placement.targetId !== targetId), ...next]
      return
    }
    const existing = placements.find((placement) => placement.id === item.placementId)
    if (!existing) return
    const sourceIndex = targetPlacements.findIndex((placement) => placement.id === item.placementId)
    const next = targetPlacements.filter((placement) => placement.id !== item.placementId)
    const insertIndex = existing.targetId === targetId && sourceIndex >= 0 && sourceIndex < targetIndex ? targetIndex - 1 : targetIndex
    next.splice(insertIndex, 0, { ...existing, targetId })
    placements = [...placements.filter((placement) => placement.targetId !== existing.targetId && placement.targetId !== targetId), ...next]
  }

  function removePlacement(placementId: string) {
    placements = placements.filter((placement) => placement.id !== placementId)
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

    <div class="taxonomy-grammar-tree"><TaxonomyGrammarTree {placements} onDrop={handleDrop} onRemove={removePlacement} onDragEnd={handleDragEnd} /></div>
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
