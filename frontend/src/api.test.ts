import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiVersionMismatchError, BackendVersionProbeError, createSeqDefVersion, getCatalogFacets, getCatalogRoot, getPart, getSeqDef, getSeqDefs, getTaxonomyDef, getTaxonomyDefHistory, getTaxonomyDefs, getCurrentTaxonomyDef, getTaxonomyNode, getTaxonomyNodeByPath, getTaxonomyNodeBySlugPath, previewSeqDefCandidate, previewSeqDefRoot, probeBackendVersion, resolveCatalogPath, resolveTaxonomyPath, searchParts, type SeqDefNode, type TaxonomyDefinitionView, type TaxonomyNodeView, type TaxonomyDefSummary } from './api'
import { catalogRootViewModel } from './catalogRoot'

afterEach(() => vi.restoreAllMocks())

describe('backend compatibility probe', () => {
  it('accepts a matching build identity', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ buildId: 'dev' }), { status: 200, headers: { 'X-UBOM-Build-ID': 'dev' } }))
    await expect(probeBackendVersion()).resolves.toBe('dev')
  })

  it('reports a changed build identity distinctly', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ buildId: 'release-2' }), { status: 200 }))
    await expect(probeBackendVersion()).rejects.toBeInstanceOf(ApiVersionMismatchError)
  })

  it('reports a missing version endpoint as a compatibility failure', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('not found', { status: 404 }))
    await expect(probeBackendVersion()).rejects.toBeInstanceOf(BackendVersionProbeError)
  })

  it('keeps ordinary API errors as ordinary errors', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: 'not found' }), { status: 404, statusText: 'Not Found' }))
    const error = await getPart('missing').catch(reason => reason)
    expect(error).toBeInstanceOf(Error)
    expect(error).not.toBeInstanceOf(BackendVersionProbeError)
    expect((error as Error).message).toBe('404 Not Found')
  })
})

describe('previewSeqDefRoot', () => {
  const root: SeqDefNode = { kind: 'literal', text: 'PN-' }

  it('posts the root and explicit bindings and returns the preview', async () => {
    const response = { id: 'preview', bindings: { family: 'SN-' }, value: 'PN-SN-' }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }))

    await expect(previewSeqDefRoot(root, { family: 'SN-' })).resolves.toEqual(response)
    expect(fetchMock).toHaveBeenCalledWith('/api/seq-defs/preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ root, bindings: { family: 'SN-' } }),
    })
  })

  it('preserves string padding in nested range nodes', async () => {
    const rangeRoot: SeqDefNode = {
      kind: 'concat',
      children: [
        { kind: 'range_radix', children: [{ kind: 'range', min: 0, max: 9, width: 2, pad: ' ' }] },
        { kind: 'bind', name: 'suffix', child: { kind: 'range', min: 0, max: 9, width: 2, pad: 'X' } },
      ],
    }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ id: 'preview', bindings: {} }), { status: 200 }))

    await previewSeqDefRoot(rangeRoot, {})

    expect(JSON.parse(fetchMock.mock.calls[0][1]?.body as string)).toEqual({ root: rangeRoot, bindings: {} })
  })

  it('uses the server error message for failed requests', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: 'invalid sequence' }), { status: 400, statusText: 'Bad Request' }))

    await expect(previewSeqDefRoot(root, {})).rejects.toThrow('invalid sequence')
  })
})

describe('sequence definition authoring API', () => {
  it('lists definitions and fetches the complete authoring document', async () => {
    const definition = { id: 'full-view', definitionId: 'full-view', version: 1, root: { kind: 'range' as const, min: 0, max: 9, width: 2, pad: ' ' } }
    const fetchMock = vi.spyOn(globalThis, 'fetch')
       .mockResolvedValueOnce(new Response(JSON.stringify([{ id: definition.id, definitionId: definition.definitionId, version: definition.version }]), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(definition), { status: 200 }))

    await expect(getSeqDefs()).resolves.toEqual([{ id: definition.id, definitionId: definition.definitionId, version: definition.version }])
    await expect(getSeqDef(definition.id)).resolves.toEqual(definition)
    expect(fetchMock.mock.calls.map(([path]) => path)).toEqual(['/api/seq-defs', '/api/seq-defs/full-view'])
  })
})

describe('taxonomy revision history API', () => {
  it('loads the immutable revision summaries for a taxonomy lineage', async () => {
    const history = [{ id: 'taxonomy-v1', definitionId: 'taxonomy', version: 1, seqDef: 'seq-v1', current: false }]
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(history), { status: 200 }))

    await expect(getTaxonomyDefHistory('taxonomy')).resolves.toEqual(history)
    expect(fetchMock).toHaveBeenCalledWith('/api/taxonomy-defs/taxonomy/history')
  })
})

