<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import GrammarEditorBlock from './GrammarEditorBlock.svelte'
  import DropZone from './DropZone.svelte'
  import type { GrammarDragPayload as DragPayload, GrammarDropTarget as DropTarget, GrammarExpr as Expr, GrammarValidationError } from './GrammarEditorModel'

  export type GrammarDragItem = DragPayload
  export type GrammarDropSlot = Extract<DropTarget, { type: 'row' }>

  export let exprs: Expr[] = []
  export let path: string[] = []
  export let container: string
  export let onDragOver: (state: DragDropState<GrammarDragItem>) => void
  export let onDragLeave: () => void
  export let onDrop: (state: DragDropState<GrammarDragItem>, target: DropTarget) => void
  export let onDragEnd: () => void
  export let onChange: (index: number, expr: Expr) => void
  export let onRemove: (index: number) => void
  export let validationErrors: GrammarValidationError[] = []

  function blockData(expr: Expr): GrammarDragItem { return { kind: 'expression', expr, sourcePath: [...path] } }
</script>

<div class="grammar-editor-row">
  {#each exprs as expr, index (expr.id)}
    <DropZone variant="gap" className={index === 0 ? 'dnd-insert-zone-edge' : ''} {container} getMode={(state) => state.draggedItem?.kind === 'palette' ? 'new' : 'move'} onDragOver={onDragOver} onDragLeave={onDragLeave} onDrop={(state) => onDrop(state, { type: 'row', path, index })} />
    <GrammarEditorBlock {expr} {path} container={container} dragData={blockData(expr)} errors={validationErrors.filter((error) => error.nodeId === expr.id)} onChange={(value) => onChange(index, value)} onRemove={() => onRemove(index)} onDrop={onDrop} onRowDragOver={onDragOver} onRowDragLeave={onDragLeave} {onDragEnd} />
  {/each}
  <DropZone variant="gap" className="dnd-insert-zone-edge" {container} getMode={(state) => state.draggedItem?.kind === 'palette' ? 'new' : 'move'} onDragOver={onDragOver} onDragLeave={onDragLeave} onDrop={(state) => onDrop(state, { type: 'row', path, index: exprs.length })} />
</div>
