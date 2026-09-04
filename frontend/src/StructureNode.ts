export type StructureNode = {
  id: string | number
  label: string
  blockType: string
  summary?: string
  /** Frontend-only grammar metadata; never serialized in taxonomy API input. */
  path?: string
  choiceAncestry?: { path: string; label: string; summary: string }[]
  alternativeLabel?: string
  alternativeSummary?: string
  capturedFields?: string[]
  targetable?: boolean
  literalOnly?: boolean
  numericDomain?: { min: number; max: number }
  children: StructureNode[]
}
