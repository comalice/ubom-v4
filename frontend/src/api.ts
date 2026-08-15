import { taxonomySlug } from './routes'

export type TaxonomyChild = { id: string; label: string }
export type TaxonomyPathItem = { id: string; label: string }
export type PartSummary = { id: string; value: string }

export type TaxonomyNodeView = {
  id: string
  label: string
  path: TaxonomyPathItem[]
  children: TaxonomyChild[]
  partNumbers: PartSummary[]
}

export type PartNumberView = {
  id: string
  value: string
  taxonomyPath: string[]
  revisions: { id: string; revision: string; bom: { partNumberId: string; partRevisionId: string }[] }[]
}

export type BomNode = {
  partNumber: PartSummary
  revisionId: string
  revision: string
  bom: BomNode[]
}

export type RevisionView = {
  id: string
  revision: string
  partNumber: PartSummary
  taxonomyPath: string[]
  bom: BomNode[]
}

async function get<T>(path: string): Promise<T> {
  const response = await fetch(path)
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`)
  return response.json() as Promise<T>
}

export const getTaxonomyNode = (taxonomyID: string, nodeID: string) =>
  get<TaxonomyNodeView>(`/api/taxonomies/${encodeURIComponent(taxonomyID)}/nodes/${encodeURIComponent(nodeID)}`)

export async function getTaxonomyNodeByPath(segments: string[]): Promise<TaxonomyNodeView> {
  const taxonomyID = 'sample-taxonomy-v1'
  let node = await getTaxonomyNode(taxonomyID, segments[0])
  for (const segment of segments.slice(1)) {
    const child = node.children.find(item => taxonomySlug(item.label) === segment)
    if (!child) throw new Error('404 Not Found')
    node = await getTaxonomyNode(taxonomyID, child.id)
  }
  return node
}

export const getPart = (value: string) => get<PartNumberView>(`/api/parts/by-value/${encodeURIComponent(value)}`)

export const getRevision = (id: string) => get<RevisionView>(`/api/revisions/${encodeURIComponent(id)}`)
