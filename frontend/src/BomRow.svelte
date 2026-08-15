<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import type { BomNode } from './api'

  export let node: BomNode
  export let depth = 0

  let expanded = true
  $: hasChildren = node.bom.length > 0
</script>

<tr class="bom-row">
  <th scope="row">
    <div class="bom-identity" style={`padding-left: ${depth * 1.25}rem`}>
      <button
        type="button"
        class:has-children={hasChildren}
        class="bom-toggle"
        aria-label={expanded ? 'Collapse BOM row' : 'Expand BOM row'}
        aria-expanded={hasChildren ? expanded : undefined}
        disabled={!hasChildren}
        on:click={() => expanded = !expanded}
      >
        {#if hasChildren}
          <span aria-hidden="true">{expanded ? '⌄' : '›'}</span>
        {/if}
      </button>
      <IdentityLink partNumber={node.partNumber.value} revision={node.revision || String(node.revisionId)} revisionID={String(node.revisionId)} />
    </div>
  </th>
  <td class="bom-unsupported" aria-label="Part number description not supported">—</td>
  <td class="bom-unsupported" aria-label="Part number status not supported">—</td>
  <td class="bom-unsupported bom-number" aria-label="Quantity not supported">—</td>
  <td class="bom-unsupported" aria-label="Unit of measure not supported">—</td>
</tr>
{#if hasChildren && expanded}
  {#each node.bom as child, index (child.revisionId + '-' + index)}
    <svelte:self node={child} depth={depth + 1} />
  {/each}
{/if}
