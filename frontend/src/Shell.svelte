<script lang="ts">
  import Breadcrumbs, { type Breadcrumb } from './Breadcrumbs.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'

  export let title = 'UBOM'
  export let fullWidth = false
  export let breadcrumbs: Breadcrumb[] = []
</script>

<svelte:head><title>{title} · UBOM</title></svelte:head>

<header class="site-header">
  <div class="header-inner">
    <PrimitiveLink className="brand" href="/">UBOM</PrimitiveLink>
    <span class="header-caption">Parts and bills of material</span>
    <PrimitiveLink className="header-catalog" href="/catalog">Catalog</PrimitiveLink>
    <PrimitiveLink className="header-artifacts" href="/artifacts">Artifacts</PrimitiveLink>
    <PrimitiveLink className="header-settings" href="/settings">Settings</PrimitiveLink>
  </div>
</header>

<main class:full-width={fullWidth} class="page-shell">
  <Breadcrumbs items={breadcrumbs} />
  <slot />
</main>
