<script lang="ts">
  import GrammarEditorBlock from './GrammarEditorBlock.svelte'
  import PlaygroundField from './PlaygroundField.svelte'
  import PlaygroundStructurePreview, { type StructureNode } from './PlaygroundStructurePreview.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'
  import { generateSequence, parseSequence, type Expr } from './grammarMock'

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
</script>

<p class="eyebrow">Grammar editor</p>
<h2>Grammar editor</h2>
<p class="grammar-editor-sketch-intro">Build a sequence by dragging blocks from the palette into the editor. Use the structure map to move through nested blocks.</p>
<section class="grammar-editor-sketch" aria-label="Grammar editor layout sketch">
  <div class="grammar-editor-sketch-palette">
    <span class="mock-section-label">Block palette</span>
    <div class="grammar-editor-sketch-palette-grid">
      {#each paletteTiles as tile}
        <SequenceExampleTile tile={{ kind: 'sequence-tile', ...tile }} variant="palette" container="grammar-editor:palette" />
      {/each}
    </div>
  </div>
  <div class="grammar-editor-sketch-layout">
    <section class="grammar-editor-sketch-canvas" aria-labelledby="grammar-editor-canvas-heading">
      <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-canvas-heading">Grammar sequence</span><span class="grammar-editor-sketch-count">3 blocks</span></div>
      <PlaygroundField filled container="grammar-editor:root" onDrop={() => {}} onDragOver={() => {}} onGapDragLeave={() => {}} onGapDrop={() => {}} onDragEnd={() => {}} onRemove={() => {}}>
        <div class="grammar-editor-sketch-row">
          {#each rootSequence as expr, index (expr.id)}<GrammarEditorBlock {expr} onChange={(value) => updateRoot(index, value)} onRemove={() => removeRoot(index)} />{/each}
        </div>
        <div class="grammar-editor-sketch-drop-zone">Drop a block here to add it to the root sequence</div>
      </PlaygroundField>
    </section>
    <aside class="grammar-editor-sketch-map" aria-labelledby="grammar-editor-map-heading">
      <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-map-heading">Structure map</span><span class="grammar-editor-sketch-count">3 blocks</span></div>
      <div class="grammar-editor-sketch-map-tree"><PlaygroundStructurePreview nodes={structureNodes} showIds={false} /></div>
      <section class="grammar-editor-sketch-test" aria-labelledby="grammar-editor-test-heading">
        <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-test-heading">Test parse</span></div>
        <label class="grammar-input-label">Input to parse<input bind:value={testInput} placeholder="Type a candidate value" /></label>
        {#if testInput}<div class:grammar-valid={!('error' in parsed)} class="grammar-parse-result">{#if 'error' in parsed}Invalid: {parsed.error}{:else}Valid input{#each Object.entries(parsed.bindings) as [name, value]}<span>{name} = {value}</span>{/each}{/if}</div>{/if}
        <div class="grammar-output"><span>Canonical generated value</span><strong>{generated || '(empty)'}</strong></div>
      </section>
    </aside>
  </div>
</section>
