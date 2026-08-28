import { describe, expect, it } from 'vitest'
import { currentTaxonomyForSeqDef, currentTaxonomySummaries } from './taxonomySelection'
import type { TaxonomyDefSummary } from './api'

const summary = (overrides: Partial<TaxonomyDefSummary>): TaxonomyDefSummary => ({
  id: 'taxonomy-v1',
  definitionId: 'components',
  version: 1,
  seqDef: 'part-number-v1',
  current: false,
  ...overrides,
})

describe('taxonomy summary selection', () => {
  it('keeps historical revisions out of the primary list', () => {
    const current = summary({ id: 'taxonomy-v2', version: 2, current: true })

    expect(currentTaxonomySummaries([summary({}), current])).toEqual([current])
  })

  it('selects the current summary for a sequence definition', () => {
    const historical = summary({ id: 'taxonomy-v1' })
    const current = summary({ id: 'taxonomy-v2', version: 2, current: true })

    expect(currentTaxonomyForSeqDef([historical, current], 'part-number-v1')).toEqual(current)
    expect(currentTaxonomyForSeqDef([historical], 'part-number-v1')).toBeUndefined()
    expect(currentTaxonomyForSeqDef([current], 'other-seq-def')).toBeUndefined()
  })
})
