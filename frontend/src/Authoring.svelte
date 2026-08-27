<script lang="ts">
  import { onMount } from 'svelte'
  import { createPartDraft, getSeqDefFields, previewPartDraft, releasePartDraft, updatePartDraft, type SeqDefField, type TaxonomyNodeView } from './api'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import SelectInput from './SelectInput.svelte'
  import TextInput from './TextInput.svelte'

  export let taxonomy: TaxonomyNodeView
  let fields: SeqDefField[] = []
  let bindings: Record<string, string> = {}
  let attributes: Record<string, string> = {}
  let draftID = ''
  let preview: { complete: boolean; value?: string; error?: string } | null = null
  let error = ''
  let saving = false

  function inputId(prefix: string, key: string, index: number) {
    const safeKey = key.replace(/[^a-zA-Z0-9_-]/g, '-')
    return `${prefix}-${safeKey}-${index}`
  }

  onMount(async () => {
    try {
      fields = await getSeqDefFields(taxonomy.seqDefId)
       for (const field of fields) if (field.kind === 'range' && field.min !== undefined) bindings[field.name] = String(field.min).padStart(field.width ?? 1, field.pad ?? '0')
    } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to load sequence fields' }
  })

  async function saveDraft() {
    error = ''
    saving = true
    try {
      if (!draftID) {
        const draft = await createPartDraft(taxonomy.seqDefId, taxonomy.taxonomyDefId, taxonomy.id)
        draftID = draft.id
      }
      await updatePartDraft(draftID, bindings, Object.entries(attributes).filter(([, value]) => value !== '').map(([attributeDefId, value]) => ({ attributeDefId, value })))
      preview = await previewPartDraft(draftID)
    } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to preview draft' }
    finally { saving = false }
  }

  async function release() {
    error = ''
    saving = true
    try {
      const result = await releasePartDraft(draftID)
      window.history.pushState({}, '', `/part/${encodeURIComponent(result.value)}`)
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to release draft' }
    finally { saving = false }
  }
</script>

<section class="authoring-panel" aria-labelledby="authoring-heading">
  <p class="eyebrow">New part draft</p>
  <h2 id="authoring-heading">Create from {taxonomy.label}</h2>
  <p class="muted">Choose the known segments. The suggested number is not reserved until release.</p>
  <form on:submit|preventDefault={saveDraft}>
    {#each fields as field, index}
      {@const id = inputId('field', field.name, index)}
        {#if field.kind === 'choice'}
          <SelectInput {id} name={field.name} label={field.name} placeholder="Choose..." options={(field.options ?? []).map(option => ({ value: option, label: option }))} bind:value={bindings[field.name]} />
        {:else}
          <TextInput {id} name={field.name} label={field.name} placeholder={field.kind === 'range' ? `1-${field.max ?? ''}` : 'Enter value'} bind:value={bindings[field.name]} />
        {/if}
    {/each}
    {#each taxonomy.attributes as attribute, index}
      {@const id = inputId('attribute', attribute.id, index)}
        {#if attribute.valueType === 'enum'}
          <SelectInput {id} name={attribute.id} label={attribute.label} required={attribute.required} placeholder="Choose..." options={(attribute.enumValues ?? []).map(option => ({ value: option, label: option }))} bind:value={attributes[attribute.id]} />
        {:else}<TextInput {id} name={attribute.id} label={attribute.label} required={attribute.required} placeholder={attribute.unit ?? 'Enter value'} bind:value={attributes[attribute.id]} />{/if}
    {/each}
    <PrimitiveButton type="submit" disabled={saving}>{saving ? 'Checking...' : 'Preview draft'}</PrimitiveButton>
  </form>
  {#if preview}
    <p class={preview.error ? 'error' : 'success'}>{preview.error ?? `Suggested: ${preview.value}`}</p>
    {#if preview.complete}<PrimitiveButton type="button" on:click={release} disabled={saving}>Release part number</PrimitiveButton>{/if}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>

<style>
  .authoring-panel :global(.primitive-field label) {
    color: var(--muted);
    font-size: .8rem;
    font-weight: 400;
  }

  .authoring-panel :global(.primitive-input) {
    padding: .65rem .7rem;
  }
</style>
