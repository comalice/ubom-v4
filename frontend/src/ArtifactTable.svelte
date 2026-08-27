<script lang="ts">
  import type { ArtifactView } from './api'
  import TableFrame from './TableFrame.svelte'

  export let artifacts: ArtifactView[] = []
</script>

<TableFrame className="artifact-table-frame">
  <table class="revision-table">
    <caption class="sr-only">Artifacts</caption>
    <thead><tr><th scope="col">Name</th><th scope="col">Kind</th><th scope="col">Source</th><th scope="col">Path</th></tr></thead>
    <tbody>
      {#each artifacts as artifact}
        <tr>
          <th scope="row"><a href={`/artifacts/${encodeURIComponent(artifact.id)}`}>{artifact.name}</a></th>
          <td>{artifact.kind}</td><td class="artifact-code">{artifact.sourceId}</td><td class="artifact-code">{artifact.path}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</TableFrame>
