<script lang="ts">
  export let id: string
  export let label: string
  export let value = ''
  export let name = ''
  export let rows = 3
  export let placeholder = ''
  export let help = ''
</script>

<div class="primitive-field">
  <label for={id}>{label}</label>
  <textarea class="primitive-input" {id} {name} {rows} {placeholder} bind:value></textarea>
  {#if help}<small>{help}</small>{/if}
</div>
