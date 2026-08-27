<script lang="ts">
  import { type DragDropState } from '@thisux/sveltednd'
  import TaxonomyGrammarNode from './TaxonomyGrammarNode.svelte'
  import { grammarNodes } from './TaxonomyEditorFixtures'
  import type { TaxonomyDragPayload, TaxonomyPlacement } from './TaxonomyEditorModel'

  export let placements: TaxonomyPlacement[] = []
  export let onDrop: (state: DragDropState<TaxonomyDragPayload>, targetId: string, targetIndex: number) => void = () => {}
  export let onRemove: (placementId: string) => void = () => {}
  export let onDragEnd: () => void = () => {}
</script>

<ul class="dnd-structure-tree taxonomy-grammar-structure-tree">
  {#each grammarNodes as node (node.id)}
    <TaxonomyGrammarNode {node} {placements} {onDrop} {onRemove} {onDragEnd} />
  {/each}
</ul>
