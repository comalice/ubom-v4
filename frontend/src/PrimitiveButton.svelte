<script lang="ts">
  export let type: 'button' | 'submit' | 'reset' = 'button'
  export let variant: 'primary' | 'secondary' | 'quiet' = 'primary'
  export let disabled = false
  export let className = ''
  export let title: string | undefined = undefined
</script>

<button {...$$restProps} class={`primitive-button primitive-button-${variant}${className ? ` ${className}` : ''}`} {type} {disabled} {title} on:click>
  <slot />
</button>
