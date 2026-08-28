<script lang="ts">
  export let type: 'button' | 'submit' | 'reset' = 'button'
  export let variant: 'primary' | 'secondary' | 'quiet' = 'primary'
  export let size: 'normal' | 'small' = 'normal'
  export let disabled = false
  export let className = ''
  export let title: string | undefined = undefined
</script>

<button {...$$restProps} class={`primitive-button primitive-button-${variant} button-size-${size}${className ? ` ${className}` : ''}`} {type} {disabled} {title} on:click>
  <slot />
</button>
