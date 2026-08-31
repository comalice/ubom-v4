import { describe, expect, it } from 'vitest'
import {
  generateGrammarExpr,
  generateGrammarSequence,
  getGrammarPreviewStatus,
  grammarPreviewKey,
  toSeqDefNode,
  toSeqDefRoot,
  type GrammarExpr,
  validateGrammarSequence,
} from './GrammarEditorModel'

const literal = (id: string, text: string): GrammarExpr => ({ id, kind: 'literal', text })

describe('GrammarEditorModel AST conversion', () => {
  it('converts every supported editor node and removes UI ids', () => {
    const expressions: GrammarExpr[] = [
      literal('literal-id', 'PN-'),
      { id: 'range-id', kind: 'range', min: 10, max: 99, width: 3, pad: '0' },
      { id: 'places-id', kind: 'places', alphabets: ['AB', '012'] },
      {
        id: 'radix-id',
        kind: 'range_radix',
        children: [
          { id: 'radix-range-id', kind: 'range', min: 1, max: 9, width: 1, pad: ' ' },
          { id: 'radix-places-id', kind: 'places', alphabets: ['XYZ'] },
        ],
      },
      {
        id: 'choice-id',
        kind: 'choice',
        children: [
          [literal('choice-a-id', 'A')],
          [literal('choice-b-id', 'B'), { id: 'choice-b-range-id', kind: 'range', min: 0, max: 9, width: 2, pad: '0' }],
        ],
      },
      {
        id: 'bind-id',
        kind: 'bind',
        name: 'serial',
        child: { id: 'bound-child-id', kind: 'literal', text: 'S' },
      },
    ]

    expect(toSeqDefRoot(expressions)).toEqual({
      kind: 'concat',
      children: [
        { kind: 'literal', text: 'PN-' },
        { kind: 'range', min: 10, max: 99, width: 3, pad: '0' },
        { kind: 'places', alphabets: ['AB', '012'] },
        {
          kind: 'range_radix',
          children: [
            { kind: 'range', min: 1, max: 9, width: 1, pad: ' ' },
            { kind: 'places', alphabets: ['XYZ'] },
          ],
        },
        {
          kind: 'choice',
          children: [
            { kind: 'literal', text: 'A' },
            { kind: 'concat', children: [{ kind: 'literal', text: 'B' }, { kind: 'range', min: 0, max: 9, width: 2, pad: '0' }] },
          ],
        },
        { kind: 'bind', name: 'serial', child: { kind: 'literal', text: 'S' } },
      ],
    })
  })

  it('converts nested bindings recursively', () => {
    const expression: GrammarExpr = {
      id: 'outer',
      kind: 'bind',
      name: 'outer',
      child: {
        id: 'inner',
        kind: 'bind',
        name: 'inner',
        child: literal('value', 'X'),
      },
    }

    expect(toSeqDefNode(expression)).toEqual({
      kind: 'bind',
      name: 'outer',
      child: { kind: 'bind', name: 'inner', child: { kind: 'literal', text: 'X' } },
    })
  })

  it('preserves empty and incomplete editor states at the boundary', () => {
    expect(toSeqDefRoot([])).toEqual({ kind: 'concat', children: [] })
    expect(toSeqDefNode({ id: 'empty-choice', kind: 'choice', children: [] })).toEqual({ kind: 'choice', children: [] })
    expect(toSeqDefNode({ id: 'empty-radix', kind: 'range_radix', children: [] })).toEqual({ kind: 'range_radix', children: [] })
    expect(toSeqDefNode({ id: 'empty-places', kind: 'places', alphabets: [] })).toEqual({ kind: 'places', alphabets: [] })
    expect(() => toSeqDefNode({ id: 'missing-child', kind: 'bind', name: 'value' })).toThrow('binding needs an expression')
  })
})

describe('GrammarEditorModel generation', () => {
  it('generates literals, ranges, places, radix sequences, choices, and bindings', () => {
    const expressions: GrammarExpr[] = [
      literal('literal', 'PN-'),
      { id: 'range', kind: 'range', min: 7, max: 9, width: 3, pad: '0' },
      { id: 'places', kind: 'places', alphabets: ['ABC', '12'] },
      { id: 'radix', kind: 'range_radix', children: [literal('radix-literal', '-'), { id: 'radix-range', kind: 'range', min: 4, max: 8, width: 2, pad: '0' }] },
      { id: 'choice', kind: 'choice', children: [[literal('first', 'A')], [literal('second', 'B'), literal('suffix', 'X')]] },
      { id: 'bind', kind: 'bind', name: 'value', child: literal('bound', 'Z') },
    ]

    expect(generateGrammarSequence(expressions)).toBe('PN-007A1-04AZ')
    expect(generateGrammarExpr(expressions[3])).toBe('-04')
  })

  it('serializes custom range padding and empty states', () => {
    expect(generateGrammarExpr({ id: 'space-padded', kind: 'range', min: 4, max: 9, width: 3, pad: ' ' })).toBe('  4')
    expect(generateGrammarExpr({ id: 'empty-literal', kind: 'literal', text: '' })).toBe('')
    expect(generateGrammarExpr({ id: 'empty-places', kind: 'places', alphabets: ['', 'ABC'] })).toBe('A')
    expect(generateGrammarExpr({ id: 'empty-choice', kind: 'choice', children: [] })).toBe('')
    expect(generateGrammarExpr({ id: 'empty-bind', kind: 'bind', name: 'value' })).toBe('')
  })
})

