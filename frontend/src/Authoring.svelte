<script lang="ts">
  import { createPart, type SeqDefView, type TaxonomyDefSummary } from './api'

  export let seqDefs: SeqDefView[] = []
  export let taxonomyDefs: TaxonomyDefSummary[] = []

  let value = ''
  let seqDef = ''
  let taxonomyDef = ''
  let message = ''
  let error = ''
  let saving = false

  $: selectedTaxonomy = taxonomyDefs.find(item => item.id === taxonomyDef)
  $: if (selectedTaxonomy) seqDef = selectedTaxonomy.seqDef

  async function submit() {
    error = ''
    message = ''
    saving = true
    try {
      const part = await createPart(value.trim(), seqDef, taxonomyDef)
      message = `Created ${part.value}`
      value = ''
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to create part'
    } finally {
      saving = false
    }
  }
</script>

<section class="authoring-panel" aria-labelledby="authoring-heading">
  <p class="eyebrow">Authoring</p>
  <h1 id="authoring-heading">Create a part number</h1>
  <p class="muted">Choose a taxonomy, then enter a value accepted by its sequence definition.</p>
  <form on:submit|preventDefault={submit}>
    <label>Taxonomy
      <select bind:value={taxonomyDef} required>
        <option value="" disabled>Select taxonomy</option>
        {#each taxonomyDefs as item}<option value={item.id}>{item.id}</option>{/each}
      </select>
    </label>
    <label>Part number
      <input bind:value={value} placeholder="PN-A0001" required />
    </label>
    <p class="muted authoring-seq">Sequence: <code>{seqDef || 'selected by taxonomy'}</code></p>
    <button type="submit" disabled={saving || !seqDef}>{saving ? 'Creating...' : 'Create part'}</button>
  </form>
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#if message}<p class="success" role="status">{message}</p>{/if}
  {#if seqDefs.length === 0 || taxonomyDefs.length === 0}<p class="muted">No authoring definitions are configured.</p>{/if}
</section>
