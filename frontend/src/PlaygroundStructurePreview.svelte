<script lang="ts">
  import type { ComponentNode } from './PlaygroundDraggable.svelte'

  export let nodes: ComponentNode[] = []
</script>

<ul class="dnd-structure-tree">
  {#each nodes as node (node.id)}
    <li>
      <span class="dnd-structure-node">{node.label}<small>#{node.id} {node.blockType}</small></span>
      {#if node.children.length}<svelte:self nodes={node.children} />{/if}
    </li>
  {:else}
    <li class="dnd-structure-empty">Empty component list</li>
  {/each}
</ul>
