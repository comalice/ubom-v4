import type { TaxonomyDefSummary } from './api'

export function currentTaxonomySummaries(summaries: TaxonomyDefSummary[]) {
  return summaries.filter(summary => summary.current)
}

export function currentTaxonomyForSeqDef(summaries: TaxonomyDefSummary[], seqDef: string) {
  return summaries.find(summary => summary.seqDef === seqDef && summary.current)
}
