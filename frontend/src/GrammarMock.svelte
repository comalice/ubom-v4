<script lang="ts">
  import { draggable } from '@thisux/sveltednd'
  import GrammarMockRow from './GrammarMockRow.svelte'
  import { createExpr, expressionTypes, generateSequence, parseSequence, type DropTarget, type Expr, type ExprKind } from './grammarMock'

  export let navigate: (path: string) => void
  let exprs: Expr[] = []
  let input = ''
  let selectedKind: ExprKind = 'literal'
  $: generated = generateSequence(exprs)
  $: parsed = parseSequence(exprs, input)

  function add() { exprs = [...exprs, createExpr(selectedKind)] }
  function paletteDrop(kind: ExprKind, target: DropTarget) {
    const expr = createExpr(kind)
    if (target.type === 'binding-child') {
      exprs = replaceExpr(exprs, target.path, found => found.kind === 'bind' ? { ...found, child: expr } : found)
      return
    }
    if (target.type === 'choice-alternative') {
      exprs = replaceExpr(exprs, target.path, found => found.kind === 'choice' ? { ...found, children: [...found.children, [expr]] } : found)
      return
    }
    exprs = insertPath(exprs, target.path, expr, target.index)
  }
  function drop(expr: Expr, sourcePath: string[], target: DropTarget) {
    if (containsTarget(expr, target)) return
    const sourceRow = getPath(exprs, sourcePath)
    const sourceIndex = sourceRow.findIndex(item => item.id === expr.id)
    if (sourceIndex < 0) return
    const without = removePath(exprs, sourcePath, expr.id)
    if (target.type === 'binding-child') { exprs = replaceExpr(without, target.path, found => found.kind === 'bind' ? { ...found, child: expr } : found); return }
    if (target.type === 'choice-alternative') { exprs = replaceExpr(without, target.path, found => found.kind === 'choice' ? { ...found, children: [...found.children, [expr]] } : found); return }
    let index = target.index
    if (samePath(sourcePath, target.path) && sourceIndex < index) index -= 1
    exprs = insertPath(without, target.path, expr, index)
  }
  function containsTarget(source: Expr, target: DropTarget) {
    const path = target.path
    if (target.type === 'row' && path.length === 0) return false
    return path.includes(source.id) || (target.type !== 'row' && path[path.length - 1] === source.id)
  }
  function samePath(a: string[], b: string[]) { return a.length === b.length && a.every((id, index) => id === b[index]) }
  function removePath(root: Expr[], path: string[], id: string): Expr[] {
    if (!path.length) return root.filter(found => found.id !== id)
    const [container, ...rest] = path
    return root.map(expr => {
      if (expr.id !== container) return expr
      if (expr.kind === 'bind') return { ...expr, child: expr.child && removePath([expr.child], rest, id)[0] }
      if (expr.kind === 'choice') {
        const alternative = Number(rest[0]); const children = [...expr.children]
        children[alternative] = removePath(children[alternative] || [], rest.slice(1), id)
        return { ...expr, children }
      }
      return expr
    })
  }
  function replaceExpr(root: Expr[], path: string[], change: (expr: Expr) => Expr): Expr[] {
    if (!path.length) return root.map(change)
    const id = path[0]
    return root.map(expr => {
      if (expr.id !== id) return expr
      if (expr.kind === 'bind' && path.length === 1) return change(expr)
      if (expr.kind === 'bind') return { ...expr, child: expr.child ? replaceExpr([expr.child], path.slice(1), change)[0] : expr.child }
      if (expr.kind === 'choice' && path.length === 1) return change(expr)
      if (expr.kind === 'choice') {
        const children = [...expr.children]; const alternative = Number(path[1])
        children[alternative] = path.length === 2 ? children[alternative].map(change) : replaceExpr(children[alternative] || [], path.slice(2), change)
        return { ...expr, children }
      }
      return expr
    })
  }
  function insertPath(root: Expr[], path: string[], expr: Expr, index: number): Expr[] {
    if (!path.length) {
      const insertAt = Math.min(index, root.length)
      return [...root.slice(0, insertAt), expr, ...root.slice(insertAt)]
    }
    const id = path[0]
    return root.map(found => {
      if (found.id !== id) return found
      if (found.kind === 'bind') {
        const child = insertPath(found.child ? [found.child] : [], path.slice(1), expr, index)
        return { ...found, child: child[0] }
      }
      if (found.kind === 'choice') {
        const children = [...found.children]
        const alternative = Number(path[1])
        if (path.length === 2) children[alternative] = insertPath(children[alternative] || [], [], expr, index)
        else children[alternative] = insertPath(children[alternative] || [], path.slice(2), expr, index)
        return { ...found, children }
      }
      return found
    })
  }
  function getPath(root: Expr[], path: string[]): Expr[] {
    if (!path.length) return root
    const id = path[0]
    for (const found of root) {
      if (found.id === id) {
        if (found.kind === 'bind') return path.length === 1 ? (found.child ? [found.child] : []) : getPath(found.child ? [found.child] : [], path.slice(1))
        if (found.kind === 'choice') {
          const alternative = found.children[Number(path[1])] || []
          return path.length === 2 ? alternative : getPath(alternative, path.slice(2))
        }
      }
      if (found.kind === 'bind' && found.child) { const nested = getPath([found.child], path); if (nested.length) return nested }
      if (found.kind === 'choice') for (const alternative of found.children) { const nested = getPath(alternative, path); if (nested.length) return nested }
    }
    return []
  }
