import type { StructureNode } from './StructureNode'
import type { SeqDefDefinition, SeqDefNode, TaxonomyDefinition, TaxonomyDefinitionInput, TaxonomyNodeInput, TaxonomyPredicate } from './api'

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
    ...(node.kind === 'bind' && node.child?.kind === 'range' && typeof node.child.max === 'number'
      ? { numericDomain: { min: node.child.min ?? 0, max: node.child.max } }
      : {}),
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
    const predicate = predicates?.[0]
    const field = predicate?.field
    if (!field) return 'grammar-root'
    const candidates: StructureNode[] = []
    const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
      if (node.label === field && node.blockType === 'binding') candidates.push(node)
      visit(node.children)
    })
    visit(grammar.nodes)
    if (candidates.length <= 1) return candidates[0] ? String(candidates[0].id) : 'grammar-root'

    const numericPredicate = predicate?.kind === 'numeric-range'
      ? { min: predicate.min ?? 0, max: predicate.max ?? 0 }
      : predicate?.kind === 'exact' && /^-?\d+(?:\.\d+)?$/.test(predicate.value ?? '')
        ? { min: Number(predicate.value), max: Number(predicate.value) }
        : undefined
    if (!numericPredicate) return 'grammar-root'

    const compatible = candidates.filter((candidate): candidate is StructureNode & { numericDomain: { min: number; max: number } } => candidate.numericDomain !== undefined
      && candidate.numericDomain.min <= numericPredicate.min
      && candidate.numericDomain.max >= numericPredicate.max)
    if (compatible.length === 1) return String(compatible[0].id)
    if (compatible.length > 1) {
      const narrowest = compatible.filter((candidate) => candidate.numericDomain
        && !compatible.some((other) => other !== candidate && other.numericDomain
          && other.numericDomain.min >= candidate.numericDomain.min
          && other.numericDomain.max <= candidate.numericDomain.max
          && (other.numericDomain.min !== candidate.numericDomain.min || other.numericDomain.max !== candidate.numericDomain.max)))
      if (narrowest.length === 1) return String(narrowest[0].id)
    }
    // Equal or incomparable compatible domains stay unresolved instead of using traversal order.
    return 'grammar-root'
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
      id: categoryID,
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

function predicateFromText(text: string | undefined): TaxonomyPredicate | undefined {
  if (!text) return undefined
  const match = text.match(/^\s*([^=]+?)\s*=\s*(.+?)\s*$/)
  if (!match) return undefined
  const field = match[1].trim()
  const value = match[2].trim()
  const range = value.match(/^(-?\d+(?:\.\d+)?)\.\.(-?\d+(?:\.\d+)?)$/)
  return range
    ? { field, kind: 'numeric-range', min: Number(range[1]), max: Number(range[2]) }
    : { field, kind: 'exact', value }
}

function placementPredicate(placement: TaxonomyPlacement): TaxonomyPredicate | undefined {
  if (placement.predicateData) {
    return placement.predicateData.kind === 'numeric-range'
      ? { field: placement.predicateData.field, kind: placement.predicateData.kind, min: placement.predicateData.min, max: placement.predicateData.max }
      : { field: placement.predicateData.field, kind: placement.predicateData.kind, value: placement.predicateData.value }
  }
  return predicateFromText(placement.predicate)
}

/** Converts the editor-only placement tree into the immutable API document. */
export function taxonomyDefinitionInputFromDraft(
  seqDef: string,
  parentId: string | undefined,
  definitions: TaxonomyAttributeDefinition[],
  draft: TaxonomyDraftNode,
): TaxonomyDefinitionInput {
  const toNode = (node: TaxonomyDraftNode): TaxonomyNodeInput => {
    const category = node.placements.find((placement) => placement.type === 'category')
    const predicates = [
      category ? placementPredicate(category) : undefined,
      ...node.rules.map(placementPredicate),
    ].filter((predicate): predicate is TaxonomyPredicate => predicate !== undefined)
    return {
      id: node.id,
      label: node.label,
      ...(predicates.length ? { predicates } : {}),
      ...(node.attributes?.length ? { attributes: node.attributes.map((assignment) => ({ ...assignment })) } : {}),
      ...(node.children.length ? { children: node.children.map(toNode) } : {}),
    }
  }
  return {
    seqDef,
    ...(parentId ? { parentId } : {}),
    attributeDefs: definitions.map((definition) => ({ ...definition, ...(definition.enumValues ? { enumValues: [...definition.enumValues] } : {}) })),
    root: toNode(draft),
  }
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
  sourceNodeId?: string
}

