export type Expr =
  | { id: string; kind: 'literal'; text: string }
  | { id: string; kind: 'number'; min: number; max: number; width: number; pad: string }
  | { id: string; kind: 'alpha'; min: string; max: string }
  | { id: string; kind: 'chars'; chars: string }
  | { id: string; kind: 'choice'; children: Expr[][] }
  | { id: string; kind: 'bind'; name: string; child?: Expr }

export type ExprKind = Expr['kind']

export type DropTarget =
  | { type: 'row'; path: string[]; index: number }
  | { type: 'binding-child'; path: string[] }
  | { type: 'choice-alternative'; path: string[] }

export type DragPayload =
  | { kind: 'expression'; expr: Expr; sourcePath: string[] }
  | { kind: 'palette'; exprKind: ExprKind }

export const expressionTypes: { kind: ExprKind; label: string; hint: string }[] = [
  { kind: 'literal', label: 'Literal', hint: 'Exact text' },
  { kind: 'number', label: 'Number', hint: 'Numeric range' },
  { kind: 'alpha', label: 'Letters', hint: 'Alphabetic range' },
  { kind: 'chars', label: 'Characters', hint: 'Allowed characters' },
  { kind: 'choice', label: 'Choice', hint: 'One alternative' },
  { kind: 'bind', label: 'Binding', hint: 'Capture a value' },
]

export function makeId() { return `${Date.now()}-${Math.random().toString(36).slice(2)}` }

export function createExpr(kind: ExprKind): Expr {
  if (kind === 'literal') return { id: makeId(), kind, text: '' }
  if (kind === 'number') return { id: makeId(), kind, min: 0, max: 9, width: 1, pad: '0' }
  if (kind === 'alpha') return { id: makeId(), kind, min: 'A', max: 'Z' }
  if (kind === 'chars') return { id: makeId(), kind, chars: 'ABC' }
  if (kind === 'choice') return { id: makeId(), kind, children: [] }
  return { id: makeId(), kind, name: '' }
}

export type ParseResult = { value: string; bindings: Record<string, string> } | { error: string }

function parseExpr(expr: Expr, input: string, offset: number, bindings: Record<string, string>): number | string {
  if (expr.kind === 'literal') return input.startsWith(expr.text, offset) ? offset + expr.text.length : `expected ${JSON.stringify(expr.text)}`
  if (expr.kind === 'number') {
    const text = input.slice(offset, offset + expr.width)
    const value = Number(text)
    if (text.length !== expr.width || !/^\d+$/.test(text) || !Number.isInteger(value) || value < expr.min || value > expr.max) return `expected ${expr.min}-${expr.max}`
    return offset + expr.width
  }
  if (expr.kind === 'alpha') {
    const value = input[offset]
    if (!value || value < expr.min || value > expr.max) return `expected ${expr.min}-${expr.max}`
    return offset + 1
  }
  if (expr.kind === 'chars') return expr.chars.includes(input[offset] || '') ? offset + 1 : `expected one of ${JSON.stringify(expr.chars)}`
  if (expr.kind === 'bind') {
    if (!expr.child) return 'binding needs an expression'
    const start = offset
    const next = parseExpr(expr.child, input, offset, bindings)
    if (typeof next === 'string') return next
    if (expr.name.trim()) bindings[expr.name.trim()] = input.slice(start, next)
    return next
  }
  for (const alternative of expr.children) {
    const trial = { ...bindings }
    let next = offset
    let error = ''
    for (const child of alternative) {
      const parsed = parseExpr(child, input, next, trial)
      if (typeof parsed === 'string') { error = parsed; break }
      next = parsed
    }
    if (!error) { Object.assign(bindings, trial); return next }
  }
  return 'no alternative matched'
}

export function parseSequence(exprs: Expr[], input: string): ParseResult {
  const bindings: Record<string, string> = {}
  let offset = 0
  for (const expr of exprs) {
    const next = parseExpr(expr, input, offset, bindings)
    if (typeof next === 'string') return { error: next }
    offset = next
  }
  return offset === input.length ? { value: input, bindings } : { error: `unexpected text at position ${offset}` }
}

export function generateExpr(expr: Expr): string {
  if (expr.kind === 'literal') return expr.text
  if (expr.kind === 'number') return String(expr.min).padStart(Math.max(1, expr.width), expr.pad || '0')
  if (expr.kind === 'alpha') return expr.min
  if (expr.kind === 'chars') return expr.chars[0] || ''
  if (expr.kind === 'bind') return expr.child ? generateExpr(expr.child) : ''
  return expr.children.length ? generateSequence(expr.children[0]) : ''
}

export function generateSequence(exprs: Expr[]) { return exprs.map(generateExpr).join('') }
