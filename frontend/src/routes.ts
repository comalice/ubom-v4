export type Route =
  | { kind: 'home' }
  | { kind: 'parts' }
  | { kind: 'artifacts' }
  | { kind: 'artifact'; id: string }
  | { kind: 'taxonomy'; segments: string[]; seqDef?: string }
  | { kind: 'part'; id: string }
  | { kind: 'revision'; partNumber: string; id: string }
  | { kind: 'settings' }
  | { kind: 'grammar-editor' }
  | { kind: 'seq-def-list' }
  | { kind: 'seq-def-new' }
  | { kind: 'seq-def'; id: string }
  | { kind: 'taxonomy-list' }
  | { kind: 'taxonomy-new'; seqDef?: string }
  | { kind: 'taxonomy-detail'; id: string }
  | { kind: 'playground'; page: 'core-components' | 'html-primitives' | 'dnd-ops' | 'block-examples' | 'grammar-blocks' | 'taxonomy-editor-components' | 'grammar-editor' | 'taxonomy-editor' | 'catalog-components' | 'catalog' | 'part-views' }
  | { kind: 'not-found' }

export function parseRoute(pathname?: string): Route {
  const currentPath = pathname ?? (typeof window !== 'undefined' ? `${window.location.pathname}${window.location.search}` : '/')
  const url = new URL(currentPath, 'http://localhost')
  const parts = url.pathname.split('/').filter(Boolean).map(decodeURIComponent)
  if (parts.length === 0) return { kind: 'home' }
  if (parts.length === 1 && parts[0] === 'parts') return { kind: 'parts' }
  if (parts.length === 1 && parts[0] === 'artifacts') return { kind: 'artifacts' }
  if (parts.length === 1 && parts[0] === 'settings') return { kind: 'settings' }
  if (parts.length === 3 && parts[0] === 'settings' && parts[1] === 'sequence-definitions' && parts[2] === 'grammar-editor') return { kind: 'grammar-editor' }
  if (parts.length === 2 && parts[0] === 'settings' && parts[1] === 'sequence-definitions') return { kind: 'seq-def-list' }
  if (parts.length === 3 && parts[0] === 'settings' && parts[1] === 'sequence-definitions' && parts[2] === 'new') return { kind: 'seq-def-new' }
  if (parts.length === 3 && parts[0] === 'settings' && parts[1] === 'sequence-definitions' && parts[2]) return { kind: 'seq-def', id: parts[2] }
  if (parts.length === 2 && parts[0] === 'settings' && parts[1] === 'taxonomies') return { kind: 'taxonomy-list' }
  if (parts.length === 3 && parts[0] === 'settings' && parts[1] === 'taxonomies' && parts[2] === 'new') {
    const selectedSeqDef = url.searchParams.get('seqDef')
    return { kind: 'taxonomy-new', ...(selectedSeqDef ? { seqDef: selectedSeqDef } : {}) }
  }
  if (parts.length === 3 && parts[0] === 'settings' && parts[1] === 'taxonomies' && parts[2]) return { kind: 'taxonomy-detail', id: parts[2] }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'dnd-ops') return { kind: 'playground', page: 'dnd-ops' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'block-examples') return { kind: 'playground', page: 'block-examples' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'grammar-blocks') return { kind: 'playground', page: 'grammar-blocks' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'taxonomy-editor-components') return { kind: 'playground', page: 'taxonomy-editor-components' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'grammar-editor') return { kind: 'playground', page: 'grammar-editor' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'taxonomy-editor') return { kind: 'playground', page: 'taxonomy-editor' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'core-components') return { kind: 'playground', page: 'core-components' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'html-primitives') return { kind: 'playground', page: 'html-primitives' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'catalog-components') return { kind: 'playground', page: 'catalog-components' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'catalog') return { kind: 'playground', page: 'catalog' }
  if (parts.length === 2 && parts[0] === 'playground' && parts[1] === 'part-views') return { kind: 'playground', page: 'part-views' }
  if (parts.length === 1 && parts[0] === 'playground') return { kind: 'playground', page: 'dnd-ops' }
  if (parts.length === 2 && parts[0] === 'artifacts') return { kind: 'artifact', id: parts[1] }
  if (parts.length >= 2 && parts[0] === 'parts') {
    const seqDef = url.searchParams.get('seqDef') ?? undefined
    return { kind: 'taxonomy', segments: parts.slice(1).length > 0 ? parts.slice(1) : ['components'], ...(seqDef ? { seqDef } : {}) }
  }
  if (parts.length === 2 && parts[0] === 'part') return { kind: 'part', id: parts[1] }
  if (parts.length === 4 && parts[0] === 'part' && parts[2] === 'revision') {
    return { kind: 'revision', partNumber: parts[1], id: parts[3] }
  }
  return { kind: 'not-found' }
}

export const taxonomySlug = (label: string) =>
  label.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')

export const taxonomyPath = (labels: string[], seqDef?: string) => {
  const path = `/parts/${labels.map(taxonomySlug).map(encodeURIComponent).join('/')}`
  return seqDef ? `${path}?seqDef=${encodeURIComponent(seqDef)}` : path
}

export const partPath = (value: string) => `/part/${encodeURIComponent(value)}`

export const revisionPath = (partNumber: string, id: string) =>
  `/part/${encodeURIComponent(partNumber)}/revision/${encodeURIComponent(id)}`
