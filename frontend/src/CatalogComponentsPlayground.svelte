<script lang="ts">
  import FilterWidget from './FilterWidget.svelte'

  const shortOptions = [
    { label: 'Steel', count: 21, selected: true },
    { label: 'Aluminum', count: 8 },
    { label: 'Plastic', count: 4 },
  ]

  const sizeOptions = Array.from({ length: 18 }, (_, index) => {
    const size = (11 + index * 5) / 1000
    return { label: `${size.toFixed(4)} in`, count: 2 + (index % 7), selected: index === 2 }
  })

  const typeOptions = [
    { label: 'Plain plug gauges', count: 18 },
    { label: 'Thread plug gauges', count: 11 },
    { label: 'Air plug gauges', count: 4, disabled: true },
    { label: 'Taper plug gauges', count: 7 },
  ]

  let infoOpen = false
</script>

<section class="catalog-components" aria-labelledby="catalog-components-heading">
  <p class="eyebrow">Catalog components</p>
  <h2 id="catalog-components-heading">Filter widget sketches</h2>
  <p class="muted catalog-components-intro">Hard-coded specimens for the catalog filter rail. These examples are deliberately live only where the interaction helps us evaluate the shape.</p>

  <section class="catalog-component-shell" aria-labelledby="filter-widget-shell-heading">
    <p class="mock-section-label">Filter widget</p>
    <h3 id="filter-widget-shell-heading">Shared shell</h3>
    <p class="muted">The shell owns the title, description, optional information affordance, and the slot for a filter implementation.</p>
    <FilterWidget
      title="Plug gauge size"
      description="Choose one or more sizes to narrow the results."
      info="Size is the nominal diameter checked by the gauge. Values are shown in inches."
    />
  </section>

  <div class="catalog-component-layout">
    <article class="catalog-component-card">
      <p class="mock-section-label">Option list</p>
      <h3>Single column</h3>
      <p class="muted">Short controlled vocabulary with counts and one selected value.</p>
      <section class="filter-group" aria-labelledby="material-heading">
        <div class="filter-heading"><h4 id="material-heading">Material</h4><span class="filter-count">33</span></div>
        <div class="filter-options">
          {#each shortOptions as option}
            <label class:selected={option.selected}><input type="checkbox" checked={option.selected} /> <span>{option.label}</span><small>{option.count}</small></label>
          {/each}
        </div>
      </section>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Option list</p>
      <h3>Two columns</h3>
      <p class="muted">The same widget can use a denser layout for short labels.</p>
      <section class="filter-group" aria-labelledby="type-heading">
        <div class="filter-heading"><h4 id="type-heading">Gauge type</h4><span class="filter-count">40</span></div>
        <div class="filter-options filter-options-two-column">
          {#each typeOptions as option}
            <label class:disabled={option.disabled}><input type="checkbox" disabled={option.disabled} /> <span>{option.label}</span><small>{option.count}</small></label>
          {/each}
        </div>
      </section>
    </article>

    <article class="catalog-component-card catalog-component-card-wide">
      <p class="mock-section-label">Option list</p>
      <h3>Scrollable values</h3>
      <p class="muted">Long engineering vocabularies stay contained inside the filter group.</p>
      <section class="filter-group" aria-labelledby="size-heading">
        <div class="filter-heading"><h4 id="size-heading">Plug gauge size</h4><button class="info-button" type="button" aria-label="More information about plug gauge size" on:click={() => infoOpen = !infoOpen}>i</button></div>
        <div class="filter-options filter-options-scroll">
          {#each sizeOptions as option}
            <label class:selected={option.selected}><input type="checkbox" checked={option.selected} /> <span>{option.label}</span><small>{option.count}</small></label>
          {/each}
        </div>
      </section>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Numeric range</p>
      <h3>Minimum and maximum</h3>
      <p class="muted">Explicit values are preferable to a slider for precise engineering data.</p>
      <section class="filter-group" aria-labelledby="diameter-heading">
        <div class="filter-heading"><h4 id="diameter-heading">Diameter</h4><span class="filter-unit">mm</span></div>
        <div class="range-fields"><label>From<input type="number" placeholder="Min" /></label><label>To<input type="number" placeholder="Max" /></label></div>
      </section>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Information</p>
      <h3>Info disclosure</h3>
      <p class="muted">A small contextual explanation attached to a filter heading.</p>
      <section class="filter-group" aria-labelledby="info-example-heading">
        <div class="filter-heading"><h4 id="info-example-heading">Gauge size</h4><button class="info-button" class:active={infoOpen} type="button" aria-expanded={infoOpen} aria-controls="filter-info" aria-label="More information about gauge size" on:click={() => infoOpen = !infoOpen}>i</button></div>
        {#if infoOpen}<div class="filter-info" id="filter-info">Size is the nominal diameter checked by the gauge. Values are shown in inches.</div>{/if}
      </section>
    </article>
  </div>
</section>

<style>
  .catalog-components-intro { max-width: 42rem; margin: 0 0 1.5rem; font-size: .75rem; }
  .catalog-component-shell { margin-bottom: 1.5rem; padding: 1rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: .3rem; }
  .catalog-component-shell h3 { margin: .2rem 0 .35rem; font: 400 1.15rem Georgia, serif; }
  .catalog-component-shell > .muted { margin: 0 0 1rem; font-size: .7rem; }
  .catalog-component-layout { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .catalog-component-card { min-width: 0; padding: 1rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: .3rem; }
  .catalog-component-card-wide { grid-column: 1 / -1; }
  .catalog-component-card h3 { margin: .2rem 0 .35rem; font: 400 1.15rem Georgia, serif; }
  .catalog-component-card > .muted { margin: 0 0 1rem; font-size: .7rem; }
  .filter-group { padding: .8rem; background: var(--ground); border: 1px solid var(--hairline-light); }
  .filter-heading { display: flex; align-items: center; gap: .4rem; margin-bottom: .55rem; }
  .filter-heading h4 { margin: 0; font: 600 .72rem ui-monospace, monospace; }
  .filter-count, .filter-unit { margin-left: auto; color: var(--muted); font: .62rem ui-monospace, monospace; }
  .filter-options { display: grid; gap: .35rem; }
  .filter-options-two-column { grid-template-columns: repeat(2, minmax(0, 1fr)); column-gap: 1rem; }
  .filter-options-scroll { max-height: 10rem; overflow-y: auto; padding-right: .5rem; }
  .filter-options label { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: .4rem; color: var(--muted); font-size: .7rem; }
  .filter-options label.selected { color: var(--ink); }
  .filter-options label.disabled { opacity: .45; }
  .filter-options input { width: .9rem; height: .9rem; margin: 0; accent-color: var(--accent); }
  .filter-options small { font: .6rem ui-monospace, monospace; }
  .info-button { display: inline-grid; width: 1rem; height: 1rem; place-items: center; padding: 0; color: var(--muted); background: transparent; border: 1px solid currentColor; border-radius: 50%; cursor: pointer; font: 600 .6rem Georgia, serif; }
  .info-button:hover, .info-button.active { color: var(--accent); }
  .range-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .6rem; }
  .range-fields label { display: grid; gap: .25rem; color: var(--muted); font-size: .65rem; }
  .range-fields input { min-width: 0; width: 100%; padding: .45rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .7rem ui-monospace, monospace; }
  .filter-info { padding: .65rem; color: var(--muted); background: var(--surface); border-left: 2px solid var(--accent); font-size: .68rem; line-height: 1.45; }
  @media (max-width: 700px) { .catalog-component-layout { grid-template-columns: 1fr; } .catalog-component-card-wide { grid-column: auto; } }
</style>
