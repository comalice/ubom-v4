<script lang="ts">
  import type { TaxonomyAttributeValueType, TaxonomyEffectiveAttribute, TaxonomyValidationError } from './TaxonomyEditorModel'

  export let attribute: TaxonomyEffectiveAttribute
  export let selectedNodeLabel: string
  export let actionLabel = ''
  export let onAction: () => void = () => {}
  export let editable = false
  export let onChange: (required: boolean) => void = () => {}
  export let onDefinitionChange: (label: string, valueType: TaxonomyAttributeValueType) => void = () => {}
  export let onRemove: () => void = () => {}
  export let errors: TaxonomyValidationError[] = []
  $: errorID = `taxonomy-attribute-error-${attribute.id}`

  function handleDefinitionChange(event: Event) {
    onDefinitionChange(attribute.label, (event.currentTarget as HTMLSelectElement).value)
  }

  function handleLabelChange(event: Event) {
    onDefinitionChange((event.currentTarget as HTMLInputElement).value, attribute.valueType)
  }

  function handleRequiredChange(event: Event) {
    onChange((event.currentTarget as HTMLInputElement).checked)
  }
</script>

<tr class:taxonomy-attribute-inherited={attribute.source !== selectedNodeLabel} class:taxonomy-attribute-override={attribute.overridden} class:taxonomy-attribute-invalid={errors.length > 0}>
  <th scope="row">{#if editable}<input aria-label={`Attribute name for ${attribute.label}`} aria-invalid={errors.length > 0} aria-describedby={errors.length > 0 ? errorID : undefined} value={attribute.label} on:change={handleLabelChange} />{:else}{attribute.label}{/if}{#if errors.length > 0}<div class="taxonomy-attribute-errors" id={errorID} role="alert">{#each errors as error}<span>{error.message}</span>{/each}</div>{/if}</th>
  <td>{#if editable}<select aria-label={`Value type for ${attribute.label}`} value={attribute.valueType} on:change={handleDefinitionChange}><option value="string">string</option><option value="decimal">decimal</option><option value="enum">enum</option></select>{:else}{attribute.valueType}{#if attribute.unit} · {attribute.unit}{/if}{/if}</td>
  <td>{#if editable}<label class="taxonomy-attribute-required"><input type="checkbox" checked={attribute.required} aria-label={`Required ${attribute.label}`} on:change={handleRequiredChange} /> Required</label>{:else}{attribute.required ? 'Required' : 'Optional'}{/if}</td>
  <td>{attribute.source}</td>
  {#if actionLabel || editable}<td><button class="taxonomy-attribute-row-action" type="button" on:click={editable ? onRemove : onAction}>{editable ? 'Remove' : actionLabel}</button></td>{/if}
</tr>
