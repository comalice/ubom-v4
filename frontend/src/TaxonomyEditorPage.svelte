<script lang="ts">
  import { getSeqDef, getTaxonomyDef, type SeqDefDefinition, type TaxonomyDefinition } from './api'
  import AsyncState from './AsyncState.svelte'
  import SectionHeading from './SectionHeading.svelte'
  import TaxonomyEditor from './TaxonomyEditor.svelte'
  import { taxonomyDraftFromDefinition, taxonomyGrammarFromSeqDef, type TaxonomyAttributeDefinition, type TaxonomyDraftNode, type TaxonomyGrammar } from './TaxonomyEditorModel'

  export let id: string | undefined = undefined
  export let seqDef: string | undefined = undefined

  let loading = true
  let error = ''
  let grammar: TaxonomyGrammar | undefined
  let draft: TaxonomyDraftNode | undefined
  let definitions: TaxonomyAttributeDefinition[] = []
  let loadedKey = ''
  let requestID = 0

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
    } catch (reason) {
      if (current === requestID) error = reason instanceof Error ? reason.message : 'Unable to load taxonomy editor'
    } finally {
      if (current === requestID) loading = false
    }
  }
</script>

<SectionHeading eyebrow="Taxonomies" title={id ? 'Edit taxonomy' : 'New taxonomy'} description="Build a taxonomy against an immutable sequence definition." />
{#if loading}
  <AsyncState state="loading" message="Loading taxonomy editor..." />
{:else if error}
  <AsyncState state="error" message={error} />
{:else if grammar && draft}
  {#key requestKey}
    <TaxonomyEditor grammar={grammar} initialAttributeDefinitions={definitions} initialDraft={draft} />
  {/key}
{/if}
