<script lang="ts">
  import DropZone, { type DropZoneMode } from './DropZone.svelte'
  import TaxonomyCategoryBlock from './TaxonomyCategoryBlock.svelte'
  import type { StructureNode } from './StructureNode'
  import type { TaxonomyGrammarCategoryBlockFixture } from './TaxonomyEditorFixtures'

  export let node: StructureNode
  export let depth = 0
  export let filled = false
  export let dropMode: DropZoneMode = 'inactive'
  export let categoryLabel = ''
  export let categoryBlock: TaxonomyGrammarCategoryBlockFixture | undefined

  function dropLabel(node: StructureNode) {
    if (node.id === 'grammar-root') return 'Drop a category at the root'
    if (node.blockType === 'binding') return `Drop a refinement for ${node.label}`
    return 'Taxonomy nodes cannot attach to this grammar item'
  }

  $: label = dropLabel(node)
</script>

<DropZone variant={filled ? 'filled' : 'slot'} mode={dropMode} className="taxonomy-drop-zone taxonomy-grammar-drop-zone dnd-insert-zone" ariaLabel={`${label} drop zone`} showHint={dropMode === 'inactive'} showLabel={depth > 1 || filled || dropMode !== 'inactive'} label={depth > 1 ? label : ''}>
  {#if filled}
    <span class="taxonomy-drop-label">{categoryLabel}</span>
    {#if categoryBlock}
      <TaxonomyCategoryBlock {...categoryBlock} />
    {/if}
  {:else if depth <= 1}
    <span class="taxonomy-drop-label">{label}</span>
  {/if}
</DropZone>
