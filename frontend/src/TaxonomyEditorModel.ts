import type { StructureNode } from './StructureNode'
import type { SeqDefDefinition, SeqDefNode, TaxonomyDefinition } from './api'

export type TaxonomyGrammar = {
  id: string
  label: string
  fields: number
  nodes: StructureNode[]
}

function sequenceSummary(node: SeqDefNode): string {
  if (node.kind === 'literal') return JSON.stringify(node.text ?? '')
  if (node.kind === 'range') return `${node.min ?? 0}..${node.max ?? 0} / width ${node.width ?? 1}`
  if (node.kind === 'places') return (node.alphabets ?? []).map((alphabet) => `[${alphabet}]`).join('')
  if (node.kind === 'bind') return `${node.name ?? 'unnamed'} = ${node.child ? sequenceSummary(node.child) : 'empty'}`
  if (node.kind === 'choice') return (node.children ?? []).map(sequenceSummary).join(' | ')
  return (node.children ?? []).map(sequenceSummary).join(' + ') || 'empty'
}

function sequenceStructureNode(node: SeqDefNode, path: string): StructureNode {
  const children = node.kind === 'bind'
    ? (node.child ? [sequenceStructureNode(node.child, `${path}-child`)] : [])
    : node.kind === 'choice'
      ? (node.children ?? []).map((child, index) => ({ id: `${path}-alternative-${index}`, label: `alternative ${String.fromCharCode(65 + index)}`, blockType: 'row', summary: sequenceSummary(child), children: [sequenceStructureNode(child, `${path}-alternative-${index}`)] }))
      : (node.children ?? []).map((child, index) => sequenceStructureNode(child, `${path}-${index}`))
  return {
    id: path,
    label: node.kind === 'bind' ? (node.name ?? 'unnamed') : node.kind,
    blockType: node.kind === 'bind' ? 'binding' : node.kind,
    summary: sequenceSummary(node),
    children,
  }
}

function sequenceFieldCount(node: SeqDefNode, names = new Set<string>()): number {
  if (node.kind === 'bind') {
    if (node.name) names.add(node.name)
    return node.child ? sequenceFieldCount(node.child, names) : names.size
  }
  if (node.kind === 'choice') return (node.children ?? []).reduce((count, child) => sequenceFieldCount(child, names), names.size)
  return (node.children ?? []).reduce((count, child) => sequenceFieldCount(child, names), names.size)
}

export function taxonomyGrammarFromSeqDef(definition: SeqDefDefinition): TaxonomyGrammar {
  const root = definition.root.kind === 'concat'
    ? (definition.root.children ?? []).map((child, index) => sequenceStructureNode(child, `grammar-${index}`))
    : [sequenceStructureNode(definition.root, 'grammar-root')]
  return { id: definition.id, label: definition.id, fields: sequenceFieldCount(definition.root), nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: definition.id, children: root }] }
}

function predicateText(predicate: NonNullable<TaxonomyDefinition['root']['predicates']>[number]): string {
  return predicate.kind === 'numeric-range' ? `${predicate.field} = ${predicate.min}..${predicate.max}` : `${predicate.field} = ${predicate.value ?? ''}`
}

export function taxonomyDraftFromDefinition(definition: TaxonomyDefinition, grammar: TaxonomyGrammar): TaxonomyDraftNode {
  const targetFor = (predicates: TaxonomyDefinition['root']['predicates']) => {
    const field = predicates?.[0]?.field
    if (!field) return 'grammar-root'
    let target = 'grammar-root'
    const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
      if (node.label === field && node.blockType === 'binding') target = String(node.id)
      visit(node.children)
    })
    visit(grammar.nodes)
    return target
  }
  const toNode = (node: TaxonomyDefinition['root'], isRoot = false): TaxonomyDraftNode => {
    const predicates = node.predicates ?? []
    const targetId = targetFor(predicates)
    const categoryID = String(node.id)
    const rules = predicates.slice(isRoot ? 0 : 1).map((predicate, index) => ({
      id: `${categoryID}-rule-${index}`,
      targetId: categoryID,
      type: 'rule' as const,
      title: predicateText(predicate),
      summary: predicateText(predicate),
      predicate: predicateText(predicate),
      predicateData: predicate.kind === 'numeric-range'
        ? { kind: predicate.kind, field: predicate.field, min: predicate.min ?? 0, max: predicate.max ?? 0 }
        : { kind: predicate.kind, field: predicate.field, value: predicate.value ?? '' },
    }))
    return {
      id: isRoot ? 'taxonomy-root' : categoryID,
      label: node.label,
      targetId,
      placements: isRoot ? [] : [{ id: categoryID, targetId, type: 'category', title: node.label, summary: predicates[0] ? predicateText(predicates[0]) : 'Category', label: node.label, nodeId: categoryID, predicate: predicates[0] ? predicateText(predicates[0]) : '' }],
      rules,
      attributes: (node.attributes ?? []).map((assignment) => ({ attributeDefId: assignment.attributeDefId, required: assignment.required })),
      children: (node.children ?? []).map((child) => toNode(child)),
    }
  }
  return toNode(definition.root, true)
}

export type TaxonomyRulePredicate =
  | { kind: 'exact'; field: string; value: string }
  | { kind: 'numeric-range'; field: string; min: number; max: number }

export type TaxonomyItemType = 'category' | 'rule'

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
  valueType: TaxonomyAttributeValueType
  unit?: string
  enumValues?: string[]
}

export type TaxonomyAttributeValueType = 'string' | 'decimal' | 'enum'

export type TaxonomyValidationError = {
  code: string
  message: string
}