describe('taxonomy node API', () => {
  it('resolves readable labels to stable IDs and rejects ambiguous segments', () => {
    const root = {
      id: 'components', label: 'Components', children: [
        { id: 'fasteners', label: 'Fasteners & Joining', children: [{ id: 'bolts', label: 'Bolts' }] },
        { id: 'duplicate', label: 'Fasteners & Joining', children: [] },
      ],
    }
    expect(resolveTaxonomyPath(root, ['components', 'bolts'])).toBeUndefined()
    expect(resolveTaxonomyPath(root, ['components', 'fasteners-joining'])).toBeUndefined()
    expect(resolveTaxonomyPath({ ...root, children: [root.children[0]] }, ['components', 'fasteners-joining', 'bolts'])).toEqual(['components', 'fasteners', 'bolts'])
  })

  it('uses resolved stable IDs for a readable path API lookup', async () => {
    const root = { id: 'components', label: 'Components', children: [{ id: 'fasteners', label: 'Fasteners & Joining', children: [] }] }
    const node = { id: 'fasteners', taxonomyDefId: 'components-v3', seqDefId: 'components-seq-v2', label: 'Fasteners & Joining', path: [], attributes: [], children: [], partNumbers: [] }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(node), { status: 200 }))

    await expect(getTaxonomyNodeBySlugPath('components-v3', root, ['components', 'fasteners-joining'])).resolves.toEqual(node)
    expect(fetchMock).toHaveBeenCalledWith('/api/taxonomies/components-v3/nodes/fasteners')
  })

  it('uses the resolved taxonomy definition ID for every node lookup', async () => {
    const root = { id: 'components', label: 'Components', children: [{ id: 'resistors', label: 'Resistors' }] }
    const resistor = { ...root, id: 'resistors', label: 'Resistors', children: [] }
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify(root), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(resistor), { status: 200 }))

    await expect(getTaxonomyNodeByPath('electronics-taxonomy-v3', ['components', 'resistors'])).resolves.toEqual(resistor)
    expect(fetchMock.mock.calls.map(([path]) => path)).toEqual([
      '/api/taxonomies/electronics-taxonomy-v3/nodes/components',
      '/api/taxonomies/electronics-taxonomy-v3/nodes/resistors',
    ])
  })
})

describe('catalog taxonomy API', () => {
  it('fetches taxonomy summaries, the current taxonomy, a detail view, and a node', async () => {
    const summary: TaxonomyDefSummary = { id: 'components-v3', definitionId: 'components', version: 3, seqDef: 'components-seq-v2', current: true }
    const definition: TaxonomyDefinitionView = { ...summary, parentId: undefined, root: { id: 'components', label: 'Components' }, seqDef: summary.seqDef, current: true }
    const node: TaxonomyNodeView = {
      id: 'resistors', taxonomyDefId: summary.id, seqDefId: summary.seqDef, label: 'Resistors', path: [{ id: 'resistors', label: 'Resistors' }],
      attributes: [], children: [], partNumbers: [],
    }
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify([summary]), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(definition), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(definition), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(node), { status: 200 }))

    await expect(getTaxonomyDefs()).resolves.toEqual([summary])
    await expect(getCurrentTaxonomyDef(summary.seqDef)).resolves.toEqual(definition)
    await expect(getTaxonomyDef(summary.id)).resolves.toEqual(definition)
    await expect(getTaxonomyNode(summary.id, node.id)).resolves.toEqual(node)
    expect(fetchMock.mock.calls.map(([path]) => path)).toEqual([
      '/api/taxonomy-defs',
      '/api/seq-defs/components-seq-v2/taxonomy',
      '/api/taxonomy-defs/components-v3',
      '/api/taxonomies/components-v3/nodes/resistors',
    ])
  })
})

