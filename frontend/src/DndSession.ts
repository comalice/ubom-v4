import { writable } from 'svelte/store'

export type DndSession = {
  kind: 'new' | 'move'
  sourceContainer: string
}

export const dndSession = writable<DndSession | null>(null)

export function beginDndSession(kind: 'new' | 'move', sourceContainer: string) {
  dndSession.set({ kind, sourceContainer })
}

export function clearDndSession() {
  dndSession.set(null)
}
