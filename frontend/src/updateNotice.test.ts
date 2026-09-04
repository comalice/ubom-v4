import { describe, expect, it } from 'vitest'
import { updateNoticeForBuild } from './updateNotice'

describe('updateNoticeForBuild', () => {
  it('returns no notice for a matching or missing backend identity', () => {
    expect(updateNoticeForBuild('build-1', 'build-1')).toBeNull()
    expect(updateNoticeForBuild('build-1', '')).toBeNull()
  })

  it('creates a notice for a changed deployment or explicit mismatch', () => {
    expect(updateNoticeForBuild('build-1', 'build-2')).toEqual({ backendBuildID: 'build-2', reason: 'updated' })
    expect(updateNoticeForBuild('build-1', 'build-2', 'mismatch')).toEqual({ backendBuildID: 'build-2', reason: 'mismatch' })
  })
})
