<script lang="ts">
  import TaxonomyGrammarNode from './TaxonomyGrammarNode.svelte'
  import { grammarNodes } from './TaxonomyEditorFixtures'
</script>

<ul class="dnd-structure-tree taxonomy-grammar-structure-tree">
  {#each grammarNodes as node (node.id)}
    <TaxonomyGrammarNode {node} />
  {/each}
</ul>