describe('GrammarEditorModel preview inputs', () => {
  it('changes when the document or binding inputs change', () => {
    const expressions = [literal('literal', 'PN-')]
    const initial = grammarPreviewKey(expressions, { family: 'SN-' })

    expect(grammarPreviewKey(expressions, { family: 'SN-' })).toBe(initial)
    expect(grammarPreviewKey([literal('literal', 'PN-2')], { family: 'SN-' })).not.toBe(initial)
    expect(grammarPreviewKey(expressions, { family: 'SN-2' })).not.toBe(initial)
  })

  it('classifies current preview lifecycle states without allowing stale results', () => {
    const base = { currentKey: 'current', previewedKey: null, preview: null, requestKey: null, requestError: '', previewing: false }
    expect(getGrammarPreviewStatus(base).kind).toBe('not-previewed')
    expect(getGrammarPreviewStatus({ ...base, requestKey: 'current', previewing: true }).kind).toBe('previewing')
    expect(getGrammarPreviewStatus({ ...base, previewedKey: 'current', preview: {} }).kind).toBe('success')
    expect(getGrammarPreviewStatus({ ...base, previewedKey: 'current', preview: { error: 'bad binding' } }).kind).toBe('server-rejected')
    expect(getGrammarPreviewStatus({ ...base, requestKey: 'current', requestError: 'Network unavailable' }).kind).toBe('request-failed')
    expect(getGrammarPreviewStatus({ ...base, previewedKey: 'old', preview: {}, requestKey: 'old', requestError: 'old error' }).kind).toBe('not-previewed')
  })
})

describe('GrammarEditorModel structural validation', () => {
  it('reports structural errors with document paths without changing the model', () => {
    const expressions: GrammarExpr[] = [
      { id: 'choice', kind: 'choice', children: [[]] },
      { id: 'first', kind: 'bind', name: 'value' },
      { id: 'second', kind: 'bind', name: 'value', child: literal('child', 'x') },
      { id: 'range', kind: 'range', min: 9, max: 2, width: 0, pad: '1' },
      { id: 'places', kind: 'places', alphabets: ['', 'AB'] },
      { id: 'radix', kind: 'range_radix', children: [] },
    ]

    const errors = validateGrammarSequence(expressions)
    expect(errors.map(({ code, path }) => ({ code, path }))).toEqual([
      { code: 'empty-choice-alternative', path: 'root[0].choice[0]' },
      { code: 'missing-bind-child', path: 'root[1].bind' },
      { code: 'duplicate-binding-name', path: 'root[2].bind' },
      { code: 'invalid-range-bounds', path: 'root[3].range' },
      { code: 'invalid-width', path: 'root[3].range' },
      { code: 'invalid-range-definition', path: 'root[3].range' },
      { code: 'invalid-character-definition', path: 'root[4].places' },
      { code: 'invalid-radix-definition', path: 'root[5].range_radix' },
    ])
    expect(expressions[3]).toEqual({ id: 'range', kind: 'range', min: 9, max: 2, width: 0, pad: '1' })
  })

  it('reports an incomplete binding without requiring AST conversion', () => {
    const expressions: GrammarExpr[] = [{ id: 'binding', kind: 'bind', name: 'value' }]

    expect(validateGrammarSequence(expressions)).toEqual([{
      code: 'missing-bind-child',
      message: 'Binding needs an expression',
      path: 'root[0].bind',
      nodeId: 'binding',
    }])
  })

  it('allows shared binding names across choice alternatives', () => {
    const expressions: GrammarExpr[] = [{
      id: 'choice',
      kind: 'choice',
      children: [
        [{ id: 'first', kind: 'bind', name: 'family', child: literal('first-value', 'A') }],
        [{ id: 'second', kind: 'bind', name: 'family', child: literal('second-value', 'B') }],
      ],
    }]

    expect(validateGrammarSequence(expressions)).toEqual([])
  })

  it('rejects duplicate binding names on one possible path', () => {
    const expressions: GrammarExpr[] = [{
      id: 'choice',
      kind: 'choice',
      children: [
        [
          { id: 'first', kind: 'bind', name: 'family', child: literal('first-value', 'A') },
          { id: 'second', kind: 'bind', name: 'family', child: literal('second-value', 'B') },
        ],
        [literal('other', 'C')],
      ],
    }]

    expect(validateGrammarSequence(expressions).map(({ code, path }) => ({ code, path }))).toEqual([
      { code: 'duplicate-binding-name', path: 'root[0].choice[0][1].bind' },
    ])
  })

  it('rejects unsupported and incomplete runtime nodes', () => {
    expect(validateGrammarSequence([{ id: 'unknown', kind: 'future' }])).toEqual([
      { code: 'unsupported-node', message: 'Unsupported grammar node kind "future"', path: 'root[0].future', nodeId: 'unknown' },
    ])
    expect(validateGrammarSequence([null, { id: 'literal', kind: 'literal' }])).toEqual([
      { code: 'unsupported-node', message: 'Unsupported or incomplete grammar node', path: 'root[0]' },
      { code: 'incomplete-node', message: 'Literal text is required', path: 'root[1].literal', nodeId: 'literal' },
    ])
  })
})
