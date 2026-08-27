export type TaxonomyRulePredicate =
  | { kind: 'exact'; field: string; value: string }
  | { kind: 'numeric-range'; field: string; min: number; max: number }

export type TaxonomyItemType = 'category' | 'rule' | 'attribute'

export type TaxonomyPaletteDragPayload = {
  kind: 'palette'
  taxonomyType: TaxonomyItemType
}

export type TaxonomyPlacedDragPayload = {
  kind: 'taxonomy'
  taxonomyType: TaxonomyItemType
  placementId: string
}

export type TaxonomyDragPayload = TaxonomyPaletteDragPayload | TaxonomyPlacedDragPayload

export type TaxonomyPlacement = {
  id: string
  targetId: string
  type: TaxonomyItemType
  title: string
  summary: string
  label?: string
  nodeId?: string
  predicate?: string
  predicateData?: TaxonomyRulePredicate
  attribute?: string
  value?: string
}

export const taxonomyEditorContainer = 'taxonomy-editor'

export function taxonomyTargetContainer(targetId: string) {
  return `${taxonomyEditorContainer}:target:${targetId}`
}
