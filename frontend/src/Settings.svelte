<script lang="ts">
  import NavigationCardList from './NavigationCardList.svelte'
  import SectionHeading from './SectionHeading.svelte'

  const items = [
    { label: 'Sequence definitions', description: 'Create and inspect immutable sequence languages.', href: '/settings/sequence-definitions' },
    { label: 'Taxonomies', description: 'Organize parts and assign their attributes.', href: '/settings/taxonomies' },
    { label: 'Part-number schemas', description: 'Connect sequence definitions to taxonomy rules.', status: 'Soon' },
    { label: 'Artifact sources', description: 'Configure external sources and captured files.', href: '/artifacts' },
  ]
</script>

<SectionHeading eyebrow="Settings" title="Settings" description="Rules and connections that shape the UBOM." />
<NavigationCardList {items} />
