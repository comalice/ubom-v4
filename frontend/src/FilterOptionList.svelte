<script lang="ts">
  export type FilterOption = {
    label: string
    value?: string
    count?: number
    selected?: boolean
    disabled?: boolean
  }

  export let options: FilterOption[] = []
  export let columns: 1 | 2 = 1
  export let scrollable = false
  export let selectedValues: string[] | undefined = undefined
  export let onChange: (values: string[]) => void = () => {}

  let localSelectedValues = options.filter(option => option.selected).map(option => option.value ?? option.label)

  $: activeSelectedValues = selectedValues ?? localSelectedValues

  function toggle(option: FilterOption) {
    const value = option.value ?? option.label
    const nextValues = activeSelectedValues.includes(value)
      ? activeSelectedValues.filter(selected => selected !== value)
      : [...activeSelectedValues, value]
    if (selectedValues === undefined) localSelectedValues = nextValues
    onChange(nextValues)
  }
</script>

<div class:filter-options-two-column={columns === 2} class:filter-options-scroll={scrollable} class="filter-option-list">
  {#each options as option}
    {@const value = option.value ?? option.label}
    <label class:selected={activeSelectedValues.includes(value)} class:disabled={option.disabled}>
      <input type="checkbox" checked={activeSelectedValues.includes(value)} disabled={option.disabled} on:change={() => toggle(option)} />
      <span>{option.label}</span>
      {#if option.count !== undefined}<small>{option.count}</small>{/if}
    </label>
  {/each}
</div>

<style>
  .filter-option-list { display: grid; gap: .35rem; }
  .filter-options-two-column { grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 1rem; }
  .filter-options-scroll { max-height: 10rem; overflow-y: auto; padding-right: .5rem; }
  .filter-option-list label { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: .4rem; color: var(--muted); font-size: .7rem; }
  .filter-option-list label.selected { color: var(--ink); }
  .filter-option-list label.disabled { opacity: .45; }
  .filter-option-list input { width: .9rem; height: .9rem; margin: 0; accent-color: var(--accent); }
  .filter-option-list small { font: .6rem ui-monospace, monospace; }
</style>
