import { describe, expect, it } from 'vitest'
import { acceptsTaxonomyCategoryChildDrop, acceptsTaxonomyRuleDrop, effectiveTaxonomyAttributes, moveTaxonomyDraftChild, moveTaxonomyDraftRule, reorderTaxonomyDraftRule, taxonomyAssignmentErrorKey, taxonomyDefinitionInputFromDraft, taxonomyDraftFromDefinition, taxonomyGrammarFromSeqDef, taxonomyTargetContainer, updateTaxonomyDraftNode, updateTaxonomyDraftRule, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyDraftNode, type TaxonomyPlacement } from './TaxonomyEditorModel'
import type { TaxonomyPredicate } from './api'

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
    const draft = taxonomyDraftFromDefinition({ id: 'tax-1', definitionId: 'tax', version: 1, seqDef: 'seq-1', attributeDefs: [], root: {
      id: 'root', label: 'Parts', attributes: [{ attributeDefId: 'finish', required: true }], children: [{ id: 'metal', label: 'Metal', predicates: [{ field: 'family', kind: 'exact', value: 'A' }], children: [] }],
    } }, grammar)
    expect(draft.attributes).toEqual([{ attributeDefId: 'finish', required: true }])
    expect(draft.children[0].targetId).toBe('grammar-root')
    expect(draft.children[0].placements[0].predicate).toBe('family = A')
  })

  it('targets duplicate numeric bindings by predicate domain', () => {
    const grammar = taxonomyGrammarFromSeqDef({ id: 'gtd-pn-v1', definitionId: 'gtd-pn', version: 1, root: { kind: 'choice', children: [
      { kind: 'concat', children: [{ kind: 'bind', name: 'family', child: { kind: 'range', min: 100, max: 999, width: 3 } }] },
      { kind: 'concat', children: [{ kind: 'bind', name: 'family', child: { kind: 'range', max: 99, width: 3 } }] },
    ] } })
    const definition = (id: string, predicate: TaxonomyPredicate) => ({
      id: 'taxonomy', definitionId: 'taxonomy', version: 1, seqDef: 'gtd-pn-v1', attributeDefs: [], root: {
        id: 'root', label: 'Parts', children: [{ id, label: id, predicates: [predicate], children: [] }],
      },
    })
    const target = (id: string, predicate: TaxonomyPredicate) => taxonomyDraftFromDefinition(definition(id, predicate), grammar).children[0]
    const nonRaw = target('family-100-999', { field: 'family', kind: 'numeric-range', min: 100, max: 999 })
    const raw = target('raw-material', { field: 'family', kind: 'numeric-range', min: 0, max: 99 })
    const exact = target('family-302', { field: 'family', kind: 'exact', value: '302' })

    expect(nonRaw.targetId).toBe('grammar-root-alternative-0-0')
    expect(raw.targetId).toBe('grammar-root-alternative-1-0')
    expect(exact.targetId).toBe('grammar-root-alternative-0-0')
    expect(nonRaw.placements[0].predicate).toBe('family = 100..999')
    expect(exact.placements[0].predicate).toBe('family = 302')
  })

  it('keeps unique binding targeting unchanged', () => {
    const grammar = taxonomyGrammarFromSeqDef({ id: 'seq-unique', definitionId: 'seq', version: 1, root: { kind: 'bind', name: 'family', child: { kind: 'range', min: 0, max: 999, width: 3 } } })
    const draft = taxonomyDraftFromDefinition({ id: 'tax-unique', definitionId: 'tax', version: 1, seqDef: 'seq-unique', attributeDefs: [], root: {
      id: 'root', label: 'Parts', children: [{ id: 'family', label: 'Family', predicates: [{ field: 'family', kind: 'numeric-range', min: 100, max: 199 }], children: [] }],
    } }, grammar)
    expect(draft.children[0].targetId).toBe('grammar-root')
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
  it('accepts palette and existing category child payloads, but requires source context', () => {
    expect(acceptsTaxonomyCategoryChildDrop({ kind: 'palette', taxonomyType: 'category' }, 'destination')).toBe(true)
    expect(acceptsTaxonomyCategoryChildDrop({ kind: 'taxonomy', taxonomyType: 'category', placementId: 'source-placement', sourceNodeId: 'source' }, 'destination')).toBe(true)
    expect(acceptsTaxonomyCategoryChildDrop({ kind: 'taxonomy', taxonomyType: 'category', placementId: 'source-placement' }, 'destination')).toBe(false)
    expect(acceptsTaxonomyCategoryChildDrop({ kind: 'taxonomy', taxonomyType: 'category', placementId: 'source-placement', sourceNodeId: 'destination' }, 'destination')).toBe(false)
  })

  it('moves an existing category under another category without changing identity', () => {
    const source: TaxonomyDraftNode = { id: 'source', label: 'Source', targetId: 'source-field', placements: [{ id: 'source-placement', targetId: 'source-field', type: 'category', title: 'Source', summary: 'Source', nodeId: 'source' }], rules: [], children: [] }
    const destination: TaxonomyDraftNode = { id: 'destination', label: 'Destination', targetId: 'destination-field', placements: [{ id: 'destination-placement', targetId: 'destination-field', type: 'category', title: 'Destination', summary: 'Destination', nodeId: 'destination' }], rules: [], children: [] }
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'root', placements: [], rules: [], children: [source, destination] }
    const result = moveTaxonomyDraftChild(draft, 'source', 'destination')

    expect(result.children.map(({ id }) => id)).toEqual(['destination'])
    expect(result.children[0].children[0]).toBe(source)
    expect(result.children[0].children[0].placements[0]).toBe(source.placements[0])
    expect(draft.children).toEqual([source, destination])
  })

  it('supports nested moves and rejects self, descendant, missing, and invalid destinations', () => {
    const leaf: TaxonomyDraftNode = { id: 'leaf', label: 'Leaf', targetId: 'leaf-field', placements: [], rules: [], children: [] }
    const branch: TaxonomyDraftNode = { id: 'branch', label: 'Branch', targetId: 'branch-field', placements: [], rules: [], children: [leaf] }
    const other: TaxonomyDraftNode = { id: 'other', label: 'Other', targetId: 'other-field', placements: [], rules: [], children: [] }
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'root', placements: [], rules: [], children: [branch, other] }

    const nested = moveTaxonomyDraftChild(draft, 'leaf', 'other')
    expect(nested.children[0].children).toEqual([])
    expect(nested.children[1].children[0]).toBe(leaf)
    expect(moveTaxonomyDraftChild(draft, 'branch', 'leaf')).toBe(draft)
    expect(moveTaxonomyDraftChild(draft, 'branch', 'branch')).toBe(draft)
    expect(moveTaxonomyDraftChild(draft, 'missing', 'other')).toBe(draft)
    expect(moveTaxonomyDraftChild(draft, 'branch', 'other', 2)).toBe(draft)
  })

  it('honors child insertion indexes and adjusts same-parent moves', () => {
    const children = ['one', 'two', 'three'].map((id) => ({ id, label: id, targetId: id, placements: [], rules: [], children: [] })) as TaxonomyDraftNode[]
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'root', placements: [], rules: [], children }
    expect(moveTaxonomyDraftChild(draft, 'three', 'root', 0).children.map(({ id }) => id)).toEqual(['three', 'one', 'two'])
    expect(moveTaxonomyDraftChild(draft, 'one', 'root', 3).children.map(({ id }) => id)).toEqual(['two', 'three', 'one'])
  })

  it('accepts palette, same-node, and cross-node existing rules, but rejects missing source context', () => {
    expect(acceptsTaxonomyRuleDrop({ kind: 'palette', taxonomyType: 'rule' })).toBe(true)
    expect(acceptsTaxonomyRuleDrop({ kind: 'taxonomy', taxonomyType: 'rule', placementId: 'rule', sourceNodeId: 'node' }, 'node')).toBe(true)
    expect(acceptsTaxonomyRuleDrop({ kind: 'taxonomy', taxonomyType: 'rule', placementId: 'rule', sourceNodeId: 'other' }, 'node')).toBe(true)
    expect(acceptsTaxonomyRuleDrop({ kind: 'taxonomy', taxonomyType: 'rule', placementId: 'rule' }, 'node')).toBe(false)
    expect(acceptsTaxonomyRuleDrop({ kind: 'palette', taxonomyType: 'category' })).toBe(false)
    expect(acceptsTaxonomyRuleDrop(undefined)).toBe(false)
  })

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

  it('reorders rules from first to last and last to first', () => {
    const rules = ['first', 'middle', 'last'].map((id) => ({ id, targetId: 'root', type: 'rule' as const, title: id, summary: id, predicateData: { kind: 'exact' as const, field: 'part', value: id } }))
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'field', placements: [], rules, children: [] }
    expect(reorderTaxonomyDraftRule(draft, 'root', 'first', 3).rules.map((rule) => rule.id)).toEqual(['middle', 'last', 'first'])
    expect(reorderTaxonomyDraftRule(draft, 'root', 'last', 0).rules.map((rule) => rule.id)).toEqual(['last', 'first', 'middle'])
  })

  it('inserts a middle rule without changing its identity or predicate', () => {
    const rule: TaxonomyPlacement = { id: 'move', targetId: 'root', type: 'rule', title: 'size = 2..8', summary: 'size = 2..8', predicateData: { kind: 'numeric-range', field: 'size', min: 2, max: 8 } }
    const other: TaxonomyPlacement = { id: 'other', targetId: 'root', type: 'rule', title: 'other', summary: 'other' }
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'field', placements: [], rules: [other, { ...rule, id: 'before' }, rule], children: [] }
    const result = reorderTaxonomyDraftRule(draft, 'root', 'move', 1)
    expect(result.rules.map((current) => current.id)).toEqual(['other', 'move', 'before'])
    expect(result.rules[1]).toBe(rule)
    expect(result.rules[1].predicateData).toEqual(rule.predicateData)
  })

  it('returns the original draft for no-op and invalid rule or node IDs', () => {
    const rule: TaxonomyPlacement = { id: 'rule', targetId: 'root', type: 'rule', title: 'rule', summary: 'rule' }
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'field', placements: [], rules: [rule], children: [] }
    expect(reorderTaxonomyDraftRule(draft, 'root', 'rule', 0)).toBe(draft)
    expect(reorderTaxonomyDraftRule(draft, 'unknown', 'rule', 0)).toBe(draft)
    expect(reorderTaxonomyDraftRule(draft, 'root', 'unknown', 0)).toBe(draft)
  })

  it('moves a rule across nested nodes while preserving predicate data and serializes it under the destination', () => {
    const rule: TaxonomyPlacement = { id: 'move', targetId: 'source', type: 'rule', title: 'Original title', summary: 'Original summary', predicate: 'source-field = value', predicateData: { kind: 'exact', field: 'source-field', value: 'value' } }
    const draft: TaxonomyDraftNode = {
      id: 'root', label: 'Root', targetId: 'grammar-root', placements: [], rules: [], children: [{
        id: 'source', label: 'Source', targetId: 'source-target', placements: [], rules: [rule], children: [],
      }, {
        id: 'branch', label: 'Branch', targetId: 'branch-target', placements: [], rules: [], children: [{
          id: 'destination', label: 'Destination', targetId: 'destination-target', placements: [{ id: 'destination-category', targetId: 'destination-target', type: 'category', title: 'Destination', summary: 'destination-category = x', label: 'Destination', nodeId: 'destination', predicateData: { kind: 'exact', field: 'destination-category', value: 'x' } }], rules: [{ id: 'existing', targetId: 'destination', type: 'rule', title: 'Existing', summary: 'Existing', predicateData: { kind: 'exact', field: 'destination-field', value: 'x' } }], children: [],
        }],
      }],
    }
    const result = moveTaxonomyDraftRule(draft, 'source', 'destination', 'move', 1)
    expect(result.children[0].rules).toEqual([])
    expect(result.children[1].children[0].rules.map(({ id }) => id)).toEqual(['existing', 'move'])
    expect(result.children[1].children[0].rules[1]).toEqual({ ...rule, targetId: 'destination' })
    expect(taxonomyDefinitionInputFromDraft('seq', undefined, [], result).root.children?.[1].children?.[0].predicates).toEqual([{ field: 'destination-category', kind: 'exact', value: 'x' }, { field: 'destination-field', kind: 'exact', value: 'x' }, { field: 'source-field', kind: 'exact', value: 'value' }])
    expect(draft.children[0].rules[0]).toBe(rule)
    expect(draft.children[1].children[0].rules).toHaveLength(1)
  })

  it('supports same-node first, middle, and last insertion gaps through move semantics', () => {
    const rules = ['first', 'middle', 'last'].map((id) => ({ id, targetId: 'node', type: 'rule' as const, title: id, summary: id }))
    const draft: TaxonomyDraftNode = { id: 'node', label: 'Node', targetId: 'field', placements: [], rules, children: [] }
    expect(moveTaxonomyDraftRule(draft, 'node', 'node', 'first', 3).rules.map(({ id }) => id)).toEqual(['middle', 'last', 'first'])
    expect(moveTaxonomyDraftRule(draft, 'node', 'node', 'middle', 0).rules.map(({ id }) => id)).toEqual(['middle', 'first', 'last'])
    expect(moveTaxonomyDraftRule(draft, 'node', 'node', 'last', 1).rules.map(({ id }) => id)).toEqual(['first', 'last', 'middle'])
  })

  it('rejects missing or ambiguous ownership and invalid destination indexes immutably', () => {
    const rule: TaxonomyPlacement = { id: 'rule', targetId: 'source', type: 'rule', title: 'Rule', summary: 'Rule' }
    const draft: TaxonomyDraftNode = { id: 'root', label: 'Root', targetId: 'root', placements: [], rules: [], children: [{ id: 'source', label: 'Source', targetId: 'source', placements: [], rules: [rule], children: [] }, { id: 'destination', label: 'Destination', targetId: 'destination', placements: [], rules: [], children: [] }] }
    expect(moveTaxonomyDraftRule(draft, 'missing', 'destination', 'rule', 0)).toBe(draft)
    expect(moveTaxonomyDraftRule(draft, 'source', 'destination', 'missing', 0)).toBe(draft)
    expect(moveTaxonomyDraftRule(draft, 'source', 'destination', 'rule', 2)).toBe(draft)
    expect(moveTaxonomyDraftRule(draft, 'source', 'destination', 'rule', -1)).toBe(draft)
    const ambiguous = { ...draft, children: [...draft.children, { ...draft.children[1], id: 'other', rules: [rule] }] }
    expect(moveTaxonomyDraftRule(ambiguous, 'source', 'destination', 'rule', 0)).toBe(ambiguous)
  })
})
