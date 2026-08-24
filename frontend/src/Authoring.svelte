<script lang="ts">
  import { onMount } from 'svelte'
  import { createPartDraft, getSeqDefFields, previewPartDraft, releasePartDraft, updatePartDraft, type SeqDefField, type TaxonomyNodeView } from './api'

  export let taxonomy: TaxonomyNodeView
  let fields: SeqDefField[] = []
  let bindings: Record<string, string> = {}
  let attributes: Record<string, string> = {}
  let draftID = ''
  let preview: { complete: boolean; value?: string; error?: string } | null = null
  let error = ''
  let saving = false

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
    {#each fields as field}
      <label>{field.name}
        {#if field.kind === 'choice'}
          <select bind:value={bindings[field.name]}>
            <option value="">Choose...</option>
            {#each field.options ?? [] as option}<option value={option}>{option}</option>{/each}
          </select>
        {:else}
          <input bind:value={bindings[field.name]} placeholder={field.kind === 'range' ? `1-${field.max ?? ''}` : 'Enter value'} />
        {/if}
      </label>
    {/each}
    {#each taxonomy.attributes as attribute}
      <label>{attribute.label}{attribute.required ? ' *' : ''}
        {#if attribute.valueType === 'enum'}
          <select bind:value={attributes[attribute.id]} required={attribute.required}>
            <option value="">Choose...</option>
            {#each attribute.enumValues ?? [] as option}<option value={option}>{option}</option>{/each}
          </select>
        {:else}<input bind:value={attributes[attribute.id]} required={attribute.required} placeholder={attribute.unit ?? 'Enter value'} />{/if}
      </label>
    {/each}
    <button type="submit" disabled={saving}>{saving ? 'Checking...' : 'Preview draft'}</button>
  </form>
  {#if preview}
    <p class={preview.error ? 'error' : 'success'}>{preview.error ?? `Suggested: ${preview.value}`}</p>
    {#if preview.complete}<button type="button" on:click={release} disabled={saving}>Release part number</button>{/if}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
</section>
