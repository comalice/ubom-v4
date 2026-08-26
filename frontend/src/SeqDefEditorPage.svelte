<script lang="ts">
  import { createSeqDef, getSeqDef, type SeqDefDefinition, type SeqDefNode } from './api'
  import GrammarEditor from './GrammarEditor.svelte'
  import { fromSeqDefRoot, validateGrammarSequence } from './GrammarEditorModel'

  export let id: string | undefined = undefined

  const newRoot: SeqDefNode = {
    kind: 'concat',
    children: [
      { kind: 'literal', text: 'PN-' },
      { kind: 'range', min: 0, max: 999, width: 3, pad: '0' },
    ],
  }

  let definition: SeqDefDefinition = { id: '', root: newRoot }
  let baseline = JSON.stringify(definition)
  let loaded = !id
  let saving = false
  let error = ''
  let documentRoot = newRoot
  let loadedID: string | undefined = undefined
  let loadRequestID = 0

  $: validationErrors = validateGrammarSequence(fromSeqDefRoot(documentRoot))
  $: dirty = JSON.stringify({ id: definition.id, root: documentRoot }) !== baseline
  $: saveDisabled = saving || !definition.id.trim() || validationErrors.length > 0 || !dirty || !!id

  // Mark the ID before awaiting so the initial reactive run is the only request.
  $: if (id && id !== loadedID) void loadDefinition(id)
  $: if (!id && loadedID) {
    loadRequestID += 1
    loadedID = undefined
    definition = { id: '', root: newRoot }
    documentRoot = newRoot
    baseline = JSON.stringify(definition)
    loaded = true
    error = ''
  }

  async function loadDefinition(definitionID: string) {
    loadedID = definitionID
    const requestID = ++loadRequestID
    loaded = false
    error = ''
    try {
      const result = await getSeqDef(definitionID)
      if (requestID !== loadRequestID) return
      definition = result
      documentRoot = definition.root
      baseline = JSON.stringify(definition)
    } catch (reason) {
      if (requestID !== loadRequestID) return
      error = reason instanceof Error ? reason.message : 'Unable to load sequence definition'
    } finally {
      if (requestID === loadRequestID) loaded = true
    }
  }

  function updateRoot(root: SeqDefNode) {
    documentRoot = root
    definition = { ...definition, root }
  }

  async function save() {
    if (saveDisabled) return
    error = ''
    saving = true
    try {
      const result = await createSeqDef({ id: definition.id.trim(), root: documentRoot })
      baseline = JSON.stringify({ id: result.id, root: documentRoot })
      definition = { id: result.id, root: documentRoot }
      window.history.pushState({}, '', `/settings/sequence-definitions/${encodeURIComponent(result.id)}`)
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to create sequence definition'
    } finally {
      saving = false
    }
  }
</script>

{#if !loaded}
  <p class="muted">Loading sequence definition...</p>
{:else if error && id && !definition.id}
  <p class="error" role="alert">{error}</p>
{:else}
  <p class="eyebrow">Sequence definition</p>
  <h1>{id ? definition.id : 'New sequence definition'}</h1>
  {#if !id}
    <form class="seq-def-toolbar" on:submit|preventDefault={save}>
      <label>ID<input bind:value={definition.id} placeholder="part-number-v1" autocomplete="off" /></label>
      <button type="submit" disabled={saveDisabled}>{saving ? 'Creating...' : 'Create definition'}</button>
    </form>
    {#if !dirty}<p class="muted">Add an ID or change the document to enable creation.</p>{/if}
  {:else}
    <p class="muted">Immutable definition. This document is read-only.</p>
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  <GrammarEditor initialRoot={documentRoot} readOnly={!!id} onChange={updateRoot} />
{/if}
