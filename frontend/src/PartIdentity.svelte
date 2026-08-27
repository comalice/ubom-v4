<script lang="ts">
  import IdentityLink from './IdentityLink.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import { taxonomyPath } from './routes'

  export let partNumber: string
  export let taxonomy: string[] = []
  export let revision: string | undefined = undefined
  export let revisionID: string | undefined = undefined
  export let attributes: { id: string; label: string; value: string; unit?: string }[] = []
</script>

<div class="part-identity">
  <h1><IdentityLink {partNumber} {revision} {revisionID} /></h1>
  {#if taxonomy.length > 0}
    <nav class="taxonomy-path" aria-label="Part category">
      {#each taxonomy as label, index}
        {#if index > 0}<span class="breadcrumb-separator" aria-hidden="true">/</span>{/if}
        <PrimitiveLink href={taxonomyPath(taxonomy.slice(0, index + 1))}>{label}</PrimitiveLink>
      {/each}
    </nav>
  {/if}
</div>

{#if attributes.length > 0}
  <section class="attribute-summary" aria-labelledby="attributes-heading">
    <h2 id="attributes-heading">Attributes</h2>
    <dl>
      {#each attributes as attribute}
        <div>
          <dt>{attribute.label}</dt>
          <dd>{attribute.value}{#if attribute.unit} {attribute.unit}{/if}</dd>
        </div>
      {/each}
    </dl>
  </section>
{/if}
