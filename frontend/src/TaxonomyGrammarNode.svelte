<script lang="ts">
  import DropZone from './DropZone.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import type { StructureNode } from './StructureTree.svelte'

  export let node: StructureNode
  export let depth = 0

  function dropLabel(node: StructureNode) {
    if (node.id === 'grammar-root') return 'Drop a category at the root'
    if (node.blockType === 'binding') return `Drop a refinement for ${node.label}`
    return 'Taxonomy nodes cannot attach to this grammar item'
  }

  function dropState(node: StructureNode) {
    if (node.id === 'grammar-root') return 'new'
    if (node.id === 'number') return 'move'
    return 'inactive'
  }

  $: state = dropState(node)
  $: label = dropLabel(node)
</script>

<li>
  <span class="dnd-structure-node">{node.label}<small>{node.summary}</small></span>
  <DropZone variant={node.id === 'number' ? 'filled' : 'slot'} mode={state} className="taxonomy-drop-zone taxonomy-grammar-drop-zone dnd-insert-zone" ariaLabel={`${label} drop zone`} showHint={state === 'inactive'} showLabel={depth > 1 || node.id === 'number' || state !== 'inactive'} label={depth > 1 ? label : ''}>
    {#if node.id === 'number'}
      <span class="taxonomy-drop-label">Drop a taxonomy node for number</span>
      <TaxonomyCategoryBlock title="High Power" summary="number = 100..199" label="High Power" nodeId="high-power" predicate="number = 100..199" addChildLabel="+ Add child category" layout="tree" />
    {:else if depth <= 1}
      <span class="taxonomy-drop-label">{label}</span>
    {/if}
  </DropZone>
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
        <svelte:self node={child} depth={depth + 1} />
      {/each}
    </ul>
  {/if}
</li>
