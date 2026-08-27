<script lang="ts">
  import TaxonomyGrammarNode from './TaxonomyGrammarNode.svelte'
  import type { StructureNode } from './StructureTree.svelte'

  const grammarNodes: StructureNode[] = [
    { id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v1', children: [
      { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
      { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [
        { id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] },
      ] },
      { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
    ] },
  ]

</script>

<ul class="dnd-structure-tree taxonomy-grammar-structure-tree">
  {#each grammarNodes as node (node.id)}
    <TaxonomyGrammarNode {node} />
  {/each}
</ul>
