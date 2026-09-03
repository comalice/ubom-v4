<script lang="ts">
  import PartTable from './PartTable.svelte'
  import type { PartListItem } from './api'

  const parts: PartListItem[] = [
    {
      id: 'part-100',
      value: 'PN-RES-0603-001',
      taxonomyPath: ['Electrical', 'Passive Components', 'Resistors'],
      status: 'active',
      attributes: [
        { id: 'resistance', label: 'Resistance', value: '10k', valueType: 'string' },
        { id: 'tolerance', label: 'Tolerance', value: '1%', valueType: 'string' },
      ],
      revisions: [{ id: 'revision-100-a', revision: 'A' }],
    },
    {
      id: 'part-101',
      value: 'PN-RES-0805-002',
      taxonomyPath: ['Electrical', 'Passive Components', 'Resistors'],
      status: 'active',
      attributes: [
        { id: 'resistance', label: 'Resistance', value: '1k', valueType: 'string' },
        { id: 'tolerance', label: 'Tolerance', value: '5%', valueType: 'string' },
      ],
      revisions: [{ id: 'revision-101-a', revision: 'A' }],
    },
  ]

  const columns = [
    { id: 'resistance', label: 'Resistance' },
    { id: 'tolerance', label: 'Tolerance' },
  ]
</script>

<section class="playground-empty-page" aria-labelledby="part-views-heading">
  <p class="eyebrow">Parts</p>
  <h1 id="part-views-heading">Part views</h1>
  <PartTable {parts} {columns} seqDef="part-number-v1" />
</section>
