<script lang="ts">
  import GrammarEditorRow, { type GrammarDragItem, type GrammarDropSlot } from './GrammarEditorRow.svelte'
  import { type DragDropState } from '@thisux/sveltednd'
  import DndField from './DndField.svelte'
  import StructureTree, { type StructureNode } from './StructureTree.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'
  import { playgroundCreateExpr as createExpr, playgroundGenerateSequence as generateSequence, playgroundParseSequence as parseSequence, type PlaygroundDragPayload as DragPayload, type PlaygroundDropTarget as DropTarget, type PlaygroundExpr as Expr, type PlaygroundExprKind as ExprKind } from './PlaygroundGrammar'

  const paletteTiles = [
    { type: 'literal', title: 'Literal', description: 'Exact text' },
    { type: 'number', title: 'Number', description: 'Numeric range' },
    { type: 'alpha', title: 'Letters', description: 'Alphabetic range' },
    { type: 'chars', title: 'Characters', description: 'Allowed characters' },
    { type: 'bind', title: 'Binding', description: 'Capture a value' },
    { type: 'choice', title: 'Choice', description: 'One alternative' },
  ]

  let rootSequence: Expr[] = [
    { id: 'literal-1', kind: 'literal', text: 'PN-' },
    { id: 'bind-1', kind: 'bind', name: 'family', child: { id: 'literal-2', kind: 'literal', text: 'SN-' } },
    { id: 'choice-1', kind: 'choice', children: [
      [{ id: 'literal-3', kind: 'literal', text: 'A' }],
      [{ id: 'literal-4', kind: 'literal', text: 'B' }, { id: 'number-1', kind: 'number', min: 0, max: 9, width: 1, pad: '0' }],
    ] },
  ]
  let structureNodes: StructureNode[] = []
  let testInput = 'PN-SN-A'

  function summary(expr: Expr): string {
    if (expr.kind === 'literal') return JSON.stringify(expr.text)
    if (expr.kind === 'number') return `${expr.min}..${expr.max} / width ${expr.width} / pad ${expr.pad}`
    if (expr.kind === 'alpha') return `${expr.min}..${expr.max}`
    if (expr.kind === 'chars') return expr.chars
    if (expr.kind === 'bind') return `${expr.name || 'unnamed'} = ${expr.child ? summary(expr.child) : 'empty'}`
    return expr.children.map((alternative) => alternative.length ? alternative.map(summary).join(' + ') : 'empty').join(' | ')
  }

  function toStructureNode(expr: Expr): StructureNode {
    const children: StructureNode[] = expr.kind === 'bind'
      ? (expr.child ? [toStructureNode(expr.child)] : [])
      : expr.kind === 'choice'
        ? expr.children.map((alternative, index) => ({ id: `${expr.id}-alternative-${index}`, label: `alternative ${String.fromCharCode(65 + index)}`, blockType: 'row', summary: alternative.map(summary).join(' + ') || 'empty', children: alternative.map(toStructureNode) }))
        : []
    return { id: expr.id, label: expr.kind, blockType: expr.kind, summary: summary(expr), children }
  }

  $: structureNodes = rootSequence.map(toStructureNode)
  $: parsed = parseSequence(rootSequence, testInput)
  $: generated = generateSequence(rootSequence)

  function updateRoot(index: number, value: Expr) { rootSequence = rootSequence.map((expr, exprIndex) => exprIndex === index ? value : expr) }
  function removeRoot(index: number) { rootSequence = rootSequence.filter((_, exprIndex) => exprIndex !== index) }
  function handleGapDragOver(state: DragDropState<GrammarDragItem>) {
    if (!state.draggedItem) return
    const target = (state.targetElement as HTMLElement | null)?.closest('.grammar-editor-sketch .dnd-insert-zone')
    const markerClass = state.draggedItem.kind === 'palette' ? 'dnd-insert-zone-new' : 'dnd-insert-zone-move'
    if (!target?.classList.contains(markerClass)) {
      clearDropMarker()
      target?.classList.add('dnd-insert-zone-active', markerClass)
    }
  }
  function handleGapDragLeave() {
    requestAnimationFrame(clearDropMarker)
  }
  function handleGapDrop(state: DragDropState<GrammarDragItem>, target: GrammarDropSlot) { handleGrammarDrop(state, target) }
  function handleGrammarDrop(state: DragDropState<DragPayload>, target: DropTarget) {
    const item = state.draggedItem
    if (!item) return
    if (item.kind === 'palette') {
      if (target.type === 'row') rootSequence = insertPath(rootSequence, target.path, createExpr(item.exprKind), target.index)
      else if (target.type === 'binding-child') rootSequence = replaceExpr(rootSequence, target.path, (expr) => expr.kind === 'bind' ? { ...expr, child: createExpr(item.exprKind) } : expr)
      else rootSequence = replaceExpr(rootSequence, target.path, (expr) => expr.kind === 'choice' ? { ...expr, children: [...expr.children, [createExpr(item.exprKind)]] } : expr)
    } else {
      const sourceRow = getPath(rootSequence, item.sourcePath)
      const sourceIndex = sourceRow.findIndex((expr) => expr.id === item.expr.id)
      if (sourceIndex < 0) return
      if (containsTarget(item.expr, target)) return
      if (target.type === 'row' && target.path.join('/') === item.sourcePath.join('/') && sourceIndex === target.index) return
      const without = removePath(rootSequence, item.sourcePath, item.expr.id)
      if (target.type === 'binding-child') rootSequence = replaceExpr(without, target.path, (expr) => expr.kind === 'bind' ? { ...expr, child: item.expr } : expr)
      else if (target.type === 'choice-alternative') rootSequence = replaceExpr(without, target.path, (expr) => expr.kind === 'choice' ? { ...expr, children: [...expr.children, [item.expr]] } : expr)
      else {
        const index = target.path.join('/') === item.sourcePath.join('/') && sourceIndex < target.index ? target.index - 1 : target.index
        rootSequence = insertPath(without, target.path, item.expr, index)
      }
    }
    clearDropMarker()
  }

  function removePath(root: Expr[], path: string[], id: string): Expr[] {
    if (!path.length) return root.filter((expr) => expr.id !== id)
    const [container, ...rest] = path
    return root.map((expr) => {
      if (expr.id !== container) return expr
      if (expr.kind === 'bind') return { ...expr, child: expr.child && removePath([expr.child], rest, id)[0] }
      if (expr.kind === 'choice') { const children = [...expr.children]; const alternative = Number(rest[0]); children[alternative] = removePath(children[alternative] || [], rest.slice(1), id); return { ...expr, children } }
      return expr
    })
  }

  function replaceExpr(root: Expr[], path: string[], change: (expr: Expr) => Expr): Expr[] {
    if (!path.length) return root.map(change)
    const id = path[0]
    return root.map((expr) => {
      if (expr.id !== id) return expr
      if (expr.kind === 'bind') return path.length === 1 ? change(expr) : { ...expr, child: expr.child ? replaceExpr([expr.child], path.slice(1), change)[0] : expr.child }
      if (expr.kind === 'choice') {
        if (path.length === 1) return change(expr)
        const children = [...expr.children]; const alternative = Number(path[1])
        children[alternative] = path.length === 2 ? children[alternative].map(change) : replaceExpr(children[alternative] || [], path.slice(2), change)
        return { ...expr, children }
      }
      return expr
    })
  }

  function insertPath(root: Expr[], path: string[], expr: Expr, index: number): Expr[] {
    if (!path.length) return [...root.slice(0, index), expr, ...root.slice(index)]
    const id = path[0]
    return root.map((found) => {
      if (found.id !== id) return found
      if (found.kind === 'bind') return { ...found, child: insertPath(found.child ? [found.child] : [], path.slice(1), expr, index)[0] }
      if (found.kind === 'choice') { const children = [...found.children]; const alternative = Number(path[1]); children[alternative] = insertPath(children[alternative] || [], path.length === 2 ? [] : path.slice(2), expr, index); return { ...found, children } }
      return found
    })
  }

  function getPath(root: Expr[], path: string[]): Expr[] {
    if (!path.length) return root
    const [id, ...rest] = path
    const found = root.find((expr) => expr.id === id)
    if (!found) return []
    if (found.kind === 'bind') return rest.length ? getPath(found.child ? [found.child] : [], rest) : (found.child ? [found.child] : [])
    if (found.kind === 'choice') { const alternative = found.children[Number(rest[0])] || []; return rest.length === 1 ? alternative : getPath(alternative, rest.slice(1)) }
    return []
  }
  function containsTarget(source: Expr, target: DropTarget): boolean {
    if (target.type === 'row' && target.path.length === 0) return false
    return target.path.includes(source.id)
  }
  function clearDropMarker() {
    document.querySelectorAll('.grammar-editor-sketch .dnd-insert-zone-active').forEach((element) => element.classList.remove('dnd-insert-zone-active', 'dnd-insert-zone-new', 'dnd-insert-zone-move'))
  }
  function handleDragEnd() { clearDropMarker() }
