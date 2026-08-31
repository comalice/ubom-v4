import { describe, expect, it } from 'vitest'
import { effectiveTaxonomyAttributes, taxonomyAssignmentErrorKey, taxonomyDefinitionInputFromDraft, taxonomyDraftFromDefinition, taxonomyGrammarFromSeqDef, taxonomyTargetContainer, updateTaxonomyDraftNode, updateTaxonomyDraftRule, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyDraftNode } from './TaxonomyEditorModel'

const definitions: TaxonomyAttributeDefinition[] = [
  { id: 'footprint', label: 'Footprint', valueType: 'enum' },
  { id: 'resistance', label: 'Resistance value', valueType: 'decimal', unit: 'ohm' },
]

describe('effectiveTaxonomyAttributes', () => {
  it('keeps inherited order while applying child overrides', () => {
    const root: TaxonomyDraftNode = {
      id: 'root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [],
      attributes: [{ attributeDefId: 'footprint', required: true }],
      children: [{
        id: 'resistors', label: 'Resistors', targetId: 'category', placements: [], rules: [],
        attributes: [{ attributeDefId: 'resistance', required: true }, { attributeDefId: 'footprint', required: false }], children: [],
      }],
    }

    expect(effectiveTaxonomyAttributes(root, 'resistors', definitions)).toEqual([
      { ...definitions[0], required: false, source: 'Resistors', overridden: true },
      { ...definitions[1], required: true, source: 'Resistors', overridden: false },
    ])
  })
})

