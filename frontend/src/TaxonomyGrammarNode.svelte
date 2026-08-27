<script lang="ts">
  import TaxonomyGrammarDropTarget from './TaxonomyGrammarDropTarget.svelte'
  import { taxonomyGrammarDropTargetFixtures } from './TaxonomyEditorFixtures'
  import type { DropZoneMode } from './DropZone.svelte'
  import type { StructureNode } from './StructureNode'

  export let node: StructureNode
  export let depth = 0

  $: dropTargetFixture = taxonomyGrammarDropTargetFixtures[String(node.id)]
  $: dropMode = dropTargetFixture?.dropMode ?? (node.id === 'grammar-root' ? 'new' : 'inactive') as DropZoneMode
</script>

<li>
  <span class="dnd-structure-node">{node.label}<small>{node.summary}</small></span>
  <TaxonomyGrammarDropTarget {node} {depth} filled={dropTargetFixture?.filled ?? false} {dropMode} categoryLabel={dropTargetFixture?.categoryLabel ?? ''} categoryBlock={dropTargetFixture?.categoryBlock} />
  {#if node.children.length}
    <ul class="dnd-structure-tree">
      {#each node.children as child (child.id)}
        <svelte:self node={child} depth={depth + 1} />
      {/each}
    </ul>
  {/if}
</li>
