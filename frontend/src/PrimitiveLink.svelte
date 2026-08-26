<script lang="ts">
  export let href: string
  export let variant: 'default' | 'muted' = 'default'
  export let current = false
  export let disabled = false
</script>

<a class="primitive-link primitive-link-{variant}" class:primitive-link-current={current} class:primitive-link-disabled={disabled} {href} aria-current={current ? 'page' : undefined} aria-disabled={disabled ? 'true' : undefined} on:click={disabled ? (event) => event.preventDefault() : undefined}>
  <slot />
</a>
