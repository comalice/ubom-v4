<script lang="ts">
  import FormField from './FormField.svelte'

  export let id: string
  export let label: string
  export let value = ''
  export let name = ''
  export let type: 'text' | 'number' = 'text'
  export let min: number | undefined = undefined
  export let placeholder = ''
  export let help = ''
  export let required = false
  export let error = ''
  export let invalid = false
  export let disabled = false
</script>

<FormField {id} {label} {help} {required} {error} {invalid} {disabled} let:describedBy>
  <input class="primitive-input" {id} {name} {type} {min} {placeholder} {required} {disabled} aria-describedby={describedBy || undefined} aria-invalid={invalid || Boolean(error) || undefined} bind:value />
</FormField>
