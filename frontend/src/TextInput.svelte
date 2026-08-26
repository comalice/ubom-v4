<script lang="ts">
  export let id: string
  export let label: string
  export let value = ''
  export let name = ''
  export let type: 'text' | 'number' = 'text'
  export let min: number | undefined = undefined
  export let placeholder = ''
  export let help = ''
  export let required = false
</script>

<div class="primitive-field">
  <label for={id}>{label}{#if required}<span aria-hidden="true">*</span>{/if}</label>
  <input class="primitive-input" {id} {name} {type} {min} {placeholder} {required} bind:value />
  {#if help}<small>{help}</small>{/if}
</div>
