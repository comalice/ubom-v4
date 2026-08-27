<script lang="ts">
  export let state: 'loading' | 'error' | 'empty'
  export let message = ''
</script>

{#if state === 'loading'}
  <p class="muted" role="status" aria-live="polite">{message}</p>
{:else if state === 'error'}
  <p class="error" role="alert">{message}</p>
{:else}
  <p class="empty-state">{message}</p>
{/if}