describe('part search API', () => {
  it('encodes part-log filters and pagination', async () => {
    const result = { items: [], total: 0, page: 2, pageSize: 25 }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(result), { status: 200 }))

    await expect(searchParts({ q: 'R-10', taxonomyNode: 'resistors', page: 2, pageSize: 25 })).resolves.toEqual(result)
    expect(fetchMock).toHaveBeenCalledWith('/api/parts?q=R-10&taxonomyNode=resistors&page=2&pageSize=25')
  })

  it('serializes catalog state and repeated exact attribute filters', async () => {
    const result = { items: [], total: 0, page: 3, pageSize: 50 }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(result), { status: 200 }))

    await expect(searchParts({
       taxonomyDef: 'components/v3', node: 'resistors', q: '10k +',
       filters: { tolerance: ['1%', '5%'], footprint: ['0603'] },
       ranges: { resistance: { min: '100', max: '10000' } },
      page: 3, pageSize: 50, sort: 'id', order: 'desc',
    })).resolves.toEqual(result)
    expect(fetchMock).toHaveBeenCalledWith('/api/parts?q=10k+%2B&taxonomyNode=resistors&taxonomyDef=components%2Fv3&attribute=footprint%3D0603&attribute=tolerance%3D1%25&attribute=tolerance%3D5%25&attributeMin.resistance=100&attributeMax.resistance=10000&page=3&pageSize=50&sort=id&order=desc')
  })

  it('sends explicit archived inclusion state when requested', async () => {
    const result = { items: [], total: 0, page: 1, pageSize: 50 }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockImplementation(async () => new Response(JSON.stringify(result), { status: 200 }))

    await searchParts({ includeArchived: false })
    await searchParts({ includeArchived: true })

    expect(fetchMock.mock.calls.map(([path]) => path)).toEqual([
      '/api/parts?includeArchived=false',
      '/api/parts?includeArchived=true',
    ])
  })
})

describe('catalog facets API', () => {
  it('encodes the catalog scope and active filters without pagination', async () => {
    const result = { categories: [{ id: 'bolts', label: 'Bolts', count: 3 }], attributes: [] }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(result), { status: 200 }))

    await expect(getCatalogFacets({ taxonomyDef: 'components/v3', node: 'fasteners', q: 'bolt', includeArchived: true, filters: { material: ['Steel', 'Aluminum'] }, ranges: { resistance: { min: '100' } }, page: 4 })).resolves.toEqual(result)
    expect(fetchMock).toHaveBeenCalledWith('/api/catalog/facets?q=bolt&taxonomyNode=fasteners&taxonomyDef=components%2Fv3&includeArchived=true&attribute=material%3DSteel&attribute=material%3DAluminum&attributeMin.resistance=100')
  })
})

describe('catalog root API and view model', () => {
  it('loads grouped root data and creates taxonomy-qualified links', async () => {
    const root = { groups: [{ taxonomyDefId: 'sample-v1', seqDefId: 'sample-seq', root: { id: 'components', label: 'Components' }, children: [{ id: 'resistors', label: 'Resistors', count: 2 }] }] }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(root), { status: 200 }))
    await expect(getCatalogRoot()).resolves.toEqual(root)
    expect(fetchMock).toHaveBeenCalledWith('/api/catalog/root')
    expect(catalogRootViewModel(root).map(group => ({ heading: group.headingHref, child: group.children[0].href, label: group.children[0].label }))).toEqual([{
      heading: '/catalog?taxonomyDef=sample-v1', child: '/catalog/resistors?taxonomyDef=sample-v1', label: 'Resistors',
    }])
  })

  it('resolves catalog paths relative to the taxonomy root', () => {
    const root = { id: 'components', label: 'Components', children: [{ id: 'resistors', label: 'Resistors' }] }
    expect(resolveCatalogPath(root, [])).toEqual(['components'])
    expect(resolveCatalogPath(root, ['resistors'])).toEqual(['components', 'resistors'])
    expect(resolveCatalogPath(root, ['components'])).toBeUndefined()
  })

  it('passes explicit archived state only when requested', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{"groups":[]}', { status: 200 }))
    await getCatalogRoot(true)
    expect(fetchMock).toHaveBeenCalledWith('/api/catalog/root?includeArchived=true')
  })
})

describe('sequence definition version API', () => {
  it('posts a root to the parent version endpoint', async () => {
    const root: SeqDefNode = { kind: 'literal', text: 'PN-2' }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ id: 'base-v2' }), { status: 201 }))
    await createSeqDefVersion('base', root)
    expect(fetchMock).toHaveBeenCalledWith('/api/seq-defs/base/versions', expect.objectContaining({ method: 'POST', body: JSON.stringify({ root }) }))
  })
})

describe('previewSeqDefCandidate', () => {
  it('posts the complete candidate value separately from binding generation', async () => {
    const response = { value: 'PN-SN-B0', bindings: { prefix: 'PN' }, fields: [{ name: 'prefix', source: 'PN', normalized: null }] }
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }))

    await expect(previewSeqDefCandidate({ kind: 'literal', text: 'PN-' }, 'PN-SN-B0')).resolves.toEqual(response)
    expect(fetchMock).toHaveBeenCalledWith('/api/seq-defs/preview/candidate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ root: { kind: 'literal', text: 'PN-' }, value: 'PN-SN-B0' }),
    })
  })

  it('uses the server error message for an invalid candidate', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: 'candidate does not match' }), { status: 400 }))

    await expect(previewSeqDefCandidate({ kind: 'literal', text: 'PN-' }, 'invalid')).rejects.toThrow('candidate does not match')
  })
})
