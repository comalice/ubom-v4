<script lang="ts">
  import PlaygroundBlockShell from './PlaygroundBlockShell.svelte'
  import PlaygroundField from './PlaygroundField.svelte'
  import PlaygroundStructurePreview, { type StructureNode } from './PlaygroundStructurePreview.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'

  const paletteTiles = [
    { type: 'literal', title: 'Literal', description: 'Exact text' },
    { type: 'number', title: 'Number', description: 'Numeric range' },
    { type: 'alpha', title: 'Letters', description: 'Alphabetic range' },
    { type: 'chars', title: 'Characters', description: 'Allowed characters' },
    { type: 'bind', title: 'Binding', description: 'Capture a value' },
    { type: 'choice', title: 'Choice', description: 'One alternative' },
  ]

  const structureNodes: StructureNode[] = [
    { id: 1, label: 'literal', blockType: 'literal', children: [] },
    { id: 2, label: 'binding', blockType: 'bind', children: [{ id: 3, label: 'literal', blockType: 'literal', children: [] }] },
    { id: 4, label: 'choice', blockType: 'choice', children: [
      { id: 5, label: 'alternative A', blockType: 'row', children: [{ id: 6, label: 'literal', blockType: 'literal', children: [] }] },
      { id: 7, label: 'alternative B', blockType: 'row', children: [
        { id: 8, label: 'literal', blockType: 'literal', children: [] },
        { id: 9, label: 'number', blockType: 'number', children: [] },
      ] },
    ] },
  ]
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
      <div class="grammar-editor-sketch-root-label">Root sequence</div>
      <PlaygroundField filled container="grammar-editor:root" onDrop={() => {}} onDragOver={() => {}} onGapDragLeave={() => {}} onGapDrop={() => {}} onDragEnd={() => {}} onRemove={() => {}}>
        <div class="grammar-editor-sketch-row">
          <PlaygroundBlockShell title="literal" summary={'"PN-"'} kind="literal" collapsible initiallyOpen={false} onRemove={() => {}}>
            <div class="grammar-editor-sketch-block-body">Text: "PN-"</div>
          </PlaygroundBlockShell>
          <PlaygroundBlockShell title="binding" summary={'family = "SN-"'} kind="bind" collapsible initiallyOpen={false} onRemove={() => {}}>
            <div class="grammar-editor-sketch-block-body">Name: family · Captures: literal</div>
          </PlaygroundBlockShell>
          <PlaygroundBlockShell title="choice" summary={'"A" | "B"'} kind="choice" collapsible initiallyOpen={false} onRemove={() => {}}>
            <div class="grammar-editor-sketch-block-body">2 alternatives</div>
          </PlaygroundBlockShell>
        </div>
        <div class="grammar-editor-sketch-drop-zone">Drop a block here to add it to the root sequence</div>
      </PlaygroundField>
    </section>
    <aside class="grammar-editor-sketch-map" aria-labelledby="grammar-editor-map-heading">
      <div class="grammar-editor-sketch-heading"><span class="mock-section-label" id="grammar-editor-map-heading">Structure map</span><span class="grammar-editor-sketch-count">3 blocks</span></div>
      <div class="grammar-editor-sketch-map-tree"><PlaygroundStructurePreview nodes={structureNodes} showIds={false} /></div>
    </aside>
  </div>
</section>