export type TaxonomyDragPayload = TaxonomyPaletteDragPayload | TaxonomyPlacedDragPayload

export function acceptsTaxonomyRuleDrop(payload: TaxonomyDragPayload | undefined, destinationNodeID?: string): boolean {
  if (payload?.kind === 'palette') return payload.taxonomyType === 'rule'
  return payload?.kind === 'taxonomy' && payload.taxonomyType === 'rule' && Boolean(destinationNodeID?.trim() && payload.sourceNodeId?.trim())
}

export function acceptsTaxonomyCategoryChildDrop(payload: TaxonomyDragPayload | undefined, destinationNodeID?: string): boolean {
  if (payload?.kind === 'palette') return payload.taxonomyType === 'category'
  return payload?.kind === 'taxonomy' && payload.taxonomyType === 'category' && Boolean(payload.sourceNodeId?.trim() && destinationNodeID?.trim() && payload.sourceNodeId !== destinationNodeID)
}

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

function predicateLabel(predicate: TaxonomyRulePredicate): string {
  return predicate.kind === 'numeric-range'
    ? `${predicate.field} = ${predicate.min}..${predicate.max}`
    : `${predicate.field} = ${predicate.value}`
}

/** Updates one draft node without changing placement IDs used by drag/drop. */
export function updateTaxonomyDraftNode(node: TaxonomyDraftNode, nodeID: string, changes: { label?: string; id?: string }): TaxonomyDraftNode {
  if (changes.id !== undefined && (!changes.id.trim() || changes.id !== nodeID && findTaxonomyDraftPath(node, changes.id).length)) return node
  if (node.id === nodeID) {
    const id = changes.id ?? node.id
    return {
      ...node,
      id,
      label: changes.label ?? node.label,
      placements: node.placements.map((placement) => placement.type === 'category'
        ? { ...placement, label: changes.label ?? placement.label, title: changes.label ?? placement.title, nodeId: id }
        : placement),
    }
  }
  const children = node.children.map((child) => updateTaxonomyDraftNode(child, nodeID, changes))
  return children.some((child, index) => child !== node.children[index]) ? { ...node, children } : node
}

/** Updates a rule and keeps its serialized predicate and UI text in sync. */
export function updateTaxonomyDraftRule(node: TaxonomyDraftNode, ruleID: string, predicate: TaxonomyRulePredicate): TaxonomyDraftNode {
  const text = predicateLabel(predicate)
  const update = (rule: TaxonomyPlacement) => rule.id === ruleID
    ? { ...rule, title: text, summary: text, predicate: text, predicateData: { ...predicate } }
    : rule
  const rules = node.rules.map(update)
  const children = node.children.map((child) => updateTaxonomyDraftRule(child, ruleID, predicate))
  return rules.some((rule, index) => rule !== node.rules[index]) || children.some((child, index) => child !== node.children[index])
    ? { ...node, rules, children }
    : node
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

export function addTaxonomyDraftChild(node: TaxonomyDraftNode, parentID: string, child: TaxonomyDraftNode, targetIndex = node.children.length): TaxonomyDraftNode {
  if (node.id === parentID) {
    if (!Number.isInteger(targetIndex) || targetIndex < 0 || targetIndex > node.children.length) return node
    const children = [...node.children]
    children.splice(targetIndex, 0, child)
    return { ...node, children }
  }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftChild(current, parentID, child, targetIndex)) }
}

