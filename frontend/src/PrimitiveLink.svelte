<script lang="ts">
  export let href: string
  export let variant: 'default' | 'muted' = 'default'
  export let current = false
  export let disabled = false
  export let className = ''
  export let title: string | undefined = undefined
  export let target: string | undefined = undefined
  export let rel: string | undefined = undefined
</script>

<a {...$$restProps} class={`primitive-link primitive-link-${variant}${className ? ` ${className}` : ''}`} class:primitive-link-current={current} class:primitive-link-disabled={disabled} {href} {title} {target} {rel} aria-current={current ? 'page' : undefined} aria-disabled={disabled ? 'true' : undefined} on:click={disabled ? (event) => event.preventDefault() : undefined}>
  <slot />
</a>
