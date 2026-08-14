<script lang="ts">
  export let title = 'UBOM'
</script>

<svelte:head><title>{title} · UBOM</title></svelte:head>

<header class="site-header">
  <div class="header-inner">
    <a class="brand" href="/parts/components">UBOM</a>
    <span class="header-caption">Parts and bills of material</span>
  </div>
</header>

<main class="page-shell">
  <slot />
</main>
