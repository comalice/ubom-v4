<script lang="ts">
  import { acceptsTaxonomyBranchDrop, taxonomyBranchContainer, taxonomyGrammarChoiceDropLabel, taxonomyGrammarChoices, taxonomyGrammarFirstBinding, type TaxonomyDragPayload, type TaxonomyGrammar } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'
  import type { DragDropState } from '@thisux/sveltednd'
  import DropZone from './DropZone.svelte'

  export let grammar: TaxonomyGrammar
  export let targetId = ''
  export let readOnly = false
  export let onBranchDrop: (state: DragDropState<TaxonomyDragPayload>, branchPath: string, targetIndex?: number) => void = () => {}
  export let onBranchFieldDrop: (state: DragDropState<TaxonomyDragPayload>, branchPath: string, field: string, targetIndex?: number) => void = () => {}

  $: choices = taxonomyGrammarChoices(grammar, targetId)

  function branchPath(alternative: StructureNode): string {
    return alternative.branchPath ?? alternative.path ?? String(alternative.id)
  }

  function getMode(state: DragDropState<TaxonomyDragPayload>) {
    return state.draggedItem?.kind === 'palette' ? 'new' as const : 'inactive' as const
  }

</script>

{#if choices.length}
  {#each choices as choice (choice.id)}
    {#each choice.children as alternative (alternative.id)}
      {@const occurrence = taxonomyGrammarFirstBinding(alternative.children)}
      {@const path = branchPath(alternative)}
      {@const targetLabel = taxonomyGrammarChoiceDropLabel(alternative, occurrence?.label)}
      {#if occurrence}
        <div class="taxonomy-branch-context">{#if !readOnly}<DropZone variant="slot" mode="inactive" className="taxonomy-branch-drop-zone" label={targetLabel} showLabel {getMode} container={taxonomyBranchContainer(path, occurrence.label)} accepts={(state) => acceptsTaxonomyBranchDrop(state.draggedItem)} onDrop={(state) => onBranchFieldDrop(state, path, occurrence.label)} ariaLabel={targetLabel} />{/if}<slot name="branch" path={path} field={occurrence.label} /></div>
      {:else}
        <div class="taxonomy-branch-context">{#if !readOnly}<DropZone variant="slot" mode="inactive" className="taxonomy-branch-drop-zone" label={targetLabel} showLabel {getMode} container={taxonomyBranchContainer(path)} accepts={(state) => acceptsTaxonomyBranchDrop(state.draggedItem)} onDrop={(state) => onBranchDrop(state, path)} ariaLabel={targetLabel} />{/if}<slot name="branch" path={path} /></div>
      {/if}
    {/each}
  {/each}
{/if}

<style>
  .taxonomy-branch-context { display: grid; gap: .35rem; min-width: 0; }
</style>
