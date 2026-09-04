export type Route =
  | { kind: 'home' }
  | { kind: 'catalog'; state?: CatalogRouteState }
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

export type CatalogRouteState = {
  taxonomyDef?: string
  /** Human-readable taxonomy slugs from the catalog root to the current node. */
  path?: string[]
  node?: string
  q?: string
  filters?: Record<string, string[]>
  ranges?: Record<string, CatalogRange>
  page?: number
  pageSize?: number
  sort?: 'value' | 'id'
  order?: 'asc' | 'desc'
}

export type CatalogRange = { min?: string; max?: string }

const catalogDefaults = {
  page: 1,
  pageSize: 50,
  sort: 'value' as const,
  order: 'asc' as const,
}

export const catalogPageSizeOptions = [25, 50, 100] as const
export const catalogServerMaxPageSize = 200
const catalogSortValues = ['value', 'id'] as const
const numericValue = /^[-+]?(?:\d+(?:\.\d*)?|\.\d+)$/

const positiveInteger = (value: string | null, fallback: number) => {
  if (!value || !/^\d+$/.test(value)) return fallback
  const parsed = Number(value)
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : fallback
}

const validPageSize = (value: string | null) => {
  const parsed = positiveInteger(value, catalogDefaults.pageSize)
  return catalogPageSizeOptions.includes(parsed as (typeof catalogPageSizeOptions)[number]) || parsed === catalogServerMaxPageSize
    ? parsed
    : catalogDefaults.pageSize
}

export function parseCatalogRouteState(search: string | URLSearchParams): CatalogRouteState {
  const params = typeof search === 'string' ? new URLSearchParams(search) : search
  const filters: Record<string, string[]> = {}
  const ranges: Record<string, CatalogRange> = {}
  for (const [key, value] of params) {
    if (key.startsWith('filter.') && key.length > 'filter.'.length && value) {
      const attribute = key.slice('filter.'.length)
      filters[attribute] = [...new Set([...(filters[attribute] ?? []), value])].sort()
    }
  }
  for (const [key, value] of params) {
    const prefix = key.startsWith('attributeMin.') ? 'attributeMin.' : key.startsWith('attributeMax.') ? 'attributeMax.' : ''
    if (!prefix || key.length === prefix.length || !value || !numericValue.test(value)) continue
    const attribute = key.slice(prefix.length)
    ranges[attribute] = { ...ranges[attribute], [prefix === 'attributeMin.' ? 'min' : 'max']: value }
  }

  const page = positiveInteger(params.get('page'), catalogDefaults.page)
  const pageSize = validPageSize(params.get('pageSize'))
  const order = params.get('order') === 'desc' ? 'desc' : catalogDefaults.order
  const sortParam = params.get('sort')
  const sort: CatalogRouteState['sort'] = sortParam === 'value' || sortParam === 'id' ? sortParam : catalogDefaults.sort
  return {
    ...(params.get('taxonomyDef') ? { taxonomyDef: params.get('taxonomyDef')! } : {}),
    ...(params.get('node') ? { node: params.get('node')! } : {}),
    ...(params.get('q') ? { q: params.get('q')! } : {}),
    ...(Object.keys(filters).length ? { filters } : {}),
    ...(Object.keys(ranges).length ? { ranges } : {}),
    ...(page !== catalogDefaults.page ? { page } : {}),
    ...(pageSize !== catalogDefaults.pageSize ? { pageSize } : {}),
    ...(sort !== catalogDefaults.sort ? { sort } : {}),
    ...(order !== catalogDefaults.order ? { order } : {}),
  }
}

export function serializeCatalogRouteState(state: CatalogRouteState = {}): string {
  const params = new URLSearchParams()
  if (state.taxonomyDef) params.set('taxonomyDef', state.taxonomyDef)
  if (state.node) params.set('node', state.node)
  if (state.q) params.set('q', state.q)
  for (const key of Object.keys(state.filters ?? {}).sort()) {
    if (!key) continue
    for (const value of [...new Set(state.filters?.[key] ?? [])].filter(Boolean).sort()) params.append(`filter.${key}`, value)
  }
  for (const key of Object.keys(state.ranges ?? {}).sort()) {
    if (!key) continue
    const range = state.ranges?.[key]
    if (range?.min && numericValue.test(range.min)) params.set(`attributeMin.${key}`, range.min)
    if (range?.max && numericValue.test(range.max)) params.set(`attributeMax.${key}`, range.max)
  }
  if (Number.isSafeInteger(state.page) && state.page! > 0 && state.page !== catalogDefaults.page) params.set('page', String(state.page))
  if (Number.isSafeInteger(state.pageSize) && state.pageSize! > 0 && (catalogPageSizeOptions.includes(state.pageSize as (typeof catalogPageSizeOptions)[number]) || state.pageSize === catalogServerMaxPageSize) && state.pageSize !== catalogDefaults.pageSize) params.set('pageSize', String(state.pageSize))
  if (state.sort && catalogSortValues.includes(state.sort) && state.sort !== catalogDefaults.sort) params.set('sort', state.sort)
  if ((state.order === 'asc' || state.order === 'desc') && state.order !== catalogDefaults.order) params.set('order', state.order)
  return params.toString()
}

export function parseRoute(pathname?: string): Route {
  const currentPath = pathname ?? (typeof window !== 'undefined' ? `${window.location.pathname}${window.location.search}` : '/')
  const url = new URL(currentPath, 'http://localhost')
  const parts = url.pathname.split('/').filter(Boolean).map(decodeURIComponent)
  if (parts.length === 0) return { kind: 'home' }
  if (parts.length >= 1 && parts[0] === 'catalog') {
    const state = parseCatalogRouteState(url.searchParams)
    const path = parts.slice(1)
    const catalogState = path.length ? { ...state, path } : state
    return Object.keys(catalogState).length ? { kind: 'catalog', state: catalogState } : { kind: 'catalog' }
  }
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

export const catalogPath = (state?: CatalogRouteState) => {
  const query = serializeCatalogRouteState(state)
  const path = state?.path?.length ? `/catalog/${state.path.map(segment => encodeURIComponent(segment)).join('/')}` : '/catalog'
  return query ? `${path}?${query}` : path
}

export const partPath = (value: string) => `/part/${encodeURIComponent(value)}`

export const revisionPath = (partNumber: string, id: string) =>
  `/part/${encodeURIComponent(partNumber)}/revision/${encodeURIComponent(id)}`
