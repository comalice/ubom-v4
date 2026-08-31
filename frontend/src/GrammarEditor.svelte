<script lang="ts">
  import GrammarEditorRow, { type GrammarDragItem, type GrammarDropSlot } from './GrammarEditorRow.svelte'
  import { type DragDropState } from '@thisux/sveltednd'
  import StructureTree from './StructureTree.svelte'
  import type { StructureNode } from './StructureNode'
  import SequenceExampleTile from './SequenceExampleTile.svelte'
  import { previewSeqDefCandidate, type SeqDefCandidatePreview, type SeqDefNode } from './api'
  import { createGrammarExpr as createExpr, fromSeqDefRoot, parseGrammarSequence as parseSequence, toSeqDefRoot, validateGrammarSequence as validateSequence, type GrammarDragPayload as DragPayload, type GrammarDropTarget as DropTarget, type GrammarExpr as Expr, type GrammarExprKind as ExprKind } from './GrammarEditorModel'

  export let initialRoot: SeqDefNode | undefined = undefined
  export let readOnly = false
  export let onChange: ((root: SeqDefNode) => void) | undefined = undefined
  export let onValidationChange: ((invalid: boolean) => void) | undefined = undefined

  const paletteTiles = [
    { type: 'literal', title: 'Literal', description: 'Exact text' },
    { type: 'range', title: 'Range', description: 'Numeric range' },
    { type: 'places', title: 'Places', description: 'Allowed characters per position' },
    { type: 'range_radix', title: 'Radix range', description: 'Ordered fields' },
    { type: 'bind', title: 'Binding', description: 'Capture a value' },
    { type: 'choice', title: 'Choice', description: 'One alternative' },
  ]

  let rootSequence: Expr[] = initialRoot ? fromSeqDefRoot(initialRoot) : [
    { id: 'literal-1', kind: 'literal', text: 'PN-' },
    { id: 'bind-1', kind: 'bind', name: 'family', child: { id: 'literal-2', kind: 'literal', text: 'SN-' } },
    { id: 'choice-1', kind: 'choice', children: [
      [{ id: 'literal-3', kind: 'literal', text: 'A' }],
      [{ id: 'literal-4', kind: 'literal', text: 'B' }, { id: 'range-1', kind: 'range', min: 0, max: 9, width: 1, pad: '0' }],
    ] },
  ]
  let structureNodes: StructureNode[] = []
  let testInput = 'PN-SN-A'
  let candidatePreview: SeqDefCandidatePreview | null = null
  let candidatePreviewKey: string | null = null
  let candidateError = ''
  let candidateErrorKey: string | null = null
  let candidateRequestKey: string | null = null
  let candidatePreviewing = false
  let candidateRequestId = 0

  function summary(expr: Expr): string {
    if (expr.kind === 'literal') return JSON.stringify(expr.text)
    if (expr.kind === 'range') return `${expr.min}..${expr.max} / width ${expr.width} / pad ${expr.pad}`
    if (expr.kind === 'places') return expr.alphabets.map((alphabet) => `[${alphabet}]`).join('')
    if (expr.kind === 'range_radix') return expr.children.map(summary).join(' + ') || 'empty'
    if (expr.kind === 'bind') return `${expr.name || 'unnamed'} = ${expr.child ? summary(expr.child) : 'empty'}`
    return expr.children.map((alternative) => alternative.length ? alternative.map(summary).join(' + ') : 'empty').join(' | ')
  }

  function toStructureNode(expr: Expr): StructureNode {
    const children: StructureNode[] = expr.kind === 'bind'
      ? (expr.child ? [toStructureNode(expr.child)] : [])
      : expr.kind === 'choice'
        ? expr.children.map((alternative, index) => ({ id: `${expr.id}-alternative-${index}`, label: `alternative ${String.fromCharCode(65 + index)}`, blockType: 'row', summary: alternative.map(summary).join(' + ') || 'empty', children: alternative.map(toStructureNode) }))
        : expr.kind === 'range_radix'
          ? expr.children.map(toStructureNode)
          : []
    return { id: expr.id, label: expr.kind, blockType: expr.kind, summary: summary(expr), children }
  }

  $: structureNodes = rootSequence.map(toStructureNode)
  $: validationErrors = validateSequence(rootSequence)
  $: parsed = parseSequence(rootSequence, testInput)
  $: candidateKey = JSON.stringify({ grammar: rootSequence, value: testInput })
  $: currentCandidatePreview = candidatePreviewKey === candidateKey ? candidatePreview : null
  $: currentCandidateError = candidateErrorKey === candidateKey ? candidateError : ''
  $: candidatePending = candidatePreviewing && candidateRequestKey === candidateKey
  $: onValidationChange?.(validationErrors.length > 0)
  $: if (validationErrors.length === 0) onChange?.(toSeqDefRoot(rootSequence))

  async function requestCandidatePreview() {
    if (validationErrors.length > 0) return
    const requestKey = candidateKey
    const requestId = ++candidateRequestId
    const root = toSeqDefRoot(rootSequence)
    candidatePreview = null
    candidatePreviewKey = null
    candidateError = ''
    candidateErrorKey = null
    candidateRequestKey = requestKey
    candidatePreviewing = true
    try {
      const result = await previewSeqDefCandidate(root, testInput)
      if (requestId === candidateRequestId && requestKey === candidateKey) {
        candidatePreview = result
        candidatePreviewKey = requestKey
      }
    } catch (reason) {
      if (requestId === candidateRequestId && requestKey === candidateKey) {
        candidateError = reason instanceof Error ? reason.message : 'Unable to preview candidate'
        candidateErrorKey = requestKey
      }
    } finally {
      if (requestId === candidateRequestId && candidateRequestKey === requestKey) candidatePreviewing = false
    }
  }

  function updateRoot(index: number, value: Expr) { if (!readOnly) rootSequence = rootSequence.map((expr, exprIndex) => exprIndex === index ? value : expr) }
  function removeRoot(index: number) { if (!readOnly) rootSequence = rootSequence.filter((_, exprIndex) => exprIndex !== index) }
  function handleGapDragOver(state: DragDropState<GrammarDragItem>) {
    if (!state.draggedItem) return
  }
  function handleGapDragLeave() {}
  function handleGapDrop(state: DragDropState<GrammarDragItem>, target: GrammarDropSlot) { handleGrammarDrop(state, target) }
  function handleGrammarDrop(state: DragDropState<DragPayload>, target: DropTarget) {
    if (readOnly) return
    const item = state.draggedItem
    if (!item) return
    if (item.kind === 'palette') {
      if (target.type === 'row') rootSequence = insertPath(rootSequence, target.path, createExpr(item.exprKind), target.index)
      else if (target.type === 'binding-child') rootSequence = replaceExpr(rootSequence, target.path, (expr) => expr.kind === 'bind' ? { ...expr, child: createExpr(item.exprKind) } : expr)
      else if (target.type === 'choice-alternative') rootSequence = replaceExpr(rootSequence, target.path, (expr) => expr.kind === 'choice' ? { ...expr, children: [...expr.children, [createExpr(item.exprKind)]] } : expr)
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
  }

  function removePath(root: Expr[], path: string[], id: string): Expr[] {
    if (!path.length) return root.filter((expr) => expr.id !== id)
    const [container, ...rest] = path
    return root.map((expr) => {
      if (expr.id !== container) return expr
      if (expr.kind === 'bind') return { ...expr, child: expr.child && removePath([expr.child], rest, id)[0] }
      if (expr.kind === 'choice') { const children = [...expr.children]; const alternative = Number(rest[0]); children[alternative] = removePath(children[alternative] || [], rest.slice(1), id); return { ...expr, children } }
      if (expr.kind === 'range_radix') return { ...expr, children: removePath(expr.children, rest, id) }
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
      if (expr.kind === 'range_radix') return { ...expr, children: path.length === 1 ? expr.children.map(change) : replaceExpr(expr.children, path.slice(1), change) }
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
      if (found.kind === 'range_radix') return { ...found, children: insertPath(found.children, path.slice(1), expr, index) }
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
    if (found.kind === 'range_radix') return getPath(found.children, rest)
    return []
  }
  function containsTarget(source: Expr, target: DropTarget): boolean {
    if (target.type === 'row' && target.path.length === 0) return false
    return target.path.includes(source.id)
  }
  function handleDragEnd() {}
</script>

<p class="eyebrow">Grammar editor</p>
<h2>Grammar editor</h2>
<p class="grammar-editor-intro">{readOnly ? 'Review the sequence and use the structure map to move through nested blocks.' : 'Build a sequence by dragging blocks from the palette into the editor. Use the structure map to move through nested blocks.'}</p>
{#if readOnly}<p class="muted" role="status">This immutable definition is read-only.</p>{/if}
<section class="grammar-editor" aria-label="Grammar editor layout sketch">
  {#if !readOnly}<div class="grammar-editor-palette">
    <span class="mock-section-label">Block palette</span>
    <div class="grammar-editor-palette-grid">
      {#each paletteTiles as tile}
        <SequenceExampleTile tile={{ kind: 'sequence-tile', ...tile }} variant="palette" container="grammar-editor:palette" dragData={{ kind: 'palette', exprKind: tile.type as ExprKind }} onDragEnd={handleDragEnd} />
      {/each}
    </div>
  </div>{/if}
  <div class="grammar-editor-layout">
    <div class="grammar-editor-canvas">
      <div class="grammar-editor-heading"><span class="mock-section-label" id="grammar-editor-canvas-heading">Grammar sequence</span><span class="grammar-editor-count">{rootSequence.length} blocks</span></div>
      {#if validationErrors.length}<section class="grammar-validation" aria-labelledby="grammar-validation-heading" role="alert"><strong id="grammar-validation-heading">Fix editor structure</strong><ul>{#each validationErrors as validationError}<li><code>{validationError.path}</code> {validationError.message}</li>{/each}</ul></section>{/if}
      <div class="grammar-editor-sequence-row">
        <GrammarEditorRow exprs={rootSequence} path={[]} container="grammar-editor:root" validationErrors={validationErrors} readOnly={readOnly} onDragOver={handleGapDragOver} onDragLeave={handleGapDragLeave} onDrop={handleGapDrop} onDragEnd={handleDragEnd} onChange={updateRoot} onRemove={removeRoot} />
      </div>
    </div>
    <aside class="grammar-editor-map" aria-labelledby="grammar-editor-map-heading">
      <div class="grammar-editor-heading"><span class="mock-section-label" id="grammar-editor-map-heading">Structure map</span><span class="grammar-editor-count">{rootSequence.length} blocks</span></div>
      <div class="grammar-editor-map-tree"><StructureTree nodes={structureNodes} showIds={false} /></div>
      <section class="grammar-editor-test" aria-labelledby="grammar-editor-test-heading">
        <div class="grammar-editor-heading"><span class="mock-section-label" id="grammar-editor-test-heading">Test parse</span></div>
        <label class="grammar-input-label">Input to parse<input bind:value={testInput} placeholder="Type a candidate value" /></label>
        {#if testInput}<div class:grammar-valid={!('error' in parsed)} class="grammar-parse-result">{#if 'error' in parsed}Invalid: {parsed.error}{:else}Valid input{#each Object.entries(parsed.bindings) as [name, value]}<span>{name} = {value}</span>{/each}{/if}</div>{/if}
        <section class="grammar-candidate-preview" aria-labelledby="grammar-candidate-preview-heading">
          <div class="grammar-editor-heading"><span class="mock-section-label" id="grammar-candidate-preview-heading">Candidate validation</span></div>
          <p class="muted">Validate the complete test input against this grammar on the server.</p>
          <button class="grammar-preview-button" type="button" disabled={candidatePreviewing || validationErrors.length > 0} onclick={requestCandidatePreview}>{candidatePending ? 'Previewing candidate...' : 'Preview candidate on server'}</button>
          <div class="grammar-preview-result" class:grammar-preview-failed={!!currentCandidateError} class:grammar-preview-pending={candidatePending} role="status" aria-live="polite">
            {#if candidatePending}
              <strong>Candidate preview in progress</strong>
              <span>Waiting for the server to validate the complete input.</span>
            {:else if currentCandidateError}
              <strong>Server error</strong>
              <span>{currentCandidateError}</span>
            {:else if currentCandidatePreview}
              <strong>Server candidate: {currentCandidatePreview.value}</strong>
              <span>Parsed bindings: {Object.entries(currentCandidatePreview.bindings).map(([name, value]) => `${name} = ${value}`).join(', ') || 'none'}</span>
              <span>Captured fields: {currentCandidatePreview.fields.map((field) => `${field.name} = ${field.source} (normalized: ${JSON.stringify(field.normalized)})`).join(', ') || 'none'}</span>
            {:else}
              <strong>Candidate not previewed</strong>
              <span>Preview the complete test input to see the authoritative server result.</span>
            {/if}
          </div>
        </section>
      </section>
    </aside>
  </div>
</section>