describe('taxonomy persisted-definition adapters', () => {
  it('renders every persisted sequence node kind into the taxonomy grammar', () => {
    const grammar = taxonomyGrammarFromSeqDef({ id: 'seq-1', definitionId: 'seq', version: 1, root: { kind: 'concat', children: [
      { kind: 'literal', text: 'PN-' },
      { kind: 'bind', name: 'family', child: { kind: 'choice', children: [{ kind: 'literal', text: 'A' }, { kind: 'places', alphabets: ['BC'] }] } },
      { kind: 'range_radix', children: [{ kind: 'range', min: 0, max: 9, width: 1 }, { kind: 'literal', text: '-' }] },
    ] } })
    expect(grammar.fields).toBe(1)
    expect(grammar.nodes[0].children.map((node) => node.blockType)).toEqual(['literal', 'binding', 'range_radix'])
    expect(grammar.nodes[0].children[1].children[0].blockType).toBe('choice')
    expect(grammar.nodes[0].children[1].children[0].children[0].blockType).toBe('row')
    expect(grammar.nodes[0].children[1].children[0].children[0].children[0].blockType).toBe('literal')
    expect(grammar.nodes[0].children[1].children[0].children[1].children[0].blockType).toBe('places')
  })

  it('maps taxonomy predicates, assignments, and descendants to editor nodes', () => {
    const grammar = taxonomyGrammarFromSeqDef({ id: 'seq-1', definitionId: 'seq', version: 1, root: { kind: 'bind', name: 'family', child: { kind: 'literal', text: 'A' } } })
    const draft = taxonomyDraftFromDefinition({ id: 'tax-1', seqDef: 'seq-1', attributeDefs: [], root: {
      id: 'root', label: 'Parts', attributes: [{ attributeDefId: 'finish', required: true }], children: [{ id: 'metal', label: 'Metal', predicates: [{ field: 'family', kind: 'exact', value: 'A' }], children: [] }],
    } }, grammar)
    expect(draft.attributes).toEqual([{ attributeDefId: 'finish', required: true }])
    expect(draft.children[0].targetId).toBe('grammar-root')
    expect(draft.children[0].placements[0].predicate).toBe('family = A')
  })

  it('serializes nested category predicates before each node\'s rules', () => {
    const draft: TaxonomyDraftNode = {
      id: 'root', label: 'Parts', targetId: 'grammar-root', placements: [],
      rules: [{ id: 'root-rule', targetId: 'root', type: 'rule', title: 'kind = part', summary: 'kind = part', predicate: 'kind = part', predicateData: { kind: 'exact', field: 'kind', value: 'part' } }],
      children: [{
        id: 'metal', label: 'Metal', targetId: 'grammar-family',
        placements: [{ id: 'metal', targetId: 'grammar-family', type: 'category', title: 'Metal', summary: 'family = metal', label: 'Metal', nodeId: 'metal', predicate: 'family = metal', predicateData: { kind: 'exact', field: 'family', value: 'metal' } }],
        rules: [{ id: 'metal-rule', targetId: 'metal', type: 'rule', title: 'finish = matte', summary: 'finish = matte', predicate: 'finish = matte', predicateData: { kind: 'exact', field: 'finish', value: 'matte' } }],
        children: [{
          id: 'resistor', label: 'Resistor', targetId: 'grammar-subtype',
          placements: [{ id: 'resistor', targetId: 'grammar-subtype', type: 'category', title: 'Resistor', summary: 'subtype = resistor', label: 'Resistor', nodeId: 'resistor', predicate: 'subtype = resistor', predicateData: { kind: 'exact', field: 'subtype', value: 'resistor' } }],
          rules: [{ id: 'resistor-rule', targetId: 'resistor', type: 'rule', title: 'resistance = 1..9', summary: 'resistance = 1..9', predicate: 'resistance = 1..9', predicateData: { kind: 'numeric-range', field: 'resistance', min: 1, max: 9 } }],
          children: [],
        }],
      }],
    }

    expect(taxonomyDefinitionInputFromDraft('seq-v1', 'tax-v1', [], draft).root).toEqual({
      id: 'root', label: 'Parts',
      predicates: [{ field: 'kind', kind: 'exact', value: 'part' }],
      children: [{
        id: 'metal', label: 'Metal',
        predicates: [
          { field: 'family', kind: 'exact', value: 'metal' },
          { field: 'finish', kind: 'exact', value: 'matte' },
        ],
        children: [{
          id: 'resistor', label: 'Resistor',
          predicates: [
            { field: 'subtype', kind: 'exact', value: 'resistor' },
            { field: 'resistance', kind: 'numeric-range', min: 1, max: 9 },
          ],
        }],
      }],
    })
  })

  it('round-trips nested predicates and rules while preserving target IDs', () => {
    const definition = {
      id: 'tax-v1', definitionId: 'tax', version: 1, seqDef: 'seq-v1', attributeDefs: [], root: {
        id: 'root', label: 'Parts', predicates: [{ field: 'kind', kind: 'exact' as const, value: 'part' }], children: [{
          id: 'metal', label: 'Metal', predicates: [{ field: 'family', kind: 'exact' as const, value: 'metal' }, { field: 'finish', kind: 'exact' as const, value: 'matte' }], children: [{
            id: 'resistor', label: 'Resistor', predicates: [{ field: 'subtype', kind: 'exact' as const, value: 'resistor' }, { field: 'resistance', kind: 'numeric-range' as const, min: 1, max: 9 }],
          }],
        }],
      },
    }
    const grammar = taxonomyGrammarFromSeqDef({ id: 'seq-v1', definitionId: 'seq', version: 1, root: { kind: 'concat', children: [
      { kind: 'bind', name: 'family', child: { kind: 'literal', text: 'M' } },
      { kind: 'bind', name: 'subtype', child: { kind: 'literal', text: 'R' } },
    ] } })
    const draft = taxonomyDraftFromDefinition(definition, grammar)
    const roundTrip = taxonomyDefinitionInputFromDraft('seq-v1', 'parent', [], draft)

    expect(roundTrip.root).toEqual(definition.root)
    expect(draft.children[0]).toMatchObject({ id: 'metal', targetId: 'grammar-0' })
    expect(draft.children[0].placements[0].targetId).toBe('grammar-0')
    expect(draft.children[0].rules.map((rule) => ({ id: rule.id, targetId: rule.targetId }))).toEqual([
      { id: 'metal-rule-0', targetId: 'metal' },
    ])
    expect(draft.children[0].children[0]).toMatchObject({ id: 'resistor', targetId: 'grammar-1' })
    expect(draft.children[0].children[0].placements[0].targetId).toBe('grammar-1')
    expect(draft.children[0].children[0].rules.map((rule) => ({ id: rule.id, targetId: rule.targetId }))).toEqual([
      { id: 'resistor-rule-0', targetId: 'resistor' },
    ])
    expect(taxonomyTargetContainer('metal')).toBe('taxonomy-editor:target:metal')
    expect(taxonomyTargetContainer('resistor')).toBe('taxonomy-editor:target:resistor')
  })
})

describe('validateTaxonomyAttributes', () => {
  const draft = (attributes: TaxonomyDraftNode['attributes'] = []): TaxonomyDraftNode => ({
    id: 'root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [], attributes, children: [],
  })

  it('reports invalid definition values with definition keys', () => {
    const result = validateTaxonomyAttributes([
      { id: 'one', label: '  ', valueType: 'enum', enumValues: [] },
      { id: 'one', label: 'Same', valueType: 'string' },
      { id: 'two', label: 'Same', valueType: 'enum', enumValues: ['', 'A', 'A'] },
    ], draft())

    expect(result.valid).toBe(false)
    expect(result.definitionErrors.one.map((error) => error.code)).toEqual(expect.arrayContaining(['empty-label', 'empty-enum', 'duplicate-id']))
    expect(result.definitionErrors.two.map((error) => error.code)).toEqual(expect.arrayContaining(['duplicate-label', 'empty-enum-value', 'duplicate-enum-value']))
  })

  it('reports missing and duplicate assignments by node and assignment key', () => {
    const result = validateTaxonomyAttributes(
      [{ id: 'known', label: 'Known', valueType: 'string' }],
      draft([{ attributeDefId: 'known', required: false }, { attributeDefId: 'known', required: true }, { attributeDefId: 'missing', required: false }]),
    )

    expect(result.nodeErrors.root[0].code).toBe('duplicate-assignment')
    expect(result.assignmentErrors[taxonomyAssignmentErrorKey('root', 0)][0].code).toBe('duplicate-assignment')
    expect(result.assignmentErrors[taxonomyAssignmentErrorKey('root', 1)][0].code).toBe('duplicate-assignment')
    expect(result.assignmentErrors[taxonomyAssignmentErrorKey('root', 2)][0].code).toBe('missing-definition')
  })

  it('keeps missing definitions visible in effective rows', () => {
    expect(effectiveTaxonomyAttributes(draft([{ attributeDefId: 'missing', required: true }]), 'root', [])).toEqual([
      { id: 'missing', label: 'Missing attribute: missing', valueType: 'string', required: true, source: 'Parts', overridden: false },
    ])
  })
})

