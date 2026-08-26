import { describe, expect, it } from 'vitest'
import {
  generateGrammarExpr,
  generateGrammarSequence,
  toSeqDefNode,
  toSeqDefRoot,
  type GrammarExpr,
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
