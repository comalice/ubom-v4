import type { StructureNode } from './StructureNode'

export type TaxonomyGrammarFixture = {
  id: string
  label: string
  fields: number
  nodes: StructureNode[]
}

const sampleGrammarNodes: StructureNode[] = [
  { id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v1', children: [
    { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
    { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [
      { id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] },
    ] },
    { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
  ] },
]

export const grammarFixtures: TaxonomyGrammarFixture[] = [
  { id: 'sample-pn-v1', label: 'Sample part number', fields: 3, nodes: sampleGrammarNodes },
  { id: 'sample-pn-v2', label: 'Extended part number', fields: 4, nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v2', children: [
    { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
    { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [{ id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] }] },
    { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
    { id: 'revision', label: 'revision', blockType: 'binding', summary: 'choice: A · B', children: [] },
  ] }] },
  { id: 'sample-pn-v3', label: 'Regional part number', fields: 4, nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v3', children: [
    { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"UB-"', children: [] },
    { id: 'region', label: 'region', blockType: 'binding', summary: 'choice: EU · US', children: [] },
    { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [{ id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] }] },
    { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
  ] }] },
]

export const grammarNodes = sampleGrammarNodes

export type TaxonomyGrammarCategoryBlockFixture = {
  title: string
  summary: string
  label: string
  nodeId: string
  predicate: string
  addChildLabel: string
  layout: 'catalogue' | 'tree'
}

export type TaxonomyGrammarDropTargetFixture = {
  filled: boolean
  dropMode: 'inactive' | 'new' | 'move'
  categoryLabel: string
  categoryBlock?: TaxonomyGrammarCategoryBlockFixture
}

export const taxonomyGrammarDropTargetFixtures: Record<string, TaxonomyGrammarDropTargetFixture> = {
  number: {
    filled: true,
    dropMode: 'move',
    categoryLabel: 'Drop a taxonomy node for number',
    categoryBlock: {
      title: 'High Power',
      summary: 'number = 100..199',
      label: 'High Power',
      nodeId: 'high-power',
      predicate: 'number = 100..199',
      addChildLabel: '+ Add child category',
      layout: 'tree',
    },
  },
}

export type TaxonomyCategoryExample = {
  title: string
  summary: string
  label: string
  nodeId: string
  predicate: string
  description: string
  example: string
}

export const taxonomyCategoryExample: TaxonomyCategoryExample = {
  title: 'category',
  summary: 'Resistors',
  label: 'Resistors',
  nodeId: 'resistors',
  predicate: 'category = A',
  description: 'A named classification node attached to a grammar location.',
  example: 'Resistors',
}

export type TaxonomyRulePredicateExample =
  | { kind: 'exact'; field: string; value: string }
  | { kind: 'numeric-range'; field: string; min: number; max: number }

export type TaxonomyRuleExample = {
  title: string
  summary: string
  predicate: TaxonomyRulePredicateExample
  description: string
  example: string
}

export const taxonomyRuleExample: TaxonomyRuleExample = {
  title: 'rule',
  summary: 'category = A',
  predicate: { kind: 'exact', field: 'category', value: 'A' },
  description: 'A predicate that narrows the meaning of a taxonomy branch using a captured field.',
  example: 'category = A',
}
