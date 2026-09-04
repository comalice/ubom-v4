import { afterEach, describe, expect, it, vi } from 'vitest'
import { catalogPath, parseCatalogRouteState, parseRoute, serializeCatalogRouteState, taxonomyPath } from './routes'

afterEach(() => vi.unstubAllGlobals())

describe('sequence definition routes', () => {
  it('parses the home page', () => {
    expect(parseRoute('/')).toEqual({ kind: 'home' })
  })
  it('parses the catalog page', () => {
    expect(parseRoute('/catalog')).toEqual({ kind: 'catalog' })
  })

  it('parses readable catalog taxonomy paths separately from query state', () => {
    expect(parseRoute('/catalog/components/fasteners?taxonomyDef=components-v3&q=bolt&page=2')).toEqual({
      kind: 'catalog',
      state: { path: ['components', 'fasteners'], taxonomyDef: 'components-v3', q: 'bolt', page: 2 },
    })
  })

  it('parses shareable catalog query state and omits defaults', () => {
    expect(parseRoute('/catalog?taxonomyDef=components-v3&node=resistors&q=10k%20%2B&filter.resistance=10k&filter.tolerance=1%25&page=2&pageSize=50&sort=updatedAt&order=desc')).toEqual({
      kind: 'catalog',
      state: {
        taxonomyDef: 'components-v3',
        node: 'resistors',
        q: '10k +',
        filters: { resistance: ['10k'], tolerance: ['1%'] },
        page: 2,
        sort: 'updatedAt',
        order: 'desc',
      },
    })
    expect(parseCatalogRouteState('?page=1&pageSize=50&sort=value&order=asc')).toEqual({})
    expect(parseCatalogRouteState('?page=1&pageSize=25&sort=value&order=asc')).toEqual({ pageSize: 25 })
  })

  it('serializes catalog state deterministically and preserves it on a round trip', () => {
    const state = { taxonomyDef: 'components-v3', filters: { material: ['Aluminum', 'Steel'], tolerance: ['1%'], resistance: ['10k'] }, q: '10k +', page: 2, order: 'desc' as const }
    expect(serializeCatalogRouteState(state)).toBe('taxonomyDef=components-v3&q=10k+%2B&filter.material=Aluminum&filter.material=Steel&filter.resistance=10k&filter.tolerance=1%25&page=2&order=desc')
    expect(catalogPath(state)).toBe('/catalog?taxonomyDef=components-v3&q=10k+%2B&filter.material=Aluminum&filter.material=Steel&filter.resistance=10k&filter.tolerance=1%25&page=2&order=desc')
    expect(serializeCatalogRouteState({ pageSize: 50 })).toBe('')
    expect(parseCatalogRouteState(serializeCatalogRouteState(state))).toEqual(state)
    expect(catalogPath()).toBe('/catalog')
    expect(catalogPath({ path: ['components', 'fasteners'], taxonomyDef: 'components-v3', q: 'bolt' })).toBe('/catalog/components/fasteners?taxonomyDef=components-v3&q=bolt')
  })

  it('serializes every catalog interaction state, including pagination and sorting', () => {
    const state = {
      taxonomyDef: 'components-v3',
      node: 'resistors',
      q: '10k',
      filters: { material: ['Steel'] },
      page: 3,
      pageSize: 25,
      sort: 'updatedAt',
      order: 'desc' as const,
    }
    expect(parseCatalogRouteState(new URL(catalogPath(state), 'http://localhost').search)).toEqual(state)
  })
  it('parses the list page', () => {
    expect(parseRoute('/settings/sequence-definitions')).toEqual({ kind: 'seq-def-list' })
  })

  it('keeps the new and existing lifecycle pages distinct from the list', () => {
    expect(parseRoute('/settings/sequence-definitions/new')).toEqual({ kind: 'seq-def-new' })
    expect(parseRoute('/settings/sequence-definitions/part-number-v1')).toEqual({ kind: 'seq-def', id: 'part-number-v1' })
  })
})

describe('taxonomy routes', () => {
  it('parses the list and new pages', () => {
    expect(parseRoute('/settings/taxonomies')).toEqual({ kind: 'taxonomy-list' })
    expect(parseRoute('/settings/taxonomies/new')).toEqual({ kind: 'taxonomy-new' })
    expect(parseRoute('/settings/taxonomies/example-taxonomy')).toEqual({ kind: 'taxonomy-detail', id: 'example-taxonomy' })
  })

  it('carries an explicit sequence definition context on parts taxonomy URLs', () => {
    expect(parseRoute('/parts/components/resistors?seqDef=electronics-v2')).toEqual({ kind: 'taxonomy', segments: ['components', 'resistors'], seqDef: 'electronics-v2' })
    expect(taxonomyPath(['Components', 'Resistors'], 'electronics-v2')).toBe('/parts/components/resistors?seqDef=electronics-v2')
  })

  it('preserves the sequence definition when parsing the current browser URL', () => {
    vi.stubGlobal('window', { location: { pathname: '/settings/taxonomies/new', search: '?seqDef=sample-pn-v1' } })
    expect(parseRoute()).toEqual({ kind: 'taxonomy-new', seqDef: 'sample-pn-v1' })
  })
})

describe('playground routes', () => {
  it('parses the core components page', () => {
    expect(parseRoute('/playground/core-components')).toEqual({ kind: 'playground', page: 'core-components' })
  })

  it('parses the HTML primitives page', () => {
    expect(parseRoute('/playground/html-primitives')).toEqual({ kind: 'playground', page: 'html-primitives' })
  })

  it('parses the taxonomy editor page', () => {
    expect(parseRoute('/playground/taxonomy-editor')).toEqual({ kind: 'playground', page: 'taxonomy-editor' })
  })

  it('parses the taxonomy editor components page', () => {
    expect(parseRoute('/playground/taxonomy-editor-components')).toEqual({ kind: 'playground', page: 'taxonomy-editor-components' })
  })

  it('parses the catalog playground pages', () => {
    expect(parseRoute('/playground/catalog-components')).toEqual({ kind: 'playground', page: 'catalog-components' })
    expect(parseRoute('/playground/catalog')).toEqual({ kind: 'playground', page: 'catalog' })
  })

  it('parses the part views playground page', () => {
    expect(parseRoute('/playground/part-views')).toEqual({ kind: 'playground', page: 'part-views' })
  })
})
