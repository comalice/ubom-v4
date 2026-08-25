<script lang="ts">
  import PlaygroundBlockShell from './PlaygroundBlockShell.svelte'
  import type { Expr } from './grammarMock'

  export let expr: Expr
  export let onChange: (expr: Expr) => void = () => {}
  export let onRemove: () => void = () => {}

  function update(value: Partial<Expr>) { onChange({ ...expr, ...value } as Expr) }
  function title() {
    return expr.kind === 'number' ? 'number' : expr.kind
  }

  function summary() {
    if (expr.kind === 'literal') return JSON.stringify(expr.text)
    if (expr.kind === 'number') return `${expr.min}..${expr.max} / width ${expr.width} / pad ${expr.pad}`
    if (expr.kind === 'alpha') return `${expr.min}..${expr.max}`
    if (expr.kind === 'chars') return expr.chars
    if (expr.kind === 'bind') return `${expr.name || 'unnamed'} = ${expr.child ? summaryFor(expr.child) : 'empty'}`
    return expr.children.map((alternative) => alternative.length ? alternative.map(summaryFor).join(' + ') : 'empty').join(' | ')
  }

  function summaryFor(value: Expr): string {
    if (value.kind === 'literal') return JSON.stringify(value.text)
    if (value.kind === 'number') return `${value.min}..${value.max}`
    if (value.kind === 'alpha') return `${value.min}..${value.max}`
    if (value.kind === 'chars') return value.chars
    if (value.kind === 'bind') return value.name || 'unnamed'
    return 'choice'
  }

  function removeChild(alternativeIndex: number, childIndex: number) {
    if (expr.kind !== 'choice') return
    update({ children: expr.children.map((alternative, index) => index === alternativeIndex ? alternative.filter((_, index) => index !== childIndex) : alternative) })
  }

  function addAlternative() {
    if (expr.kind === 'choice') update({ children: [...expr.children, []] })
  }

  function removeAlternative(index: number) {
    if (expr.kind === 'choice') update({ children: expr.children.filter((_, alternativeIndex) => alternativeIndex !== index) })
  }
</script>

<PlaygroundBlockShell title={title()} summary={summary()} kind={expr.kind} collapsible initiallyOpen={false} onRemove={onRemove}>
  {#if expr.kind === 'literal'}
    <div class="grammar-expr-body"><label>Text<input value={expr.text} aria-label="Literal text" on:input={(event) => update({ text: (event.currentTarget as HTMLInputElement).value })} /></label></div>
  {:else if expr.kind === 'number'}
    <div class="grammar-expr-body grammar-number-fields"><label>Min<input type="number" value={expr.min} aria-label="Number minimum" on:input={(event) => update({ min: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Max<input type="number" value={expr.max} aria-label="Number maximum" on:input={(event) => update({ max: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Width<input type="number" value={expr.width} aria-label="Number width" on:input={(event) => update({ width: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Pad<input maxlength="1" value={expr.pad} aria-label="Number padding" on:input={(event) => update({ pad: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'alpha'}
    <div class="grammar-expr-body grammar-two-fields"><label>From<input maxlength="1" value={expr.min} aria-label="Letters minimum" on:input={(event) => update({ min: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label><span>to</span><label>To<input maxlength="1" value={expr.max} aria-label="Letters maximum" on:input={(event) => update({ max: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'chars'}
    <div class="grammar-expr-body"><label>Allowed characters<input value={expr.chars} aria-label="Allowed characters" on:input={(event) => update({ chars: (event.currentTarget as HTMLInputElement).value })} /></label></div>
  {:else if expr.kind === 'bind'}
    <div class="grammar-expr-body"><label>Name<input value={expr.name} aria-label="Binding name" on:input={(event) => update({ name: (event.currentTarget as HTMLInputElement).value })} /></label><div class="grammar-child-label">Captures</div><div class="grammar-binding-target">{#if expr.child}<svelte:self expr={expr.child} onChange={(value) => update({ child: value })} onRemove={() => update({ child: undefined })} />{:else}<span class="grammar-row-empty">Drop a block here to capture an expression</span>{/if}</div></div>
  {:else}
    <div class="grammar-expr-body grammar-choice"><div class="grammar-child-label">Alternatives</div>{#each expr.children as alternative, alternativeIndex}<div class="grammar-alternative"><span class="grammar-alternative-label">{String.fromCharCode(65 + alternativeIndex)}</span><div class="grammar-choice-block">{#each alternative as child, childIndex}<svelte:self expr={child} onChange={(value) => { const children = expr.children.map((row, index) => index === alternativeIndex ? row.map((item, index) => index === childIndex ? value : item) : row); update({ children }) }} onRemove={() => removeChild(alternativeIndex, childIndex)} />{:else}<span class="grammar-row-empty">Drop blocks here to create an alternative</span>{/each}</div><button class="grammar-remove grammar-remove-alternative" type="button" aria-label="Remove alternative {String.fromCharCode(65 + alternativeIndex)}" on:click={() => removeAlternative(alternativeIndex)}>x</button></div>{/each}<button class="grammar-add" type="button" on:click={addAlternative}>+ add alternative</button></div>
  {/if}
</PlaygroundBlockShell>
