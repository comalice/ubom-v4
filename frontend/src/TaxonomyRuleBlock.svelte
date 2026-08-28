<script context="module" lang="ts">
  export type { TaxonomyRulePredicate } from './TaxonomyEditorModel'
</script>

<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import type { TaxonomyDragPayload, TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let title: string
  export let field: string
  export let predicate: TaxonomyRulePredicate
  export let container = ''
  export let dragData: TaxonomyDragPayload | undefined = undefined
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
  let matchType: TaxonomyRulePredicate['kind'] = predicate.kind
  let exactValue = predicate.value ?? ''
  let rangeMin = predicate.min ?? 0
  let rangeMax = predicate.max ?? 0
  $: ruleSummary = matchType === 'exact' ? `${field} = ${exactValue}` : `${field} in ${rangeMin}..${rangeMax}`
</script>

  <BlockShell title={title} summary={ruleSummary} kind="taxonomy" collapsible initiallyOpen={true} {container} {dragData} dragHandle=".block-shell-drag" interactive={['input', 'select', 'button']} {onDragEnd} {onRemove}>
  <div class="taxonomy-block-body">
    <label>Match type<select bind:value={matchType} aria-label="Rule match type"><option value="exact">Exact value</option><option value="numeric-range">Numeric range</option></select></label>
    {#if matchType === 'exact'}
      <label>Value<input bind:value={exactValue} aria-label="Rule value" /></label>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{ruleSummary}</strong></div>
    {:else}
      <div class="taxonomy-rule-range"><label>Minimum<input type="number" bind:value={rangeMin} aria-label="Rule minimum" /></label><label>Maximum<input type="number" bind:value={rangeMax} aria-label="Rule maximum" /></label></div>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{ruleSummary}</strong></div>
    {/if}
  </div>
</BlockShell>
