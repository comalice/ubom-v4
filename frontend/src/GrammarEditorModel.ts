import type { SeqDefNode } from './api'

export type GrammarExpr =
  | { id: string; kind: 'literal'; text: string }
  | { id: string; kind: 'range'; min: number; max: number; width: number; pad: string }
  | { id: string; kind: 'places'; alphabets: string[] }
  | { id: string; kind: 'range_radix'; children: GrammarExpr[] }
  | { id: string; kind: 'choice'; children: GrammarExpr[][] }
  | { id: string; kind: 'bind'; name: string; child?: GrammarExpr }

export type GrammarExprKind = GrammarExpr['kind']
export type GrammarValidationError = {
  code: string
  message: string
  path: string
  nodeId?: string
}

const isObject = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null

export function validateGrammarSequence(exprs: unknown): GrammarValidationError[] {
  const errors: GrammarValidationError[] = []

  function add(code: string, message: string, path: string, nodeId?: string) {
    errors.push({ code, message, path, ...(nodeId ? { nodeId } : {}) })
  }

  // Return the binding names present on each possible successful path. Choice
  // alternatives start from the same names but do not share captures.
  function visit(value: unknown, path: string, nodeId?: string, paths: Set<string>[] = [new Set()]): Set<string>[] {
    if (!isObject(value) || typeof value.kind !== 'string') {
      add('unsupported-node', 'Unsupported or incomplete grammar node', path, nodeId)
      return paths
    }
    const kind = value.kind
    const id = typeof value.id === 'string' ? value.id : nodeId
    const nodePath = `${path}.${kind}`
    if (kind === 'literal') {
      if (typeof value.text !== 'string') add('incomplete-node', 'Literal text is required', nodePath, id)
    } else if (kind === 'range') {
      const min = value.min
      const max = value.max
      const width = value.width
      if (typeof min !== 'number' || !Number.isInteger(min) || min < 0 || typeof max !== 'number' || !Number.isInteger(max) || max < min) {
        add('invalid-range-bounds', 'Range bounds must be non-negative integers with min no greater than max', nodePath, id)
      }
      if (typeof width !== 'number' || !Number.isInteger(width) || width < 1) {
        add('invalid-width', 'Range width must be a positive integer', nodePath, id)
      } else if (typeof max === 'number' && Number.isInteger(max) && String(max).length > width) {
        add('invalid-width', 'Range width is too small for the maximum value', nodePath, id)
      }
      if (typeof value.pad !== 'string' || [...value.pad].length > 1 || (value.pad.length === 1 && value.pad >= '1' && value.pad <= '9')) {
        add('invalid-range-definition', 'Range padding must be one character and cannot be a digit from 1 to 9', nodePath, id)
      }
    } else if (kind === 'places') {
      if (!Array.isArray(value.alphabets) || value.alphabets.length === 0 || value.alphabets.some((alphabet) => typeof alphabet !== 'string' || alphabet.length === 0)) {
        add('invalid-character-definition', 'Places must contain one or more non-empty character alphabets', nodePath, id)
      }
    } else if (kind === 'bind') {
      if (typeof value.name !== 'string' || !value.name.trim()) add('empty-binding-name', 'Binding name cannot be empty', nodePath, id)
      else if (/\s/.test(value.name)) add('invalid-binding-name', 'Binding name cannot contain whitespace', nodePath, id)
      else {
        paths = paths.map((names) => {
          const next = new Set(names)
          if (names.has(value.name as string)) add('duplicate-binding-name', `Binding name "${value.name}" is already used`, nodePath, id)
          else next.add(value.name as string)
          return next
        })
      }
      if (!('child' in value) || value.child === undefined || value.child === null) add('missing-bind-child', 'Binding needs an expression', nodePath, id)
      else paths = visit(value.child, nodePath, id, paths)
    } else if (kind === 'range_radix') {
      if (!Array.isArray(value.children) || value.children.length === 0) add('invalid-radix-definition', 'Radix range must contain one or more fields', nodePath, id)
      else value.children.forEach((child, index) => { paths = visit(child, `${nodePath}[${index}]`, undefined, paths) })
    } else if (kind === 'choice') {
      if (!Array.isArray(value.children) || value.children.length === 0) add('empty-choice', 'Choice must contain one or more alternatives', nodePath, id)
      else {
        const incomingPaths = paths
        let choicePaths: Set<string>[] = []
        value.children.forEach((alternative, index) => {
          const alternativePath = `${nodePath}[${index}]`
          if (!Array.isArray(alternative) || alternative.length === 0) add('empty-choice-alternative', 'Choice alternatives cannot be empty', alternativePath, id)
          else {
            let alternativePaths = incomingPaths.map((names) => new Set(names))
            alternative.forEach((child, childIndex) => { alternativePaths = visit(child, `${alternativePath}[${childIndex}]`, undefined, alternativePaths) })
            choicePaths.push(...alternativePaths)
          }
        })
        paths = choicePaths.length ? choicePaths : incomingPaths
      }
    } else {
      add('unsupported-node', `Unsupported grammar node kind "${kind}"`, nodePath, id)
    }
    return paths
  }

  if (!Array.isArray(exprs)) add('incomplete-node', 'Grammar root must be a sequence', 'root')
  else {
    let paths: Set<string>[] = [new Set()]
    exprs.forEach((expr, index) => { paths = visit(expr, `root[${index}]`, undefined, paths) })
  }
  return errors
}

