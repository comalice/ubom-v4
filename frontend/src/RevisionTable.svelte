<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import TableFrame from './TableFrame.svelte'
  import type { PartNumberView } from './api'

  export let partNumber: PartNumberView['value']
  export let revisions: PartNumberView['revisions'] = []
</script>

<TableFrame>
  <table class="revision-table">
    <caption class="sr-only">Part revisions</caption>
    <thead>
      <tr>
        <th scope="col">Rev</th>
        <th scope="col">Status</th>
        <th scope="col">ECO/ECR</th>
        <th scope="col">Effectivity date</th>
      </tr>
    </thead>
    <tbody>
      {#each revisions as item}
        <tr>
          <th scope="row"><IdentityLink partNumber={partNumber} revision={item.revision} revisionID={item.id} /></th>
          <td class="unsupported" aria-label="Status not supported">—</td>
          <td class="unsupported" aria-label="ECO or ECR not supported">—</td>
          <td class="unsupported" aria-label="Effectivity date not supported">—</td>
        </tr>
      {/each}
    </tbody>
  </table>
</TableFrame>
