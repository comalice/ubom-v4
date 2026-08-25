export type PlaygroundExpr =
  | { id: string; kind: 'literal'; text: string }
  | { id: string; kind: 'number'; min: number; max: number; width: number; pad: string }
  | { id: string; kind: 'alpha'; min: string; max: string }
  | { id: string; kind: 'chars'; chars: string }
  | { id: string; kind: 'choice'; children: PlaygroundExpr[][] }
  | { id: string; kind: 'bind'; name: string; child?: PlaygroundExpr }

export type PlaygroundExprKind = PlaygroundExpr['kind']
export type PlaygroundDropTarget =
  | { type: 'row'; path: string[]; index: number }
  | { type: 'binding-child'; path: string[] }
  | { type: 'choice-alternative'; path: string[] }

export type PlaygroundDragPayload =
  | { kind: 'expression'; expr: PlaygroundExpr; sourcePath: string[] }
  | { kind: 'palette'; exprKind: PlaygroundExprKind }

export type PlaygroundParseResult = { value: string; bindings: Record<string, string> } | { error: string }

export function playgroundMakeId() { return `${Date.now()}-${Math.random().toString(36).slice(2)}` }

export function playgroundCreateExpr(kind: PlaygroundExprKind): PlaygroundExpr {
  if (kind === 'literal') return { id: playgroundMakeId(), kind, text: '' }
  if (kind === 'number') return { id: playgroundMakeId(), kind, min: 0, max: 9, width: 1, pad: '0' }
  if (kind === 'alpha') return { id: playgroundMakeId(), kind, min: 'A', max: 'Z' }
  if (kind === 'chars') return { id: playgroundMakeId(), kind, chars: 'ABC' }
  if (kind === 'choice') return { id: playgroundMakeId(), kind, children: [] }
  return { id: playgroundMakeId(), kind, name: '' }
}

function playgroundParseExpr(expr: PlaygroundExpr, input: string, offset: number, bindings: Record<string, string>): number | string {
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
    const next = playgroundParseExpr(expr.child, input, offset, bindings)
    if (typeof next === 'string') return next
    if (expr.name.trim()) bindings[expr.name.trim()] = input.slice(start, next)
    return next
  }
  for (const alternative of expr.children) {
    const trial = { ...bindings }
    let next = offset
    let error = ''
    for (const child of alternative) {
      const parsed = playgroundParseExpr(child, input, next, trial)
      if (typeof parsed === 'string') { error = parsed; break }
      next = parsed
    }
    if (!error) { Object.assign(bindings, trial); return next }
  }
  return 'no alternative matched'
}

export function playgroundParseSequence(exprs: PlaygroundExpr[], input: string): PlaygroundParseResult {
  const bindings: Record<string, string> = {}
  let offset = 0
  for (const expr of exprs) {
    const next = playgroundParseExpr(expr, input, offset, bindings)
    if (typeof next === 'string') return { error: next }
    offset = next
  }
  return offset === input.length ? { value: input, bindings } : { error: `unexpected text at position ${offset}` }
}

export function playgroundGenerateExpr(expr: PlaygroundExpr): string {
  if (expr.kind === 'literal') return expr.text
  if (expr.kind === 'number') return String(expr.min).padStart(Math.max(1, expr.width), expr.pad || '0')
  if (expr.kind === 'alpha') return expr.min
  if (expr.kind === 'chars') return expr.chars[0] || ''
  if (expr.kind === 'bind') return expr.child ? playgroundGenerateExpr(expr.child) : ''
  return expr.children.length ? playgroundGenerateSequence(expr.children[0]) : ''
}

export function playgroundGenerateSequence(exprs: PlaygroundExpr[]) { return exprs.map(playgroundGenerateExpr).join('') }
