export type Route =
  | { kind: 'taxonomy'; taxonomyID: string; nodeID: string }
  | { kind: 'part'; id: string }
  | { kind: 'revision'; id: string }
  | { kind: 'not-found' }

export function parseRoute(pathname = window.location.pathname): Route {
  const parts = pathname.split('/').filter(Boolean).map(decodeURIComponent)
  if (parts.length === 4 && parts[0] === 'taxonomies' && parts[2] === 'nodes') {
    return { kind: 'taxonomy', taxonomyID: parts[1], nodeID: parts[3] }
  }
  if (parts.length === 2 && parts[0] === 'parts') return { kind: 'part', id: parts[1] }
  if (parts.length === 2 && parts[0] === 'revisions') return { kind: 'revision', id: parts[1] }
  return { kind: 'not-found' }
}

export const taxonomyPath = (taxonomyID: string, nodeID: string) =>
  `/taxonomies/${encodeURIComponent(taxonomyID)}/nodes/${encodeURIComponent(nodeID)}`

export const partPath = (id: string) => `/parts/${encodeURIComponent(id)}`

export const revisionPath = (id: string) => `/revisions/${encodeURIComponent(id)}`
