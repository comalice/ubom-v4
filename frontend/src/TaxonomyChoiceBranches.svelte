<script lang="ts">
  import { taxonomyBranchContainer, taxonomyGrammarChoiceDropLabel, taxonomyGrammarChoices, taxonomyGrammarFirstBinding, type TaxonomyDragPayload, type TaxonomyGrammar } from './TaxonomyEditorModel'
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
        <div class="taxonomy-branch-context">{#if !readOnly}<DropZone variant="slot" mode="inactive" label={targetLabel} showLabel container={taxonomyBranchContainer(path, occurrence.label)} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDrop={(state) => onBranchFieldDrop(state, path, occurrence.label)} ariaLabel={targetLabel} />{/if}<slot name="branch" path={path} field={occurrence.label} /></div>
      {:else}
        <div class="taxonomy-branch-context">{#if !readOnly}<DropZone variant="slot" mode="inactive" label={targetLabel} showLabel container={taxonomyBranchContainer(path)} accepts={(state) => state.draggedItem?.kind === 'palette' && state.draggedItem.taxonomyType === 'category'} onDrop={(state) => onBranchDrop(state, path)} ariaLabel={targetLabel} />{/if}<slot name="branch" path={path} /></div>
      {/if}
    {/each}
  {/each}
{/if}

<style>
  .taxonomy-branch-context { display: grid; gap: .35rem; }
</style>
