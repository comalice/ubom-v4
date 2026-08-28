<script lang="ts">
  import AsyncState from './AsyncState.svelte'
  import SectionHeading from './SectionHeading.svelte'

  export let id: string | undefined = undefined
</script>

<SectionHeading eyebrow="Taxonomies" title={id ? 'Edit taxonomy' : 'New taxonomy'} description="A persisted taxonomy editor is not wired yet." />
<AsyncState state="empty" message={id ? 'Taxonomy details are not available yet.' : 'Choose a sequence definition to begin. Taxonomy editing will be available in a later tranche.'} />
