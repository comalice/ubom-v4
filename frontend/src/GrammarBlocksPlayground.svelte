<script lang="ts">
  import PlaygroundBlockShell from './PlaygroundBlockShell.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'
</script>

<p class="eyebrow">Grammar blocks</p>
<h2>Grammar blocks</h2>
<section class="block-examples" aria-labelledby="basic-expressions-heading">
  <span class="mock-section-label" id="basic-expressions-heading">Basic expressions</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'literal', title: 'Literal', description: 'Exact text' }} variant="palette" container="grammar-examples:palette" />
      <PlaygroundBlockShell title="literal" summary={'"PN-"'} kind="literal" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body"><label>Text<input value="PN-" aria-label="Literal text" /></label></div>
      </PlaygroundBlockShell>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'number', title: 'Number', description: 'Numeric range' }} variant="palette" container="grammar-examples:palette" />
      <PlaygroundBlockShell title="number" summary="0..9 / width 2 / pad 0" kind="number" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body grammar-number-fields"><label>Min<input value="0" aria-label="Number minimum" /></label><label>Max<input value="9" aria-label="Number maximum" /></label><label>Width<input value="2" aria-label="Number width" /></label><label>Pad<input value="0" aria-label="Number padding" /></label></div>
      </PlaygroundBlockShell>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'alpha', title: 'Letters', description: 'Alphabetic range' }} variant="palette" container="grammar-examples:palette" />
      <PlaygroundBlockShell title="letters" summary="A..Z" kind="alpha" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body grammar-two-fields"><label>From<input value="A" aria-label="Letters minimum" /></label><span>to</span><label>To<input value="Z" aria-label="Letters maximum" /></label></div>
      </PlaygroundBlockShell>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'chars', title: 'Characters', description: 'Allowed characters' }} variant="palette" container="grammar-examples:palette" />
      <PlaygroundBlockShell title="characters" summary="ABC012" kind="chars" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body"><label>Allowed characters<input value="ABC012" aria-label="Allowed characters" /></label></div>
      </PlaygroundBlockShell>
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="binding-blocks-heading">
  <span class="mock-section-label" id="binding-blocks-heading">Binding blocks</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'bind', title: 'Binding', description: 'Capture a value' }} variant="palette" container="grammar-examples:palette" />
      <PlaygroundBlockShell title="binding" summary={'family = "PN-"'} kind="bind" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body">
          <label>Name<input value="family" aria-label="Binding name" /></label>
          <div class="grammar-child-label">Captures</div>
          <div class="grammar-binding-target">
            <PlaygroundBlockShell title="literal" summary={'"PN-"'} kind="literal" size="small" collapsible initiallyOpen={false} onRemove={() => {}}>
              <div class="grammar-expr-body"><label>Text<input value="PN-" aria-label="Captured literal text" /></label></div>
            </PlaygroundBlockShell>
          </div>
        </div>
      </PlaygroundBlockShell>
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="choice-blocks-heading">
  <span class="mock-section-label" id="choice-blocks-heading">Choice blocks</span>
</section>
