<script lang="ts">
  export let title = 'UBOM'
  export let fullWidth = false
</script>

<svelte:head><title>{title} · UBOM</title></svelte:head>

<header class="site-header">
  <div class="header-inner">
    <a class="brand" href="/parts">UBOM</a>
    <span class="header-caption">Parts and bills of material</span>
  </div>
</header>

<main class:full-width={fullWidth} class="page-shell">
  <slot />
</main>
