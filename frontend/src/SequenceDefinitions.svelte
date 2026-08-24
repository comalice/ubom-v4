<script lang="ts">
  import SeqDefCanvas, { type SeqDefComponent } from './SeqDefCanvas.svelte'
  import { createSeqDef, getSeqDefs, previewSeqDef, type SeqDefNode, type SeqDefView } from './api'
  export let creating = false
  export let navigate: (path: string) => void
  let definitions: SeqDefView[] = []
  let components: SeqDefComponent[] = []
  let inactiveComponents: SeqDefComponent[] = []
  let selectedIndex: number | null = null
  let name = ''
  let bindings: Record<string, string> = {}
  let backendPreview = ''
  let error = ''
  let saving = false
  let loaded = false

  $: fieldNames = [...new Set(components.map(component => component.fieldName).filter(Boolean) as string[])]
  $: for (const field of fieldNames) if (!(field in bindings)) bindings[field] = ''

  function validateComponent(component: SeqDefComponent, index: number): string | null {
    const label = `Component ${index + 1}`
    if (component.kind === 'range') {
      if (!/^\d+$/.test(component.min || '') || !/^\d+$/.test(component.max || '')) return `${label}: range minimum and maximum must be whole numbers`
      const min = Number(component.min); const max = Number(component.max); const width = component.width ?? 1
      if (min > max) return `${label}: range minimum must not exceed maximum`
      if (!Number.isInteger(width) || width < 1) return `${label}: range width must be at least 1`
      if (String(max).length > width) return `${label}: range width is too small for the maximum`
      const pad = component.pad ?? '0'
      if (pad.length !== 1 || (pad >= '0' && pad <= '9' && pad !== '0')) return `${label}: pad must be one non-decimal character or 0`
    }
    if (component.kind === 'choice') {
      if (!component.options?.length) return `${label}: select needs at least one option`
      if (component.options.some(option => !option.trim())) return `${label}: select options cannot be empty`
    }
    if (component.kind === 'places' && (!component.alphabets?.length || component.alphabets.some(alphabet => !alphabet.trim()))) return `${label}: places needs non-empty alphabets`
    return null
  }
  $: validationError = components.length ? components.map(validateComponent).find(Boolean) || '' : 'Add at least one component'
  // This is the explicit UI-to-backend adapter; editor metadata is not persisted.
  function ast(component: SeqDefComponent): SeqDefNode {
    let node: SeqDefNode
    if (component.kind === 'literal') node = { kind: 'literal', text: component.text || '' }
    else if (component.kind === 'range' && /^\d+$/.test(component.min || '') && /^\d+$/.test(component.max || '')) node = { kind: 'range', min: Number(component.min), max: Number(component.max), width: component.width || String(component.max).length, pad: component.pad && component.pad !== '0' ? component.pad.charCodeAt(0) : undefined }
    else if (component.kind === 'choice') node = { kind: 'choice', children: (component.options || []).map(text => ({ kind: 'literal', text })) }
    else if (component.kind === 'places') node = { kind: 'places', alphabets: component.alphabets || [] }
    else throw new Error('Range requires numeric minimum and maximum values')
    return component.fieldName?.trim() ? { kind: 'bind', name: component.fieldName.trim(), child: node } : node
  }
  function serialize(): SeqDefNode { return { kind: 'concat', children: components.map(ast) } }
  async function save() { saving = true; error = ''; try { await createSeqDef(name.trim(), serialize()); navigate('/settings/sequence-definitions') } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to save sequence definition' } finally { saving = false } }
  async function preview() { backendPreview = ''; error = ''; try { const result = await previewSeqDef(serialize(), bindings); backendPreview = result.error ? `Error: ${result.error}` : result.value || '' } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to preview sequence definition' } }
  async function load() { try { definitions = await getSeqDefs() } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to load sequence definitions' } }
  $: if (!creating && !loaded) { loaded = true; load() }
</script>

{#if !creating}
  <p class="eyebrow">Settings</p><h1>Sequence definitions</h1><a class="button-link" href="/settings/sequence-definitions/new">New sequence definition</a>
  {#if error}<p class="error">{error}</p>{/if}
  {#if definitions.length}<div class="part-list seq-def-list">{#each definitions as definition}<div class="part-row"><span class="part-value">{definition.id}</span><span class="muted">SeqDef</span></div>{/each}</div>{:else}<p class="empty-state">No sequence definitions.</p>{/if}
{:else}
  <div class="generator-page">
     <div class="generator-topbar"><button class="icon-link" on:click={() => navigate('/settings/sequence-definitions')}>☰</button><button class="topbar-button" on:click={save} disabled={saving || !name.trim() || validationError}>Save</button><button class="topbar-button" on:click={preview} disabled={!name.trim() || validationError}>Backend preview</button><span class="topbar-caption">{name || 'New sequence definition'}</span></div>
    <SeqDefCanvas bind:components bind:inactiveComponents bind:selectedIndex onComponentsChange={(value) => components = value} onInactiveChange={(value) => inactiveComponents = value} onSelectedIndexChange={(value) => selectedIndex = value} />
      <div class="powerline-bar"><input class="sequence-name" bind:value={name} placeholder="sequence_name" /><div class="powerline-actions"><button on:click={preview} disabled={!name.trim() || validationError}>Preview</button></div></div>
    {#if fieldNames.length}<div class="preview-fields"><h2>Preview bindings</h2>{#each fieldNames as field}<label>{field}<input bind:value={bindings[field]} /></label>{/each}</div>{/if}
     {#if validationError}<p class="error" role="alert">{validationError}</p>{/if}{#if backendPreview}<p class="success artifact-code">Backend preview: {backendPreview}</p>{/if}{#if error}<p class="error" role="alert">{error}</p>{/if}
     <div class="form-actions"><button class="primary-button" on:click={save} disabled={saving || !name.trim() || validationError}>Create definition</button><button class="quiet-button" on:click={() => navigate('/settings/sequence-definitions')}>Cancel</button></div>
  </div>
{/if}