/** Moves an existing category node without changing its placement or descendants. */
export function moveTaxonomyDraftChild(node: TaxonomyDraftNode, sourceNodeID: string, destinationNodeID: string, destinationIndex?: number): TaxonomyDraftNode {
  if (!sourceNodeID || !destinationNodeID || sourceNodeID === node.id || destinationIndex !== undefined && (!Number.isInteger(destinationIndex) || destinationIndex < 0)) return node

  const nodes: TaxonomyDraftNode[] = []
  const parents = new Map<TaxonomyDraftNode, TaxonomyDraftNode | undefined>()
  const visit = (current: TaxonomyDraftNode, parent?: TaxonomyDraftNode) => {
    nodes.push(current)
    parents.set(current, parent)
    current.children.forEach((child) => visit(child, current))
  }
  visit(node)

  const sources = nodes.filter((current) => current.id === sourceNodeID)
  const destinations = nodes.filter((current) => current.id === destinationNodeID)
  if (sources.length !== 1 || destinations.length !== 1) return node
  const source = sources[0]
  const destination = destinations[0]
  const insertionIndex = destinationIndex ?? destination.children.length
  if (insertionIndex > destination.children.length) return node

  // A destination below the source would make the tree cyclic.
  for (let current: TaxonomyDraftNode | undefined = destination; current; current = parents.get(current)) {
    if (current === source) return node
  }

  const sourceParent = parents.get(source)
  if (!sourceParent) return node
  const sourceIndex = sourceParent.children.indexOf(source)
  const adjustedIndex = sourceParent === destination && sourceIndex < insertionIndex ? insertionIndex - 1 : insertionIndex
  if (sourceParent === destination && adjustedIndex === sourceIndex) return node

  const remove = (current: TaxonomyDraftNode): TaxonomyDraftNode => ({
    ...current,
    children: current.children.filter((child) => child !== source).map(remove),
  })
  const insert = (current: TaxonomyDraftNode): TaxonomyDraftNode => {
    if (current.id === destinationNodeID) {
      const children = [...current.children]
      children.splice(adjustedIndex, 0, source)
      return { ...current, children }
    }
    const children = current.children.map(insert)
    return children.some((child, index) => child !== current.children[index]) ? { ...current, children } : current
  }
  return insert(remove(node))
}

/** Moves a category and keeps its grammar target metadata in sync. */
export function moveTaxonomyDraftCategory(
  node: TaxonomyDraftNode,
  sourceNodeID: string,
  destinationNodeID: string,
  targetId: string,
  destinationIndex?: number,
): TaxonomyDraftNode {
  if (!targetId.trim()) return node
  const nodes: TaxonomyDraftNode[] = []
  const parents = new Map<TaxonomyDraftNode, TaxonomyDraftNode | undefined>()
  const visit = (current: TaxonomyDraftNode, parent?: TaxonomyDraftNode) => {
    nodes.push(current)
    parents.set(current, parent)
    current.children.forEach((child) => visit(child, current))
  }
  visit(node)
  const sources = nodes.filter((current) => current.id === sourceNodeID)
  const destinations = nodes.filter((current) => current.id === destinationNodeID)
  if (sources.length !== 1 || destinations.length !== 1 || sources[0] === destinations[0]) return node
  const source = sources[0]
  const destination = destinations[0]
  if (destinationIndex !== undefined && (!Number.isInteger(destinationIndex) || destinationIndex < 0 || destinationIndex > destination.children.length)) return node
  for (let current: TaxonomyDraftNode | undefined = destination; current; current = parents.get(current)) if (current === source) return node
  const moved = moveTaxonomyDraftChild(node, sourceNodeID, destinationNodeID, destinationIndex)

  const update = (current: TaxonomyDraftNode): TaxonomyDraftNode => {
    if (current.id === sourceNodeID) {
      const placements = current.placements.map((placement) => {
        if (placement.type !== 'category') return placement
        const predicateData = placement.predicateData
          ? { ...placement.predicateData, field: targetId }
          : undefined
        const predicate = predicateData
          ? predicateData.kind === 'numeric-range'
            ? `${targetId} = ${predicateData.min}..${predicateData.max}`
            : `${targetId} = ${predicateData.value}`
          : placement.predicate?.replace(/^\s*[^=]+?\s*=\s*/, `${targetId} = `)
        return { ...placement, targetId, predicate, ...(predicateData ? { predicateData } : {}) }
      })
      return { ...current, targetId, placements }
    }
    const children = current.children.map(update)
    return children.some((child, index) => child !== current.children[index]) ? { ...current, children } : current
  }
  return update(moved)
}

