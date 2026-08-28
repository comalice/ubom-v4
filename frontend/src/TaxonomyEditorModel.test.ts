import { describe, expect, it } from 'vitest'
import { effectiveTaxonomyAttributes, taxonomyAssignmentErrorKey, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyDraftNode } from './TaxonomyEditorModel'

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
