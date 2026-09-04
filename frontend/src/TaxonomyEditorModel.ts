import type { StructureNode } from './StructureNode'
import type { SeqDefDefinition, SeqDefNode, TaxonomyDefinition, TaxonomyDefinitionInput, TaxonomyNodeInput, TaxonomyPredicate } from './api'

export type TaxonomyGrammar = {
  id: string
  label: string
  fields: number
  nodes: StructureNode[]
}

export type TaxonomyTargetResolution =
  | { status: 'unique'; field: string; targetId: string }
  | { status: 'ambiguous'; field: string; candidates: string[] }
  | { status: 'unresolved'; field: string }

export const taxonomyAmbiguousTarget = (field: string) => `ambiguous:${field}`
export const taxonomyUnresolvedTarget = (field: string) => `unresolved:${field}`
export function taxonomyTargetIsUsable(targetId: string): boolean {
  return !targetId.startsWith('ambiguous:') && !targetId.startsWith('unresolved:')
}

/** Returns binding names in their grammar order, without duplicate lanes. */
export function taxonomyGrammarBindingFields(grammar: TaxonomyGrammar): string[] {
  const fields: string[] = []
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (node.blockType === 'binding' && node.label && !fields.includes(node.label)) fields.push(node.label)
    visit(node.children)
  })
  visit(grammar.nodes)
  return fields
}

/** Resolves persisted structure targets and new field-name targets to a binding name. */
export function taxonomyTargetField(grammar: TaxonomyGrammar, targetId: string): string | undefined {
  const fields = taxonomyGrammarBindingFields(grammar)
  if (fields.includes(targetId)) return targetId
  let field: string | undefined
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (!field && String(node.id) === targetId && node.blockType === 'binding') field = node.label
    visit(node.children)
  })
  visit(grammar.nodes)
  return field
}

export function taxonomyBranchPath(grammar: TaxonomyGrammar, targetId: string): string | undefined {
  let result: string | undefined
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (!result && (String(node.id) === targetId || node.branchPath === targetId) && node.branchPath) result = node.branchPath
    visit(node.children)
  })
  visit(grammar.nodes)
  return result
}

/** Returns the canonical branch that owns a branch-scoped draft node. */
export function taxonomyDraftBranchPath(node: TaxonomyDraftNode): string | undefined {
  return node.branch?.path[node.branch.path.length - 1]
}

/**
 * Finds the fields available below a taxonomy target. Root targets expose the
 * first field; unresolved targets expose all fields as a useful recovery path.
 */
export function taxonomyContextFields(grammar: TaxonomyGrammar, targetId: string): string[] {
  const fields = taxonomyGrammarBindingFields(grammar)
  if (!fields.length) return []
  const targets: StructureNode[] = []
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (String(node.id) === targetId || node.label === targetId || node.branchPath === targetId || node.path === targetId) targets.push(node)
    visit(node.children)
  })
  visit(grammar.nodes)
  const target = targets.find((node) => node.blockType === 'binding')
  if (target) {
    const index = fields.indexOf(target.label)
    return index < 0 ? [] : fields.slice(index + 1)
  }
  const branch = targets.find((node) => node.branchPath === targetId)
  if (branch) return branch.capturedFields ?? []
  const root = targets.find((node) => node.id === 'grammar-root' || node.blockType === 'sequence')
  return root || targetId === 'grammar-root' ? fields.slice(0, 1) : fields
}

function sequenceSummary(node: SeqDefNode): string {
  if (node.kind === 'literal') return JSON.stringify(node.text ?? '')
  if (node.kind === 'range') return `${node.min ?? 0}..${node.max ?? 0} / width ${node.width ?? 1}`
  if (node.kind === 'places') return (node.alphabets ?? []).map((alphabet) => `[${alphabet}]`).join('')
  if (node.kind === 'bind') return `${node.name ?? 'unnamed'} = ${node.child ? sequenceSummary(node.child) : 'empty'}`
  if (node.kind === 'choice') return (node.children ?? []).map(sequenceSummary).join(' | ')
  return (node.children ?? []).map(sequenceSummary).join(' + ') || 'empty'
}

