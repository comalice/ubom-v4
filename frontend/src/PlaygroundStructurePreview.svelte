<script lang="ts">
  export type StructureNode = { id: number; label: string; blockType: string; children: StructureNode[] }

  export let nodes: StructureNode[] = []
  export let showIds = true
</script>

<ul class="dnd-structure-tree">
  {#each nodes as node (node.id)}
    <li>
       <span class="dnd-structure-node">{node.label}{#if showIds}<small>#{node.id} {node.blockType}</small>{/if}</span>
      {#if node.children.length}<svelte:self nodes={node.children} />{/if}
    </li>
  {:else}
    <li class="dnd-structure-empty">Empty component list</li>
  {/each}
</ul>
