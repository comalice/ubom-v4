<script lang="ts">
  import TaxonomyGrammarDropTarget from './TaxonomyGrammarDropTarget.svelte'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let depth = 0

</script>

<li>
  <span class="dnd-structure-node">{node.label}<small>{node.summary}</small></span>
  <TaxonomyGrammarDropTarget {node} {depth} />
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
        <svelte:self node={child} depth={depth + 1} />
      {/each}
    </ul>
  {/if}
</li>
