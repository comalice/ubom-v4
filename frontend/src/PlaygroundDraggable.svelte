<script lang="ts">
  import { draggable } from '@thisux/sveltednd'

  export type DraggableComponent = { kind: 'component'; label: string; id?: number; nodeId?: number; sourceParentId?: number | null }
  export let item: DraggableComponent
  export let container: string
  export let variant: 'palette' | 'placed'
  export let description = ''
  export let dragData: DraggableComponent = item
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

<div class={variant === 'palette' ? 'dnd-palette-card' : 'dnd-placed-component'} use:draggable={{ container, dragData, callbacks: { onDragEnd } }}>
  <span class="dnd-grip" aria-hidden="true">::</span>
  <span>
    <strong>{item.label}</strong>
    {#if description}<small>{description}</small>{/if}
  </span>
  {#if onRemove}<button type="button" aria-label="Remove {item.label}" on:click={onRemove}>x</button>{/if}
</div>
