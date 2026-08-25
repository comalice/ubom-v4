<script lang="ts">
  import { draggable } from '@thisux/sveltednd'

  export type SequenceTile = { kind: 'sequence-tile'; type: string; title: string; description: string; id?: number; nodeId?: number }
  export let tile: SequenceTile
  export let variant: 'palette' | 'dropped'
  export let container: string
  export let onDragEnd: () => void = () => {}
  export let onRemove: (() => void) | undefined = undefined
</script>

<div class={variant === 'palette' ? 'sequence-example-tile sequence-example-palette-tile' : 'sequence-example-tile sequence-example-dropped-tile'} class:sequence-example-type-select={tile.type === 'select'} class:sequence-example-type-range={tile.type === 'range'} use:draggable={{ container, dragData: tile, callbacks: { onDragEnd } }}>
  <span class="sequence-example-dot" aria-hidden="true"></span>
  <span>
    <strong>{tile.title}</strong>
    {#if variant === 'palette'}<small>{tile.description}</small>{/if}
  </span>
  {#if onRemove}<button type="button" aria-label="Remove {tile.title}" on:click={onRemove}><span class="sequence-example-remove-mark" aria-hidden="true"></span></button>{/if}
</div>
