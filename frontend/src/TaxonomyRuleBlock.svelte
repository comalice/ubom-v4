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
  export let onPredicateChange: (predicate: TaxonomyRulePredicate) => void = () => {}
  export let readOnly = false
  let matchType: TaxonomyRulePredicate['kind'] = predicate.kind
  let exactValue = predicate.value ?? ''
  let rangeMin = predicate.min ?? 0
  let rangeMax = predicate.max ?? 0
  $: effectiveField = predicate.field || field
  $: ruleSummary = matchType === 'exact' ? `${effectiveField} = ${exactValue}` : `${effectiveField} in ${rangeMin}..${rangeMax}`
  function emitPredicate() {
    if (matchType === 'numeric-range' && (!Number.isFinite(rangeMin) || !Number.isFinite(rangeMax))) return
    onPredicateChange(matchType === 'exact' ? { kind: 'exact', field: effectiveField, value: exactValue } : { kind: 'numeric-range', field: effectiveField, min: rangeMin, max: rangeMax })
  }
</script>

  <BlockShell title={title} summary={ruleSummary} kind="taxonomy" collapsible initiallyOpen={true} container={readOnly || dragData?.kind === 'taxonomy' ? '' : container} dragData={readOnly || dragData?.kind === 'taxonomy' ? undefined : dragData} showDragHandle={dragData?.kind !== 'taxonomy'} dragHandle=".block-shell-drag" interactive={['input', 'select', 'button']} {onDragEnd} onRemove={readOnly ? undefined : onRemove}>
  <div class="taxonomy-block-body">
   <label>Match type<select bind:value={matchType} disabled={readOnly} on:change={emitPredicate} aria-label="Rule match type"><option value="exact">Exact value</option><option value="numeric-range">Numeric range</option></select></label>
    {#if matchType === 'exact'}
       <label>Value<input bind:value={exactValue} readonly={readOnly} on:input={emitPredicate} aria-label="Rule value" /></label>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{ruleSummary}</strong></div>
    {:else}
       <div class="taxonomy-rule-range"><label>Minimum<input type="number" bind:value={rangeMin} readonly={readOnly} on:input={emitPredicate} aria-label="Rule minimum" /></label><label>Maximum<input type="number" bind:value={rangeMax} readonly={readOnly} on:input={emitPredicate} aria-label="Rule maximum" /></label></div>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{ruleSummary}</strong></div>
    {/if}
  </div>
</BlockShell>
