<script lang="ts">
  import { taxonomyGrammarAlternatives, type TaxonomyGrammar } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let grammar: TaxonomyGrammar

  $: alternatives = taxonomyGrammarAlternatives(grammar)

  function bindings(nodes: StructureNode[]): StructureNode[] {
    return nodes.flatMap((node) => [
      ...(node.blockType === 'binding' ? [node] : []),
      ...bindings(node.children),
    ])
  }

  function branchPath(alternative: StructureNode): string {
    return [...(alternative.choiceAncestry ?? []).map((choice) => choice.path), alternative.path ?? String(alternative.id)].join(' / ')
  }
</script>

{#if alternatives.length}
  <section class="taxonomy-choice-branches" aria-label="Sequence definition choices">
    <div class="taxonomy-choice-branches-heading">
      <span class="mock-section-label">Sequence choices</span>
      <small>Each path is frontend display metadata; taxonomy predicates remain field-based.</small>
    </div>
    {#each alternatives as alternative (alternative.id)}
      {@const occurrences = bindings(alternative.children)}
      <article class="taxonomy-choice-branch" class:taxonomy-choice-literal-only={alternative.literalOnly}>
        <header>
          <strong>{alternative.label}</strong>
          <span>{alternative.summary}</span>
        </header>
        <small class="taxonomy-choice-path">path: {branchPath(alternative)}</small>
        {#if alternative.choiceAncestry?.length}
          <small class="taxonomy-choice-ancestry">ancestry: {alternative.choiceAncestry.map((choice) => `${choice.path} (${choice.summary})`).join(' -> ')}</small>
        {/if}
        {#if occurrences.length}
          <div class="taxonomy-choice-bindings" aria-label={`Captured bindings for ${alternative.label}`}>
            {#each occurrences as occurrence (occurrence.path ?? occurrence.id)}
              <span class="taxonomy-choice-binding">
                <b>{occurrence.label}</b>
                <small>{occurrence.path}</small>
                {#if occurrence.numericDomain}<small>domain {occurrence.numericDomain.min}..{occurrence.numericDomain.max}</small>{/if}
              </span>
            {/each}
          </div>
        {:else}
          <small class="taxonomy-choice-nontargetable">literal-only branch - no captured field target</small>
        {/if}
      </article>
    {/each}
  </section>
{/if}

<style>
  .taxonomy-choice-branches { display: grid; gap: .55rem; padding: .7rem; border: 1px solid var(--hairline); border-radius: .25rem; }
  .taxonomy-choice-branches-heading, .taxonomy-choice-branch, .taxonomy-choice-bindings, .taxonomy-choice-binding { display: grid; gap: .2rem; }
  .taxonomy-choice-branches-heading small, .taxonomy-choice-branch > small, .taxonomy-choice-binding small { color: var(--muted); font-size: .68rem; }
  .taxonomy-choice-branch { padding: .55rem; background: var(--ground); border-left: .2rem solid var(--accent); }
  .taxonomy-choice-literal-only { border-left-color: var(--muted); }
  .taxonomy-choice-branch header { display: grid; gap: .15rem; }
  .taxonomy-choice-branch header span { color: var(--muted); font-size: .78rem; }
  .taxonomy-choice-path, .taxonomy-choice-ancestry, .taxonomy-choice-binding small { font-family: ui-monospace, monospace; overflow-wrap: anywhere; }
  .taxonomy-choice-bindings { grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr)); margin-top: .2rem; }
  .taxonomy-choice-binding { padding: .35rem; border: 1px solid var(--hairline); }
  .taxonomy-choice-nontargetable { color: var(--muted); font-style: italic; }
</style>
