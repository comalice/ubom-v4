import { taxonomySlug } from './routes'

export type TaxonomyChild = { id: string; label: string }
export type TaxonomyPathItem = { id: string; label: string }
export type PartSummary = { id: string; value: string }
export type RevisionSummary = { id: string; revision: string }
export type PartAttribute = { id: string; label: string; value: string; valueType: string; unit?: string }
export type AttributeColumn = { id: string; label: string; unit?: string }
export type PartListItem = {
  id: string
  value: string
  taxonomyPath: string[]
  attributes: PartAttribute[]
  revisions: RevisionSummary[]
}

export type TaxonomyNodeView = {
  id: string
  taxonomyDefId: string
  seqDefId: string
  label: string
  path: TaxonomyPathItem[]
  attributes: (AttributeColumn & { valueType: string; required: boolean })[]
  children: TaxonomyChild[]
  partNumbers: PartListItem[]
}

export type PartNumberView = {
  id: string
  value: string
  taxonomyPath: string[]
  attributes: { id: string; label: string; value: string; valueType: string; unit?: string }[]
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

export type SourceView = { id: string; name: string; kind: string }
export type ArtifactView = { id: string; name: string; kind: string; sourceId: string; path: string }
export type SnapshotView = {
  id: string
  artifactId: string
  reference: string
  commit?: string
  snapshotKey: string
  digest: string
  mediaType: string
  size: number
}
export type SeqDefView = { id: string }
export type TaxonomyDefSummary = { id: string; seqDef: string }

async function get<T>(path: string): Promise<T> {
  const response = await fetch(path)
  if (!response.ok) throw new Error(`${response.status} ${response.statusText}`)
  return response.json() as Promise<T>
}

async function post<T>(path: string, value: unknown): Promise<T> {
  const response = await fetch(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(value) })
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: string } | null
    throw new Error(body?.error ?? `${response.status} ${response.statusText}`)
  }
  return response.json() as Promise<T>
}

async function patch<T>(path: string, value: unknown): Promise<T> {
  const response = await fetch(path, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(value) })
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: string } | null
    throw new Error(body?.error ?? `${response.status} ${response.statusText}`)
  }
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

export const getParts = () => get<PartListItem[]>('/api/parts')

export const getRevision = (id: string) => get<RevisionView>(`/api/revisions/${encodeURIComponent(id)}`)

export const getSources = () => get<SourceView[]>('/api/artifact-sources')
export const getArtifacts = () => get<ArtifactView[]>('/api/artifacts')
export const getArtifact = (id: string) => get<ArtifactView>(`/api/artifacts/${encodeURIComponent(id)}`)
export const getSnapshots = (id: string) => get<SnapshotView[]>(`/api/artifacts/${encodeURIComponent(id)}/snapshots`)
export const getSeqDefs = () => get<SeqDefView[]>('/api/seq-defs')
export type SeqDefNode =
  | { kind: 'literal'; text: string }
  | { kind: 'range'; min: number; max: number; width?: number; pad?: number }
  | { kind: 'choice' | 'concat' | 'branch' | 'range_radix'; children: SeqDefNode[] }
  | { kind: 'bind'; name: string; child: SeqDefNode }
  | { kind: 'places'; alphabets: string[] }
export const createSeqDef = (id: string, root: SeqDefNode) => post<{ id: string }>('/api/seq-defs', { id, root })
export const previewSeqDef = (root: SeqDefNode, bindings: Record<string, string>) =>
  post<{ id: string; bindings: Record<string, string>; value?: string; error?: string }>('/api/seq-defs/preview', { root, bindings })
export const getTaxonomyDefs = () => get<TaxonomyDefSummary[]>('/api/taxonomy-defs')
export const createPart = (value: string, seqDef: string, taxonomyDef: string) =>
  post<{ id: string; value: string }>('/api/parts', { value, seqDef, taxonomyDef })
export type SeqDefField = { name: string; kind: string; options?: string[]; min?: number; max?: number; width?: number; pad?: string }
export const getSeqDefFields = (id: string) => get<SeqDefField[]>(`/api/seq-defs/${encodeURIComponent(id)}/fields`)
export const createPartDraft = (seqDef: string, taxonomyDef: string, contextNodeId: string) =>
  post<{ id: string; status: string }>('/api/part-drafts', { seqDef, taxonomyDef, contextNodeId })
export const updatePartDraft = (id: string, bindings: Record<string, string>, attributes: { attributeDefId: string; value: string }[]) =>
  patch<{ id: string }>(`/api/part-drafts/${encodeURIComponent(id)}`, { bindings, attributes })
export const previewPartDraft = (id: string) => post<{ id: string; complete: boolean; value?: string; error?: string; suggestedBindings?: Record<string, string> }>(`/api/part-drafts/${encodeURIComponent(id)}/preview`, {})
export const releasePartDraft = (id: string) => post<{ id: string; value: string }>(`/api/part-drafts/${encodeURIComponent(id)}/release`, {})
