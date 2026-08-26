<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import SequenceExampleTile from './SequenceExampleTile.svelte'

  type ChoiceBlock = { kind: 'literal' | 'range'; summary: string }
  type RadixRangePosition =
    | { kind: 'range'; min: number; max: number; width: number; pad: string }
    | { kind: 'places'; chars: string }
  let choiceAlternatives: ChoiceBlock[][] = [
    [{ kind: 'literal', summary: '"PN-"' }],
      [{ kind: 'literal', summary: '"SN-"' }, { kind: 'range', summary: '00..99 / width 2 / pad 0' }],
  ]
  let radixRangePositions: RadixRangePosition[] = [
    { kind: 'places', chars: 'ABC' },
    { kind: 'range', min: 0, max: 9, width: 1, pad: '0' },
    { kind: 'places', chars: 'XYZ' },
  ]

  $: choiceSummary = choiceAlternatives.map((alternative) => alternative.length ? alternative.map((block) => block.summary).join(' + ') : 'empty').join(' | ')
  $: radixRangeSummary = radixRangePositions.map((position) => position.kind === 'range'
    ? `${position.min}..${position.max} / width ${position.width} / pad ${position.pad}`
    : `[${position.chars}]`).join(' + ')
  $: radixRangeExample = radixRangePositions.map((position) => position.kind === 'range'
    ? String(position.min).padStart(Math.max(1, position.width), position.pad || '0')
    : position.chars[0] || '(empty)').join('')
  $: choiceExample = choiceAlternatives[0]?.map((block) => block.kind === 'literal' ? block.summary.replace(/^"|"$/g, '') : '00').join('') || '(empty)'
  function addChoiceAlternative() { choiceAlternatives = [...choiceAlternatives, []] }
  function updateRadixRangePosition(index: number, changes: Partial<RadixRangePosition>) {
    radixRangePositions = radixRangePositions.map((position, positionIndex) => positionIndex === index ? { ...position, ...changes } : position)
  }
  function addRadixRangePosition() {
    radixRangePositions = [...radixRangePositions, { kind: 'range', min: 0, max: 9, width: 1, pad: '0' }]
  }
  function removeRadixRangePosition(index: number) {
    if (radixRangePositions.length <= 1) return
    radixRangePositions = radixRangePositions.filter((_, positionIndex) => positionIndex !== index)
  }
</script>

