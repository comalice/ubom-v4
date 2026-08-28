<script lang="ts">
  import { createTaxonomyDef, getSeqDef, getTaxonomyDef, getTaxonomyDefHistory, type SeqDefDefinition, type TaxonomyDefinitionView, type TaxonomyDefSummary } from './api'
  import AsyncState from './AsyncState.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import SectionHeading from './SectionHeading.svelte'
  import TaxonomyEditor from './TaxonomyEditor.svelte'
  import { taxonomyDefinitionInputFromDraft, taxonomyDraftFromDefinition, taxonomyGrammarFromSeqDef, validateTaxonomyAttributes, type TaxonomyAttributeDefinition, type TaxonomyDraftNode, type TaxonomyGrammar } from './TaxonomyEditorModel'

  export let id: string | undefined = undefined
  export let seqDef: string | undefined = undefined

  let loading = true
  let error = ''
  let grammar: TaxonomyGrammar | undefined
  let draft: TaxonomyDraftNode | undefined
  let definitions: TaxonomyAttributeDefinition[] = []
  let loadedKey = ''
  let requestID = 0
  let saving = false
  let saveError = ''
  let editorDraft: TaxonomyDraftNode | undefined
  let editorDefinitions: TaxonomyAttributeDefinition[] = []
  let baseline = ''
  let history: TaxonomyDefSummary[] = []
  let historyLoading = false
  let historyError = ''
  let currentRevision = true

  $: requestKey = id ? `taxonomy:${id}` : `new:${seqDef ?? ''}`
  $: if (requestKey !== loadedKey) void load(requestKey)

  async function load(key: string) {
    loadedKey = key
    const current = ++requestID
    loading = true
    error = ''
    grammar = undefined
    draft = undefined
    history = []
    historyError = ''
    historyLoading = false
    currentRevision = true
    try {
      let sequence: SeqDefDefinition
      let taxonomy: TaxonomyDefinitionView | undefined
      if (id) {
        taxonomy = await getTaxonomyDef(id)
        sequence = await getSeqDef(taxonomy.seqDef)
        currentRevision = taxonomy.current
        historyLoading = true
        try {
          history = await getTaxonomyDefHistory(taxonomy.definitionId)
        } catch (reason) {
          if (current === requestID) historyError = reason instanceof Error ? reason.message : 'Unable to load revision history'
        } finally {
          if (current === requestID) historyLoading = false
        }
      } else if (seqDef) {
        sequence = await getSeqDef(seqDef)
      } else {
        throw new Error('Choose a sequence definition to begin.')
      }
      if (current !== requestID) return
      const nextGrammar = taxonomyGrammarFromSeqDef(sequence)
      grammar = nextGrammar
      definitions = (taxonomy?.attributeDefs ?? []).map((definition) => ({ id: definition.id, label: definition.label, valueType: definition.valueType as TaxonomyAttributeDefinition['valueType'], unit: definition.unit, enumValues: definition.enumValues }))
      draft = taxonomy ? taxonomyDraftFromDefinition(taxonomy, nextGrammar) : { id: 'taxonomy-root', label: 'Parts', targetId: 'grammar-root', placements: [], rules: [], attributes: [], children: [] }
      editorDraft = draft
      editorDefinitions = definitions
      baseline = JSON.stringify({ definitions, draft })
    } catch (reason) {
      if (current === requestID) error = reason instanceof Error ? reason.message : 'Unable to load taxonomy editor'
    } finally {
      if (current === requestID) loading = false
    }
  }

  $: validation = editorDraft ? validateTaxonomyAttributes(editorDefinitions, editorDraft) : { valid: false }
  $: dirty = editorDraft ? JSON.stringify({ definitions: editorDefinitions, draft: editorDraft }) !== baseline : false
  $: saveDisabled = saving || !editorDraft || !validation.valid || (id ? !dirty || !currentRevision : !seqDef)

  function updateEditor(draftValue: TaxonomyDraftNode, definitionsValue: TaxonomyAttributeDefinition[]) {
    editorDraft = draftValue
    editorDefinitions = definitionsValue
  }

  async function save() {
    if (saveDisabled || !editorDraft) return
    saving = true
    saveError = ''
    try {
      const result = await createTaxonomyDef(taxonomyDefinitionInputFromDraft(seqDef ?? grammar!.id, id, editorDefinitions, editorDraft))
      window.history.pushState({}, '', `/settings/taxonomies/${encodeURIComponent(result.id)}`)
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (reason) {
      const message = reason instanceof Error ? reason.message : 'Unable to save taxonomy'
      saveError = /conflict|current|parent/i.test(message)
        ? `Save conflict: the taxonomy changed since it was opened. Reload the current revision before saving. (${message})`
        : message
    } finally {
      saving = false
    }
  }
</script>

<SectionHeading eyebrow="Taxonomies" title={id ? 'Edit taxonomy' : 'New taxonomy'} description="Build a taxonomy against an immutable sequence definition." />
{#if loading}
  <AsyncState state="loading" message="Loading taxonomy editor..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if grammar && draft}
  <div class="taxonomy-page-toolbar">
    <div><p class="muted">{id ? `Revision opened from ${id}` : 'New taxonomy draft'}</p>{#if !currentRevision}<p class="muted" role="status">Historical revision; immutable.</p>{:else if saveError}<p class="error" role="alert">{saveError}</p>{:else if saving}<p class="muted" role="status">Saving taxonomy revision...</p>{:else if !dirty}<p class="muted">No unsaved changes.</p>{/if}</div>
    {#if currentRevision}<PrimitiveButton type="button" disabled={saveDisabled} on:click={save}>{saving ? 'Saving...' : 'Save taxonomy'}</PrimitiveButton>{/if}
  </div>
   {#if id}
     <section class="taxonomy-revision-history" aria-labelledby="taxonomy-revision-history-heading">
       <h3 id="taxonomy-revision-history-heading">Revision history</h3>
       {#if historyLoading}<p class="muted" role="status">Loading revision history...</p>
       {:else if historyError}<p class="muted" role="status">Revision history is unavailable. {historyError}</p>
       {:else if history.length}<ul>{#each history as revision (revision.id)}<li><a href={`/settings/taxonomies/${encodeURIComponent(revision.id)}`}>Revision {revision.version}</a> <span class="muted">{revision.id}</span>{#if revision.current}<span aria-label="current revision">Current</span>{/if}</li>{/each}</ul>
       {:else}<p class="muted">No revision history available.</p>{/if}
     </section>
  {/if}
  {#key requestKey}
    <TaxonomyEditor grammar={grammar} initialAttributeDefinitions={definitions} initialDraft={draft} readOnly={!currentRevision} onChange={updateEditor} />
  {/key}
{/if}
