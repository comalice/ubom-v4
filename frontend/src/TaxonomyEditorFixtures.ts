import type { StructureNode } from './StructureNode'

export const grammarNodes: StructureNode[] = [
  { id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v1', children: [
    { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
    { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [
      { id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] },
    ] },
    { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
  ] },
]