export type GrammarDropTarget =
  | { type: 'row'; path: string[]; index: number }
  | { type: 'binding-child'; path: string[] }
  | { type: 'choice-alternative'; path: string[] }

export type GrammarDragPayload =
  | { kind: 'expression'; expr: GrammarExpr; sourcePath: string[] }
  | { kind: 'palette'; exprKind: GrammarExprKind }

export type GrammarParseResult = { value: string; bindings: Record<string, string> } | { error: string }

export type GrammarPreviewStatus =
  | { kind: 'not-previewed'; label: 'Not previewed'; message: string }
  | { kind: 'previewing'; label: 'Preview in progress'; message: string }
  | { kind: 'success'; label: 'Preview successful'; message: string }
  | { kind: 'server-rejected'; label: 'Server rejected preview'; message: string }
  | { kind: 'request-failed'; label: 'Preview request failed'; message: string }

export type GrammarPreviewStatusInput = {
  currentKey: string
  previewedKey: string | null
  preview: { error?: string } | null
  requestKey: string | null
  requestError: string
  previewing: boolean
}

export function getGrammarPreviewStatus(input: GrammarPreviewStatusInput): GrammarPreviewStatus {
  if (input.requestKey === input.currentKey && input.previewing) {
    return { kind: 'previewing', label: 'Preview in progress', message: 'Waiting for the server preview.' }
  }
  if (input.requestKey === input.currentKey && input.requestError) {
    return { kind: 'request-failed', label: 'Preview request failed', message: input.requestError }
  }
  if (input.previewedKey !== input.currentKey || !input.preview) {
    return { kind: 'not-previewed', label: 'Not previewed', message: 'Preview this sequence on the server to see the canonical value.' }
  }
  if (input.preview.error) {
    return { kind: 'server-rejected', label: 'Server rejected preview', message: input.preview.error }
  }
  return { kind: 'success', label: 'Preview successful', message: 'The server returned a canonical value.' }
}

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

// Include the editor document and its preview bindings so a response cannot be
// presented as current after either one changes.
export function grammarPreviewKey(exprs: GrammarExpr[], bindings: Record<string, string>): string {
  return JSON.stringify({
    exprs,
    bindings: Object.fromEntries(Object.entries(bindings).sort(([left], [right]) => left.localeCompare(right))),
  })
}

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

export function fromSeqDefRoot(root: SeqDefNode): GrammarExpr[] {
  const fromNode = (node: SeqDefNode): GrammarExpr => {
    const id = makeGrammarId()
    if (node.kind === 'literal') return { id, kind: node.kind, text: node.text ?? '' }
    if (node.kind === 'range') return { id, kind: node.kind, min: node.min ?? 0, max: node.max ?? 0, width: node.width ?? 1, pad: node.pad ?? '0' }
    if (node.kind === 'places') return { id, kind: node.kind, alphabets: node.alphabets ?? [] }
    if (node.kind === 'bind') return { id, kind: node.kind, name: node.name ?? '', child: node.child ? fromNode(node.child) : undefined }
    if (node.kind === 'choice') return { id, kind: node.kind, children: (node.children ?? []).map(child => child.kind === 'concat' ? (child.children ?? []).map(fromNode) : [fromNode(child)]) }
    if (node.kind === 'concat') return { id, kind: 'range_radix', children: (node.children ?? []).map(fromNode) }
    return { id, kind: node.kind, children: (node.children ?? []).map(fromNode) }
  }
  if (root.kind === 'concat') return (root.children ?? []).map(fromNode)
  return [fromNode(root)]
}
