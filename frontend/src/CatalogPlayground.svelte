<script lang="ts">
  import Breadcrumbs from './Breadcrumbs.svelte'
  import CatalogBrowser from './CatalogBrowser.svelte'
  import CategoryList, { type CategoryListItem } from './CategoryList.svelte'
  import CategoryTileGrid from './CategoryTileGrid.svelte'
  import FilterOptionList, { type FilterOption } from './FilterOptionList.svelte'
  import FilterRange from './FilterRange.svelte'
  import FilterWidget from './FilterWidget.svelte'
  import PartTable from './PartTable.svelte'
  import type { PartListItem } from './api'

  const categories: CategoryListItem[] = [
    { id: 'electrical-lighting', label: 'Electrical & Lighting', href: '/parts/electrical-and-lighting', count: 42 },
    { id: 'fastening-joining', label: 'Fastening & Joining', href: '/parts/fastening-and-joining', count: 67, selected: true },
    { id: 'measuring-inspecting', label: 'Measuring & Inspecting', href: '/parts/measuring-and-inspecting', count: 31 },
    { id: 'mechanical-components', label: 'Mechanical Components', href: '/parts/mechanical-components', count: 54 },
    { id: 'raw-materials', label: 'Raw Materials', href: '/parts/raw-materials', count: 14 },
    { id: 'safety-supplies', label: 'Safety Supplies', href: '/parts/safety-supplies', count: 28 },
  ]

  const materials: FilterOption[] = [
    { label: 'Steel', count: 21 },
    { label: 'Aluminum', count: 8 },
    { label: 'Plastic', count: 4 },
  ]

  const gaugeTypes: FilterOption[] = [
    { label: 'Plain plug gauges', count: 18 },
    { label: 'Thread plug gauges', count: 11 },
    { label: 'Air plug gauges', count: 4, disabled: true },
    { label: 'Taper plug gauges', count: 7 },
  ]

  let selectedMaterials = ['Steel']
  let selectedGaugeTypes: string[] = []
  let diameterMinimum = ''
  let diameterMaximum = ''

  $: activeFilterLabels = [
    ...selectedMaterials.map(value => `Material: ${value}`),
    ...selectedGaugeTypes.map(value => `Gauge type: ${value}`),
    ...(diameterMinimum ? [`Diameter from ${diameterMinimum} mm`] : []),
    ...(diameterMaximum ? [`Diameter to ${diameterMaximum} mm`] : []),
  ]

  function clearFilters() {
    selectedMaterials = []
    selectedGaugeTypes = []
    diameterMinimum = ''
    diameterMaximum = ''
  }

  function updateDiameter(range: { minimum: string; maximum: string }) {
    diameterMinimum = range.minimum
    diameterMaximum = range.maximum
  }

  function removeFilter(filter: string) {
    if (filter.startsWith('Material: ')) selectedMaterials = selectedMaterials.filter(value => `Material: ${value}` !== filter)
    else if (filter.startsWith('Gauge type: ')) selectedGaugeTypes = selectedGaugeTypes.filter(value => `Gauge type: ${value}` !== filter)
    else if (filter.startsWith('Diameter from ')) diameterMinimum = ''
    else if (filter.startsWith('Diameter to ')) diameterMaximum = ''
  }

  const fasteningTiles = [
    { id: 'bolts', title: 'Bolts', href: '/parts/bolts', description: 'Machine bolts and structural fasteners.', image: 'IMG', partCount: 48 },
    { id: 'screws', title: 'Screws', href: '/parts/screws', description: 'Thread-forming and machine screws.', image: 'IMG', partCount: 82 },
    { id: 'nuts', title: 'Nuts', href: '/parts/nuts', description: 'Hex, lock, wing, and specialty nuts.', image: 'IMG', partCount: 35 },
    { id: 'washers', title: 'Washers', href: '/parts/washers', description: 'Spacers, shims, and load distribution.', partCount: 26 },
    { id: 'threaded-rod', title: 'Threaded Rod', href: '/parts/threaded-rod', description: 'Continuous-thread rod and studs.', partCount: 19 },
    { id: 'rivets', title: 'Rivets', href: '/parts/rivets', description: 'Permanent joining hardware.', partCount: 31 },
  ]

  const adhesiveTiles = [
    { id: 'adhesives', title: 'Adhesives', href: '/parts/adhesives', description: 'Structural and general-purpose bonding.', image: 'IMG', partCount: 44 },
    { id: 'adhesive-tape', title: 'Adhesive Tape', href: '/parts/adhesive-tape', description: 'Double-sided, transfer, and mounting tape.', image: 'IMG', partCount: 29 },
    { id: 'threadlockers', title: 'Threadlockers', href: '/parts/threadlockers', description: 'Threadlocking compounds and sealants.', partCount: 17 },
  ]

  const weldingTiles = [
    { id: 'welding', title: 'Welding', href: '/parts/welding', description: 'Welding equipment and consumables.', image: 'IMG', partCount: 57 },
    { id: 'brazing', title: 'Brazing', href: '/parts/brazing', description: 'Brazing rods, alloys, and fluxes.', partCount: 22 },
    { id: 'soldering', title: 'Soldering', href: '/parts/soldering', description: 'Solder, tools, and accessories.', partCount: 38 },
  ]

  const parts: PartListItem[] = [
    { id: 'part-200', value: 'PN-BOLT-M6-001', taxonomyPath: ['Fastening & Joining', 'Bolts'], status: 'active', attributes: [{ id: 'material', label: 'Material', value: 'Steel', valueType: 'enum' }], revisions: [{ id: 'revision-200-a', revision: 'A' }] },
    { id: 'part-201', value: 'PN-BOLT-M8-002', taxonomyPath: ['Fastening & Joining', 'Bolts'], status: 'active', attributes: [{ id: 'material', label: 'Material', value: 'Steel', valueType: 'enum' }], revisions: [{ id: 'revision-201-a', revision: 'A' }] },
  ]
