<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import { taxonomyPath } from './routes'
  import type { PartListItem } from './api'

  export let parts: PartListItem[] = []
</script>

<div class="table-frame part-table-frame">
  <table class="part-table">
    <caption class="sr-only">Part numbers</caption>
    <thead>
      <tr>
        <th scope="col">PN</th>
        <th scope="col">Description</th>
        <th scope="col">Taxonomy path</th>
        <th scope="col">Latest revision</th>
      </tr>
    </thead>
    <tbody>
      {#each parts as part}
        <tr>
          <th scope="row"><IdentityLink partNumber={part.value} /></th>
          <td class="unsupported" aria-label="Description not supported">—</td>
          <td class="part-taxonomy">
            {#each part.taxonomyPath as label, index}
              {#if index > 0}<span class="breadcrumb-separator" aria-hidden="true"> / </span>{/if}
              <a href={taxonomyPath(part.taxonomyPath.slice(0, index + 1))}>{label}</a>
            {/each}
          </td>
          <td>
            {#if part.revisions.length > 0}
              {@const latest = part.revisions[part.revisions.length - 1]}
              <IdentityLink partNumber={part.value} revision={latest.revision} revisionID={latest.id} />
            {:else}
              <span class="muted">No revisions</span>
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
