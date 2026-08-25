<script lang="ts">
  import SequenceMockRow from './SequenceMockRow.svelte'
  import { blockTypes, createSequenceSampler, createBlock, sampleSampler, type BlockType, type SequenceBlock, type SequenceSampler } from './sequenceMock'

  export let navigate: (path: string) => void
  let blocks: SequenceBlock[] = []
  let name = 'untitled_sequence'
  let editingName = false
  let previewBeginning = 12
  let previewEnding = 4

  $: sampler = createSequenceSampler(blocks)
  $: totalCount = sampler.count

  $: previewGroups = (() => {
    const branches = blocks.flatMap((block, index) => block.type === 'branch' ? [{ block, index }] : [])
    if (!branches.length) return [{ label: '', sampler }]
    const groups: { label: string; sampler: SequenceSampler }[] = []
    function addGroup(branchIndex: number, selected: Map<number, number>, labels: string[]) {
      if (branchIndex === branches.length) {
        const variant = blocks.map((block, index) => {
          const armIndex = selected.get(index)
          if (block.type !== 'branch' || armIndex === undefined) return block
          return { ...block, arms: [block.arms[armIndex]] }
        })
        groups.push({ label: labels.join(' / '), sampler: createSequenceSampler(variant) })
        return
      }
      const branch = branches[branchIndex]
      branch.block.arms.forEach((arm, armIndex) => addGroup(branchIndex + 1, new Map(selected).set(branch.index, armIndex), [...labels, arm.label || 'Unnamed arm']))
    }
    addGroup(0, new Map(), [])
    return groups
  })()

  function shownValues(valueSampler: SequenceSampler) { return sampleSampler(valueSampler, previewBeginning, previewEnding) }
  function formatCount(value: bigint) { return value.toLocaleString() }

  function add(type: BlockType) { blocks = [...blocks, createBlock(type)] }
  function moveBlock(block: SequenceBlock, sourcePath: string[], targetPath: string[], index: number) {
    if (sourcePath.length < targetPath.length && sourcePath.every((value, pathIndex) => value === targetPath[pathIndex])) return
    let moved = false
    function changeRow(current: SequenceBlock[], path: string[], action: (row: SequenceBlock[]) => SequenceBlock[]): SequenceBlock[] {
      if (!path.length) return action(current)
      return current.map(item => item.type === 'branch' && item.arms.some(arm => arm.id === path[0])
        ? { ...item, arms: item.arms.map(arm => arm.id === path[0] ? { ...arm, blocks: changeRow(arm.blocks, path.slice(1), action) } : arm) }
        : item)
    }
    let next = changeRow(blocks, sourcePath, row => {
      const sourceIndex = row.findIndex(item => item.id === block.id)
      if (sourceIndex < 0) return row
      moved = true
      return row.filter((_, itemIndex) => itemIndex !== sourceIndex)
    })
    if (moved) next = changeRow(next, targetPath, row => { const result = [...row]; result.splice(Math.min(index, result.length), 0, block); return result })
    blocks = next
  }
  function focusNameInput(node: HTMLInputElement) { node.focus(); node.select() }
  function paletteDrag(event: DragEvent, type: BlockType) {
    event.dataTransfer?.setData('palette-type', type)
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'copy'
  }
</script>

<div class="mock-workspace">
  <header class="mock-workspace-header">
    <div class="mock-title"><span class="mock-title-mark">SEQ</span><span class="mock-slash">/</span><span>Sequence generator</span></div>
    <span class="mock-name-separator">·</span>
    {#if editingName}
      <input class="mock-name-input" use:focusNameInput value={name} aria-label="Sequence name" on:input={(event) => name = (event.currentTarget as HTMLInputElement).value} on:blur={() => { if (!name.trim()) name = 'untitled_sequence'; editingName = false }} on:keydown={(event) => event.key === 'Enter' && (event.currentTarget as HTMLInputElement).blur()} />
    {:else}
      <button class="mock-name-display" type="button" on:click={() => editingName = true}><span>{name}</span><span class="mock-name-icon" aria-hidden="true">✎</span></button>
    {/if}
    <span class="mock-local-badge">UI MOCK · NOT CONNECTED</span>
    <span class="mock-spacer"></span>
    {#if totalCount > 0n}<span class="mock-total">{formatCount(totalCount)} generated</span>{/if}
    <button class="mock-clear" type="button" on:click={() => blocks = []} disabled={!blocks.length}>Clear</button>
    <button class="mock-close" type="button" on:click={() => navigate('/settings')}>Close</button>
  </header>

  <div class="mock-workspace-body">
    <aside class="mock-palette">
      <span class="mock-section-label">Blocks</span>
      <span class="mock-palette-help">Click to append or drag into the sequence.</span>
      {#each blockTypes as item}
        <button class="mock-palette-item mock-type-{item.type}" type="button" draggable="true" on:dragstart={(event) => paletteDrag(event, item.type)} on:click={() => add(item.type)}>
          <span class="mock-palette-dot"></span><span><strong>{item.label}</strong><small>{item.hint}</small></span>
        </button>
      {/each}
      <div class="mock-model-note"><strong>Local model</strong><span>Changes stay in this page until we agree on the interaction.</span></div>
    </aside>

    <main class="mock-main">
      <section class="mock-sequence-panel">
        <div class="mock-panel-heading"><div><span class="mock-section-label">Sequence</span><h1>Build an identifier</h1></div>{#if blocks.length}<span class="mock-block-total">{blocks.length} block{blocks.length === 1 ? '' : 's'}</span>{/if}</div>
        <SequenceMockRow {blocks} onChange={(value) => blocks = value} onMove={moveBlock} />
      </section>
      <section class="mock-preview-panel">
        <div class="mock-preview-heading"><div><span class="mock-section-label">Live preview</span><h2>Generated values</h2></div>{#if totalCount > 0n}<span class="mock-preview-count">{formatCount(totalCount)} total</span>{/if}</div>
        <div class="mock-preview-controls"><label>Show beginning<input type="number" min="0" bind:value={previewBeginning} /></label><label>Show ending<input type="number" min="0" bind:value={previewEnding} /></label></div>
        {#if totalCount > 0n}
          <div class="mock-results">{#each previewGroups as group}{#if previewGroups.length > 1}<div class="mock-branch-divider"><span>{group.label}</span><small>{formatCount(group.sampler.count)} generated values below</small></div>{/if}{#each shownValues(group.sampler) as result, index}<span title={result || '(empty)'}>{result || '(empty)'}</span>{/each}{/each}</div>
        {:else}<div class="mock-preview-empty">Build a sequence above to see output.</div>{/if}
      </section>
    </main>
  </div>
</div>
