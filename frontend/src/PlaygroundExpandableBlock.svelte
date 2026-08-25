<script lang="ts">
  export let title = 'Literal'
  export let initiallyOpen = true
  let open = initiallyOpen
</script>

<article class="example-expandable-block" class:example-expandable-block-open={open}>
  <header class="example-expandable-header">
    <strong>{title}</strong>
    <button class="example-expandable-toggle" type="button" aria-label={open ? `Collapse ${title}` : `Expand ${title}`} aria-expanded={open} on:click={() => open = !open}>
      <span aria-hidden="true"></span>
    </button>
  </header>
  {#if open}<div class="example-expandable-body" aria-label="Empty block body"></div>{/if}
</article>
