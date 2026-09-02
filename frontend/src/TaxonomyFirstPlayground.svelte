<script lang="ts">
  import TaxonomyEditor from './TaxonomyEditor.svelte'
  import { grammarNodes } from './TaxonomyEditorFixtures'
  import type { TaxonomyDraftNode, TaxonomyGrammar } from './TaxonomyEditorModel'

  const grammar: TaxonomyGrammar = { id: 'gtd-pn-taxonomy-v1', label: 'Part Numbers', fields: 3, nodes: grammarNodes }
  const draft: TaxonomyDraftNode = { id: 'taxonomy-root', label: 'Part Numbers', targetId: String(grammar.nodes[0]?.id ?? 'grammar-root'), placements: [], rules: [], children: [] }
</script>

<section class="taxonomy-first-draft" aria-labelledby="taxonomy-first-heading">
  <div class="taxonomy-first-heading">
    <div><p class="eyebrow">Interaction concept</p><h2 id="taxonomy-first-heading">Taxonomy-first authoring</h2><p class="taxonomy-first-intro">A taxonomy tree is the work surface. The grammar remains available as a quiet map of fields to classify by.</p></div>
    <span class="taxonomy-first-status">Exploration draft</span>
  </div>
  <TaxonomyEditor {grammar} initialDraft={draft} />
</section>

<style>
  .taxonomy-first-draft { max-width: 1280px; }.taxonomy-first-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 2rem; margin-bottom: 1.5rem; }.taxonomy-first-heading h2 { margin: .15rem 0 .55rem; letter-spacing: -.035em; font-size: clamp(2rem, 4vw, 3rem); line-height: 1; }.taxonomy-first-intro { max-width: 650px; margin: 0; color: var(--muted); line-height: 1.5; }.taxonomy-first-status { padding: .35rem .55rem; color: var(--muted); border: 1px solid var(--hairline); border-radius: .2rem; font: .6rem ui-monospace, monospace; text-transform: uppercase; letter-spacing: .08em; white-space: nowrap; }
  @media (max-width: 680px) { .taxonomy-first-heading { display: block; }.taxonomy-first-status { display: inline-block; margin-top: 1rem; } }
</style>
