<script lang="ts">
  import ControlButton from './ControlButton.svelte'
  export let title = 'Literal'
  export let initiallyOpen = true
  let open = initiallyOpen
</script>

<article class="example-expandable-block" class:example-expandable-block-open={open}>
  <header class="example-expandable-header">
    <strong>{title}</strong>
    <ControlButton kind="chevron" className="example-expandable-toggle" label={open ? `Collapse ${title}` : `Expand ${title}`} expanded={open} onClick={() => open = !open} />
  </header>
  {#if open}<div class="example-expandable-body" aria-label="Empty block body"></div>{/if}
</article>