</script>

<p class="eyebrow">Grammar editor</p>
<h2>Grammar editor</h2>
<p class="grammar-editor-sketch-intro">Build a sequence by dragging blocks from the palette into the editor. Use the structure map to move through nested blocks.</p>
<section class="grammar-editor-sketch" aria-label="Grammar editor layout sketch">
  <div class="grammar-editor-sketch-palette">
    <span class="mock-section-label">Block palette</span>
    <div class="grammar-editor-sketch-palette-grid">
      {#each paletteTiles as tile}
        <SequenceExampleTile tile={{ kind: 'sequence-tile', ...tile }} variant="palette" container="grammar-editor:palette" dragData={{ kind: 'palette', exprKind: tile.type as ExprKind }} onDragEnd={handleDragEnd} />
      {/each}
    </div>
  </div>
  <div class="grammar-editor-sketch-layout">
    <section class="grammar-editor-sketch-canvas" aria-labelledby="grammar-editor-canvas-heading">
      <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-canvas-heading">Grammar sequence</span><span class="grammar-editor-sketch-count">3 blocks</span></div>
      <DndField filled container="grammar-editor:root" onDrop={() => {}} onDragOver={() => {}} onGapDragLeave={() => {}} onGapDrop={() => {}} onDragEnd={() => {}} onRemove={() => {}}>
        <div class="grammar-editor-sketch-row">
          <GrammarEditorRow exprs={rootSequence} path={[]} container="grammar-editor:root" onDragOver={handleGapDragOver} onDragLeave={handleGapDragLeave} onDrop={handleGapDrop} onDragEnd={handleDragEnd} onChange={updateRoot} onRemove={removeRoot} />
        </div>
        <div class="grammar-editor-sketch-drop-zone">Drop a block here to add it to the root sequence</div>
      </DndField>
    </section>
    <aside class="grammar-editor-sketch-map" aria-labelledby="grammar-editor-map-heading">
      <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-map-heading">Structure map</span><span class="grammar-editor-sketch-count">3 blocks</span></div>
      <div class="grammar-editor-sketch-map-tree"><StructureTree nodes={structureNodes} showIds={false} /></div>
      <section class="grammar-editor-sketch-test" aria-labelledby="grammar-editor-test-heading">
        <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-test-heading">Test parse</span></div>
        <label class="grammar-input-label">Input to parse<input bind:value={testInput} placeholder="Type a candidate value" /></label>
        {#if testInput}<div class:grammar-valid={!('error' in parsed)} class="grammar-parse-result">{#if 'error' in parsed}Invalid: {parsed.error}{:else}Valid input{#each Object.entries(parsed.bindings) as [name, value]}<span>{name} = {value}</span>{/each}{/if}</div>{/if}
        <div class="grammar-output"><span>Canonical generated value</span><strong>{generated || '(empty)'}</strong></div>
      </section>
    </aside>
  </div>
</section>
