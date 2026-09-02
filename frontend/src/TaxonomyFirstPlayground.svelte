<script lang="ts">
  import type { DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'
  import TaxonomyGrammarCategory from './TaxonomyGrammarCategory.svelte'
  import TaxonomyPalette from './TaxonomyPalette.svelte'
  import { addTaxonomyDraftChild, addTaxonomyDraftRule, findTaxonomyDraftNode, findTaxonomyDraftPath, flattenTaxonomyDraft, moveTaxonomyDraftChild, moveTaxonomyDraftRule, removeTaxonomyDraftRule, taxonomyTargetContainer, updateTaxonomyDraftNode, updateTaxonomyDraftRule, type TaxonomyDraftNode, type TaxonomyDragPayload, type TaxonomyPlacement, type TaxonomyRulePredicate } from './TaxonomyEditorModel'

  let placementNumber = 1
  let taxonomyDraft: TaxonomyDraftNode = {
    id: 'taxonomy-root', label: 'Part Numbers', targetId: 'grammar-root', placements: [], rules: [], children: [
      { id: 'family-302', label: '302', targetId: 'family', placements: [{ id: 'family-302', targetId: 'family', type: 'category', title: '302', summary: 'family = 302', label: '302', nodeId: 'family-302', predicate: 'family = 302', predicateData: { kind: 'exact', field: 'family', value: '302' } }], rules: [], children: [
        { id: 'family-302-group-00012', label: '00012', targetId: 'group', placements: [{ id: 'family-302-group-00012', targetId: 'group', type: 'category', title: '00012', summary: 'group = 00012', label: '00012', nodeId: 'family-302-group-00012', predicate: 'group = 00012', predicateData: { kind: 'exact', field: 'group', value: '00012' } }], rules: [], children: [] },
        { id: 'family-302-group-00018', label: '00018', targetId: 'group', placements: [{ id: 'family-302-group-00018', targetId: 'group', type: 'category', title: '00018', summary: 'group = 00018', label: '00018', nodeId: 'family-302-group-00018', predicate: 'group = 00018', predicateData: { kind: 'exact', field: 'group', value: '00018' } }], rules: [], children: [] },
      ] },
      { id: 'family-410', label: '410', targetId: 'family', placements: [{ id: 'family-410', targetId: 'family', type: 'category', title: '410', summary: 'family = 410', label: '410', nodeId: 'family-410', predicate: 'family = 410', predicateData: { kind: 'exact', field: 'family', value: '410' } }], rules: [], children: [] },
    ],
  }
  let selectedNodeID = 'family-302'
  let selectedGrammarField = 'family'

  $: placements = flattenTaxonomyDraft(taxonomyDraft)
  $: selectedNode = findTaxonomyDraftNode(taxonomyDraft, selectedNodeID) ?? taxonomyDraft
  $: selectedPath = findTaxonomyDraftPath(taxonomyDraft, selectedNode.id)
  $: effectivePredicate = selectedPath.slice(1).map((node) => node.placements.find((placement) => placement.type === 'category')?.predicate).filter(Boolean).join(' AND ')

  function newPlacement(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, targetId: string): TaxonomyPlacement {
    const id = `taxonomy-${payload.taxonomyType}-${placementNumber++}`
    if (payload.taxonomyType === 'category') return { id, targetId, type: 'category', title: 'New category', summary: `Attached to ${targetId}`, label: 'New category', nodeId: id, predicate: `${targetId} = A`, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
    return { id, targetId, type: 'rule', title: 'New rule', summary: `${targetId} = A`, predicate: `${targetId} = A`, predicateData: { kind: 'exact', field: targetId, value: 'A' } }
  }

  function addCategory(payload: Extract<TaxonomyDragPayload, { kind: 'palette' }>, parentID: string, targetId: string) {
    const placement = newPlacement(payload, targetId)
    taxonomyDraft = addTaxonomyDraftChild(taxonomyDraft, parentID, { id: placement.id, label: placement.label ?? placement.title, targetId, placements: [placement], rules: [], children: [] })
    selectedNodeID = placement.id
  }

  function handleRootDrop(state: DragDropState<TaxonomyDragPayload>) {
    const item = state.draggedItem
    if (item?.kind === 'palette' && item.taxonomyType === 'category') addCategory(item, taxonomyDraft.id, selectedGrammarField)
  }

  function handleChildDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string) {
    const item = state.draggedItem
    if (!item) return
    if (item.kind === 'palette' && item.taxonomyType === 'category') addCategory(item, parentID, selectedGrammarField)
    else if (item.kind === 'taxonomy' && item.taxonomyType === 'category' && item.sourceNodeId) taxonomyDraft = moveTaxonomyDraftChild(taxonomyDraft, item.sourceNodeId, parentID)
  }

  function handleRuleDrop(state: DragDropState<TaxonomyDragPayload>, parentID: string, targetIndex: number) {
    const item = state.draggedItem
    const parent = findTaxonomyDraftNode(taxonomyDraft, parentID)
    if (!item || !parent) return
    if (item.kind === 'palette' && item.taxonomyType === 'rule') taxonomyDraft = addTaxonomyDraftRule(taxonomyDraft, parentID, newPlacement(item, parent.targetId), targetIndex)
    else if (item.kind === 'taxonomy' && item.taxonomyType === 'rule' && item.sourceNodeId) taxonomyDraft = moveTaxonomyDraftRule(taxonomyDraft, item.sourceNodeId, parentID, item.placementId, targetIndex)
  }

  function removeTaxonomyDraftCategorySubtree(node: TaxonomyDraftNode, placementID: string): TaxonomyDraftNode {
    const children = node.children
      .filter((child) => !child.placements.some((placement) => placement.type === 'category' && placement.id === placementID))
      .map((child) => removeTaxonomyDraftCategorySubtree(child, placementID))
    return children.length !== node.children.length || children.some((child, index) => child !== node.children[index]) ? { ...node, children } : node
  }

  function removePlacement(placementID: string) {
    const categoryNode = findTaxonomyDraftNode(taxonomyDraft, placementID)
    const categoryPlacement = categoryNode?.placements.find((placement) => placement.id === placementID)
    if (categoryPlacement?.type === 'category' && categoryNode) {
      const selectedPath = findTaxonomyDraftPath(taxonomyDraft, selectedNodeID)
      const selectedCategory = selectedPath.some((node) => node === categoryNode)
      const categoryPath = findTaxonomyDraftPath(taxonomyDraft, categoryNode.id)
      const parent = categoryPath.length > 1 ? categoryPath[categoryPath.length - 2] : undefined
      taxonomyDraft = removeTaxonomyDraftCategorySubtree(taxonomyDraft, placementID)
      if (selectedCategory) selectedNodeID = parent?.id ?? taxonomyDraft.id
    } else {
      taxonomyDraft = removeTaxonomyDraftRule(taxonomyDraft, placementID)
    }
  }

  function updateNodeLabel(nodeID: string, label: string) { taxonomyDraft = updateTaxonomyDraftNode(taxonomyDraft, nodeID, { label }) }
  function updateNodeID(nodeID: string, id: string) { taxonomyDraft = updateTaxonomyDraftNode(taxonomyDraft, nodeID, { id }) }
  function updateRulePredicate(ruleID: string, predicate: TaxonomyRulePredicate) { taxonomyDraft = updateTaxonomyDraftRule(taxonomyDraft, ruleID, predicate) }
</script>

<section class="taxonomy-first-draft" aria-labelledby="taxonomy-first-heading">
  <div class="taxonomy-first-heading">
    <div><p class="eyebrow">Interaction concept</p><h2 id="taxonomy-first-heading">Taxonomy-first authoring</h2><p class="taxonomy-first-intro">A taxonomy tree is the work surface. The grammar remains visible as a quiet map of fields you can classify by.</p></div>
    <span class="taxonomy-first-status">Exploration draft</span>
  </div>

  <div class="taxonomy-first-workflow taxonomy-editor">
    <section class="taxonomy-first-authoring-row" aria-label="Authoring blocks">
      <TaxonomyPalette />
      <label class="taxonomy-first-field-picker">New category grammar field
        <select bind:value={selectedGrammarField} aria-label="New category grammar field">
          <option value="family">family</option>
          <option value="group">group</option>
          <option value="number">number</option>
        </select>
      </label>
    </section>
    <main class="taxonomy-editor-canvas">
      <div class="taxonomy-editor-heading"><div><span class="mock-section-label">Taxonomy tree</span><h3>Part Numbers</h3></div><span class="taxonomy-editor-count">gtd-pn-taxonomy-v1 · local draft</span></div>
      <p class="taxonomy-grammar-note">Select a category to focus it. Categories are the hierarchy; rules stay local to the selected category.</p>
      <div class="taxonomy-grammar-tree">
        <div class="taxonomy-first-root"><span class="taxonomy-grammar-kind">root scope</span><strong>Part Numbers</strong><small>all values matching gtd-pn-v1</small></div>
        <DropZone variant="slot" mode="inactive" container={taxonomyTargetContainer(taxonomyDraft.id)} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDrop={handleRootDrop} className="taxonomy-first-root-drop" ariaLabel="Drop a first category here" />
        <div class="taxonomy-first-branches taxonomy-grammar-structure-tree">
          {#each taxonomyDraft.children as child (child.id)}
            {#if child.placements[0]}
              <div class:selected={selectedNodeID === child.id} class="taxonomy-first-category" role="button" tabindex="0" on:click={(event) => { if (!(event.target as HTMLElement).closest('.block-shell-remove')) selectedNodeID = child.id }} on:keydown={(event) => event.key === 'Enter' && (selectedNodeID = child.id)}>
                <TaxonomyGrammarCategory placement={child.placements[0]} draftNode={child} parentContainer={taxonomyTargetContainer(taxonomyDraft.id)} onChildDrop={handleChildDrop} onRuleDrop={handleRuleDrop} onRemove={removePlacement} onAttributes={(nodeID) => selectedNodeID = nodeID} onNodeLabelChange={updateNodeLabel} onNodeIdChange={updateNodeID} onRulePredicateChange={updateRulePredicate} />
              </div>
            {/if}
          {/each}
        </div>
      </div>
      <div class="taxonomy-first-effective"><span class="mock-section-label">Effective condition for {selectedNode.label}</span><code>{effectivePredicate || 'any valid part number'}</code><small>Inherited from the taxonomy path.</small></div>
    </main>
  </div>
</section>

<style>
  .taxonomy-first-draft { max-width: 1280px; }.taxonomy-first-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 2rem; margin-bottom: 1.5rem; }.taxonomy-first-heading h2 { margin: .15rem 0 .55rem; letter-spacing: -.035em; font-size: clamp(2rem, 4vw, 3rem); line-height: 1; }.taxonomy-first-intro { max-width: 650px; margin: 0; color: var(--muted); line-height: 1.5; }.taxonomy-first-status { padding: .35rem .55rem; color: var(--muted); border: 1px solid var(--hairline); border-radius: .2rem; font: .6rem ui-monospace, monospace; text-transform: uppercase; letter-spacing: .08em; white-space: nowrap; }
  .taxonomy-first-workflow { display: grid; gap: 1rem; }
  .taxonomy-first-authoring-row { display: grid; gap: 1rem; background: var(--ground); }
  .taxonomy-first-field-picker { display: grid; gap: .3rem; align-content: start; max-width: 14rem; padding: 0 1rem 1rem; color: var(--muted); font: .65rem ui-monospace, monospace; }
  .taxonomy-first-field-picker select { padding: .45rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: inherit; }
  .taxonomy-first-root { display: grid; gap: .15rem; padding: .6rem .55rem; border-bottom: 1px solid var(--hairline); }
  .taxonomy-first-root small { color: var(--muted); font: .65rem ui-monospace, monospace; }
  .taxonomy-first-branches { display: grid; gap: .5rem; padding: .5rem 0 0 1.15rem; border-left: 1px solid var(--hairline); }
  .taxonomy-first-category { padding: .15rem; border: 1px solid transparent; border-radius: .25rem; cursor: pointer; }
  .taxonomy-first-category:hover, .taxonomy-first-category.selected { border-color: var(--accent); background: var(--accent-light); }
  .taxonomy-first-effective { display: grid; gap: .35rem; margin-top: .7rem; padding: .7rem; color: var(--surface); background: var(--ink); border-radius: .2rem; }
  .taxonomy-first-effective .mock-section-label { color: var(--muted); }
  .taxonomy-first-effective code { color: var(--surface); font: .78rem ui-monospace, monospace; }
  .taxonomy-first-effective small { color: var(--muted); font: .65rem ui-monospace, monospace; }
  @media (max-width: 680px) {
    .taxonomy-first-heading { display: block; }
    .taxonomy-first-status { display: inline-block; margin-top: 1rem; }
    .taxonomy-editor-heading { align-items: flex-start; flex-wrap: wrap; }
  }
</style>