</script>

<CatalogBrowser title="Browse parts" subtitle="Navigate by category, then narrow the current scope with relevant attributes.">
  <form class="catalog-search" slot="search" on:submit|preventDefault>
    <label for="playground-catalog-search-input">Search parts</label>
    <div class="catalog-search-controls">
      <input id="playground-catalog-search-input" placeholder="Part number or description" />
      <button type="submit">Search</button>
    </div>
  </form>
  <svelte:fragment slot="sidebar">
    <div class="catalog-panel-content">
      <section class="catalog-panel-section" aria-labelledby="catalog-categories-heading">
        <p class="catalog-panel-kicker">Browse catalog</p>
        <h2 id="catalog-categories-heading">Categories</h2>
        <CategoryList categories={categories} />
      </section>
      <section class="catalog-panel-section" aria-labelledby="catalog-filters-heading">
        <div class="catalog-panel-heading"><h2 id="catalog-filters-heading">Filter by</h2><button type="button" on:click={clearFilters}>Clear</button></div>
        <FilterWidget title="Material" description="">
          <FilterOptionList options={materials} selectedValues={selectedMaterials} onChange={values => selectedMaterials = values} />
        </FilterWidget>
        <FilterWidget title="Gauge type" description="">
          <FilterOptionList options={gaugeTypes} columns={2} selectedValues={selectedGaugeTypes} onChange={values => selectedGaugeTypes = values} />
        </FilterWidget>
        <FilterWidget title="Diameter" description="">
          <FilterRange unit="mm" minimum={diameterMinimum} maximum={diameterMaximum} onChange={updateDiameter} />
        </FilterWidget>
      </section>
    </div>

  </svelte:fragment>
    <div class="catalog-main">
      <Breadcrumbs items={[{ label: 'Home', href: '/' }, { label: 'Catalog', href: '/catalog' }, { label: 'Fastening & Joining' }]} />
      <div class="catalog-main-heading"><div><p class="eyebrow">Category</p><h2>Fastening &amp; Joining</h2></div><span class="catalog-result-count">67 parts</span></div>
      {#if activeFilterLabels.length > 0}
        <div class="catalog-active-filters" aria-label="Active filters"><span class="catalog-active-filters-label">Active filters</span>{#each activeFilterLabels as filter}<span class="catalog-filter-chip">{filter}<button type="button" aria-label={`Remove ${filter}`} on:click={() => removeFilter(filter)}>x</button></span>{/each}</div>
      {/if}
      <section class="catalog-category-section" aria-labelledby="fastening-heading">
        <div class="catalog-section-heading"><h3 id="fastening-heading">Fasteners</h3><span>6 categories</span></div>
        <CategoryTileGrid tiles={fasteningTiles} />
      </section>
      <section class="catalog-category-section" aria-labelledby="adhesive-heading">
        <div class="catalog-section-heading"><h3 id="adhesive-heading">Adhesives &amp; Tape</h3><span>3 categories</span></div>
        <CategoryTileGrid tiles={adhesiveTiles} />
      </section>
      <section class="catalog-category-section" aria-labelledby="welding-heading">
        <div class="catalog-section-heading"><h3 id="welding-heading">Welding, Brazing &amp; Soldering</h3><span>3 categories</span></div>
        <CategoryTileGrid tiles={weldingTiles} />
      </section>
      <section class="catalog-parts-section" aria-labelledby="catalog-parts-heading">
        <div class="catalog-section-heading"><h3 id="catalog-parts-heading">Parts in this category</h3><span>2 shown</span></div>
        <PartTable {parts} columns={[{ id: 'material', label: 'Material' }]} seqDef="part-number-v1" />
      </section>
    </div>
</CatalogBrowser>

<style>
  .catalog-search { display: grid; flex: 0 1 20rem; gap: .3rem; color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-search-controls { display: flex; min-width: 0; gap: .45rem; }
  .catalog-search input { min-width: 0; flex: 1; padding: .6rem .7rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .75rem Inter, sans-serif; }
  .catalog-search input:focus { border-color: var(--accent); outline: none; }
  .catalog-search button { flex: none; padding: .6rem .8rem; color: var(--surface); background: var(--accent); border: 0; border-radius: .2rem; cursor: pointer; font: .7rem ui-monospace, monospace; }
  .catalog-search button:hover, .catalog-search button:focus-visible { background: var(--ink); outline: none; }
  .catalog-panel-content { min-width: 0; }
  .catalog-panel-section { padding: 1rem; }
  .catalog-panel-section + .catalog-panel-section { border-top: 1px solid var(--hairline); }
  .catalog-panel-kicker { margin: 0 0 .35rem; color: var(--accent); font: 600 .65rem ui-monospace, monospace; letter-spacing: .1em; text-transform: uppercase; }
  .catalog-panel-content h2 { margin: 0 0 .7rem; font: 400 1.15rem Georgia, serif; }
  .catalog-panel-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .5rem; }
  .catalog-panel-heading button { padding: 0; color: var(--accent); background: none; border: 0; cursor: pointer; font: .62rem ui-monospace, monospace; }
  .catalog-main-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; }
  .catalog-main-heading h2 { margin: .2rem 0 0; font: 400 1.6rem Georgia, serif; }
  .catalog-result-count, .catalog-section-heading span { color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-active-filters { display: flex; flex-wrap: wrap; align-items: center; gap: .4rem; margin-top: 1rem; }
  .catalog-active-filters-label { color: var(--muted); font: .62rem ui-monospace, monospace; }
  .catalog-filter-chip { display: inline-flex; align-items: center; gap: .35rem; padding: .25rem .35rem .25rem .5rem; color: var(--ink); background: var(--accent-light); border: 1px solid var(--hairline); border-radius: 1rem; font-size: .68rem; }
  .catalog-filter-chip button { display: grid; width: .9rem; height: .9rem; place-items: center; padding: 0; color: var(--muted); background: transparent; border: 0; border-radius: 50%; cursor: pointer; font: .65rem ui-monospace, monospace; }
  .catalog-filter-chip button:hover { color: var(--ink); background: var(--surface); }
  .catalog-category-section, .catalog-parts-section { margin-top: 1.75rem; }
  .catalog-section-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: .7rem; padding-bottom: .4rem; border-bottom: 1px solid var(--hairline); }
  .catalog-section-heading h3 { margin: 0; font: 400 1.1rem Georgia, serif; }
  @media (max-width: 700px) { .catalog-search { flex: none; width: 100%; max-width: none; } }
</style>
