<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import TableFrame from './TableFrame.svelte'
  import { taxonomyPath } from './routes'
  import type { AttributeColumn, PartAttribute, PartListItem } from './api'

  export let parts: PartListItem[] = []
  export let columns: AttributeColumn[] = []

  function attributeValue(part: PartListItem, id: string): PartAttribute | undefined {
    return part.attributes.find(attribute => attribute.id === id)
  }
</script>

<TableFrame className="part-table-frame">
  <table class="part-table">
    <caption class="sr-only">Part numbers</caption>
    <thead>
      <tr>
        <th scope="col">PN</th>
        <th scope="col">Description</th>
        <th scope="col">Taxonomy path</th>
        {#each columns as column}
          <th scope="col">{column.label}{#if column.unit} ({column.unit}){/if}</th>
        {/each}
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
              <PrimitiveLink href={taxonomyPath(part.taxonomyPath.slice(0, index + 1))}>{label}</PrimitiveLink>
            {/each}
          </td>
          {#each columns as column}
            {@const attribute = attributeValue(part, column.id)}
            <td>{attribute?.value ?? '—'}</td>
          {/each}
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
</TableFrame>
