export type TaxonomyChild = { id: string; label: string }
export type PartSummary = { id: string; value: string }

export type TaxonomyNodeView = {
  id: string
  label: string
  children: TaxonomyChild[]
  partNumbers: PartSummary[]
}

export type PartNumberView = {
  id: string
  value: string
  taxonomyPath: string[]
  revisions: { id: string; bom: { partNumberId: string; partRevisionId: string }[] }[]
}

export type BomNode = {
  partNumber: PartSummary
  revisionId: string
  bom: BomNode[]
}

export type RevisionView = {
  id: string
  partNumber: PartSummary
  bom: BomNode[]
}

async function get<T>(path: string): Promise<T> {
  const response = await fetch(path)
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`)
  return response.json() as Promise<T>
}

export const getTaxonomyNode = (taxonomyID: string, nodeID: string) =>
  get<TaxonomyNodeView>(`/api/taxonomies/${encodeURIComponent(taxonomyID)}/nodes/${encodeURIComponent(nodeID)}`)

export const getPart = (id: string) => get<PartNumberView>(`/api/parts/${encodeURIComponent(id)}`)

export const getRevision = (id: string) => get<RevisionView>(`/api/revisions/${encodeURIComponent(id)}`)
