export type Route =
  | { kind: 'parts' }
  | { kind: 'taxonomy'; segments: string[] }
  | { kind: 'part'; id: string }
  | { kind: 'revision'; partNumber: string; id: string }
  | { kind: 'not-found' }

export function parseRoute(pathname = window.location.pathname): Route {
  const parts = pathname.split('/').filter(Boolean).map(decodeURIComponent)
  if (parts.length === 1 && parts[0] === 'parts') return { kind: 'parts' }
  if (parts.length >= 2 && parts[0] === 'parts') {
    return { kind: 'taxonomy', segments: parts.slice(1).length > 0 ? parts.slice(1) : ['components'] }
  }
  if (parts.length === 2 && parts[0] === 'part') return { kind: 'part', id: parts[1] }
  if (parts.length === 4 && parts[0] === 'part' && parts[2] === 'revision') {
    return { kind: 'revision', partNumber: parts[1], id: parts[3] }
  }
  return { kind: 'not-found' }
}

export const taxonomySlug = (label: string) =>
  label.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')

export const taxonomyPath = (labels: string[]) =>
  `/parts/${labels.map(taxonomySlug).map(encodeURIComponent).join('/')}`

export const partPath = (value: string) => `/part/${encodeURIComponent(value)}`

export const revisionPath = (partNumber: string, id: string) =>
  `/part/${encodeURIComponent(partNumber)}/revision/${encodeURIComponent(id)}`