export type TaxonomyValidationResult = {
  valid: boolean
  definitionErrors: Record<string, TaxonomyValidationError[]>
  assignmentErrors: Record<string, TaxonomyValidationError[]>
  nodeErrors: Record<string, TaxonomyValidationError[]>
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

export function taxonomyAssignmentErrorKey(nodeID: string, assignmentIndex: number) {
  return `${nodeID}:${assignmentIndex}`
}

export function validateTaxonomyAttributes(definitions: TaxonomyAttributeDefinition[], root: TaxonomyDraftNode): TaxonomyValidationResult {
  const definitionErrors: Record<string, TaxonomyValidationError[]> = {}
  const assignmentErrors: Record<string, TaxonomyValidationError[]> = {}
  const nodeErrors: Record<string, TaxonomyValidationError[]> = {}
  const add = (target: Record<string, TaxonomyValidationError[]>, key: string, code: string, message: string) => {
    ;(target[key] ??= []).push({ code, message })
  }
  const definitionIDs = new Map<string, number>()
  const labels = new Map<string, string>()
  for (const definition of definitions) {
    definitionIDs.set(definition.id, (definitionIDs.get(definition.id) ?? 0) + 1)
    if (!definition.label.trim()) add(definitionErrors, definition.id, 'empty-label', 'Attribute label cannot be empty')
    const labelKey = definition.label.trim()
    if (labelKey) {
      if (labels.has(labelKey)) {
        add(definitionErrors, definition.id, 'duplicate-label', `Attribute label "${labelKey}" is already used`)
        add(definitionErrors, labels.get(labelKey)!, 'duplicate-label', `Attribute label "${labelKey}" is already used`)
      } else labels.set(labelKey, definition.id)
    }
    if (!['string', 'decimal', 'enum'].includes(definition.valueType)) add(definitionErrors, definition.id, 'invalid-value-type', 'Value type must be string, decimal, or enum')
    if (definition.valueType === 'enum') {
      if (!definition.enumValues?.some((value) => value.trim())) add(definitionErrors, definition.id, 'empty-enum', 'Enum attributes need one or more non-empty values')
      const values = new Map<string, number>()
      for (const value of definition.enumValues ?? []) {
        const normalized = value.trim()
        if (!normalized) add(definitionErrors, definition.id, 'empty-enum-value', 'Enum values cannot be empty')
        else if (values.has(normalized)) add(definitionErrors, definition.id, 'duplicate-enum-value', `Enum value "${normalized}" is duplicated`)
        else values.set(normalized, 1)
      }
    }
  }
  for (const [id, count] of definitionIDs) if (count > 1) add(definitionErrors, id, 'duplicate-id', `Attribute ID "${id}" is duplicated`)

  const visit = (node: TaxonomyDraftNode) => {
    const seen = new Map<string, number>()
    for (const [index, assignment] of (node.attributes ?? []).entries()) {
      const key = taxonomyAssignmentErrorKey(node.id, index)
      if (!definitions.some((definition) => definition.id === assignment.attributeDefId)) add(assignmentErrors, key, 'missing-definition', `Assignment references missing attribute definition "${assignment.attributeDefId}"`)
      const firstIndex = seen.get(assignment.attributeDefId)
      if (firstIndex !== undefined) {
        const message = `Attribute "${assignment.attributeDefId}" is assigned more than once on this node`
        add(assignmentErrors, key, 'duplicate-assignment', message)
        add(assignmentErrors, taxonomyAssignmentErrorKey(node.id, firstIndex), 'duplicate-assignment', message)
        add(nodeErrors, node.id, 'duplicate-assignment', message)
      } else seen.set(assignment.attributeDefId, index)
    }
    node.children.forEach(visit)
  }
  visit(root)
  const hasErrors = [definitionErrors, assignmentErrors, nodeErrors].some((errorsByKey) => Object.values(errorsByKey).some((errors) => errors.length > 0))
  return { valid: !hasErrors, definitionErrors, assignmentErrors, nodeErrors }
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
      if (!definition) {
        effective.set(assignment.attributeDefId, { id: assignment.attributeDefId, label: `Missing attribute: ${assignment.attributeDefId}`, valueType: 'string', required: assignment.required, source: node.label, overridden: false })
      } else effective.set(assignment.attributeDefId, { ...definition, required: assignment.required, source: node.label, overridden: node !== path[0] && effective.has(assignment.attributeDefId) })
    }
  }
  return [...effective.values()]
}

export function findTaxonomyDraftPath(node: TaxonomyDraftNode, nodeID: string, path: TaxonomyDraftNode[] = []): TaxonomyDraftNode[] {
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

export function addTaxonomyAttribute(node: TaxonomyDraftNode, nodeID: string, assignment: TaxonomyAttributeAssignment): TaxonomyDraftNode {
  if (node.id === nodeID) return { ...node, attributes: [...(node.attributes ?? []), assignment] }
  return { ...node, children: node.children.map((child) => addTaxonomyAttribute(child, nodeID, assignment)) }
}

export function updateTaxonomyAttribute(node: TaxonomyDraftNode, nodeID: string, attributeDefID: string, assignment: TaxonomyAttributeAssignment): TaxonomyDraftNode {
  if (node.id === nodeID) return { ...node, attributes: (node.attributes ?? []).map((current) => current.attributeDefId === attributeDefID ? assignment : current) }
  return { ...node, children: node.children.map((child) => updateTaxonomyAttribute(child, nodeID, attributeDefID, assignment)) }
}

export function removeTaxonomyAttribute(node: TaxonomyDraftNode, nodeID: string, attributeDefID: string): TaxonomyDraftNode {
  if (node.id === nodeID) return { ...node, attributes: (node.attributes ?? []).filter((assignment) => assignment.attributeDefId !== attributeDefID) }
  return { ...node, children: node.children.map((child) => removeTaxonomyAttribute(child, nodeID, attributeDefID)) }
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
