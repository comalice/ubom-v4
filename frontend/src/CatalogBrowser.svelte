<script lang="ts">
  export let title: string
  export let subtitle = ''
  export let eyebrow = 'Catalog'
  export let labelledBy = 'catalog-browser-heading'
</script>

<section class="catalog-browser" aria-labelledby={labelledBy}>
  <div class="catalog-browser-masthead">
    <div>
      <p class="eyebrow">{eyebrow}</p>
      <h1 id={labelledBy}>{title}</h1>
      {#if subtitle}<p class="muted">{subtitle}</p>{/if}
    </div>
    <slot name="search" />
  </div>

  <div class="catalog-browser-frame">
    <aside class="catalog-browser-sidebar" aria-label="Catalog navigation and filters">
      <slot name="sidebar" />
    </aside>
    <main class="catalog-browser-main" aria-label="Catalog content">
      <slot />
    </main>
  </div>
</section>

<style>
  .catalog-browser { min-width: 0; }
  .catalog-browser-masthead { display: flex; align-items: end; justify-content: space-between; gap: 1.5rem; margin-bottom: 1.5rem; }
  .catalog-browser-masthead h1 { margin-bottom: .35rem; }
  .catalog-browser-masthead p { margin: 0; font-size: .75rem; }
  .catalog-browser-frame { display: grid; grid-template-columns: 15rem minmax(0, 1fr); gap: 1.5rem; }
  .catalog-browser-sidebar { align-self: start; min-width: 0; background: var(--surface); border: 1px solid var(--hairline); }
  .catalog-browser-main { min-width: 0; }
  @media (max-width: 900px) { .catalog-browser-frame { grid-template-columns: 13rem minmax(0, 1fr); gap: 1rem; } }
  @media (max-width: 700px) { .catalog-browser-masthead { align-items: stretch; flex-direction: column; } .catalog-browser-frame { grid-template-columns: 1fr; } }
</style>
