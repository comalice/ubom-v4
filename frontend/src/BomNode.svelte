<script lang="ts">
  import type { BomNode } from './api'
  export let node: BomNode
</script>

<li>
  <a href={revisionPath(node.partNumber.value, node.revisionId)}>{node.partNumber.value}</a>
  <span class="muted">(revision {node.revisionId})</span>
  {#if node.bom.length > 0}
    <ul>{#each node.bom as child (child.revisionId)}<svelte:self node={child} />{/each}</ul>
  {/if}
</li>

<script context="module" lang="ts">
  import { revisionPath } from './routes'
</script>
