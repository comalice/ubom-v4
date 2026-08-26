import { describe, expect, it } from 'vitest'
import { parseRoute } from './routes'

describe('sequence definition routes', () => {
  it('parses the new and existing lifecycle pages', () => {
    expect(parseRoute('/settings/sequence-definitions/new')).toEqual({ kind: 'seq-def-new' })
    expect(parseRoute('/settings/sequence-definitions/part-number-v1')).toEqual({ kind: 'seq-def', id: 'part-number-v1' })
  })
})
