<script lang="ts">
  import SequenceMockRow from './SequenceMockRow.svelte'
  import { tick } from 'svelte'
  import { blockCount, id, type BranchArm, type SequenceBlock } from './sequenceMock'

  export let block: SequenceBlock
  export let depth = 0
  export let rowPath: string[] = []
  export let onChange: (block: SequenceBlock) => void
  export let onDelete: () => void
  export let onDragStart: (event: DragEvent) => void
  export let onMove: (block: SequenceBlock, sourcePath: string[], targetPath: string[], index: number) => void
  let open = true
  let optionInputs: (HTMLInputElement | null)[] = []
  let columnInputs: (HTMLInputElement | null)[] = []
  let armInputs: (HTMLInputElement | null)[] = []

  function update(value: Partial<SequenceBlock>) { onChange({ ...block, ...value } as SequenceBlock) }
  function updateOption(index: number, value: string) { const options = [...(block.type === 'select' ? block.options : [])]; options[index] = value; update({ options } as Partial<SequenceBlock>) }
  function updateColumn(index: number, key: 'from' | 'to', value: string) {
    if (block.type !== 'range') return
    const columns = [...block.columns]; columns[index] = { ...columns[index], [key]: value }; update({ columns } as Partial<SequenceBlock>)
  }
  function updateColumnChars(index: number, chars: string) {
    if (block.type !== 'range' || block.columns[index].kind !== 'chars') return
    const columns = [...block.columns]; columns[index] = { ...columns[index], chars }; update({ columns } as Partial<SequenceBlock>)
  }
  function updateColumnKind(index: number, kind: 'number' | 'chars') {
    if (block.type !== 'range') return
    const column = block.columns[index]
    const columns = [...block.columns]
    columns[index] = kind === 'number'
      ? { id: column.id, kind, from: '0', to: '9' }
      : { id: column.id, kind, chars: 'ABC' }
    update({ columns } as Partial<SequenceBlock>)
  }
  function updateArm(index: number, arm: BranchArm) {
    if (block.type !== 'branch') return
    const arms = [...block.arms]; arms[index] = arm; update({ arms } as Partial<SequenceBlock>)
  }
  async function addOption() {
    if (block.type !== 'select') return
    const newIndex = block.options.length
    update({ options: [...block.options, ''] } as Partial<SequenceBlock>)
    await tick()
    optionInputs[newIndex]?.focus()
  }
  async function addColumn() {
    if (block.type !== 'range') return
    const newIndex = block.columns.length
    update({ columns: [...block.columns, { id: id(), kind: 'number', from: '0', to: '9' }] } as Partial<SequenceBlock>)
    await tick()
    columnInputs[newIndex]?.focus()
  }
  async function addArm() {
    if (block.type !== 'branch') return
    const newIndex = block.arms.length
    update({ arms: [...block.arms, { id: id(), label: String.fromCharCode(65 + newIndex), blocks: [] }] } as Partial<SequenceBlock>)
    await tick()
    armInputs[newIndex]?.focus()
  }
</script>

