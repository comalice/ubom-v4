<script context="module" lang="ts">
  export type { TaxonomyRulePredicate } from './TaxonomyEditorModel'
</script>

<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import type { TaxonomyDragPayload, TaxonomyRulePredicate } from './TaxonomyEditorModel'

  export let title: string
  export let summary: string
  export let predicate: TaxonomyRulePredicate
  export let container = ''
  export let dragData: TaxonomyDragPayload | undefined = undefined
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

<BlockShell title={title} summary={summary} kind="taxonomy" collapsible initiallyOpen={true} {container} {dragData} {onDragEnd} {onRemove}>
  <div class="taxonomy-block-body">
    <label>Grammar field<select value={predicate.field} aria-label="Rule grammar field"><option>category</option><option>number</option></select></label>
    <label>Match type<select value={predicate.kind} aria-label="Rule match type"><option value="exact">Exact value</option><option value="numeric-range">Numeric range</option></select></label>
    {#if predicate.kind === 'exact'}
      <label>Value<input value={predicate.value} aria-label="Rule value" /></label>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{predicate.field} = {predicate.value}</strong></div>
    {:else}
      <label>Range<input value={`${predicate.min}..${predicate.max}`} aria-label="Rule numeric range" /></label>
      <div class="taxonomy-block-rule"><span>Predicate</span><strong>{predicate.field} in {predicate.min}..{predicate.max}</strong></div>
    {/if}
  </div>
</BlockShell>
