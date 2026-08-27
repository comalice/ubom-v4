<script lang="ts">
  import type { SnapshotView } from './api'
  import TableFrame from './TableFrame.svelte'

  export let snapshots: SnapshotView[] = []
</script>

<TableFrame><table class="revision-table"><caption class="sr-only">Artifact snapshots</caption><thead><tr><th scope="col">Reference</th><th scope="col">Digest</th><th scope="col">Size</th><th scope="col">Media type</th></tr></thead><tbody>{#each snapshots as snapshot}<tr><th scope="row">{snapshot.reference || 'Unspecified'}</th><td class="artifact-code">{snapshot.digest}</td><td>{snapshot.size} bytes</td><td>{snapshot.mediaType}</td></tr>{/each}</tbody></table></TableFrame>
