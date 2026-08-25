<script lang="ts">
  import PlaygroundBlockShell from './PlaygroundBlockShell.svelte'
  import type { Expr } from './grammarMock'

  export let expr: Expr
  export let onRemove: () => void = () => {}

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
</script>

<PlaygroundBlockShell title={title()} summary={summary()} kind={expr.kind} collapsible initiallyOpen={false} onRemove={onRemove}>
  {#if expr.kind === 'literal'}
    <div class="grammar-expr-body"><label>Text<input value={expr.text} aria-label="Literal text" /></label></div>
  {:else if expr.kind === 'number'}
    <div class="grammar-expr-body grammar-number-fields"><label>Min<input value={expr.min} aria-label="Number minimum" /></label><label>Max<input value={expr.max} aria-label="Number maximum" /></label><label>Width<input value={expr.width} aria-label="Number width" /></label><label>Pad<input value={expr.pad} aria-label="Number padding" /></label></div>
  {:else if expr.kind === 'alpha'}
    <div class="grammar-expr-body grammar-two-fields"><label>From<input value={expr.min} aria-label="Letters minimum" /></label><span>to</span><label>To<input value={expr.max} aria-label="Letters maximum" /></label></div>
  {:else if expr.kind === 'chars'}
    <div class="grammar-expr-body"><label>Allowed characters<input value={expr.chars} aria-label="Allowed characters" /></label></div>
  {:else if expr.kind === 'bind'}
    <div class="grammar-expr-body"><label>Name<input value={expr.name} aria-label="Binding name" /></label><div class="grammar-child-label">Captures</div><div class="grammar-binding-target">{#if expr.child}<svelte:self expr={expr.child} />{:else}<span class="grammar-row-empty">Drop a block here to capture an expression</span>{/if}</div></div>
  {:else}
    <div class="grammar-expr-body grammar-choice"><div class="grammar-child-label">Alternatives</div>{#each expr.children as alternative, alternativeIndex}<div class="grammar-alternative"><span class="grammar-alternative-label">{String.fromCharCode(65 + alternativeIndex)}</span><div class="grammar-choice-block">{#each alternative as child}<svelte:self expr={child} />{:else}<span class="grammar-row-empty">Drop blocks here to create an alternative</span>{/each}</div></div>{/each}<button class="grammar-add" type="button">+ add alternative</button></div>
  {/if}
</PlaygroundBlockShell>
