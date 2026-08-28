<script lang="ts">
  export let kind: 'chevron' | 'close' | 'attributes'
  export let label: string
  export let expanded: boolean | undefined = undefined
  export let className = ''
  export let onClick: () => void = () => {}
</script>

<button class={`control-button control-${kind} ${className}`} type="button" aria-label={label} aria-expanded={kind === 'chevron' ? expanded : undefined} on:click={onClick}>
  <span aria-hidden="true">{kind === 'attributes' ? 'attributes' : ''}</span>
</button>
