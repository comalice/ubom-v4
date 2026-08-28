<script lang="ts">
  import Breadcrumbs from './Breadcrumbs.svelte'
  import BlockShell from './BlockShell.svelte'
  import ControlButton from './ControlButton.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import ExpandableBlock from './ExpandableBlock.svelte'
  import SequenceExampleTile, { type SequenceTile } from './SequenceExampleTile.svelte'
  import IdentityLink from './IdentityLink.svelte'
  import PartIdentity from './PartIdentity.svelte'
  import StructureTree from './StructureTree.svelte'
  import type { StructureNode } from './StructureNode'
  import DndDraggable, { type ComponentNode } from './DndDraggable.svelte'
  import DndField from './DndField.svelte'
  import DndComponentList, { type DropSlot } from './DndComponentList.svelte'
  import type { DragDropState } from '@thisux/sveltednd'
  import BasicBlock from './BasicBlock.svelte'
  import NestedBlock from './NestedBlock.svelte'
  import NewDropMarker from './NewDropMarker.svelte'
  import MoveDropMarker from './MoveDropMarker.svelte'
  import InactiveDropMarker from './InactiveDropMarker.svelte'
  import PartTable from './PartTable.svelte'
  import BomTable from './BomTable.svelte'
  import ArtifactTable from './ArtifactTable.svelte'
  import type { ArtifactView, BomNode, PartListItem } from './api'

  const tile: SequenceTile = { kind: 'sequence-tile', type: 'select', title: 'Plant code', description: 'Choose from a fixed set of plant identifiers' }
  const component: ComponentNode = { kind: 'component', label: 'Part number', blockType: 'basic', id: 1, children: [] }
  const nestedComponent: ComponentNode = { kind: 'component', label: 'Assembly', blockType: 'nested', id: 2, children: [component] }
  const structure: StructureNode[] = [{ id: 1, label: 'Part number', blockType: 'basic', summary: 'PN', children: [{ id: 2, label: 'Revision', blockType: 'basic', summary: 'Rev', children: [] }] }]
  const parts: PartListItem[] = [{ id: 'part-100', value: 'PN-100', taxonomyPath: ['Components', 'Fasteners'], attributes: [{ id: 'material', label: 'Material', value: 'Steel', valueType: 'text' }], revisions: [{ id: 'revision-1', revision: 'A' }] }]
  const bom: BomNode[] = [{ partNumber: { id: 'part-100', value: 'PN-100' }, revisionId: 'revision-1', revision: 'A', bom: [] }]
  const artifacts: ArtifactView[] = [{ id: 'artifact-1', name: 'parts.csv', kind: 'csv', sourceId: 'demo-source', path: '/imports/parts.csv' }]
  const noop = () => {}
  const remove = () => {}
  const handleDrop = (_state: DragDropState<any>) => {}
  const handleGapDrop = (_slot: DropSlot, _state: DragDropState<any>) => {}
  const handleRemove = (_id: number) => {}
</script>