</script>

  <div class="grammar-mock">
  <header class="grammar-mock-header"><div><span class="mock-section-label">Grammar workbench</span><h1>Define a sequence language</h1></div><span class="mock-local-badge">UI MOCK · NOT CONNECTED</span><span class="mock-spacer"></span><button class="mock-close" type="button" on:click={() => navigate('/settings')}>Close</button></header>
  <div class="grammar-mock-intro">Click a palette item to append it, or drag it to an insertion point. Use the grip to move expressions. Drop on a Binding child to replace it, or on Choice alternatives to create one.</div>
  <main class="grammar-mock-main">
     <section class="grammar-editor"><div class="grammar-palette"><span class="mock-section-label">Expression palette</span>{#each expressionTypes as item}<div class="grammar-palette-item" role="button" tabindex="0" use:draggable={{ container: `palette:${item.kind}`, dragData: { kind: 'palette', exprKind: item.kind } }} on:click={() => { selectedKind = item.kind; add() }} on:keydown={(event) => (event.key === 'Enter' || event.key === ' ') && (event.preventDefault(), selectedKind = item.kind, add())}><span aria-hidden="true">::</span><span><strong>{item.label}</strong><small>{item.hint}</small></span></div>{/each}</div><div class="grammar-section-heading"><span class="mock-section-label">Grammar sequence</span><span>{exprs.length} expression{exprs.length === 1 ? '' : 's'}</span></div><GrammarMockRow exprs={exprs} onChange={(value) => exprs = value} onDrop={(expr, source, target) => drop(expr, source, target)} onPalette={paletteDrop} /></section>
    <section class="grammar-preview"><div class="grammar-section-heading"><div><span class="mock-section-label">Operations</span><h2>Parse and generate</h2></div></div><label class="grammar-input-label">Input to parse<input bind:value={input} placeholder="Type a candidate value" /></label>{#if input}<div class:grammar-valid={!('error' in parsed)} class="grammar-parse-result">{#if 'error' in parsed}Invalid: {parsed.error}{:else}Valid input{#each Object.entries(parsed.bindings) as [name, value]}<span>{name} = {value}</span>{/each}{/if}</div>{/if}<div class="grammar-output"><span>Canonical generated value</span><strong>{generated || '(empty)'}</strong></div><div class="grammar-model-note"><strong>Semantic model</strong><span>Rows concatenate. Choice selects an alternative. Binding captures its child.</span></div></section>
  </main>
</div>
