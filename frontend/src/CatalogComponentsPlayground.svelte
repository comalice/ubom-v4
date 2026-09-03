<script lang="ts">
  import FilterWidget from './FilterWidget.svelte'
  import FilterOptionList, { type FilterOption } from './FilterOptionList.svelte'
  import FilterRange from './FilterRange.svelte'

  const shortOptions: FilterOption[] = [
    { label: 'Steel', count: 21, selected: true },
    { label: 'Aluminum', count: 8 },
    { label: 'Plastic', count: 4 },
  ]

  const sizeOptions: FilterOption[] = Array.from({ length: 18 }, (_, index) => {
    const size = (11 + index * 5) / 1000
    return { label: `${size.toFixed(4)} in`, count: 2 + (index % 7), selected: index === 2 }
  })

  const typeOptions: FilterOption[] = [
    { label: 'Plain plug gauges', count: 18 },
    { label: 'Thread plug gauges', count: 11 },
    { label: 'Air plug gauges', count: 4, disabled: true },
    { label: 'Taper plug gauges', count: 7 },
  ]

</script>

<section class="catalog-components" aria-labelledby="catalog-components-heading">
  <p class="eyebrow">Catalog components</p>
  <h2 id="catalog-components-heading">Filter widget sketches</h2>
  <p class="muted catalog-components-intro">Hard-coded specimens for the catalog filter rail. These examples are deliberately live only where the interaction helps us evaluate the shape.</p>

  <section class="catalog-component-shell" aria-labelledby="filter-widget-shell-heading">
    <p class="mock-section-label">Filter widget</p>
    <h3 id="filter-widget-shell-heading">Shared shell</h3>
    <p class="muted">The shell owns the title, description, optional information affordance, and the slot for a filter implementation.</p>
    <FilterWidget title="Plug gauge size" description="Choose one or more sizes to narrow the results." info="Size is the nominal diameter checked by the gauge. Values are shown in inches." />
  </section>

  <div class="catalog-component-layout">
    <article class="catalog-component-card">
      <p class="mock-section-label">Option list</p>
      <h3>Single column</h3>
      <p class="muted">Short controlled vocabulary with counts and one selected value.</p>
      <FilterWidget title="Material" description="Choose one or more materials.">
        <FilterOptionList options={shortOptions} />
      </FilterWidget>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Option list</p>
      <h3>Two columns</h3>
      <p class="muted">The same widget can use a denser layout for short labels.</p>
      <FilterWidget title="Gauge type" description="Choose one or more gauge types.">
        <FilterOptionList options={typeOptions} columns={2} />
      </FilterWidget>
    </article>

    <article class="catalog-component-card catalog-component-card-wide">
      <p class="mock-section-label">Option list</p>
      <h3>Scrollable values</h3>
      <p class="muted">Long engineering vocabularies stay contained inside the filter group.</p>
      <FilterWidget title="Plug gauge size" description="Choose one or more sizes to narrow the results." info="Size is the nominal diameter checked by the gauge. Values are shown in inches.">
        <FilterOptionList options={sizeOptions} scrollable />
      </FilterWidget>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Numeric range</p>
      <h3>Minimum and maximum</h3>
      <p class="muted">Explicit values are preferable to a slider for precise engineering data.</p>
      <FilterWidget title="Diameter" description="Enter an inclusive range in millimeters.">
        <FilterRange unit="mm" />
      </FilterWidget>
    </article>

    <article class="catalog-component-card">
      <p class="mock-section-label">Information</p>
      <h3>Info disclosure</h3>
      <p class="muted">A small contextual explanation attached to a filter heading.</p>
      <FilterWidget title="Gauge size" description="An optional explanation can be opened from the title." info="Size is the nominal diameter checked by the gauge. Values are shown in inches." />
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
  @media (max-width: 700px) { .catalog-component-layout { grid-template-columns: 1fr; } .catalog-component-card-wide { grid-column: auto; } }
</style>
