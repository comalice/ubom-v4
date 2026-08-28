<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import TaxonomyGrammarNode from './TaxonomyGrammarNode.svelte'
  import { grammarNodes as defaultGrammarNodes } from './TaxonomyEditorFixtures'
  import type { TaxonomyDraftNode, TaxonomyDragPayload, TaxonomyPlacement } from './TaxonomyEditorModel'
  import type { StructureNode } from './StructureNode'

  export let placements: TaxonomyPlacement[] = []
  export let taxonomyDraft: TaxonomyDraftNode | undefined
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onChildDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRuleDrop: (state: DragDropState<TaxonomyDragPayload>, parentID: string) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
  export let onAttributes: (nodeID: string) => void = () => {}
  export let grammarNodes: StructureNode[] = defaultGrammarNodes
</script>

<ul class="dnd-structure-tree taxonomy-grammar-structure-tree">
  {#each grammarNodes as node (node.id)}
    <TaxonomyGrammarNode {node} {placements} {taxonomyDraft} {onDrop} {onChildDrop} {onRuleDrop} {onRemove} {onDragEnd} {onAttributes} />
  {/each}
</ul>
