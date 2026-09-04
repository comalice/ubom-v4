export type UpdateNotice = {
  backendBuildID: string
  reason: 'updated' | 'mismatch'
}

export function updateNoticeForBuild(frontendBuildID: string, backendBuildID: string, reason: UpdateNotice['reason'] = 'updated'): UpdateNotice | null {
  if (!backendBuildID || frontendBuildID === backendBuildID) return null
  return { backendBuildID, reason }
}
