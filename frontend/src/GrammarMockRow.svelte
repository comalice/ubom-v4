<script lang="ts">
  import GrammarMockExpr from './GrammarMockExpr.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import { type DropTarget, type DragPayload, type Expr, type ExprKind } from './grammarMock'

  export let exprs: Expr[] = []
  export let path: string[] = []
  export let depth = 0
  export let emptyLabel = 'Drop expressions here'
  export let onChange: (exprs: Expr[]) => void
  export let onDrop: (expr: Expr, sourcePath: string[], target: DropTarget) => void
  export let onPalette: (kind: ExprKind, target: DropTarget) => void
  const rowContainer = `row:${path.join('.') || 'root'}`
  function targetFor(state: DragDropState<DragPayload>): DropTarget | null {
    const id = state.targetContainer
    if (!id) return null
    if (id === rowContainer) return { type: 'row', path, index: exprs.length }
    if (id.startsWith('expr:')) {
      const index = exprs.findIndex(expr => `expr:${expr.id}` === id)
      if (index < 0) return null
      return { type: 'row', path, index: state.dropPosition === 'after' ? index + 1 : index }
    }
    return null
  }
  function handleDrop(state: DragDropState<DragPayload>, target: DropTarget) {
    if (state.draggedItem.kind === 'palette') onPalette(state.draggedItem.exprKind, target)
    else onDrop(state.draggedItem.expr, state.draggedItem.sourcePath, target)
  }
  function handleExpressionDrop(state: DragDropState<DragPayload>) {
    const target = targetFor(state)
    if (target) handleDrop(state, target)
  }
  function update(index: number, expr: Expr) { const next = [...exprs]; next[index] = expr; onChange(next) }
  function remove(index: number) { onChange(exprs.filter((_, itemIndex) => itemIndex !== index)) }
</script>

  <div class="grammar-row" class:grammar-row-nested={depth > 0} role="list">
  {#if !exprs.length}<span class="grammar-row-empty" use:droppable={{ container: rowContainer, callbacks: { onDrop: handleDrop.bind(null) }, attributes: { dragOverClass: 'grammar-dnd-over' } }}>{emptyLabel}</span>{/if}
  {#each exprs as expr, index (expr.id)}
    <div class="grammar-expr-wrap">
      <div class="grammar-expr-drop-target" use:droppable={{ container: `expr:${expr.id}`, direction: 'horizontal', callbacks: { onDrop: handleExpressionDrop }, attributes: { dragOverClass: 'grammar-dnd-over' } }}><GrammarMockExpr {expr} {path} {depth} onChange={(value) => update(index, value)} onDelete={() => remove(index)} onDrop={(item, source, target) => onDrop(item, source, target)} onPalette={(kind, target) => onPalette(kind, target)} /></div>
    </div>
  {/each}
  {#if exprs.length}<div class="grammar-row-end" use:droppable={{ container: rowContainer, direction: 'horizontal', callbacks: { onDrop: handleDrop }, attributes: { dragOverClass: 'grammar-dnd-over' } }} aria-label="Insert at end"></div>{/if}
  </div>