export function removeTaxonomyDraftCategorySubtree(node: TaxonomyDraftNode, placementID: string): TaxonomyDraftNode {
  const children = node.children
    .filter((child) => !child.placements.some((placement) => placement.type === 'category' && placement.id === placementID))
    .map((child) => removeTaxonomyDraftCategorySubtree(child, placementID))
  return children.length !== node.children.length || children.some((child, index) => child !== node.children[index]) ? { ...node, children } : node
}

export function addTaxonomyDraftRule(node: TaxonomyDraftNode, parentID: string, rule: TaxonomyPlacement, targetIndex = node.rules.length): TaxonomyDraftNode {
  if (node.id === parentID) {
    const rules = [...node.rules]
    rules.splice(Math.max(0, Math.min(targetIndex, rules.length)), 0, rule)
    return { ...node, rules }
  }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftRule(current, parentID, rule)) }
}

/** Reorders an existing rule within one taxonomy node without changing its identity. */
export function reorderTaxonomyDraftRule(node: TaxonomyDraftNode, nodeID: string, ruleID: string, targetIndex: number): TaxonomyDraftNode {
  if (node.id === nodeID) {
    const sourceIndex = node.rules.findIndex((rule) => rule.id === ruleID)
    if (sourceIndex < 0 || !Number.isInteger(targetIndex) || targetIndex < 0 || targetIndex > node.rules.length) return node
    const adjustedIndex = sourceIndex < targetIndex ? targetIndex - 1 : targetIndex
    if (adjustedIndex === sourceIndex) return node
    const rules = [...node.rules]
    const [rule] = rules.splice(sourceIndex, 1)
    rules.splice(adjustedIndex, 0, rule)
    return { ...node, rules }
  }
  const children = node.children.map((child) => reorderTaxonomyDraftRule(child, nodeID, ruleID, targetIndex))
  return children.some((child, index) => child !== node.children[index]) ? { ...node, children } : node
}

/** Moves an existing rule between taxonomy nodes without rewriting its predicate. */
export function moveTaxonomyDraftRule(node: TaxonomyDraftNode, sourceNodeID: string, destinationNodeID: string, ruleID: string, destinationIndex: number): TaxonomyDraftNode {
  if (!sourceNodeID || !destinationNodeID || !ruleID || !Number.isInteger(destinationIndex) || destinationIndex < 0) return node

  const nodes: TaxonomyDraftNode[] = []
  const visit = (current: TaxonomyDraftNode) => {
    nodes.push(current)
    current.children.forEach(visit)
  }
  visit(node)

  const sourceNodes = nodes.filter((current) => current.id === sourceNodeID)
  const destinationNodes = nodes.filter((current) => current.id === destinationNodeID)
  const owners = nodes.filter((current) => current.rules.some((rule) => rule.id === ruleID))
  const occurrences = nodes.reduce((count, current) => count + current.rules.filter((rule) => rule.id === ruleID).length, 0)
  if (sourceNodes.length !== 1 || destinationNodes.length !== 1 || owners.length !== 1 || owners[0] !== sourceNodes[0] || occurrences !== 1) return node

  const source = sourceNodes[0]
  const destination = destinationNodes[0]
  if (destinationIndex > destination.rules.length) return node
  if (source === destination) return reorderTaxonomyDraftRule(node, sourceNodeID, ruleID, destinationIndex)

  const rule = source.rules.find((current) => current.id === ruleID)
  if (!rule) return node
  const movedRule = { ...rule, targetId: destinationNodeID }
  const replace = (current: TaxonomyDraftNode): TaxonomyDraftNode => {
    const nextChildren = current.children.map(replace)
    const nextRules = current === source
      ? current.rules.filter((currentRule) => currentRule.id !== ruleID)
      : current === destination
        ? [...current.rules.slice(0, destinationIndex), movedRule, ...current.rules.slice(destinationIndex)]
        : current.rules
    return nextRules !== current.rules || nextChildren.some((child, index) => child !== current.children[index])
      ? { ...current, rules: nextRules, children: nextChildren }
      : current
  }
  return replace(node)
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
