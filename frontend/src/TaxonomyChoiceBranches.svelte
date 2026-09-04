<script lang="ts">
  import { taxonomyGrammarChoiceDropLabel, taxonomyGrammarChoices, taxonomyGrammarFirstBinding, type TaxonomyDragPayload, type TaxonomyGrammar } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'
  import type { DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'

  export let grammar: TaxonomyGrammar
  export let targetId = ''
  export let readOnly = false
  export let onBranchDrop: (state: DragDropState<TaxonomyDragPayload>, branchPath: string) => void = () => {}
  export let onBranchFieldDrop: (state: DragDropState<TaxonomyDragPayload>, branchPath: string, field: string) => void = () => {}

  $: choices = taxonomyGrammarChoices(grammar, targetId)

  function branchPath(alternative: StructureNode): string {
    return alternative.branchPath ?? alternative.path ?? String(alternative.id)
  }

</script>

{#if choices.length}
  {#each choices as choice (choice.id)}
    {#each choice.children as alternative (alternative.id)}
      {@const occurrence = taxonomyGrammarFirstBinding(alternative.children)}
      {@const path = branchPath(alternative)}
      {@const targetLabel = taxonomyGrammarChoiceDropLabel(alternative, occurrence?.label)}
      {#if occurrence}
        {#if !readOnly}
          <DropZone variant="slot" mode="inactive" label={targetLabel} showLabel container={`taxonomy-editor:branch:${path}:field:${occurrence.label}`} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDrop={(state) => onBranchFieldDrop(state, path, occurrence.label)} ariaLabel={targetLabel} />
        {/if}
      {:else if !readOnly}
        <DropZone variant="slot" mode="inactive" label={targetLabel} showLabel container={`taxonomy-editor:branch:${path}`} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDrop={(state) => onBranchDrop(state, path)} ariaLabel={targetLabel} />
      {/if}
    {/each}
  {/each}
{/if}
