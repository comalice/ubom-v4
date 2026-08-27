<script lang="ts">
  export let id: string
  export let label: string
  export let help = ''
  export let required = false
  export let error = ''
  export let invalid = false
  export let disabled = false

  $: describedBy = [help && `${id}-help`, error && `${id}-error`].filter(Boolean).join(' ')
  $: fieldInvalid = invalid || Boolean(error)
</script>

<div class="primitive-field" aria-disabled={disabled || undefined}>
  <label for={id}>{label}{#if required}<span aria-hidden="true">*</span>{/if}</label>
  <slot {describedBy} />
  {#if help}<small id={`${id}-help`}>{help}</small>{/if}
  {#if error}<small id={`${id}-error`}>{error}</small>{/if}
</div>
