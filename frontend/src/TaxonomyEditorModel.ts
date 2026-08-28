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

export type TaxonomyAttributeDefinition = {
  id: string
  label: string
  valueType: string
  unit?: string
  enumValues?: string[]
}

export type TaxonomyAttributeAssignment = {
  attributeDefId: string
  required: boolean
}

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
  rules: TaxonomyPlacement[]
  attributes?: TaxonomyAttributeAssignment[]
  children: TaxonomyDraftNode[]
}

export type TaxonomyEffectiveAttribute = TaxonomyAttributeDefinition & {
  required: boolean
  source: string
  overridden: boolean
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

export function effectiveTaxonomyAttributes(root: TaxonomyDraftNode, nodeID: string, definitions: TaxonomyAttributeDefinition[]): TaxonomyEffectiveAttribute[] {
  const path = findTaxonomyDraftPath(root, nodeID)
  const definitionByID = new Map(definitions.map((definition) => [definition.id, definition]))
  const effective = new Map<string, TaxonomyEffectiveAttribute>()
  for (const node of path) {
    for (const assignment of node.attributes ?? []) {
      const definition = definitionByID.get(assignment.attributeDefId)
      if (!definition) continue
      effective.set(assignment.attributeDefId, { ...definition, required: assignment.required, source: node.label, overridden: node !== path[0] && effective.has(assignment.attributeDefId) })
    }
  }
  return [...effective.values()]
}

function findTaxonomyDraftPath(node: TaxonomyDraftNode, nodeID: string, path: TaxonomyDraftNode[] = []): TaxonomyDraftNode[] {
  const nextPath = [...path, node]
  if (node.id === nodeID) return nextPath
  for (const child of node.children) {
    const childPath = findTaxonomyDraftPath(child, nodeID, nextPath)
    if (childPath.length) return childPath
  }
  return []
}

export function addTaxonomyDraftChild(node: TaxonomyDraftNode, parentID: string, child: TaxonomyDraftNode): TaxonomyDraftNode {
  if (node.id === parentID) return { ...node, children: [...node.children, child] }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftChild(current, parentID, child)) }
}

export function addTaxonomyDraftRule(node: TaxonomyDraftNode, parentID: string, rule: TaxonomyPlacement): TaxonomyDraftNode {
  if (node.id === parentID) return { ...node, rules: [...node.rules, rule] }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftRule(current, parentID, rule)) }
}

export function removeTaxonomyDraftRule(node: TaxonomyDraftNode, ruleID: string): TaxonomyDraftNode {
  return {
    ...node,
    rules: node.rules.filter((rule) => rule.id !== ruleID),
    children: node.children.map((child) => removeTaxonomyDraftRule(child, ruleID)),
  }
}

function replace(node: TaxonomyDraftNode, placements: TaxonomyPlacement[], existingIDs: Set<string>): TaxonomyDraftNode {
  const nodeIDs = new Set(node.placements.map((placement) => placement.id))
  return {
    ...node,
    placements: placements.filter((placement) => nodeIDs.has(placement.id) || (!existingIDs.has(placement.id) && node.id === 'taxonomy-root')),
    children: node.children.map((child) => replace(child, placements, existingIDs)),
  }
}
