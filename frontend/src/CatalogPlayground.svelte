<script lang="ts">
  import CategoryList, { type CategoryListItem } from './CategoryList.svelte'
  import CategoryTile from './CategoryTile.svelte'
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
    { label: 'Steel', count: 21, selected: true },
    { label: 'Aluminum', count: 8 },
    { label: 'Plastic', count: 4 },
  ]

  const gaugeTypes: FilterOption[] = [
    { label: 'Plain plug gauges', count: 18 },
    { label: 'Thread plug gauges', count: 11 },
    { label: 'Air plug gauges', count: 4, disabled: true },
    { label: 'Taper plug gauges', count: 7 },
  ]

  const fasteningTiles = [
    { title: 'Bolts', description: 'Machine bolts and structural fasteners.', image: 'IMG', partCount: 48 },
    { title: 'Screws', description: 'Thread-forming and machine screws.', image: 'IMG', partCount: 82 },
    { title: 'Nuts', description: 'Hex, lock, wing, and specialty nuts.', image: 'IMG', partCount: 35 },
    { title: 'Washers', description: 'Spacers, shims, and load distribution.', partCount: 26 },
    { title: 'Threaded Rod', description: 'Continuous-thread rod and studs.', partCount: 19 },
    { title: 'Rivets', description: 'Permanent joining hardware.', partCount: 31 },
  ]

  const adhesiveTiles = [
    { title: 'Adhesives', description: 'Structural and general-purpose bonding.', image: 'IMG', partCount: 44 },
    { title: 'Adhesive Tape', description: 'Double-sided, transfer, and mounting tape.', image: 'IMG', partCount: 29 },
    { title: 'Threadlockers', description: 'Threadlocking compounds and sealants.', partCount: 17 },
  ]

  const weldingTiles = [
    { title: 'Welding', description: 'Welding equipment and consumables.', image: 'IMG', partCount: 57 },
    { title: 'Brazing', description: 'Brazing rods, alloys, and fluxes.', partCount: 22 },
    { title: 'Soldering', description: 'Solder, tools, and accessories.', partCount: 38 },
  ]

  const parts: PartListItem[] = [
    { id: 'part-200', value: 'PN-BOLT-M6-001', taxonomyPath: ['Fastening & Joining', 'Bolts'], status: 'active', attributes: [{ id: 'material', label: 'Material', value: 'Steel', valueType: 'enum' }], revisions: [{ id: 'revision-200-a', revision: 'A' }] },
    { id: 'part-201', value: 'PN-BOLT-M8-002', taxonomyPath: ['Fastening & Joining', 'Bolts'], status: 'active', attributes: [{ id: 'material', label: 'Material', value: 'Steel', valueType: 'enum' }], revisions: [{ id: 'revision-201-a', revision: 'A' }] },
  ]
</script>

