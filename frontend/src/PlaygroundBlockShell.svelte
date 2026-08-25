<script lang="ts">
  import { draggable } from '@thisux/sveltednd'
  import PlaygroundControlButton from './PlaygroundControlButton.svelte'

  export let title = ''
  export let summary = ''
  export let kind = ''
  export let size: 'normal' | 'small' = 'normal'
  export let item: unknown = undefined
  export let dragData: unknown = item
  export let container = ''
  export let collapsible = false
  export let initiallyOpen = true
  export let onDragEnd: () => void = () => {}
  export let dragHandle = ''
  export let interactive: string[] = []
  export let onRemove: (() => void) | undefined = undefined
  let open = initiallyOpen
</script>

<article class:dnd-size-small={size === 'small'} class:playground-block-shell-draggable={container !== ''} class:playground-block-shell-collapsible={collapsible} class:playground-block-shell-removable={onRemove !== undefined} class:playground-block-shell-open={open} class="playground-block-shell grammar-kind-{kind}" use:draggable={{ container: container || 'playground-block-shell', dragData, handle: dragHandle || undefined, interactive, disabled: container === '', callbacks: { onDragEnd } }}>
  <header class="playground-block-shell-header">
    <span class="grammar-drag" aria-hidden="true">::</span>
    <strong>{title}</strong>
    {#if summary}<span class="playground-block-shell-summary">{summary}</span>{/if}
    {#if collapsible}<PlaygroundControlButton kind="chevron" className="playground-block-shell-toggle" label={open ? `Collapse ${title}` : `Expand ${title}`} expanded={open} onClick={() => open = !open} />{/if}
    {#if onRemove}<PlaygroundControlButton kind="close" className="playground-block-shell-remove" label="Remove {title}" onClick={onRemove} />{/if}
  </header>
  {#if !collapsible || open}<div class="playground-block-shell-body"><slot /></div>{/if}
</article>
