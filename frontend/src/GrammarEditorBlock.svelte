<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import GrammarEditorRow from './GrammarEditorRow.svelte'
  import GrammarEditorDropTarget from './GrammarEditorDropTarget.svelte'
  import BlockShell from './BlockShell.svelte'
  import type { GrammarDragPayload, GrammarDropTarget, GrammarExpr as Expr, GrammarValidationError } from './GrammarEditorModel'

  export let expr: Expr
  export let onChange: (expr: Expr) => void = () => {}
  export let onRemove: () => void = () => {}
  export let container = ''
  export let dragData: unknown = expr
  export let onDragEnd: () => void = () => {}
  export let path: string[] = []
  export let onDrop: (state: DragDropState<GrammarDragPayload>, target: GrammarDropTarget) => void = () => {}
  export let onRowDragOver: (state: DragDropState<GrammarDragPayload>) => void = () => {}
  export let onRowDragLeave: () => void = () => {}
  export let errors: GrammarValidationError[] = []

  function update(value: Partial<Expr>) { onChange({ ...expr, ...value } as Expr) }
  function summaryFor(value: Expr): string {
    if (value.kind === 'literal') return JSON.stringify(value.text)
    if (value.kind === 'range') return `${value.min}..${value.max} / width ${value.width} / pad ${value.pad}`
    if (value.kind === 'places') return value.alphabets.map((alphabet) => `[${alphabet}]`).join('')
    if (value.kind === 'range_radix') return value.children.map(summaryFor).join(' + ') || 'empty'
    if (value.kind === 'bind') return `${value.name || 'unnamed'} = ${value.child ? summaryFor(value.child) : 'empty'}`
    return value.children.map((alternative) => alternative.length ? alternative.map(summaryFor).join(' + ') : 'empty').join(' | ')
  }

  $: blockTitle = expr.kind === 'range_radix' ? 'range_radix' : expr.kind
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

<BlockShell title={blockTitle} summary={blockSummary} kind={expr.kind} errors={errors.map((error) => error.message)} collapsible initiallyOpen={false} onRemove={onRemove} {container} {dragData} dragHandle=".block-shell-drag" interactive={['input', 'button']} {onDragEnd}>
  {#if expr.kind === 'literal'}
    <div class="grammar-expr-body"><label>Text<input value={expr.text} aria-label="Literal text" on:input={(event) => update({ text: (event.currentTarget as HTMLInputElement).value })} /></label></div>
  {:else if expr.kind === 'range'}
    <div class="grammar-expr-body grammar-number-fields"><label>Min<input type="number" value={expr.min} aria-label="Range minimum" on:input={(event) => update({ min: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Max<input type="number" value={expr.max} aria-label="Range maximum" on:input={(event) => update({ max: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Width<input type="number" min="1" value={expr.width} aria-label="Range width" on:input={(event) => update({ width: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Pad<input maxlength="1" value={expr.pad} aria-label="Range padding" on:input={(event) => update({ pad: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'places'}
    <div class="grammar-expr-body"><label>Allowed characters<input value={expr.alphabets[0] || ''} aria-label="Places allowed characters" on:input={(event) => update({ alphabets: [(event.currentTarget as HTMLInputElement).value] })} /></label></div>
  {:else if expr.kind === 'range_radix'}
      <div class="grammar-expr-body grammar-radix-range"><div class="grammar-child-label">Ordered fields</div><div class="grammar-binding-target"><GrammarEditorRow exprs={expr.children} path={[...path, expr.id]} validationErrors={errors} container={`grammar-editor:range-radix-row:${expr.id}`} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} onDragEnd={onDragEnd} onChange={(index, value) => update({ children: expr.children.map((child, childIndex) => childIndex === index ? value : child) })} onRemove={(index) => update({ children: expr.children.filter((_, childIndex) => childIndex !== index) })} /></div></div>
  {:else if expr.kind === 'bind'}
      <div class="grammar-expr-body"><label>Name<input value={expr.name} aria-label="Binding name" on:input={(event) => update({ name: (event.currentTarget as HTMLInputElement).value })} /></label><div class="grammar-child-label">Captures</div><div class="grammar-binding-target"><GrammarEditorDropTarget container={`grammar-editor:binding-slot:${expr.id}`} target={{ type: 'binding-child', path: [...path, expr.id] }} label="Replace captured expression" showLabel getMode={(state) => state.draggedItem?.kind === 'palette' ? 'new' : 'move'} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} /><GrammarEditorRow exprs={expr.child ? [expr.child] : []} path={[...path, expr.id]} validationErrors={errors} container={`grammar-editor:binding-row:${expr.id}`} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} onDragEnd={onDragEnd} onChange={(index, value) => { if (index === 0) update({ child: value }) }} onRemove={() => update({ child: undefined })} showGaps={false} /></div></div>
  {:else}
      <div class="grammar-expr-body grammar-choice"><div class="grammar-child-label">Alternatives</div>{#each expr.children as alternative, alternativeIndex}<div class="grammar-alternative"><span class="grammar-alternative-label">{String.fromCharCode(65 + alternativeIndex)}</span><div class="grammar-choice-block"><GrammarEditorRow exprs={alternative} path={[...path, expr.id, String(alternativeIndex)]} validationErrors={errors} container={`grammar-editor:choice-row:${expr.id}:${alternativeIndex}`} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} onDragEnd={onDragEnd} onChange={(index, value) => { const children = expr.children.map((row, rowIndex) => rowIndex === alternativeIndex ? row.map((item, itemIndex) => itemIndex === index ? value : item) : row); update({ children }) }} onRemove={(index) => removeChild(alternativeIndex, index)} /></div><button class="grammar-remove grammar-remove-alternative" type="button" aria-label="Remove alternative {String.fromCharCode(65 + alternativeIndex)}" on:click={() => removeAlternative(alternativeIndex)}>x</button></div>{/each}<GrammarEditorDropTarget container={`grammar-editor:choice-slot:${expr.id}`} target={{ type: 'choice-alternative', path: [...path, expr.id] }} label="Add a choice alternative" showLabel getMode={(state) => state.draggedItem?.kind === 'palette' ? 'new' : 'move'} onDragOver={onRowDragOver} onDragLeave={onRowDragLeave} onDrop={onDrop} /><button class="grammar-add" type="button" on:click={addAlternative}>+ add alternative</button></div>
  {/if}
</BlockShell>