<section class="catalog-playground" aria-labelledby="catalog-heading">
  <div class="catalog-masthead">
    <div>
      <p class="eyebrow">Catalog</p>
      <h1 id="catalog-heading">Browse parts</h1>
      <p class="muted">Navigate by category, then narrow the current scope with relevant attributes.</p>
    </div>
    <label class="catalog-search">Search parts<input placeholder="Part number or description" /></label>
  </div>

  <div class="catalog-playground-frame">
    <aside class="catalog-panel" aria-label="Catalog navigation and filters">
      <section class="catalog-panel-section" aria-labelledby="catalog-categories-heading">
        <p class="catalog-panel-kicker">Browse catalog</p>
        <h2 id="catalog-categories-heading">Categories</h2>
        <CategoryList categories={categories} />
      </section>
      <section class="catalog-panel-section" aria-labelledby="catalog-filters-heading">
        <div class="catalog-panel-heading"><h2 id="catalog-filters-heading">Filter by</h2><button type="button">Clear</button></div>
        <FilterWidget title="Material" description="">
          <FilterOptionList options={materials} />
        </FilterWidget>
        <FilterWidget title="Gauge type" description="">
          <FilterOptionList options={gaugeTypes} columns={2} />
        </FilterWidget>
        <FilterWidget title="Diameter" description="">
          <FilterRange unit="mm" />
        </FilterWidget>
      </section>
    </aside>

    <main class="catalog-main" aria-label="Catalog content">
      <div class="catalog-breadcrumbs"><span>Home</span><span aria-hidden="true">/</span><strong>Fastening &amp; Joining</strong></div>
      <div class="catalog-main-heading"><div><p class="eyebrow">Category</p><h2>Fastening &amp; Joining</h2></div><span class="catalog-result-count">67 parts</span></div>
      <section class="catalog-category-section" aria-labelledby="fastening-heading">
        <div class="catalog-section-heading"><h3 id="fastening-heading">Fasteners</h3><span>6 categories</span></div>
        <div class="catalog-tile-grid">{#each fasteningTiles as tile}<CategoryTile {...tile} size="standard" />{/each}</div>
      </section>
      <section class="catalog-category-section" aria-labelledby="adhesive-heading">
        <div class="catalog-section-heading"><h3 id="adhesive-heading">Adhesives &amp; Tape</h3><span>3 categories</span></div>
        <div class="catalog-tile-grid catalog-tile-grid-small">{#each adhesiveTiles as tile}<CategoryTile {...tile} size="compact" />{/each}</div>
      </section>
      <section class="catalog-category-section" aria-labelledby="welding-heading">
        <div class="catalog-section-heading"><h3 id="welding-heading">Welding, Brazing &amp; Soldering</h3><span>3 categories</span></div>
        <div class="catalog-tile-grid catalog-tile-grid-small">{#each weldingTiles as tile}<CategoryTile {...tile} size="compact" />{/each}</div>
      </section>
      <section class="catalog-parts-section" aria-labelledby="catalog-parts-heading">
        <div class="catalog-section-heading"><h3 id="catalog-parts-heading">Parts in this category</h3><span>2 shown</span></div>
        <PartTable {parts} columns={[{ id: 'material', label: 'Material' }]} seqDef="part-number-v1" />
      </section>
    </main>
  </div>
</section>

<style>
  .catalog-playground { min-width: 0; }
  .catalog-masthead { display: flex; align-items: end; justify-content: space-between; gap: 1.5rem; margin-bottom: 1.5rem; }
  .catalog-masthead h1 { margin-bottom: .35rem; }
  .catalog-masthead p { margin: 0; font-size: .75rem; }
  .catalog-search { display: grid; flex: 0 1 20rem; gap: .3rem; color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-search input { min-width: 0; padding: .6rem .7rem; color: var(--ink); background: var(--surface); border: 1px solid var(--hairline); border-radius: .2rem; font: .75rem Inter, sans-serif; }
  .catalog-search input:focus { border-color: var(--accent); outline: none; }
  .catalog-playground-frame { display: grid; grid-template-columns: 15rem minmax(0, 1fr); gap: 1.5rem; }
  .catalog-panel { align-self: start; background: var(--surface); border: 1px solid var(--hairline); }
  .catalog-panel-section { padding: 1rem; }
  .catalog-panel-section + .catalog-panel-section { border-top: 1px solid var(--hairline); }
  .catalog-panel-kicker { margin: 0 0 .35rem; color: var(--accent); font: 600 .65rem ui-monospace, monospace; letter-spacing: .1em; text-transform: uppercase; }
  .catalog-panel h2 { margin: 0 0 .7rem; font: 400 1.15rem Georgia, serif; }
  .catalog-panel-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .5rem; }
  .catalog-panel-heading button { padding: 0; color: var(--accent); background: none; border: 0; cursor: pointer; font: .62rem ui-monospace, monospace; }
  .catalog-main { min-width: 0; }
  .catalog-breadcrumbs { display: flex; flex-wrap: wrap; gap: .4rem; margin-bottom: 1rem; color: var(--muted); font-size: .7rem; }
  .catalog-main-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; }
  .catalog-main-heading h2 { margin: .2rem 0 0; font: 400 1.6rem Georgia, serif; }
  .catalog-result-count, .catalog-section-heading span { color: var(--muted); font: .65rem ui-monospace, monospace; }
  .catalog-category-section, .catalog-parts-section { margin-top: 1.75rem; }
  .catalog-section-heading { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: .7rem; padding-bottom: .4rem; border-bottom: 1px solid var(--hairline); }
  .catalog-section-heading h3 { margin: 0; font: 400 1.1rem Georgia, serif; }
  .catalog-tile-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; }
  .catalog-tile-grid-small { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  @media (max-width: 900px) { .catalog-playground-frame { grid-template-columns: 13rem minmax(0, 1fr); gap: 1rem; } .catalog-tile-grid, .catalog-tile-grid-small { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  @media (max-width: 700px) { .catalog-masthead { align-items: stretch; flex-direction: column; } .catalog-search { max-width: none; } .catalog-playground-frame { grid-template-columns: 1fr; } .catalog-tile-grid, .catalog-tile-grid-small { grid-template-columns: 1fr; } }
</style>
