<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import type { BomNode } from './api'
  export let node: BomNode
</script>

<li>
  <IdentityLink partNumber={node.partNumber.value} revision={String(node.revisionId)} />
  {#if node.bom.length > 0}
    <ul>{#each node.bom as child (child.revisionId)}<svelte:self node={child} />{/each}</ul>
  {/if}
</li>
