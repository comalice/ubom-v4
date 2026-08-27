<script lang="ts">
  import FormField from './FormField.svelte'

  export type SelectOption = { value: string; label: string }
  export let id: string
  export let label: string
  export let options: SelectOption[] = []
  export let value = ''
  export let name = ''
  export let placeholder = ''
  export let help = ''
  export let required = false
  export let error = ''
  export let invalid = false
  export let disabled = false
</script>

<FormField {id} {label} {help} {required} {error} {invalid} {disabled} let:describedBy>
  <select class="primitive-input" {id} {name} {required} {disabled} aria-describedby={describedBy || undefined} aria-invalid={invalid || Boolean(error) || undefined} bind:value>
    {#if placeholder}<option value="">{placeholder}</option>{/if}
    {#each options as option}<option value={option.value}>{option.label}</option>{/each}
  </select>
</FormField>
