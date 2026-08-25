<script lang="ts">
  export type StructureNode = { id: string | number; label: string; blockType: string; summary?: string; children: StructureNode[] }

  export let nodes: StructureNode[] = []
  export let showIds = true
</script>

<ul class="dnd-structure-tree">
  {#each nodes as node (node.id)}
    <li>
       <span class="dnd-structure-node">{node.label}<small>{#if showIds}#{node.id} {node.blockType}{:else}{node.summary}{/if}</small></span>
      {#if node.children.length}<svelte:self nodes={node.children} showIds={showIds} />{/if}
    </li>
  {:else}
    <li class="dnd-structure-empty">Empty component list</li>
  {/each}
</ul>
