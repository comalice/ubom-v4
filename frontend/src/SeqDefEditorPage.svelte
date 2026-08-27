<script lang="ts">
  import { createSeqDef, createSeqDefVersion, getSeqDef, type SeqDefDefinition, type SeqDefNode } from './api'
  import GrammarEditor from './GrammarEditor.svelte'
  import AsyncState from './AsyncState.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import { fromSeqDefRoot, validateGrammarSequence } from './GrammarEditorModel'

  export let id: string | undefined = undefined

  const newRoot: SeqDefNode = {
    kind: 'concat',
    children: [
      { kind: 'literal', text: 'PN-' },
      { kind: 'range', min: 0, max: 999, width: 3, pad: '0' },
    ],
  }

  let definition: SeqDefDefinition = { id: '', definitionId: '', version: 1, root: newRoot }
  let baseline = JSON.stringify({ id: definition.id, root: definition.root })
  let loaded = !id
  let saving = false
  let editingVersion = false
  let error = ''
  let documentRoot = newRoot
  let editorInvalid = false
  let loadedID: string | undefined = undefined
  let loadRequestID = 0

  $: validationErrors = validateGrammarSequence(fromSeqDefRoot(documentRoot))
  $: dirty = JSON.stringify({ id: definition.id, root: documentRoot }) !== baseline
  $: saveDisabled = saving || !definition.id.trim() || validationErrors.length > 0 || editorInvalid || !dirty || (!!id && !editingVersion)

  // Mark the ID before awaiting so the initial reactive run is the only request.
  $: if (id && id !== loadedID) void loadDefinition(id)
  $: if (!id && loadedID) {
    loadRequestID += 1
    loadedID = undefined
    definition = { id: '', definitionId: '', version: 1, root: newRoot }
    documentRoot = newRoot
    baseline = JSON.stringify({ id: definition.id, root: definition.root })
    loaded = true
    error = ''
  }

  async function loadDefinition(definitionID: string) {
    loadedID = definitionID
    editingVersion = false
    const requestID = ++loadRequestID
    loaded = false
    error = ''
    try {
      const result = await getSeqDef(definitionID)
      if (requestID !== loadRequestID) return
      definition = result
      documentRoot = definition.root
      baseline = JSON.stringify({ id: definition.id, root: definition.root })
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

  function updateEditorValidation(invalid: boolean) {
    editorInvalid = invalid
  }

  async function save() {
    if (saveDisabled) return
    error = ''
    saving = true
    try {
      const result = await createSeqDef({ id: definition.id.trim(), root: documentRoot })
      baseline = JSON.stringify({ id: result.id, root: documentRoot })
      definition = { ...definition, id: result.id, definitionId: result.definitionId, version: result.version, root: documentRoot }
      window.history.pushState({}, '', `/settings/sequence-definitions/${encodeURIComponent(result.id)}`)
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to create sequence definition'
    } finally {
      saving = false
    }
  }

  function beginVersionEdit() {
    editingVersion = true
    baseline = JSON.stringify({ id: definition.id, root: documentRoot })
    error = ''
  }

  async function saveAsVersion() {
    if (saveDisabled) return
    saving = true
    error = ''
    try {
      const result = await createSeqDefVersion(definition.id, documentRoot)
      editingVersion = false
      window.history.pushState({}, '', `/settings/sequence-definitions/${encodeURIComponent(result.id)}`)
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (reason) {
      error = reason instanceof Error ? reason.message : 'Unable to create new sequence definition version'
    } finally { saving = false }
  }
</script>

{#if !loaded}
  <AsyncState state="loading" message="Loading sequence definition..." />
{:else if error && id && !definition.id}
  <AsyncState state="error" message={error} />
{:else}
  <p class="eyebrow">Sequence definition</p>
  <h1>{id ? definition.id : 'New sequence definition'}</h1>
  {#if id}<p class="muted">Definition {definition.definitionId}, version {definition.version}</p>{/if}
  {#if !id}
    <form class="seq-def-toolbar" on:submit|preventDefault={save}>
      <label>ID<input bind:value={definition.id} placeholder="part-number-v1" autocomplete="off" /></label>
      <PrimitiveButton type="submit" disabled={saveDisabled}>{saving ? 'Creating...' : 'Create definition'}</PrimitiveButton>
    </form>
    {#if !dirty}<p class="muted">Add an ID or change the document to enable creation.</p>{/if}
  {:else}
    {#if !editingVersion}
      <p class="muted">Immutable definition. This document is read-only.</p>
      <PrimitiveButton type="button" on:click={beginVersionEdit}>Save as new version</PrimitiveButton>
    {:else}
      <form class="seq-def-toolbar" on:submit|preventDefault={saveAsVersion}>
        <p class="muted">Editing a copy. The next version will be created automatically.</p>
        <PrimitiveButton type="submit" disabled={saveDisabled}>{saving ? 'Saving...' : 'Save as new version'}</PrimitiveButton>
      </form>
    {/if}
  {/if}
  {#if error}<p class="error" role="alert">{error}</p>{/if}
  {#key id ? `${id}:${definition.id}:${definition.version}` : 'new'}
    <GrammarEditor initialRoot={documentRoot} readOnly={!!id && !editingVersion} onChange={updateRoot} onValidationChange={updateEditorValidation} />
  {/key}
{/if}
