<script lang="ts">
  import SequenceMockRow from './SequenceMockRow.svelte'
  import SequenceMockBlock from './SequenceMockBlock.svelte'
  import { createBlock, type BlockType, type SequenceBlock } from './sequenceMock'

  export let blocks: SequenceBlock[] = []
  export let depth = 0
  export let rowPath: string[] = []
  export let onChange: (blocks: SequenceBlock[]) => void
  export let onMove: (block: SequenceBlock, sourcePath: string[], targetPath: string[], index: number) => void
  let dropIndex: number | null = null
  let draggedId = ''
  let paletteType: BlockType | '' = ''
  let blockEls: (HTMLElement | null)[] = []

  function calcDropIndex(clientX: number) {
    for (let index = 0; index < blockEls.length; index += 1) {
      const element = blockEls[index]
      if (element && clientX < element.getBoundingClientRect().left + element.getBoundingClientRect().width / 2) return index
    }
    return blocks.length
  }

  function dragOver(event: DragEvent, index: number) {
    event.preventDefault()
    autoScroll(event)
    dropIndex = index
    paletteType = (event.dataTransfer?.getData('palette-type') || '') as BlockType | ''
    draggedId = event.dataTransfer?.getData('block-id') || ''
  }

  function autoScroll(event: DragEvent) {
    let element = event.currentTarget as HTMLElement | null
    while (element) {
      if (element.scrollHeight > element.clientHeight) {
        const rect = element.getBoundingClientRect()
        const edge = 48
        if (event.clientY < rect.top + edge) element.scrollBy({ top: -12 })
        else if (event.clientY > rect.bottom - edge) element.scrollBy({ top: 12 })
        return
      }
      element = element.parentElement
    }
  }

  function drop(event: DragEvent, index = blocks.length) {
    event.preventDefault()
    const type = (event.dataTransfer?.getData('palette-type') || paletteType) as BlockType | ''
    const sourceId = event.dataTransfer?.getData('block-id') || draggedId
    const next = [...blocks]
    if (type) next.splice(index, 0, createBlock(type))
    else if (sourceId) {
      const source = next.findIndex(block => block.id === sourceId)
      if (source >= 0) {
        const [block] = next.splice(source, 1)
        next.splice(source < index ? index - 1 : index, 0, block)
      } else {
        const data = event.dataTransfer?.getData('block-data')
        const sourcePath = event.dataTransfer?.getData('block-path')
        if (!data || !sourcePath) return resetDrop()
        try {
          onMove(JSON.parse(data) as SequenceBlock, JSON.parse(sourcePath) as string[], rowPath, index)
        } catch {
          return resetDrop()
        }
        resetDrop()
        return
      }
    }
    onChange(next)
    resetDrop()
  }

  function resetDrop() { dropIndex = null; draggedId = ''; paletteType = '' }
  function update(index: number, block: SequenceBlock) { const next = [...blocks]; next[index] = block; onChange(next) }
  function remove(index: number) { onChange(blocks.filter((_, itemIndex) => itemIndex !== index)) }
</script>

<div class="mock-row" class:mock-row-nested={depth > 0} role="list" on:dragover|stopPropagation={(event) => dragOver(event, calcDropIndex(event.clientX))} on:drop|stopPropagation={(event) => drop(event)} on:dragleave={(event) => { if (!event.currentTarget.contains(event.relatedTarget as Node)) resetDrop() }}>
  {#if !blocks.length && dropIndex === null}<span class="mock-row-empty">{depth ? 'Drop blocks into this arm' : 'Click a block or drag it here to start'}</span>{/if}
  {#each blocks as block, index (block.id)}
    <div class="mock-drop-slot">
      <span class="mock-drop-marker" class:mock-drop-marker-active={dropIndex === index}></span>
      <div class="mock-block-wrap" role="listitem" bind:this={blockEls[index]} on:dragover|stopPropagation={(event) => dragOver(event, calcDropIndex(event.clientX))} on:drop|stopPropagation={(event) => drop(event, calcDropIndex(event.clientX))}>
        <SequenceMockBlock {block} {depth} {rowPath} onChange={(value) => update(index, value)} onDelete={() => remove(index)} onMove={onMove} onDragStart={(event) => { event.dataTransfer?.setData('block-id', block.id); event.dataTransfer?.setData('block-data', JSON.stringify(block)); event.dataTransfer?.setData('block-path', JSON.stringify(rowPath)); draggedId = block.id }} />
      </div>
    </div>
  {/each}
  <div class="mock-drop-slot"><span class="mock-drop-marker" class:mock-drop-marker-active={dropIndex === blocks.length}></span></div>
</div>