<p class="eyebrow">Grammar blocks</p>
<h2>Grammar blocks</h2>
<section class="block-examples" aria-labelledby="basic-expressions-heading">
  <span class="mock-section-label" id="basic-expressions-heading">Basic expressions</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'literal', title: 'Literal', description: 'Exact text' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="literal" summary={'"PN-"'} kind="literal" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body"><label>Text<input value="PN-" aria-label="Literal text" /></label></div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Matches exact text.</p><div class="grammar-block-example"><span>Example output</span><code>PN-</code></div></div>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'range', title: 'Range', description: 'Numeric range' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="range" summary="0..9 / width 2 / pad 0" kind="range" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body grammar-number-fields"><label>Min<input value="0" aria-label="Range minimum" /></label><label>Max<input value="9" aria-label="Range maximum" /></label><label>Width<input value="2" aria-label="Range width" /></label><label>Pad<input value="0" aria-label="Range padding" /></label></div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Matches a numeric value within an inclusive range.</p><div class="grammar-block-example"><span>Example output</span><code>00</code></div></div>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'range_radix', title: 'Radix range', description: 'Ordered numeric positions' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="range_radix" summary={radixRangeSummary} kind="range_radix" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body grammar-radix-range">
          <div class="grammar-child-label">Ordered positions</div>
          {#each radixRangePositions as position, positionIndex}
            <div class="grammar-radix-position">
              <span class="grammar-alternative-label" aria-hidden="true">{positionIndex + 1}</span>
              <div class="grammar-radix-position-fields">
                <label>Field type<select value={position.kind} aria-label={`Radix range position ${positionIndex + 1} field type`} on:change={(event) => {
                  const kind = (event.currentTarget as HTMLSelectElement).value
                  radixRangePositions = radixRangePositions.map((current, currentIndex) => currentIndex === positionIndex
                    ? kind === 'places' ? { kind: 'places', chars: 'ABC' } : { kind: 'range', min: 0, max: 9, width: 1, pad: '0' }
                    : current)
                }}><option value="range">Numeric range</option><option value="places">Character set</option></select></label>
                {#if position.kind === 'range'}
                  <div class="grammar-number-fields">
                    <label>Min<input type="number" value={position.min} aria-label={`Radix range position ${positionIndex + 1} minimum`} on:input={(event) => updateRadixRangePosition(positionIndex, { min: Number((event.currentTarget as HTMLInputElement).value) })} /></label>
                    <label>Max<input type="number" value={position.max} aria-label={`Radix range position ${positionIndex + 1} maximum`} on:input={(event) => updateRadixRangePosition(positionIndex, { max: Number((event.currentTarget as HTMLInputElement).value) })} /></label>
                    <label>Width<input type="number" min="1" value={position.width} aria-label={`Radix range position ${positionIndex + 1} width`} on:input={(event) => updateRadixRangePosition(positionIndex, { width: Number((event.currentTarget as HTMLInputElement).value) })} /></label>
                    <label>Pad<input maxlength="1" value={position.pad} aria-label={`Radix range position ${positionIndex + 1} padding`} on:input={(event) => updateRadixRangePosition(positionIndex, { pad: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label>
                  </div>
                {:else}
                  <label>Allowed characters<input value={position.chars} aria-label={`Radix range position ${positionIndex + 1} allowed characters`} on:input={(event) => updateRadixRangePosition(positionIndex, { chars: (event.currentTarget as HTMLInputElement).value })} /></label>
                {/if}
              </div>
              {#if radixRangePositions.length > 1}<button class="grammar-remove grammar-remove-alternative" type="button" aria-label={`Remove radix range position ${positionIndex + 1}`} on:click={() => removeRadixRangePosition(positionIndex)}>x</button>{/if}
            </div>
          {/each}
          <button class="grammar-add" type="button" on:click={addRadixRangePosition}>+ add position</button>
        </div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Evaluates ordered fields from left to right.</p><div class="grammar-block-example"><span>Example output</span><code>{radixRangeExample}</code></div></div>
    </article>
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'places', title: 'Places', description: 'Allowed characters per position' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="places" summary="[ABC012]" kind="places" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body"><label>Allowed characters<input value="ABC012" aria-label="Places allowed characters" /></label></div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Matches one character from the configured set.</p><div class="grammar-block-example"><span>Example output</span><code>A</code></div></div>
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="binding-blocks-heading">
  <span class="mock-section-label" id="binding-blocks-heading">Binding blocks</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'bind', title: 'Binding', description: 'Capture a value' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="binding" summary={'family = "PN-"'} kind="bind" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body">
          <label>Name<input value="family" aria-label="Binding name" /></label>
          <div class="grammar-child-label">Captures</div>
          <div class="grammar-binding-target">
            <BlockShell title="literal" summary={'"PN-"'} kind="literal" size="small" collapsible initiallyOpen={false} onRemove={() => {}}>
              <div class="grammar-expr-body"><label>Text<input value="PN-" aria-label="Captured literal text" /></label></div>
            </BlockShell>
          </div>
        </div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Captures the child value under a name.</p><div class="grammar-block-example"><span>Example output</span><code>family = PN-</code></div></div>
    </article>
  </div>
</section>
<section class="block-examples" aria-labelledby="choice-blocks-heading">
  <span class="mock-section-label" id="choice-blocks-heading">Choice blocks</span>
  <div class="grammar-basic-example-list">
    <article class="grammar-basic-example">
      <SequenceExampleTile tile={{ kind: 'sequence-tile', type: 'choice', title: 'Choice', description: 'One alternative' }} variant="palette" container="grammar-examples:palette" />
      <BlockShell title="choice" summary={choiceSummary} kind="choice" collapsible initiallyOpen={false} onRemove={() => {}}>
        <div class="grammar-expr-body grammar-choice">
          <div class="grammar-child-label">Alternatives</div>
          {#each choiceAlternatives as alternative, alternativeIndex}
            <div class="grammar-alternative">
              <span class="grammar-alternative-label">{String.fromCharCode(65 + alternativeIndex)}</span>
              <div class="grammar-choice-block">
                {#if alternative.length}
                  {#each alternative as block, blockIndex}
                    {#if block.kind === 'literal'}
                      <BlockShell title="literal" summary={block.summary} kind="literal" size="small" collapsible initiallyOpen={false} onRemove={() => {}}>
                        <div class="grammar-expr-body"><label>Text<input value="{blockIndex === 0 ? 'PN-' : 'SN-'}" aria-label="Choice alternative literal text" /></label></div>
                      </BlockShell>
                    {:else}
                      <BlockShell title="range" summary={block.summary} kind="range" size="small" collapsible initiallyOpen={false} onRemove={() => {}}>
                        <div class="grammar-expr-body grammar-number-fields"><label>Min<input value="0" aria-label="Choice alternative range minimum" /></label><label>Max<input value="99" aria-label="Choice alternative range maximum" /></label><label>Width<input value="2" aria-label="Choice alternative range width" /></label><label>Pad<input value="0" aria-label="Choice alternative range padding" /></label></div>
                      </BlockShell>
                    {/if}
                  {/each}
                {:else}
                  <span class="grammar-row-empty">Drop blocks here to create an alternative</span>
                {/if}
              </div>
            </div>
          {/each}
          <button class="grammar-add" type="button" on:click={addChoiceAlternative}>+ add alternative</button>
        </div>
      </BlockShell>
      <div class="grammar-block-meta"><p class="grammar-block-description">Matches one of the configured alternatives.</p><div class="grammar-block-example"><span>Example output</span><code>{choiceExample}</code></div></div>
    </article>
  </div>
</section>
