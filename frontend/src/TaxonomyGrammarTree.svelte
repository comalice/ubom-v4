<script lang="ts">
  import BlockShell from './BlockShell.svelte'
  import DropZone from './DropZone.svelte'
  import type { StructureNode } from './StructureTree.svelte'

  const grammarNodes: StructureNode[] = [
    { id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: 'sample-pn-v1', children: [
      { id: 'literal-prefix', label: 'Literal', blockType: 'literal', summary: '"PN-"', children: [] },
      { id: 'category', label: 'category', blockType: 'binding', summary: 'choice: A · B', children: [
        { id: 'category-values', label: 'Choice values', blockType: 'choice', summary: 'A · B', children: [] },
      ] },
      { id: 'number', label: 'number', blockType: 'binding', summary: 'range: 0000..9999', children: [] },
    ] },
  ]

  function dropLabel(node: StructureNode) {
    if (node.id === 'grammar-root') return 'Drop a category at the root'
    if (node.blockType === 'binding') return `Drop a refinement for ${node.label}`
    return 'Taxonomy nodes cannot attach to this grammar item'
  }

  function dropState(node: StructureNode) {
    if (node.id === 'grammar-root') return 'new'
    if (node.id === 'number') return 'move'
    return 'inactive'
  }
</script>

<ul class="dnd-structure-tree taxonomy-grammar-structure-tree">
  {#each grammarNodes as node (node.id)}
    <li>
      <span class="dnd-structure-node">{node.label}<small>{node.summary}</small></span>
      <DropZone variant={node.id === 'number' ? 'filled' : 'slot'} mode={dropState(node)} className="taxonomy-drop-zone taxonomy-grammar-drop-zone dnd-insert-zone" ariaLabel={`${dropLabel(node)} drop zone`} showHint={dropState(node) === 'inactive'} showLabel={node.id === 'number' || dropState(node) !== 'inactive'}>
        {#if node.id === 'number'}
          <span class="taxonomy-drop-label">Drop a taxonomy node for number</span>
          <BlockShell title="Category" summary="category = A" kind="taxonomy" collapsible onRemove={() => {}}>
            <div class="taxonomy-block-fields">
              <label>Label<input value="Resistors" /></label>
              <label>Node ID<input value="resistors" /></label>
            </div>
            <div class="taxonomy-block-meta"><span>Rule</span><strong>category = A</strong></div>
            <button class="taxonomy-block-add" type="button">+ Add child category</button>
          </BlockShell>
        {:else}
          <span class="taxonomy-drop-label">{dropLabel(node)}</span>
        {/if}
      </DropZone>
      {#if node.children.length}
        <ul class="dnd-structure-tree">
          {#each node.children as child (child.id)}
            <li>
              <span class="dnd-structure-node">{child.label}<small>{child.summary}</small></span>
              <DropZone variant={child.id === 'number' ? 'filled' : 'slot'} mode={dropState(child)} className="taxonomy-drop-zone taxonomy-grammar-drop-zone dnd-insert-zone" ariaLabel={`${dropLabel(child)} drop zone`} showHint={dropState(child) === 'inactive'} showLabel={child.id === 'number' || dropState(child) !== 'inactive'}>
                {#if child.id === 'number'}
                  <span class="taxonomy-drop-label">Drop a taxonomy node for number</span>
                  <BlockShell title="High Power" summary="number = 100..199" kind="taxonomy" collapsible onRemove={() => {}}>
                    <div class="taxonomy-block-fields">
                      <label>Label<input value="High Power" /></label>
                      <label>Node ID<input value="high-power" /></label>
                    </div>
                    <div class="taxonomy-block-meta"><span>Rule</span><strong>number = 100..199</strong></div>
                    <button class="taxonomy-block-add" type="button">+ Add child category</button>
                  </BlockShell>
                {:else}
                  <span class="taxonomy-drop-label">{dropLabel(child)}</span>
                {/if}
              </DropZone>
              {#if child.children.length}
                <ul class="dnd-structure-tree">
                  {#each child.children as grandchild (grandchild.id)}
                    <li>
                      <span class="dnd-structure-node">{grandchild.label}<small>{grandchild.summary}</small></span>
                      <DropZone variant="slot" mode="inactive" className="taxonomy-drop-zone taxonomy-grammar-drop-zone dnd-insert-zone" label={dropLabel(grandchild)} ariaLabel={`${dropLabel(grandchild)} drop zone`} showHint showLabel />
                    </li>
                  {/each}
                </ul>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </li>
  {/each}
</ul>
