<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import { taxonomyPath } from './routes'

  export let partNumber: string
  export let taxonomy: string[] = []
  export let revision: string | undefined = undefined
  export let revisionID: string | undefined = undefined
</script>

<div class="part-identity">
  <h1><IdentityLink {partNumber} {revision} {revisionID} /></h1>
  {#if taxonomy.length > 0}
    <nav class="taxonomy-path" aria-label="Part category">
      {#each taxonomy as label, index}
        {#if index > 0}<span class="breadcrumb-separator" aria-hidden="true">/</span>{/if}
        <a href={taxonomyPath(taxonomy.slice(0, index + 1))}>{label}</a>
      {/each}
    </nav>
  {/if}
</div>
