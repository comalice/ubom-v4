export type BlockType = 'literal' | 'select' | 'range' | 'branch'

export type SequenceBlock =
  | { id: string; type: 'literal'; value: string; fieldName?: string }
  | { id: string; type: 'select'; options: string[]; fieldName?: string }
  | { id: string; type: 'range'; columns: RangeColumn[]; padding: number; padChar: string; fieldName?: string }
  | { id: string; type: 'branch'; arms: BranchArm[]; fieldName?: string }

export type RangeColumn =
  | { id: string; kind: 'number'; from: string; to: string }
  | { id: string; kind: 'chars'; chars: string }
export type BranchArm = { id: string; label: string; blocks: SequenceBlock[] }

export const blockTypes: { type: BlockType; label: string; hint: string }[] = [
  { type: 'literal', label: 'Literal', hint: 'Fixed text' },
  { type: 'select', label: 'Select', hint: 'Pick one of N' },
  { type: 'range', label: 'Pattern', hint: 'Rollover columns' },
  { type: 'branch', label: 'Branch', hint: 'Alternation / OR' },
]

export function id() { return `${Date.now()}-${Math.random().toString(36).slice(2)}` }

export function createBlock(type: BlockType): SequenceBlock {
  if (type === 'literal') return { id: id(), type, value: '' }
  if (type === 'select') return { id: id(), type, options: ['', ''] }
  if (type === 'range') return { id: id(), type, columns: [{ id: id(), kind: 'number', from: '0', to: '9' }], padding: 0, padChar: '0' }
  return { id: id(), type, arms: [{ id: id(), label: 'A', blocks: [] }, { id: id(), label: 'B', blocks: [] }] }
}

export type SequenceSampler = { count: bigint; valueAt: (index: bigint) => string }

type ColumnSampler = { count: bigint; valueAt: (index: bigint) => string }

function columnSampler(column: RangeColumn, padding: number, padChar: string): ColumnSampler {
  if (column.kind === 'chars') {
    const chars = [...column.chars]
    return { count: BigInt(chars.length), valueAt: index => chars[Number(index)] || '' }
  }
  const from = column.from.trim(); const to = column.to.trim()
  if (/^\d+$/.test(from) && /^\d+$/.test(to)) {
    const start = BigInt(from); const end = BigInt(to); const count = end >= start ? end - start + 1n : 0n
    const width = padding > 0 ? Math.max(from.length, to.length, padding) : Math.max(from.length, to.length, 1)
    return { count, valueAt: index => (start + index).toString().padStart(width, padChar || '0') }
  }
  return { count: 0n, valueAt: () => '' }
}

function blockSampler(block: SequenceBlock): SequenceSampler {
  if (block.type === 'literal') return { count: 1n, valueAt: () => block.value }
  if (block.type === 'select') {
    const options = block.options.filter(option => option.trim())
    return { count: BigInt(options.length), valueAt: index => options[Number(index)] || '' }
  }
  if (block.type === 'range') {
    const columns = block.columns.map(column => columnSampler(column, block.padding, block.padChar))
    const count = columns.reduce((total, column) => total * column.count, 1n)
    return { count, valueAt: index => {
      const values: string[] = []
      for (let columnIndex = columns.length - 1; columnIndex >= 0; columnIndex -= 1) {
        const column = columns[columnIndex]
        values[columnIndex] = column.valueAt(index % column.count)
        index /= column.count
      }
      return values.join('')
    } }
  }
  const arms = block.arms.map(arm => createSequenceSampler(arm.blocks.length ? arm.blocks : [{ id: arm.id, type: 'literal', value: '' }]))
  const count = arms.reduce((total, arm) => total + arm.count, 0n)
  return { count, valueAt: index => {
    for (const arm of arms) { if (index < arm.count) return arm.valueAt(index); index -= arm.count }
    return ''
  } }
}

export function createSequenceSampler(blocks: SequenceBlock[]): SequenceSampler {
  if (!blocks.length) return { count: 0n, valueAt: () => '' }
  const samplers = blocks.map(blockSampler)
  const count = samplers.reduce((total, sampler) => total * sampler.count, 1n)
  return { count, valueAt: index => samplers.map((sampler, childIndex) => {
    const divisor = samplers.slice(childIndex + 1).reduce((total, child) => total * child.count, 1n)
    return sampler.valueAt(divisor ? index / divisor % sampler.count : index % sampler.count)
  }).join('') }
}

export function sampleSampler(sampler: SequenceSampler, beginning: number, ending: number) {
  const values: string[] = []
  const first = BigInt(Math.max(0, Math.floor(beginning) || 0)); const last = BigInt(Math.max(0, Math.floor(ending) || 0))
  const firstCount = first < sampler.count ? first : sampler.count
  for (let index = 0n; index < firstCount; index += 1n) values.push(sampler.valueAt(index))
  const lastStart = sampler.count > last ? sampler.count - last : 0n
  for (let index = lastStart; index < sampler.count; index += 1n) if (index >= firstCount) values.push(sampler.valueAt(index))
  return values
}

export function blockCount(block: SequenceBlock) { return blockSampler(block).count }