function sequenceStructureNode(node: SeqDefNode, path: string, choiceAncestry: StructureNode['choiceAncestry'] = [], displayPath = path): StructureNode {
  const children = node.kind === 'bind'
    ? (node.child ? [sequenceStructureNode(node.child, `${path}/bind`, choiceAncestry, `${displayPath}-child`)] : [])
    : node.kind === 'choice'
      ? (node.children ?? []).map((child, index) => {
        const alternativePath = `${displayPath}-alternative-${index}`
        const canonicalBranchPath = `${path}/choice[${index}]`
        const alternativeLabel = `alternative ${String.fromCharCode(65 + index)}`
        const alternativeSummary = sequenceSummary(child)
        const ancestry = [...choiceAncestry, { path: canonicalBranchPath, label: 'choice', summary: sequenceSummary(node) }]
        return { id: alternativePath, label: alternativeLabel, blockType: 'row', summary: alternativeSummary, path: alternativePath, branchPath: canonicalBranchPath, choiceAncestry, alternativeLabel, alternativeSummary, targetable: true, literalOnly: !containsBinding(child), capturedFields: bindingNames(child), children: [sequenceStructureNode(child, canonicalBranchPath, ancestry, alternativePath)] }
      })
      : (node.children ?? []).map((child, index) => sequenceStructureNode(child, `${path}/${index}`, choiceAncestry, `${displayPath}-${index}`))
  return {
    id: displayPath,
    label: node.kind === 'bind' ? (node.name ?? 'unnamed') : node.kind,
    blockType: node.kind === 'bind' ? 'binding' : node.kind,
    summary: sequenceSummary(node),
    path: displayPath,
    choiceAncestry,
    targetable: node.kind === 'bind',
    literalOnly: node.kind !== 'bind' && !containsBinding(node),
    capturedFields: bindingNames(node),
    ...(node.kind === 'bind' && node.child?.kind === 'range' && typeof node.child.max === 'number'
      ? { numericDomain: { min: node.child.min ?? 0, max: node.child.max } }
      : {}),
    children,
  }
}

function containsBinding(node: SeqDefNode): boolean {
  if (node.kind === 'bind') return true
  return (node.child ? [node.child] : node.children ?? []).some(containsBinding)
}

function bindingNames(node: SeqDefNode): string[] {
  const names: string[] = []
  const visit = (current: SeqDefNode) => {
    if (current.kind === 'bind' && current.name && !names.includes(current.name)) names.push(current.name)
    ;(current.child ? [current.child] : current.children ?? []).forEach(visit)
  }
  visit(node)
  return names
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
    ? (definition.root.children ?? []).map((child, index) => sequenceStructureNode(child, `root/${index}`, [], `grammar-${index}`))
    : [sequenceStructureNode(definition.root, 'root', [], 'grammar-root')]
  return { id: definition.id, label: definition.id, fields: sequenceFieldCount(definition.root), nodes: [{ id: 'grammar-root', label: 'Part number', blockType: 'sequence', summary: definition.id, children: root }] }
}

export function taxonomyGrammarAlternatives(grammar: TaxonomyGrammar): StructureNode[] {
  const alternatives: StructureNode[] = []
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (node.alternativeLabel) alternatives.push(node)
    visit(node.children)
  })
  visit(grammar.nodes)
  return alternatives
}

export function taxonomyGrammarNode(grammar: TaxonomyGrammar, targetId: string): StructureNode | undefined {
  let result: StructureNode | undefined
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (!result && (String(node.id) === targetId || node.label === targetId || node.branchPath === targetId || node.path === targetId)) result = node
    visit(node.children)
  })
  visit(grammar.nodes)
  return result
}

/** Returns the nearest choices owned by a taxonomy grammar target. */
export function taxonomyGrammarChoices(grammar: TaxonomyGrammar, targetId: string): StructureNode[] {
  const target = taxonomyGrammarNode(grammar, targetId)
  if (!target) return []
  const choices: StructureNode[] = []
  const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
    if (node.blockType === 'choice') choices.push(node)
    else visit(node.children)
  })
  if (target.blockType === 'choice') choices.push(target)
  else visit(target.children)
  return choices
}

/** Returns the first captured binding in a choice alternative's grammar order. */
export function taxonomyGrammarFirstBinding(nodes: StructureNode[]): StructureNode | undefined {
  for (const node of nodes) {
    if (node.blockType === 'binding') return node
    const binding = taxonomyGrammarFirstBinding(node.children)
    if (binding) return binding
  }
  return undefined
}

