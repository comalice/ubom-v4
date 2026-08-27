<script lang="ts">
  export let level: 'h1' | 'h2' = 'h1'
  export let eyebrow = ''
  export let title = ''
  export let description = ''
  export let headingId = ''
</script>

{#if eyebrow}<p class="eyebrow">{eyebrow}</p>{/if}
<svelte:element this={level} id={headingId || undefined}>{title}</svelte:element>
{#if description}<p class="muted settings-intro">{description}</p>{/if}
