import { describe, expect, it } from 'vitest'
import { parseRoute } from './routes'

describe('sequence definition routes', () => {
  it('parses the list page', () => {
    expect(parseRoute('/settings/sequence-definitions')).toEqual({ kind: 'seq-def-list' })
  })

  it('keeps the new and existing lifecycle pages distinct from the list', () => {
    expect(parseRoute('/settings/sequence-definitions/new')).toEqual({ kind: 'seq-def-new' })
    expect(parseRoute('/settings/sequence-definitions/part-number-v1')).toEqual({ kind: 'seq-def', id: 'part-number-v1' })
  })
})

describe('playground routes', () => {
  it('parses the core components page', () => {
    expect(parseRoute('/playground/core-components')).toEqual({ kind: 'playground', page: 'core-components' })
  })

  it('parses the HTML primitives page', () => {
    expect(parseRoute('/playground/html-primitives')).toEqual({ kind: 'playground', page: 'html-primitives' })
  })
})
