<script lang="ts">
  import { type TaxonomyChild, type TaxonomyPathItem } from './api'
  import { taxonomyPath as buildTaxonomyPath } from './routes'

  export let taxonomyPath: TaxonomyPathItem[]
  export let childCategories: TaxonomyChild[]
</script>

{#if childCategories.length > 0}
  <section aria-labelledby="categories-heading">
    <h2 id="categories-heading">Categories</h2>
    <div class="cards">
      {#each childCategories as child}
        <a class="card" href={buildTaxonomyPath([...taxonomyPath.map(item => item.label), child.label])}>
          <span>{child.label}</span>
          <span class="card-arrow" aria-hidden="true">→</span>
        </a>
      {/each}
    </div>
  </section>
{/if}
