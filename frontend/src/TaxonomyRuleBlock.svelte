<script context="module" lang="ts">
  export type TaxonomyRulePredicate =
    | { kind: 'exact'; field: string; value: string }
    | { kind: 'numeric-range'; field: string; min: number; max: number }
</script>

<script lang="ts">
  import BlockShell from './BlockShell.svelte'

  export let title: string
  export let summary: string
  export let predicate: TaxonomyRulePredicate
</script>

<BlockShell title={title} summary={summary} kind="taxonomy" collapsible initiallyOpen={true} onRemove={() => {}}>
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
