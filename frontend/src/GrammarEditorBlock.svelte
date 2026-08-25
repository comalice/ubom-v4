<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import GrammarEditorRow from './GrammarEditorRow.svelte'
  import PlaygroundGrammarDropTarget from './PlaygroundGrammarDropTarget.svelte'
  import PlaygroundBlockShell from './PlaygroundBlockShell.svelte'
  import type { PlaygroundDragPayload, PlaygroundDropTarget, PlaygroundExpr as Expr } from './PlaygroundGrammar'

  export let expr: Expr
  export let onChange: (expr: Expr) => void = () => {}
  export let onRemove: () => void = () => {}
  export let container = ''
  export let dragData: unknown = expr
  export let onDragEnd: () => void = () => {}
  export let path: string[] = []
  export let onDrop: (state: DragDropState<PlaygroundDragPayload>, target: PlaygroundDropTarget) => void = () => {}
  export let onRowDragOver: (state: DragDropState<PlaygroundDragPayload>) => void = () => {}
  export let onRowDragLeave: () => void = () => {}

  function update(value: Partial<Expr>) { onChange({ ...expr, ...value } as Expr) }
  function summaryFor(value: Expr): string {
    if (value.kind === 'literal') return JSON.stringify(value.text)
    if (value.kind === 'number') return `${value.min}..${value.max} / width ${value.width} / pad ${value.pad}`
    if (value.kind === 'alpha') return `${value.min}..${value.max}`
    if (value.kind === 'chars') return value.chars
    if (value.kind === 'bind') return `${value.name || 'unnamed'} = ${value.child ? summaryFor(value.child) : 'empty'}`
    return value.children.map((alternative) => alternative.length ? alternative.map(summaryFor).join(' + ') : 'empty').join(' | ')
  }

  $: blockTitle = expr.kind === 'number' ? 'number' : expr.kind
  $: blockSummary = summaryFor(expr)

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

<PlaygroundBlockShell title={blockTitle} summary={blockSummary} kind={expr.kind} collapsible initiallyOpen={false} onRemove={onRemove} {container} {dragData} dragHandle=".grammar-drag" interactive={['input', 'button']} {onDragEnd}>
  {#if expr.kind === 'literal'}
    <div class="grammar-expr-body"><label>Text<input value={expr.text} aria-label="Literal text" on:input={(event) => update({ text: (event.currentTarget as HTMLInputElement).value })} /></label></div>
  {:else if expr.kind === 'number'}
    <div class="grammar-expr-body grammar-number-fields"><label>Min<input type="number" value={expr.min} aria-label="Number minimum" on:input={(event) => update({ min: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Max<input type="number" value={expr.max} aria-label="Number maximum" on:input={(event) => update({ max: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Width<input type="number" value={expr.width} aria-label="Number width" on:input={(event) => update({ width: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Pad<input maxlength="1" value={expr.pad} aria-label="Number padding" on:input={(event) => update({ pad: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'alpha'}
    <div class="grammar-expr-body grammar-two-fields"><label>From<input maxlength="1" value={expr.min} aria-label="Letters minimum" on:input={(event) => update({ min: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label><span>to</span><label>To<input maxlength="1" value={expr.max} aria-label="Letters maximum" on:input={(event) => update({ max: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'chars'}
    <div class="grammar-expr-body"><label>Allowed characters<input value={expr.chars} aria-label="Allowed characters" on:input={(event) => update({ chars: (event.currentTarget as HTMLInputElement).value })} /></label></div>
  {:else if expr.kind === 'bind'}
    <div class="grammar-expr-body"><label>Name<input value={expr.name} aria-label="Binding name" on:input={(event) => update({ name: (event.currentTarget as HTMLInputElement).value })} /></label><div class="grammar-child-label">Captures</div><div class="grammar-binding-target"><PlaygroundGrammarDropTarget container={`grammar-editor:binding-slot:${expr.id}`} target={{ type: 'binding-child', path: [...path, expr.id] }} label="Replace captured expression" showLabel onDrop={onDrop} /><GrammarEditorRow exprs={expr.child ? [expr.child] : []} path={[...path, expr.id]} container={`grammar-editor:binding-row:${expr.id}`} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} onDragEnd={onDragEnd} onChange={(index, value) => { if (index === 0) update({ child: value }) }} onRemove={() => update({ child: undefined })} /></div></div>
  {:else}
    <div class="grammar-expr-body grammar-choice"><div class="grammar-child-label">Alternatives</div>{#each expr.children as alternative, alternativeIndex}<div class="grammar-alternative"><span class="grammar-alternative-label">{String.fromCharCode(65 + alternativeIndex)}</span><div class="grammar-choice-block"><GrammarEditorRow exprs={alternative} path={[...path, expr.id, String(alternativeIndex)]} container={`grammar-editor:choice-row:${expr.id}:${alternativeIndex}`} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} onDragEnd={onDragEnd} onChange={(index, value) => { const children = expr.children.map((row, rowIndex) => rowIndex === alternativeIndex ? row.map((item, itemIndex) => itemIndex === index ? value : item) : row); update({ children }) }} onRemove={(index) => removeChild(alternativeIndex, index)} /></div><button class="grammar-remove grammar-remove-alternative" type="button" aria-label="Remove alternative {String.fromCharCode(65 + alternativeIndex)}" on:click={() => removeAlternative(alternativeIndex)}>x</button></div>{/each}<PlaygroundGrammarDropTarget container={`grammar-editor:choice-slot:${expr.id}`} target={{ type: 'choice-alternative', path: [...path, expr.id] }} label="Add a choice alternative" showLabel onDrop={onDrop} /><button class="grammar-add" type="button" on:click={addAlternative}>+ add alternative</button></div>
  {/if}
</PlaygroundBlockShell>
