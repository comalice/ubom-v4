<script lang="ts">
  import { partPath, revisionPath } from './routes'

  export let partNumber: string
  export let revision: string | undefined = undefined
  export let revisionID: string | undefined = undefined

  let feedback: 'idle' | 'copied' | 'failed' = 'idle'

  $: label = revision ? `${partNumber} @ ${revision}` : partNumber
  $: href = revision ? revisionPath(partNumber, revisionID ?? revision) : partPath(partNumber)

  async function copyIdentity() {
    try {
      if (!navigator.clipboard) throw new Error('Clipboard unavailable')
      await navigator.clipboard.writeText(label)
      feedback = 'copied'
    } catch {
      feedback = 'failed'
    }
    window.setTimeout(() => feedback = 'idle', 1800)
  }
</script>

<span class="identity-link">
  <a href={href} aria-label={revision ? `Open part ${partNumber}, revision ${revision}` : `Open part ${partNumber}`}>
    {label}
  </a>
  <button
    type="button"
    class="copy-button"
    aria-label={`Copy ${label}`}
    title={feedback === 'copied' ? 'Copied' : feedback === 'failed' ? 'Copy failed' : `Copy ${label}`}
    on:click|stopPropagation={copyIdentity}
  >
    {#if feedback === 'copied'}
      <span aria-hidden="true" class="copy-feedback">OK</span>
    {:else}
      <svg aria-hidden="true" width="13" height="13" viewBox="0 0 16 16" fill="none">
        <rect x="5.5" y="5.5" width="7" height="7" rx="1" stroke="currentColor" stroke-width="1.2" />
        <path d="M10.5 5.5V4.5C10.5 3.95 10.05 3.5 9.5 3.5H4.5C3.95 3.5 3.5 3.95 3.5 4.5V9.5C3.5 10.05 3.95 10.5 4.5 10.5H5.5" stroke="currentColor" stroke-width="1.2" />
      </svg>
    {/if}
  </button>
  <span class="sr-only" role="status" aria-live="polite">
    {feedback === 'copied' ? `${label} copied` : feedback === 'failed' ? `Unable to copy ${label}` : ''}
  </span>
</span>
