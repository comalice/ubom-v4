<script lang="ts">
  import { taxonomyGrammarBranchLabel, taxonomyGrammarChoices, type TaxonomyGrammar } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let grammar: TaxonomyGrammar
  export let targetId = ''

  $: choices = taxonomyGrammarChoices(grammar, targetId)

  function branchPath(alternative: StructureNode): string {
    return alternative.branchPath ?? alternative.path ?? String(alternative.id)
  }

</script>

{#if choices.length}
  {#each choices as choice (choice.id)}
    {#each choice.children as alternative (alternative.id)}
      {@const path = branchPath(alternative)}
      {@const branchLabel = taxonomyGrammarBranchLabel(alternative)}
      <div class="taxonomy-branch-context"><slot name="branch" path={path} branchLabel={branchLabel} /></div>
    {/each}
  {/each}
{/if}

<style>
  .taxonomy-branch-context { display: grid; gap: .35rem; min-width: 0; }
</style>