<article class="mock-block mock-type-{block.type}" class:mock-block-open={open} class:mock-branch={block.type === 'branch'}>
  <header class="mock-block-header" role="group" draggable="true" on:dragstart={onDragStart}>
    <span class="mock-drag" aria-hidden="true">::</span><span class="mock-dot"></span><strong>{block.type === 'range' ? 'pattern' : block.type}</strong><span class="mock-count">x{blockCount(block)}</span>
    {#if block.type !== 'branch'}<button class="mock-icon" type="button" aria-label="Toggle {block.type} editor" on:click={() => open = !open}>{open ? '^' : 'v'}</button>{/if}
    <button class="mock-icon mock-delete" type="button" aria-label="Delete {block.type}" on:click={onDelete}>x</button>
  </header>

  {#if !open && block.type !== 'branch'}
     <div class="mock-summary">{block.type === 'literal' ? (block.value || 'empty') : block.type === 'select' ? block.options.filter(Boolean).join(' | ') || 'none' : block.columns.map(column => column.kind === 'chars' ? `[${column.chars}]` : `[${column.from}-${column.to}]`).join('')}</div>
  {/if}

  {#if open && block.type === 'literal'}
    <div class="mock-config"><input value={block.value} on:input={(event) => update({ value: (event.currentTarget as HTMLInputElement).value })} placeholder="fixed text" /></div>
  {:else if open && block.type === 'select'}
    <div class="mock-config mock-stack">{#each block.options as option, index}<div class="mock-inline"><input bind:this={optionInputs[index]} value={option} on:input={(event) => updateOption(index, (event.currentTarget as HTMLInputElement).value)} placeholder="option {index + 1}" /><button class="mock-icon" type="button" on:click={() => update({ options: block.options.filter((_, itemIndex) => itemIndex !== index) } as Partial<SequenceBlock>)}>x</button></div>{/each}<button class="mock-add" type="button" on:click={addOption}>+ add option</button></div>
  {:else if open && block.type === 'range'}
    <div class="mock-config mock-stack">{#each block.columns as column, index}<div class="mock-inline"><span class="mock-index">[{index}]</span><select class="mock-column-kind" value={column.kind} aria-label="Pattern column type" on:change={(event) => updateColumnKind(index, (event.currentTarget as HTMLSelectElement).value as 'number' | 'chars')}><option value="number">number</option><option value="chars">chars</option></select>{#if column.kind === 'number'}<input bind:this={columnInputs[index]} value={column.from} on:input={(event) => updateColumn(index, 'from', (event.currentTarget as HTMLInputElement).value)} placeholder="from" /><span class="mock-dash">-</span><input value={column.to} on:input={(event) => updateColumn(index, 'to', (event.currentTarget as HTMLInputElement).value)} placeholder="to" />{:else}<input class="mock-chars-input" value={column.chars} on:input={(event) => updateColumnChars(index, (event.currentTarget as HTMLInputElement).value)} placeholder="allowed characters" />{/if}{#if block.columns.length > 1}<button class="mock-icon" type="button" aria-label="Delete pattern column" on:click={() => update({ columns: block.columns.filter((_, itemIndex) => itemIndex !== index) } as Partial<SequenceBlock>)}>x</button>{/if}</div>{/each}<button class="mock-add" type="button" on:click={addColumn}>+ add column</button><div class="mock-range-options"><label>Padding width<input type="number" min="0" value={block.padding} on:input={(event) => update({ padding: Number((event.currentTarget as HTMLInputElement).value) } as Partial<SequenceBlock>)} /></label><label>Padding character<input maxlength="1" value={block.padChar} on:input={(event) => update({ padChar: (event.currentTarget as HTMLInputElement).value.slice(0, 1) } as Partial<SequenceBlock>)} /></label></div></div>
  {:else if open && block.type === 'branch'}
    <div class="mock-config mock-stack">{#each block.arms as arm, index (arm.id)}<div class="mock-arm"><div class="mock-arm-header"><span>arm</span><input bind:this={armInputs[index]} tabindex="0" value={arm.label} on:input={(event) => updateArm(index, { ...arm, label: (event.currentTarget as HTMLInputElement).value })} /><span class="mock-count">x{blockCount({ id: arm.id, type: 'branch', arms: [arm] })}</span>{#if block.arms.length > 2}<button class="mock-icon" type="button" on:click={() => update({ arms: block.arms.filter((_, itemIndex) => itemIndex !== index) } as Partial<SequenceBlock>)}>x</button>{/if}</div><SequenceMockRow blocks={arm.blocks} rowPath={[...rowPath, arm.id]} depth={depth + 1} onChange={(blocks) => updateArm(index, { ...arm, blocks })} onMove={onMove} /></div>{/each}<button class="mock-add" type="button" on:click={addArm}>+ add arm</button></div>
  {/if}
  {#if open}<div class="mock-field-divider"></div><div class="mock-field-config"><label>Field name <span>(optional)</span><input value={block.fieldName || ''} on:input={(event) => update({ fieldName: (event.currentTarget as HTMLInputElement).value } as Partial<SequenceBlock>)} placeholder="e.g. family" /></label></div>{/if}
</article>
