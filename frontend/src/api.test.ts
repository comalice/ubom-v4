import { afterEach, describe, expect, it, vi } from 'vitest'
import { previewSeqDefRoot, type SeqDefNode } from './api'

afterEach(() => vi.restoreAllMocks())

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
