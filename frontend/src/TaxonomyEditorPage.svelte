<script lang="ts">
  import { createTaxonomyDef, getSeqDef, getTaxonomyDef, type SeqDefDefinition, type TaxonomyDefinition } from './api'
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

  $: requestKey = id ? `taxonomy:${id}` : `new:${seqDef ?? ''}`
  $: if (requestKey !== loadedKey) void load(requestKey)

  async function load(key: string) {
    loadedKey = key
    const current = ++requestID
    loading = true
    error = ''
    grammar = undefined
    draft = undefined
    try {
      let sequence: SeqDefDefinition
      let taxonomy: TaxonomyDefinition | undefined
      if (id) {
        taxonomy = await getTaxonomyDef(id)
        sequence = await getSeqDef(taxonomy.seqDef)
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
  $: saveDisabled = saving || !editorDraft || !validation.valid || (id ? !dirty : !seqDef)

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
    <div><p class="muted">{id ? `Revision opened from ${id}` : 'New taxonomy draft'}</p>{#if saveError}<p class="error" role="alert">{saveError}</p>{:else if saving}<p class="muted" role="status">Saving taxonomy revision...</p>{:else if !dirty}<p class="muted">No unsaved changes.</p>{/if}</div>
    <PrimitiveButton type="button" disabled={saveDisabled} on:click={save}>{saving ? 'Saving...' : 'Save taxonomy'}</PrimitiveButton>
  </div>
  {#key requestKey}
    <TaxonomyEditor grammar={grammar} initialAttributeDefinitions={definitions} initialDraft={draft} onChange={updateEditor} />
  {/key}
{/if}
