<script lang="ts">
  import type { TaxonomyEffectiveAttribute } from './TaxonomyEditorModel'

  export let attribute: TaxonomyEffectiveAttribute
  export let selectedNodeLabel: string
  export let actionLabel = ''
  export let onAction: () => void = () => {}
</script>

<tr class:taxonomy-attribute-inherited={attribute.source !== selectedNodeLabel} class:taxonomy-attribute-override={attribute.overridden}>
  <th scope="row">{attribute.label}</th>
  <td>{attribute.valueType}{#if attribute.unit} · {attribute.unit}{/if}</td>
  <td>{attribute.required ? 'Required' : 'Optional'}</td>
  <td>{attribute.source}</td>
  {#if actionLabel}<td><button class="taxonomy-attribute-row-action" type="button" on:click={onAction}>{actionLabel}</button></td>{/if}
</tr>
