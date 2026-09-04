export type EventSourceTarget = Pick<Window, 'addEventListener' | 'removeEventListener'>
export type ClosableEventSource = Pick<EventSource, 'close'>

export function closeEventSourceOnPageHide(source: ClosableEventSource, target: EventSourceTarget): () => void {
  let closed = false
  const close = () => {
    if (closed) return
    closed = true
    target.removeEventListener('pagehide', close)
    source.close()
  }
  target.addEventListener('pagehide', close)
  return close
}
