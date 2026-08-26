<script lang="ts">
  export let type: 'button' | 'submit' | 'reset' = 'button'
  export let variant: 'primary' | 'secondary' | 'quiet' = 'primary'
  export let disabled = false
</script>

<button class="primitive-button primitive-button-{variant}" {type} {disabled} on:click>
  <slot />
</button>
