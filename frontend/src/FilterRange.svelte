<script lang="ts">
  export let unit = ''
  export let minimum: string | undefined = undefined
  export let maximum: string | undefined = undefined
  export let onChange: (range: { minimum: string; maximum: string }) => void = () => {}
  let localMinimum = ''
  let localMaximum = ''

  $: activeMinimum = minimum ?? localMinimum
  $: activeMaximum = maximum ?? localMaximum

  function updateMinimum(event: Event) {
    const nextMinimum = (event.currentTarget as HTMLInputElement).value
    if (minimum === undefined) localMinimum = nextMinimum
    onChange({ minimum: nextMinimum, maximum: activeMaximum })
  }

  function updateMaximum(event: Event) {
    const nextMaximum = (event.currentTarget as HTMLInputElement).value
    if (maximum === undefined) localMaximum = nextMaximum
    onChange({ minimum: activeMinimum, maximum: nextMaximum })
  }
</script>

<div class="filter-range">
  <div class="filter-range-heading">{#if unit}<span>{unit}</span>{/if}</div>
  <div class="filter-range-fields">
    <label>From<input type="number" placeholder="Min" value={activeMinimum} on:input={updateMinimum} /></label>
    <label>To<input type="number" placeholder="Max" value={activeMaximum} on:input={updateMaximum} /></label>
  </div>
</div>

<style>
  .filter-range-heading { min-height: .1rem; color: var(--muted); text-align: right; font: .62rem ui-monospace, monospace; }
  .filter-range-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .6rem; }
  .filter-range-fields label { display: grid; gap: .25rem; color: var(--muted); font-size: .65rem; }
  .filter-range-fields input { min-width: 0; width: 100%; padding: .45rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .7rem ui-monospace, monospace; }
</style>
