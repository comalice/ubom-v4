<script lang="ts">
  import { draggable } from '@thisux/sveltednd'

  export let title = ''
  export let kind = ''
  export let size: 'normal' | 'small' = 'normal'
  export let item: unknown = undefined
  export let dragData: unknown = item
  export let container = ''
  export let collapsible = false
  export let initiallyOpen = true
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
  let open = initiallyOpen
</script>

<article class:dnd-size-small={size === 'small'} class:playground-block-shell-draggable={container !== ''} class:playground-block-shell-collapsible={collapsible} class:playground-block-shell-open={open} class="playground-block-shell grammar-kind-{kind}" use:draggable={{ container: container || 'playground-block-shell', dragData, disabled: container === '', callbacks: { onDragEnd } }}>
  <header class="playground-block-shell-header">
    <span class="grammar-drag" aria-hidden="true">::</span>
    <strong>{title}</strong>
    {#if collapsible}<button class="playground-block-shell-toggle" type="button" aria-label={open ? `Collapse ${title}` : `Expand ${title}`} aria-expanded={open} on:click={() => open = !open}><span aria-hidden="true"></span></button>{/if}
    {#if onRemove}<button class="playground-block-shell-remove" type="button" aria-label="Remove {title}" on:click={onRemove}>x</button>{/if}
  </header>
  {#if !collapsible || open}<div class="playground-block-shell-body"><slot /></div>{/if}
</article>
