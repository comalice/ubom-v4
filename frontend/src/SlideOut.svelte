<script lang="ts">
  export let open = false
  export let title: string
  export let eyebrow = ''
  export let onClose: () => void = () => {}

  function handleKeydown(event: KeyboardEvent) {
    if (open && event.key === 'Escape') onClose()
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class:slide-out-open={open} class="slide-out" aria-hidden={!open}>
  <button class="slide-out-backdrop" type="button" aria-label="Close panel" tabindex={open ? 0 : -1} on:click={onClose}></button>
  <div class="slide-out-panel" role="dialog" aria-modal="true" aria-labelledby="slide-out-title">
    <header class="slide-out-header">
      <div>{#if eyebrow}<span class="mock-section-label">{eyebrow}</span>{/if}<h2 id="slide-out-title">{title}</h2></div>
      <button class="slide-out-close" type="button" aria-label={`Close ${title}`} tabindex={open ? 0 : -1} on:click={onClose}>close</button>
    </header>
    <div class="slide-out-body"><slot /></div>
  </div>
</div>
