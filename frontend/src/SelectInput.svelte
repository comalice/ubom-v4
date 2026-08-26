<script lang="ts">
  export type SelectOption = { value: string; label: string }
  export let id: string
  export let label: string
  export let options: SelectOption[] = []
  export let value = ''
  export let name = ''
  export let help = ''
</script>

<div class="primitive-field">
  <label for={id}>{label}</label>
  <select class="primitive-input" {id} {name} bind:value>
    {#each options as option}<option value={option.value}>{option.label}</option>{/each}
  </select>
  {#if help}<small>{help}</small>{/if}
</div>
