<script lang="ts">
  export let title: string
  export let id = 'filter-widget'
  export let description = ''
  export let info = ''
  let infoOpen = false
</script>

<section class="filter-widget" aria-labelledby={`${id}-title`}>
  <div class="filter-widget-heading">
    <h4 id={`${id}-title`}>{title}</h4>
    {#if info}
      <button class:active={infoOpen} type="button" aria-label={`More information about ${title}`} aria-expanded={infoOpen} aria-controls={`${id}-info`} on:click={() => infoOpen = !infoOpen}>i</button>
    {/if}
  </div>
  {#if description}<p class="filter-widget-description">{description}</p>{/if}
  {#if info && infoOpen}<div class="filter-widget-info" id={`${id}-info`}>{info}</div>{/if}
  <div class="filter-widget-body"><slot /></div>
</section>

<style>
  .filter-widget { padding: .8rem; background: var(--ground); border: 1px solid var(--hairline-light); }
  .filter-widget-heading { display: flex; align-items: center; gap: .4rem; }
  .filter-widget-heading h4 { margin: 0; font: 600 .72rem ui-monospace, monospace; }
  .filter-widget-heading button { display: inline-grid; width: 1rem; height: 1rem; margin-left: auto; place-items: center; padding: 0; color: var(--muted); background: transparent; border: 1px solid currentColor; border-radius: 50%; cursor: pointer; font: 600 .6rem Georgia, serif; }
  .filter-widget-heading button:hover, .filter-widget-heading button.active { color: var(--accent); }
  .filter-widget-description { margin: .45rem 0 0; color: var(--muted); font-size: .68rem; line-height: 1.45; }
  .filter-widget-info { margin-top: .65rem; padding: .65rem; color: var(--muted); background: var(--surface); border-left: 2px solid var(--accent); font-size: .68rem; line-height: 1.45; }
  .filter-widget-body { margin-top: .7rem; }
</style>
