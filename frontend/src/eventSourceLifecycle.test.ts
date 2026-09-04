import { describe, expect, it } from 'vitest'
import { closeEventSourceOnPageHide, type EventSourceTarget } from './eventSourceLifecycle'

class FakePageLifecycle implements EventSourceTarget {
  private pagehideListeners = new Set<() => void>()

  addEventListener(type: string, listener: EventListenerOrEventListenerObject): void {
    if (type === 'pagehide' && typeof listener === 'function') this.pagehideListeners.add(listener as () => void)
  }

  removeEventListener(type: string, listener: EventListenerOrEventListenerObject): void {
    if (type === 'pagehide' && typeof listener === 'function') this.pagehideListeners.delete(listener as () => void)
  }

  dispatchPageHide() {
    for (const listener of [...this.pagehideListeners]) listener()
  }
}

describe('closeEventSourceOnPageHide', () => {
  it('closes every duplicate SSE connection when the page hides', () => {
    const target = new FakePageLifecycle()
    const sources = [{ close: () => undefined }, { close: () => undefined }]
    let closeCount = 0
    for (const source of sources) {
      source.close = () => { closeCount += 1 }
      closeEventSourceOnPageHide(source, target)
    }

    target.dispatchPageHide()
    target.dispatchPageHide()

    expect(closeCount).toBe(2)
  })

  it('makes explicit cleanup idempotent', () => {
    const target = new FakePageLifecycle()
    let closeCount = 0
    const close = closeEventSourceOnPageHide({ close: () => { closeCount += 1 } }, target)

    close()
    close()
    target.dispatchPageHide()

    expect(closeCount).toBe(1)
  })
})
