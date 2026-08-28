<script lang="ts">
  export let kind: 'chevron' | 'close'
  export let label: string
  export let size: 'normal' | 'small' = 'normal'
  export let expanded: boolean | undefined = undefined
  export let className = ''
  export let onClick: () => void = () => {}
</script>

<button class={`control-button control-${kind} button-size-${size} ${className}`} type="button" aria-label={label} aria-expanded={kind === 'chevron' ? expanded : undefined} on:click={onClick}>
  <span aria-hidden="true"></span>
</button>