<section class="core-components" aria-labelledby="core-components-heading">
  <p class="eyebrow">Core components</p>
  <h2 id="core-components-heading">Core components</h2>
  <p class="muted core-components-intro">A catalog of the reusable interface pieces currently used across UBOM. Each example is live where interaction is meaningful.</p>

  <section class="core-component-group" aria-labelledby="navigation-heading">
    <div class="core-component-heading"><div><p class="mock-section-label">Navigation and identity</p><h3 id="navigation-heading">Navigation and identity</h3></div><span class="core-component-count">4 components</span></div>
    <div class="core-component-grid">
      <article class="core-component-card"><h4>Breadcrumbs</h4><p class="muted">Linked ancestors and the current page.</p><Breadcrumbs items={[{ label: 'Parts', href: '/parts' }, { label: 'Fasteners' }, { label: 'PN-100' }]} /></article>
      <article class="core-component-card"><h4>IdentityLink</h4><p class="muted">Part and revision identity with copy feedback.</p><IdentityLink partNumber="PN-100" revision="A" revisionID="revision-1" /></article>
      <article class="core-component-card"><h4>PartIdentity</h4><p class="muted">Identity, taxonomy path, and attributes.</p><PartIdentity partNumber="PN-100" taxonomy={['Components', 'Fasteners']} revision="A" attributes={[{ id: 'material', label: 'Material', value: 'Steel' }]} /></article>
      <article class="core-component-card"><h4>StructureTree</h4><p class="muted">Recursive structure display.</p><StructureTree nodes={structure} showIds={false} /></article>
    </div>
  </section>

  <section class="core-component-group" aria-labelledby="controls-heading">
    <div class="core-component-heading"><div><p class="mock-section-label">Feedback</p><h3 id="controls-heading">Feedback</h3></div><span class="core-component-count">5 components</span></div>
    <div class="core-component-grid">
      <article class="core-component-card"><h4>NewDropMarker</h4><p class="muted">Marker for creating a new item.</p><div class="core-marker-demo"><NewDropMarker /></div></article>
      <article class="core-component-card"><h4>MoveDropMarker</h4><p class="muted">Marker for moving an existing item.</p><div class="core-marker-demo"><MoveDropMarker /></div></article>
      <article class="core-component-card"><h4>InactiveDropMarker</h4><p class="muted">Neutral marker for an unavailable target.</p><div class="core-marker-demo"><InactiveDropMarker /></div></article>
      <article class="core-component-card"><h4>ControlButton</h4><p class="muted">Small icon actions and text actions for block headers.</p><div class="core-control-button-row"><PrimitiveButton variant="secondary" size="small">attributes</PrimitiveButton><ControlButton kind="chevron" label="Collapse block" expanded={true} /><ControlButton kind="close" label="Remove block" /></div></article>
      <article class="core-component-card"><h4>Empty and status states</h4><p class="muted">Existing shared patterns used by lists and tables.</p><div class="playground-empty core-mini-empty">No components to show.</div></article>
    </div>
  </section>

  <section class="core-component-group" aria-labelledby="blocks-heading">
    <div class="core-component-heading"><div><p class="mock-section-label">Blocks and tiles</p><h3 id="blocks-heading">Blocks and tiles</h3></div><span class="core-component-count">6 components</span></div>
    <div class="core-component-grid">
      <article class="core-component-card"><h4>BlockShell</h4><p class="muted">Shared draggable block chrome.</p><BlockShell title="Literal" summary="A fixed text value" kind="literal" collapsible onRemove={remove} errors={['Example validation message']}><span class="core-block-content">Block content</span></BlockShell></article>
      <article class="core-component-card"><h4>ExpandableBlock</h4><p class="muted">Standalone expandable example block.</p><ExpandableBlock title="Expandable example" /></article>
      <article class="core-component-card"><h4>SequenceExampleTile</h4><p class="muted">Palette and dropped sequence tiles.</p><div class="core-stack"><SequenceExampleTile {tile} variant="palette" container="core-components-palette" /><SequenceExampleTile {tile} variant="dropped" container="core-components-dropped" onRemove={remove} /></div></article>
      <article class="core-component-card"><h4>DndDraggable</h4><p class="muted">Generic draggable palette and placed items.</p><div class="core-stack"><DndDraggable item={component} variant="palette" container="core-components-dnd" title="Palette item" description="Drag source" /><DndDraggable item={component} variant="placed" container="core-components-dnd" onRemove={remove} /></div></article>
      <article class="core-component-card"><h4>BasicBlock</h4><p class="muted">Component-node adapter for a basic block.</p><BasicBlock item={component} parentId={null} container="core-components-blocks" onRemove={remove} /></article>
      <article class="core-component-card"><h4>NestedBlock</h4><p class="muted">Component-node adapter for nested content.</p><NestedBlock item={nestedComponent} parentId={null} container="core-components-blocks" onDragEnd={noop} onRemove={remove}><span class="core-block-content">Nested content</span></NestedBlock></article>
    </div>
  </section>

  <section class="core-component-group" aria-labelledby="dnd-heading">
    <div class="core-component-heading"><div><p class="mock-section-label">DnD foundations</p><h3 id="dnd-heading">DnD foundations</h3></div><span class="core-component-count">2 components</span></div>
    <div class="core-component-grid">
      <article class="core-component-card"><h4>DndField</h4><p class="muted">An empty drop field ready to receive a component.</p><DndField onDrop={handleDrop} onDragOver={noop} onGapDragLeave={noop} onGapDrop={handleGapDrop} onDragEnd={noop} onRemove={handleRemove} /></article>
      <article class="core-component-card"><h4>DndComponentList</h4><p class="muted">A reusable list of placed components and insertion zones.</p><DndComponentList nodes={[]} container="core-components-list" onDragOver={handleDrop} onDragLeave={noop} onDrop={handleGapDrop} onDragEnd={noop} onRemove={handleRemove} /></article>
    </div>
  </section>

  <section class="core-component-group" aria-labelledby="data-heading">
    <div class="core-component-heading"><div><p class="mock-section-label">Data display</p><h3 id="data-heading">Data display</h3></div><span class="core-component-count">3 components</span></div>
    <div class="core-component-table-stack">
      <article class="core-component-card"><h4>PartTable</h4><p class="muted">Part numbers with taxonomy and revision columns.</p><PartTable {parts} columns={[{ id: 'material', label: 'Material' }]} /></article>
      <article class="core-component-card"><h4>BomTable</h4><p class="muted">Recursive, expandable bill of materials.</p><BomTable nodes={bom} /></article>
      <article class="core-component-card"><h4>ArtifactTable</h4><p class="muted">Captured source artifact listing.</p><ArtifactTable {artifacts} /></article>
    </div>
  </section>

</section>
