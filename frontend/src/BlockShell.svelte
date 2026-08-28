<script lang="ts">
  import { draggable } from '@thisux/sveltednd'
  import ControlButton from './ControlButton.svelte'

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
  export let errors: string[] = []
  let open = initiallyOpen

  function stopNestedDragEvent(event: Event) { event.stopPropagation() }
</script>

<article data-dnd-kind={dragData && typeof dragData === 'object' && 'kind' in dragData && (dragData as { kind?: string }).kind === 'palette' ? 'new' : 'move'} data-dnd-container={container} on:pointerdown={stopNestedDragEvent} on:dragstart={stopNestedDragEvent} class:component-size-small={size === 'small'} class:block-shell-draggable={container !== ''} class:block-shell-collapsible={collapsible} class:block-shell-removable={onRemove !== undefined} class:block-shell-open={open} class="block-shell grammar-kind-{kind}" use:draggable={{ container: container || 'block-shell', dragData, handle: dragHandle || undefined, interactive, disabled: container === '', callbacks: { onDragEnd } }}>
  <header class="block-shell-header">
    <div class="block-shell-drag">
      <span aria-hidden="true">::</span>
      <strong>{title}</strong>
      {#if summary}<span class="block-shell-summary">{summary}</span>{/if}
    </div>
    <div class="block-shell-header-actions"><slot name="header-actions" /></div>
    {#if collapsible}<ControlButton kind="chevron" className="block-shell-toggle" label={open ? `Collapse ${title}` : `Expand ${title}`} expanded={open} onClick={() => open = !open} />{/if}
    {#if errors.length}<span class="block-shell-error-count">{errors.length} error{errors.length === 1 ? '' : 's'}</span>{/if}
    {#if onRemove}<ControlButton kind="close" className="block-shell-remove" label="Remove {title}" onClick={onRemove} />{/if}
  </header>
  {#if errors.length}<ul class="block-shell-errors">{#each errors as error}<li>{error}</li>{/each}</ul>{/if}
  {#if !collapsible || open}<div class="block-shell-body"><slot /></div>{/if}
</article>
