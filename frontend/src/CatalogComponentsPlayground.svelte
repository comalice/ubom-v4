<script lang="ts">
  import FilterWidget from './FilterWidget.svelte'
  import FilterOptionList, { type FilterOption } from './FilterOptionList.svelte'
  import FilterRange from './FilterRange.svelte'
  import CategoryList, { type CategoryListItem } from './CategoryList.svelte'

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

  const categories: CategoryListItem[] = [
    { label: 'Abrading & Polishing', count: 24 },
    { label: 'Electrical & Lighting', count: 42 },
    { label: 'Fastening & Joining', count: 67, selected: true },
    { label: 'Measuring & Inspecting', count: 31 },
    { label: 'Pipe, Tubing, Hose & Fittings', count: 19 },
    { label: 'Raw Materials', count: 14 },
  ]

</script>

<section class="catalog-components" aria-labelledby="catalog-components-heading">
  <p class="eyebrow">Catalog components</p>
  <h2 id="catalog-components-heading">Filter widget sketches</h2>
  <p class="muted catalog-components-intro">Hard-coded specimens for the catalog filter rail. These examples are deliberately live only where the interaction helps us evaluate the shape.</p>

  <section class="catalog-components-group" aria-labelledby="catalog-navigation-heading">
    <div class="catalog-components-group-heading">
      <p class="mock-section-label">Catalog navigation</p>
      <h3 id="catalog-navigation-heading">Navigation specimen</h3>
      <p class="muted">A compact left-rail list for the current catalog scope.</p>
    </div>
    <article class="catalog-category-list-frame" aria-labelledby="category-list-heading">
      <h4 id="category-list-heading">Categories List</h4>
      <CategoryList categories={categories} />
    </article>
  </section>

  <section class="catalog-components-group" aria-labelledby="filter-shell-heading">
    <div class="catalog-components-group-heading">
      <p class="mock-section-label">Filter widget</p>
      <h3 id="filter-shell-heading">Shared shell</h3>
      <p class="muted">The shell owns the title, description, optional information affordance, and the slot for a filter implementation.</p>
    </div>
    <div class="catalog-component-layout catalog-shell-layout">
      <article class="catalog-component-card">
        <h4>Base shell</h4>
        <FilterWidget title="Plug gauge size" description="Choose one or more sizes to narrow the results." info="Size is the nominal diameter checked by the gauge. Values are shown in inches." />
      </article>

      <article class="catalog-component-card">
        <p class="mock-section-label">Information</p>
        <h4>Info disclosure</h4>
        <p class="muted">A small contextual explanation attached to a filter heading.</p>
        <FilterWidget title="Gauge size" description="An optional explanation can be opened from the title." info="Size is the nominal diameter checked by the gauge. Values are shown in inches." />
      </article>
    </div>
  </section>

  <section class="catalog-components-group" aria-labelledby="filter-controls-heading">
    <div class="catalog-components-group-heading">
      <p class="mock-section-label">Filter controls</p>
      <h3 id="filter-controls-heading">Value specimens</h3>
      <p class="muted">The same shell supports compact choices, dense values, scrolling lists, and precise numeric input.</p>
    </div>
    <div class="catalog-component-layout">
      <article class="catalog-component-card">
        <p class="mock-section-label">Option list</p>
        <h4>Single column</h4>
        <p class="muted">Short controlled vocabulary with counts and one selected value.</p>
        <FilterWidget title="Material" description="Choose one or more materials.">
          <FilterOptionList options={shortOptions} />
        </FilterWidget>
      </article>

      <article class="catalog-component-card">
        <p class="mock-section-label">Option list</p>
        <h4>Two columns</h4>
        <p class="muted">The same widget can use a denser layout for short labels.</p>
        <FilterWidget title="Gauge type" description="Choose one or more gauge types.">
          <FilterOptionList options={typeOptions} columns={2} />
        </FilterWidget>
      </article>

      <article class="catalog-component-card catalog-component-card-wide">
        <p class="mock-section-label">Option list</p>
        <h4>Scrollable values</h4>
        <p class="muted">Long engineering vocabularies stay contained inside the filter group.</p>
        <FilterWidget title="Plug gauge size" description="Choose one or more sizes to narrow the results." info="Size is the nominal diameter checked by the gauge. Values are shown in inches.">
          <FilterOptionList options={sizeOptions} scrollable />
        </FilterWidget>
      </article>

      <article class="catalog-component-card">
        <p class="mock-section-label">Numeric range</p>
        <h4>Minimum and maximum</h4>
        <p class="muted">Explicit values are preferable to a slider for precise engineering data.</p>
        <FilterWidget title="Diameter" description="Enter an inclusive range in millimeters.">
          <FilterRange unit="mm" />
        </FilterWidget>
      </article>
    </div>
  </section>
</section>

<style>
  .catalog-components-intro { max-width: 42rem; margin: 0 0 1.5rem; font-size: .75rem; }
  .catalog-components-group { margin-bottom: 2rem; }
  .catalog-components-group-heading { max-width: 42rem; margin-bottom: 1rem; }
  .catalog-components-group-heading h3 { margin: .2rem 0 .35rem; font: 400 1.15rem Georgia, serif; }
  .catalog-components-group-heading > .muted { margin: 0; font-size: .7rem; }
  .catalog-category-list-frame { max-width: 24rem; padding: 1rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: .3rem; }
  .catalog-category-list-frame h4, .catalog-component-card h4 { margin: 0 0 .75rem; font: 600 .72rem ui-monospace, monospace; }
  .catalog-component-layout { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; }
  .catalog-component-card { min-width: 0; padding: 1rem; background: var(--surface); border: 1px solid var(--hairline); border-radius: .3rem; }
  .catalog-component-card-wide { grid-column: 1 / -1; }
  .catalog-component-card > .muted { margin: 0 0 1rem; font-size: .7rem; }
  @media (max-width: 700px) { .catalog-component-layout { grid-template-columns: 1fr; } .catalog-component-card-wide { grid-column: auto; } }
</style>
