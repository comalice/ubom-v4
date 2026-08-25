<script lang="ts">
  import GrammarMockRow from './GrammarMockRow.svelte'
  import GrammarDropTarget from './GrammarDropTarget.svelte'
  import { draggable, type DragDropState } from '@thisux/sveltednd'
  import { type DropTarget, type DragPayload, type Expr, type ExprKind } from './grammarMock'
  export let expr: Expr
  export let path: string[] = []
  export let depth = 0
  export let onChange: (expr: Expr) => void
  export let onDelete: () => void
  export let onDrop: (expr: Expr, sourcePath: string[], target: DropTarget) => void
  export let onPalette: (kind: ExprKind, target: DropTarget) => void
  function update(value: Partial<Expr>) { onChange({ ...expr, ...value } as Expr) }
  function addChoice() { if (expr.kind === 'choice') update({ children: [...expr.children, []] }) }
  function updateAlternative(index: number, children: Expr[]) { if (expr.kind === 'choice') { const alternatives = [...expr.children]; alternatives[index] = children; update({ children: alternatives }) } }
  function removeAlternative(index: number) { if (expr.kind === 'choice') update({ children: expr.children.filter((_, alternativeIndex) => alternativeIndex !== index) }) }
  function internalDrop(state: DragDropState<DragPayload>, target: DropTarget) { if (state.draggedItem.kind === 'palette') onPalette(state.draggedItem.exprKind, target); else onDrop(state.draggedItem.expr, state.draggedItem.sourcePath, target) }
</script>

<article class="grammar-expr grammar-kind-{expr.kind}" use:draggable={{ container: `expr:${expr.id}`, dragData: { kind: 'expression', expr, sourcePath: path }, handle: '.grammar-drag-handle', interactive: ['input', 'button'] }}>
  <header class="grammar-expr-header" role="group"><span class="grammar-drag grammar-drag-handle" aria-label="Drag expression" role="button" tabindex="0">::</span><span class="grammar-expr-kind">{expr.kind === 'number' ? 'number' : expr.kind}</span><button class="grammar-remove" type="button" aria-label="Remove {expr.kind}" on:click={onDelete}>x</button></header>
  {#if expr.kind === 'literal'}
    <div class="grammar-expr-body"><label>Text<input value={expr.text} on:input={(event) => update({ text: (event.currentTarget as HTMLInputElement).value })} placeholder="PN-" /></label></div>
  {:else if expr.kind === 'number'}
    <div class="grammar-expr-body grammar-number-fields"><label>Min<input type="number" value={expr.min} on:input={(event) => update({ min: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Max<input type="number" value={expr.max} on:input={(event) => update({ max: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Width<input type="number" min="1" value={expr.width} on:input={(event) => update({ width: Number((event.currentTarget as HTMLInputElement).value) })} /></label><label>Pad<input maxlength="1" value={expr.pad} on:input={(event) => update({ pad: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'alpha'}
    <div class="grammar-expr-body grammar-two-fields"><label>From<input maxlength="1" value={expr.min} on:input={(event) => update({ min: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label><span>to</span><label>To<input maxlength="1" value={expr.max} on:input={(event) => update({ max: (event.currentTarget as HTMLInputElement).value.slice(0, 1) })} /></label></div>
  {:else if expr.kind === 'chars'}
    <div class="grammar-expr-body"><label>Allowed characters<input value={expr.chars} on:input={(event) => update({ chars: (event.currentTarget as HTMLInputElement).value })} placeholder="ABC012" /></label></div>
  {:else if expr.kind === 'bind'}
     <div class="grammar-expr-body"><label>Name<input value={expr.name} on:input={(event) => update({ name: (event.currentTarget as HTMLInputElement).value })} placeholder="family" /></label><div class="grammar-child-label">Captures</div><div class="grammar-binding-target"><GrammarDropTarget container={`binding:${expr.id}`} target={{ type: 'binding-child', path: [...path, expr.id] }} label="Replace captured expression" showLabel onDrop={internalDrop} /><GrammarMockRow exprs={expr.child ? [expr.child] : []} path={[...path, expr.id]} depth={depth + 1} emptyLabel={expr.child ? 'Drop here to replace the captured expression' : 'Drop here to capture an expression'} onChange={(children) => update({ child: children[0] })} {onDrop} {onPalette} /></div></div>
  {:else}
     <div class="grammar-expr-body grammar-choice"><div class="grammar-child-label">Alternatives</div>{#each expr.children as alternative, index (index)}<div class="grammar-alternative"><span class="grammar-alternative-label">{String.fromCharCode(65 + index)}</span><GrammarMockRow exprs={alternative} path={[...path, expr.id, String(index)]} depth={depth + 1} onChange={(children) => updateAlternative(index, children)} {onDrop} {onPalette} /><button class="grammar-remove grammar-remove-alternative" type="button" aria-label="Remove alternative {String.fromCharCode(65 + index)}" on:click={() => removeAlternative(index)}>x</button></div>{/each}<GrammarDropTarget container={`choice:${expr.id}`} target={{ type: 'choice-alternative', path: [...path, expr.id] }} label="Drop here to add an alternative" showLabel onDrop={internalDrop} /><button class="grammar-add" type="button" on:click={addChoice}>+ add alternative</button></div>
  {/if}
</article>