describe('taxonomyDefinitionInputFromDraft', () => {
  it('serializes IDs, order, assignments, category predicates, and rules without UI fields', () => {
    const root: TaxonomyDraftNode = {
      id: 'root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [{ id: 'rule', targetId: 'root', type: 'rule', title: 'ignored', summary: 'ignored', predicateData: { kind: 'numeric-range', field: 'size', min: 1, max: 9 } }],
      attributes: [{ attributeDefId: 'finish', required: true }], children: [{
        id: 'metal', label: 'Metal', targetId: 'family', placements: [{ id: 'category', targetId: 'family', type: 'category', title: 'ignored', summary: 'ignored', label: 'Metal', nodeId: 'metal', predicate: 'family = METAL' }], rules: [], attributes: [], children: [],
      }],
    }
    expect(taxonomyDefinitionInputFromDraft('seq-v1', 'tax-v1', definitions, root)).toEqual({
      seqDef: 'seq-v1', parentId: 'tax-v1', attributeDefs: definitions, root: { id: 'root', label: 'Parts', attributes: [{ attributeDefId: 'finish', required: true }], predicates: [{ field: 'size', kind: 'numeric-range', min: 1, max: 9 }], children: [{ id: 'metal', label: 'Metal', predicates: [{ field: 'family', kind: 'exact', value: 'METAL' }] }] },
    })
  })
})

describe('taxonomy draft editing', () => {
  it('updates a node label and category placement without changing placement identity', () => {
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'grammar-root', placements: [], rules: [], children: [{ id: 'child', label: 'Old', targetId: 'field', placements: [{ id: 'placement', targetId: 'field', type: 'category', title: 'Old', summary: 'Category', label: 'Old', nodeId: 'child' }], rules: [], children: [] }] }
    const result = updateTaxonomyDraftNode(draft, 'child', { label: 'New' })
    expect(result.children[0].label).toBe('New')
    expect(result.children[0].placements[0]).toMatchObject({ id: 'placement', label: 'New', title: 'New', nodeId: 'child' })
    expect(result.children[0].placements[0]).not.toBe(draft.children[0].placements[0])
  })

  it('updates a valid node ID while preserving the placement ID', () => {
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'grammar-root', placements: [], rules: [], children: [{ id: 'child', label: 'Child', targetId: 'field', placements: [{ id: 'placement', targetId: 'field', type: 'category', title: 'Child', summary: 'Category', label: 'Child', nodeId: 'child' }], rules: [], children: [] }] }
    const result = updateTaxonomyDraftNode(draft, 'child', { id: 'renamed' })
    expect(result.children[0].id).toBe('renamed')
    expect(result.children[0].placements[0]).toMatchObject({ id: 'placement', nodeId: 'renamed' })
  })

  it('rejects empty and duplicate node IDs', () => {
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'grammar-root', placements: [], rules: [], children: [{ id: 'child', label: 'Child', targetId: 'field', placements: [], rules: [], children: [] }] }
    const duplicate = updateTaxonomyDraftNode(draft, 'child', { id: 'root' })
    const empty = updateTaxonomyDraftNode(draft, 'child', { id: ' ' })
    expect(duplicate).toBe(draft)
    expect(empty).toBe(draft)
  })

  it('updates rule predicate data and display text immutably', () => {
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'grammar-root', placements: [], rules: [{ id: 'rule', targetId: 'root', type: 'rule', title: 'old', summary: 'old' }], children: [] }
    const result = updateTaxonomyDraftRule(draft, 'rule', { kind: 'numeric-range', field: 'size', min: 2, max: 8 })
    expect(result.rules[0]).toMatchObject({ title: 'size = 2..8', summary: 'size = 2..8', predicate: 'size = 2..8', predicateData: { kind: 'numeric-range', min: 2, max: 8 } })
    expect(result.rules[0]).not.toBe(draft.rules[0])
  })
})
