import type { SeqDefNode } from './api'

export type GrammarExpr =
  | { id: string; kind: 'literal'; text: string }
  | { id: string; kind: 'range'; min: number; max: number; width: number; pad: string }
  | { id: string; kind: 'places'; alphabets: string[] }
  | { id: string; kind: 'range_radix'; children: GrammarExpr[] }
  | { id: string; kind: 'choice'; children: GrammarExpr[][] }
  | { id: string; kind: 'bind'; name: string; child?: GrammarExpr }

export type GrammarExprKind = GrammarExpr['kind']
export type GrammarDropTarget =
  | { type: 'row'; path: string[]; index: number }
  | { type: 'binding-child'; path: string[] }
  | { type: 'choice-alternative'; path: string[] }
  | { type: 'range-radix-children'; path: string[] }

export type GrammarDragPayload =
  | { kind: 'expression'; expr: GrammarExpr; sourcePath: string[] }
  | { kind: 'palette'; exprKind: GrammarExprKind }

export type GrammarParseResult = { value: string; bindings: Record<string, string> } | { error: string }

export function makeGrammarId() { return `${Date.now()}-${Math.random().toString(36).slice(2)}` }

export function createGrammarExpr(kind: GrammarExprKind): GrammarExpr {
  if (kind === 'literal') return { id: makeGrammarId(), kind, text: '' }
  if (kind === 'range') return { id: makeGrammarId(), kind, min: 0, max: 9, width: 1, pad: '0' }
  if (kind === 'places') return { id: makeGrammarId(), kind, alphabets: ['ABC'] }
  if (kind === 'range_radix') return { id: makeGrammarId(), kind, children: [] }
  if (kind === 'choice') return { id: makeGrammarId(), kind, children: [] }
  return { id: makeGrammarId(), kind, name: '' }
}

function parseGrammarExpr(expr: GrammarExpr, input: string, offset: number, bindings: Record<string, string>): number | string {
  if (expr.kind === 'literal') return input.startsWith(expr.text, offset) ? offset + expr.text.length : `expected ${JSON.stringify(expr.text)}`
  if (expr.kind === 'range') {
    const text = input.slice(offset, offset + expr.width)
    const value = Number(text)
    if (text.length !== expr.width || !/^\d+$/.test(text) || !Number.isInteger(value) || value < expr.min || value > expr.max) return `expected ${expr.min}-${expr.max}`
    return offset + expr.width
  }
  if (expr.kind === 'places') {
    let next = offset
    for (const alphabet of expr.alphabets) {
      if (!alphabet || !input[next] || !alphabet.includes(input[next])) return `expected one of ${JSON.stringify(alphabet)}`
      next++
    }
    return next
  }
  if (expr.kind === 'bind') {
    if (!expr.child) return 'binding needs an expression'
    const start = offset
    const next = parseGrammarExpr(expr.child, input, offset, bindings)
    if (typeof next === 'string') return next
    if (expr.name.trim()) bindings[expr.name.trim()] = input.slice(start, next)
    return next
  }
  if (expr.kind === 'range_radix') {
    let next = offset
    for (const child of expr.children) {
      const parsed = parseGrammarExpr(child, input, next, bindings)
      if (typeof parsed === 'string') return parsed
      next = parsed
    }
    return next
  }
  for (const alternative of expr.children) {
    const trial = { ...bindings }
    let next = offset
    let error = ''
    for (const child of alternative) {
      const parsed = parseGrammarExpr(child, input, next, trial)
      if (typeof parsed === 'string') { error = parsed; break }
      next = parsed
    }
    if (!error) { Object.assign(bindings, trial); return next }
  }
  return 'no alternative matched'
}

export function parseGrammarSequence(exprs: GrammarExpr[], input: string): GrammarParseResult {
  const bindings: Record<string, string> = {}
  let offset = 0
  for (const expr of exprs) {
    const next = parseGrammarExpr(expr, input, offset, bindings)
    if (typeof next === 'string') return { error: next }
    offset = next
  }
  return offset === input.length ? { value: input, bindings } : { error: `unexpected text at position ${offset}` }
}

export function generateGrammarExpr(expr: GrammarExpr): string {
  if (expr.kind === 'literal') return expr.text
  if (expr.kind === 'range') return String(expr.min).padStart(Math.max(1, expr.width), expr.pad || '0')
  if (expr.kind === 'places') return expr.alphabets.map((alphabet) => alphabet[0] || '').join('')
  if (expr.kind === 'range_radix') return generateGrammarSequence(expr.children)
  if (expr.kind === 'bind') return expr.child ? generateGrammarExpr(expr.child) : ''
  return expr.children.length ? generateGrammarSequence(expr.children[0]) : ''
}

export function generateGrammarSequence(exprs: GrammarExpr[]) { return exprs.map(generateGrammarExpr).join('') }

// The editor model is intentionally UI-friendly; this adapter is the only
// boundary where it is converted to the server's recursive SeqDef AST.
export function toSeqDefNode(expr: GrammarExpr): SeqDefNode {
  if (expr.kind === 'literal') return { kind: 'literal', text: expr.text }
  if (expr.kind === 'range') return { kind: 'range', min: expr.min, max: expr.max, width: expr.width, pad: expr.pad }
  if (expr.kind === 'places') return { kind: 'places', alphabets: expr.alphabets }
  if (expr.kind === 'range_radix') return { kind: 'range_radix', children: expr.children.map(toSeqDefNode) }
  if (expr.kind === 'bind') {
    if (!expr.child) throw new Error('binding needs an expression')
    return { kind: 'bind', name: expr.name, child: toSeqDefNode(expr.child) }
  }
  if (expr.kind === 'choice') {
    return {
      kind: 'choice',
      children: expr.children.map((alternative) => alternative.length === 1
        ? toSeqDefNode(alternative[0])
        : { kind: 'concat', children: alternative.map(toSeqDefNode) }),
    }
  }
  throw new Error('unsupported grammar expression')
}

export function toSeqDefRoot(exprs: GrammarExpr[]): SeqDefNode {
  return { kind: 'concat', children: exprs.map(toSeqDefNode) }
}
