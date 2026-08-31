export type StructureNode = {
  id: string | number
  label: string
  blockType: string
  summary?: string
  /** Frontend-only domain used to disambiguate repeated numeric bindings. */
  numericDomain?: { min: number; max: number }
  children: StructureNode[]
}
