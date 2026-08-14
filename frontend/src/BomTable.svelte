<script lang="ts">
  import BomRow from './BomRow.svelte'
  import type { BomNode } from './api'

  export let nodes: BomNode[] = []
</script>

<div class="table-frame bom-table-frame">
  <table class="bom-table">
    <caption class="sr-only">Indented bill of materials</caption>
    <thead>
      <tr>
        <th scope="col">Part / revision</th>
        <th scope="col">PN description</th>
        <th scope="col">PN status</th>
        <th scope="col">Qty</th>
        <th scope="col">UOM</th>
      </tr>
    </thead>
    <tbody>
      {#each nodes as node, index (node.revisionId + '-' + index)}
        <BomRow {node} />
      {/each}
    </tbody>
  </table>
</div>
