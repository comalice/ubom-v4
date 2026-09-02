<script lang="ts">
  import { type TaxonomyChild, type TaxonomyPathItem } from './api'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import { taxonomyPath as buildTaxonomyPath } from './routes'

  export let taxonomyPath: TaxonomyPathItem[]
  export let childCategories: TaxonomyChild[]
  export let seqDef: string | undefined = undefined
</script>

{#if childCategories.length > 0}
  <section aria-labelledby="categories-heading">
    <h2 id="categories-heading">Categories</h2>
    <div class="cards">
      {#each childCategories as child}
        <PrimitiveLink className="card" href={buildTaxonomyPath([...taxonomyPath.map(item => item.label), child.label], seqDef)}>
          <span>{child.label}</span>
          <span class="card-arrow" aria-hidden="true">→</span>
        </PrimitiveLink>
      {/each}
    </div>
  </section>
{/if}