export function taxonomyGrammarFieldHasChoiceAlternatives(grammar: TaxonomyGrammar, targetId: string, field: string): boolean {
  return taxonomyGrammarChoices(grammar, targetId).some((choice) => choice.children.some((alternative) => taxonomyGrammarFirstBinding(alternative.children)?.label === field))
}

export function taxonomyGrammarChoiceDropLabel(alternative: StructureNode, field?: string): string {
  const label = alternative.alternativeLabel?.replace(/^alternative\s+/i, '') ?? alternative.label
  return field ? `${field} - ${label}` : `branch - ${label}`
}

export function taxonomyGrammarBranchLabel(alternative: StructureNode): string {
  const field = taxonomyGrammarFirstBinding(alternative.children)?.label
  const label = alternative.alternativeLabel?.replace(/^alternative\s+/i, '') ?? alternative.label
  return field ? `Branch - ${field} - ${label}` : `Branch - ${label}`
}

function predicateText(predicate: NonNullable<TaxonomyDefinition['root']['predicates']>[number]): string {
  return predicate.kind === 'numeric-range' ? `${predicate.field} = ${predicate.min}..${predicate.max}` : `${predicate.field} = ${predicate.value ?? ''}`
}

export function taxonomyDraftFromDefinition(definition: TaxonomyDefinition, grammar: TaxonomyGrammar): TaxonomyDraftNode {
  const targetFor = (node: TaxonomyDefinition['root']): { targetId: string; resolution?: TaxonomyTargetResolution } => {
    if (node.branch?.path?.length) return { targetId: node.branch.path[node.branch.path.length - 1] }
    const predicates = node.predicates
    const predicate = predicates?.[0]
    const field = predicate?.field
    if (!field) return { targetId: 'grammar-root' }
    const candidates: StructureNode[] = []
    const visit = (nodes: StructureNode[]) => nodes.forEach((node) => {
      if (node.label === field && node.blockType === 'binding') candidates.push(node)
      visit(node.children)
    })
    visit(grammar.nodes)
    if (candidates.length === 1) return { targetId: String(candidates[0].id), resolution: { status: 'unique', field, targetId: String(candidates[0].id) } }
    if (!candidates.length) return { targetId: taxonomyUnresolvedTarget(field), resolution: { status: 'unresolved', field } }

    const numericPredicate = predicate?.kind === 'numeric-range'
      ? { min: predicate.min ?? 0, max: predicate.max ?? 0 }
      : predicate?.kind === 'exact' && /^-?\d+(?:\.\d+)?$/.test(predicate.value ?? '')
        ? { min: Number(predicate.value), max: Number(predicate.value) }
        : undefined
    if (!numericPredicate) return { targetId: taxonomyAmbiguousTarget(field), resolution: { status: 'ambiguous', field, candidates: candidates.map((candidate) => String(candidate.id)) } }

    const compatible = candidates.filter((candidate): candidate is StructureNode & { numericDomain: { min: number; max: number } } => candidate.numericDomain !== undefined
      && candidate.numericDomain.min <= numericPredicate.min
      && candidate.numericDomain.max >= numericPredicate.max)
    if (compatible.length === 1) return { targetId: String(compatible[0].id), resolution: { status: 'unique', field, targetId: String(compatible[0].id) } }
    if (compatible.length > 1) {
      const narrowest = compatible.filter((candidate) => candidate.numericDomain
        && !compatible.some((other) => other !== candidate && other.numericDomain
          && other.numericDomain.min >= candidate.numericDomain.min
          && other.numericDomain.max <= candidate.numericDomain.max
          && (other.numericDomain.min !== candidate.numericDomain.min || other.numericDomain.max !== candidate.numericDomain.max)))
      if (narrowest.length === 1) return { targetId: String(narrowest[0].id), resolution: { status: 'unique', field, targetId: String(narrowest[0].id) } }
    }
    // Equal or incomparable compatible domains remain ambiguous instead of using traversal order.
    return { targetId: taxonomyAmbiguousTarget(field), resolution: { status: 'ambiguous', field, candidates: candidates.map((candidate) => String(candidate.id)) } }
  }
  const toNode = (node: TaxonomyDefinition['root'], isRoot = false): TaxonomyDraftNode => {
    const predicates = node.predicates ?? []
    const target = targetFor(node)
    const targetId = target.targetId
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
      ...(node.branch ? { branch: { path: [...node.branch.path] } } : {}),
      ...(target.resolution ? { targetResolution: target.resolution } : {}),
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
    const categoryPredicate = category ? placementPredicate(category) : undefined
    const rules = node.rules.map(placementPredicate).filter((predicate): predicate is TaxonomyPredicate => predicate !== undefined)
    // A category's first predicate is the placeholder/category match. A rule
    // for that same field is the edited replacement, not an additional
    // predicate (the server permits only one predicate per field).
    const predicates = [
      ...(categoryPredicate && !rules.some((rule) => rule.field === categoryPredicate.field) ? [categoryPredicate] : []),
      ...rules,
    ]
    return {
      id: node.id,
      label: node.label,
      ...(node.branch ? { branch: { path: [...node.branch.path] } } : {}),
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

/** Branch slots create new scoped categories; existing categories stay in the normal tree. */
export function acceptsTaxonomyBranchDrop(payload: TaxonomyDragPayload | undefined): boolean {
  return payload?.kind === 'palette' && payload.taxonomyType === 'category'
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
  targetResolution?: TaxonomyTargetResolution
  branch?: { path: string[] }
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

export function taxonomyBranchContainer(branchPath: string, field?: string) {
  return `${taxonomyEditorContainer}:branch:${branchPath}${field ? `:field:${field}` : ''}`
}

export function flattenTaxonomyDraft(node: TaxonomyDraftNode): TaxonomyPlacement[] {
  return [...node.placements, ...node.children.flatMap(flattenTaxonomyDraft)]
}

export function replaceTaxonomyDraftPlacements(node: TaxonomyDraftNode, placements: TaxonomyPlacement[]): TaxonomyDraftNode {
  const existingIDs = new Set(flattenTaxonomyDraft(node).map((placement) => placement.id))
  return replace(node, placements, existingIDs)
}

export function findTaxonomyDraftNode(node: TaxonomyDraftNode, placementID: string): TaxonomyDraftNode | undefined {
  if (node.id === placementID) return node
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

/** Creates a branch-scoped category below the active taxonomy node. */
export function addTaxonomyDraftBranchChild(node: TaxonomyDraftNode, parentID: string, id: string, branchPath: string, label = 'New category', branchIndex?: number): TaxonomyDraftNode {
  if (!parentID.trim() || !id.trim() || !branchPath.trim()) return node
  const placement: TaxonomyPlacement = { id, targetId: branchPath, type: 'category', title: label, summary: `Branch ${branchPath}`, label, nodeId: id }
  const child = { id, label, targetId: branchPath, branch: { path: [branchPath] }, placements: [placement], rules: [], children: [] }
  return addTaxonomyDraftBranchNode(node, parentID, child, branchPath, branchIndex)
}

/** Creates a branch-scoped category with its first captured-field predicate. */
export function addTaxonomyDraftBranchFieldChild(node: TaxonomyDraftNode, parentID: string, id: string, branchPath: string, field: string, label = 'New category', branchIndex?: number): TaxonomyDraftNode {
  if (!field.trim()) return node
  const predicate = `${field} = A`
  const placement: TaxonomyPlacement = { id, targetId: field, type: 'category', title: label, summary: predicate, label, nodeId: id, predicate, predicateData: { kind: 'exact', field, value: 'A' } }
  return addTaxonomyDraftBranchNode(node, parentID, { id, label, targetId: branchPath, branch: { path: [branchPath] }, placements: [placement], rules: [], children: [] }, branchPath, branchIndex)
}

function addTaxonomyDraftBranchNode(node: TaxonomyDraftNode, parentID: string, child: TaxonomyDraftNode, branchPath: string, branchIndex?: number): TaxonomyDraftNode {
  if (node.id === parentID) {
    const branchChildren = node.children.filter((current) => taxonomyDraftBranchPath(current) === branchPath)
    const localIndex = branchIndex === undefined ? branchChildren.length : branchIndex
    if (!Number.isInteger(localIndex) || localIndex < 0 || localIndex > branchChildren.length) return node
    const insertionIndex = localIndex === branchChildren.length
      ? (branchChildren.length ? node.children.indexOf(branchChildren[branchChildren.length - 1]) + 1 : node.children.length)
      : node.children.indexOf(branchChildren[localIndex])
    const children = [...node.children]
    children.splice(insertionIndex, 0, child)
    return { ...node, children }
  }
  return { ...node, children: node.children.map((current) => addTaxonomyDraftBranchNode(current, parentID, child, branchPath, branchIndex)) }
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
