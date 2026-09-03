import { afterEach, describe, expect, it, vi } from 'vitest'
import { parseRoute, taxonomyPath } from './routes'

afterEach(() => vi.unstubAllGlobals())

describe('sequence definition routes', () => {
  it('parses the home page', () => {
    expect(parseRoute('/')).toEqual({ kind: 'home' })
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
})
