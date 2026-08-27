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

export type TaxonomyDraftNode = {
  id: string
  label: string
  targetId: string
  placements: TaxonomyPlacement[]
  children: TaxonomyDraftNode[]
}

export const taxonomyEditorContainer = 'taxonomy-editor'

export function taxonomyTargetContainer(targetId: string) {
  return `${taxonomyEditorContainer}:target:${targetId}`
}

export function flattenTaxonomyDraft(node: TaxonomyDraftNode): TaxonomyPlacement[] {
  return [...node.placements, ...node.children.flatMap(flattenTaxonomyDraft)]
}

export function replaceTaxonomyDraftPlacements(node: TaxonomyDraftNode, placements: TaxonomyPlacement[]): TaxonomyDraftNode {
  const existingIDs = new Set(flattenTaxonomyDraft(node).map((placement) => placement.id))
  return replace(node, placements, existingIDs)
}

export function findTaxonomyDraftNode(node: TaxonomyDraftNode, placementID: string): TaxonomyDraftNode | undefined {
  if (node.placements.some((placement) => placement.id === placementID)) return node
  return node.children.map((child) => findTaxonomyDraftNode(child, placementID)).find(Boolean)
}

export function addTaxonomyDraftChild(node: TaxonomyDraftNode, parentID: string, child: TaxonomyDraftNode): TaxonomyDraftNode {
  if (node.id === parentID) return { ...node, children: [...node.children, child] }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftChild(current, parentID, child)) }
}

function replace(node: TaxonomyDraftNode, placements: TaxonomyPlacement[], existingIDs: Set<string>): TaxonomyDraftNode {
  const nodeIDs = new Set(node.placements.map((placement) => placement.id))
  return {
    ...node,
    placements: placements.filter((placement) => nodeIDs.has(placement.id) || (!existingIDs.has(placement.id) && node.id === 'taxonomy-root')),
    children: node.children.map((child) => replace(child, placements, existingIDs)),
  }
}
