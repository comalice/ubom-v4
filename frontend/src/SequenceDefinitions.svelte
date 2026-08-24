<script lang="ts">
  import SeqDefCanvas, { type SequenceComponent } from './SeqDefCanvas.svelte'
  import { createSeqDef, getSeqDefs, previewSeqDef, type SeqDefNode, type SeqDefView } from './api'
  export let creating = false
  export let navigate: (path: string) => void
  let definitions: SeqDefView[] = []
  let components: SequenceComponent[] = []
  let inactiveComponents: SequenceComponent[] = []
  let selectedIndex: number | null = null
  let selectedOptions: Record<number, number> = {}
  let name = ''
  let bindings: Record<string, string> = {}
  let results: string[] = []
  let parseValue = ''
  let incrementValue = ''
  let fileInput: HTMLInputElement
  let backendPreview = ''
  let error = ''
  let saving = false
  let loaded = false

  $: fieldNames = [...new Set(components.map(component => component.fieldName).filter(Boolean) as string[])]
  $: for (const field of fieldNames) if (!(field in bindings)) bindings[field] = ''

  function values(component: SequenceComponent): string[] {
    if (component.type === 'literal') return [component.value || '']
    if (component.type === 'sequence') return (component.values || []).filter(Boolean).filter((_, index) => index % (component.step || 1) === 0)
    if (component.type === 'select') return component.options || []
    const start = component.start || ''; const end = component.end || ''; const step = component.step || 1; const output: string[] = []
    if (/^\d+$/.test(start) && /^\d+$/.test(end)) for (let i = Number(start); i <= Number(end) && output.length < 100; i += step) output.push(String(i).padStart(component.padding || 0, '0'))
    else if (start.length === 1 && end.length === 1) for (let i = start.charCodeAt(0); i <= end.charCodeAt(0) && output.length < 100; i += step) output.push(String.fromCharCode(i))
    return output
  }
  function generate(count: number, start?: string): string[] {
    if (!components.length) return []
    const all = components.map(values); const indices = all.map(() => 0)
    if (start) { let rest = start; components.forEach((component, index) => { const match = all[index].findIndex(value => rest.startsWith(value)); if (match >= 0) { indices[index] = match; rest = rest.slice(all[index][match].length) } }) }
    components.forEach((component, index) => { if (component.type === 'select' && selectedOptions[index] !== undefined) indices[index] = selectedOptions[index] })
    const output: string[] = []
    for (let i = 0; i < count; i += 1) { output.push(indices.map((value, index) => all[index][value] || '').join('')); let carry = true; for (let j = indices.length - 1; j >= 0 && carry; j -= 1) { if (components[j].type === 'literal' || (components[j].type === 'select' && selectedOptions[j] !== undefined)) continue; indices[j] += 1; if (indices[j] >= all[j].length) indices[j] = 0; else carry = false } if (carry) break }
    return output
  }
  function parse(input: string): boolean { const generated = generate(100); return generated.includes(input) }
  function increment(input: string): string | null { const generated = generate(200); const index = generated.indexOf(input); return index >= 0 ? generated[index + 1] || null : null }
  function grammar(component: SequenceComponent): string { if (component.type === 'literal') return `'${component.value || ''}'`; if (component.type === 'range') return `[${component.start || ''}-${component.end || ''}]`; if (component.type === 'sequence') return `[${(component.values || []).join(', ')}]`; return `{${(component.options || []).join(', ')}}` }
  function importGrammar(text: string) { components = text.trim().replace(/^\w+\s*:=\s*/, '').split(/\s*\+\s*/).flatMap((part): SequenceComponent[] => { const value = part.trim(); if (value.startsWith("'") && value.endsWith("'")) return [{ type: 'literal', value: value.slice(1, -1) }]; if (value.startsWith('{') && value.endsWith('}')) return [{ type: 'select', options: value.slice(1, -1).split(',').map(item => item.trim()) }]; if (value.startsWith('[') && value.endsWith(']')) { const body = value.slice(1, -1); if (body.includes(',')) return [{ type: 'sequence', values: body.split(',').map(item => item.trim()) }]; const [start, end] = body.split('-'); return [{ type: 'range', start, end, padding: /^\d+$/.test(start) && start.length > 1 ? start.length : 0 }]; } return [] }) }
  function ast(component: SequenceComponent): SeqDefNode {
    let node: SeqDefNode
    if (component.type === 'literal') node = { kind: 'literal', text: component.value || '' }
    else if (component.type === 'range' && /^\d+$/.test(component.start || '') && /^\d+$/.test(component.end || '')) node = { kind: 'range', min: Number(component.start), max: Number(component.end), width: component.padding || String(component.end).length }
    else if (component.type === 'select' || component.type === 'sequence') node = { kind: 'choice', children: (component.type === 'select' ? component.options : component.values || []).map(text => ({ kind: 'literal', text })) }
    else node = { kind: 'literal', text: component.start || '' }
    return component.fieldName?.trim() ? { kind: 'bind', name: component.fieldName.trim(), child: node } : node
  }
  $: root = { kind: 'concat', children: components.map(ast) } as SeqDefNode
  async function save() { saving = true; error = ''; try { await createSeqDef(name.trim(), root); navigate('/settings/sequence-definitions') } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to save sequence definition' } finally { saving = false } }
  async function preview() { backendPreview = ''; error = ''; try { const result = await previewSeqDef(name.trim(), bindings); backendPreview = result.error ? `Error: ${result.error}` : result.value || '' } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to preview sequence definition' } }
  async function load() { try { definitions = await getSeqDefs() } catch (reason) { error = reason instanceof Error ? reason.message : 'Unable to load sequence definitions' } }
  function download() { const blob = new Blob([components.map(grammar).join(' + ')], { type: 'text/plain' }); const link = document.createElement('a'); link.href = URL.createObjectURL(blob); link.download = `${name || 'sequence'}.txt`; link.click(); URL.revokeObjectURL(link.href) }
  $: if (!creating && !loaded) { loaded = true; load() }
</script>

{#if !creating}
  <p class="eyebrow">Settings</p><h1>Sequence definitions</h1><a class="button-link" href="/settings/sequence-definitions/new">New sequence definition</a>
  {#if error}<p class="error">{error}</p>{/if}
  {#if definitions.length}<div class="part-list seq-def-list">{#each definitions as definition}<div class="part-row"><span class="part-value">{definition.id}</span><span class="muted">SeqDef</span></div>{/each}</div>{:else}<p class="empty-state">No sequence definitions.</p>{/if}
{:else}
  <div class="generator-page">
    <div class="generator-topbar"><button class="icon-link" on:click={() => navigate('/settings/sequence-definitions')}>☰</button><button class="topbar-button" on:click={save} disabled={saving || !name.trim()}>Save</button><button class="topbar-button" on:click={preview} disabled={!name.trim()}>Backend preview</button><span class="topbar-caption">{name || 'New sequence definition'}</span></div>
    <SeqDefCanvas bind:components bind:inactiveComponents bind:selectedIndex bind:selectedOptions onComponentsChange={(value) => components = value} onInactiveChange={(value) => inactiveComponents = value} onSelectedIndexChange={(value) => selectedIndex = value} onSelectedOptionsChange={(value) => selectedOptions = value} />
    <div class="powerline-bar"><input bind:this={fileInput} type="file" accept=".txt" hidden on:change={async (event) => { const file = (event.currentTarget as HTMLInputElement).files?.[0]; if (file) importGrammar(await file.text()); (event.currentTarget as HTMLInputElement).value = '' }} /><input class="sequence-name" bind:value={name} placeholder="sequence_name" /><code>{components.length ? components.map(grammar).join(' + ') : 'No components'}</code><div class="powerline-actions"><button on:click={() => fileInput.click()}>Import</button><button on:click={download} disabled={!components.length}>Export</button><button on:click={preview} disabled={!name.trim()}>Preview</button><button class="primary-button" on:click={() => results = generate(20)} disabled={!components.length}>Generate</button></div></div>
    {#if results.length}<div class="results-panel"><div class="results-header"><b>Generated Sequences ({results.length})</b><button on:click={() => results = []}>Clear</button></div><div class="results-grid">{#each results as result}<button on:click={() => navigator.clipboard?.writeText(result)}>{result}<span>⧉</span></button>{/each}</div></div>{/if}
    <div class="generator-controls"><div><label>Parse &amp; Jump<input bind:value={parseValue} placeholder="e.g. A-005" /></label><button on:click={() => results = parse(parseValue) ? [parseValue, ...generate(10, parseValue)] : []} disabled={!parseValue || !components.length}>Parse</button></div><div><label>Increment<input bind:value={incrementValue} placeholder="e.g. Z-998" /></label><button on:click={() => { const next = increment(incrementValue); results = next ? [next] : [] }} disabled={!incrementValue || !components.length}>Increment</button></div></div>
    {#if fieldNames.length}<div class="preview-fields"><h2>Preview bindings</h2>{#each fieldNames as field}<label>{field}<input bind:value={bindings[field]} /></label>{/each}</div>{/if}
    {#if backendPreview}<p class="success artifact-code">Backend preview: {backendPreview}</p>{/if}{#if error}<p class="error">{error}</p>{/if}
    <div class="form-actions"><button class="primary-button" on:click={save} disabled={saving || !name.trim()}>Create definition</button><button class="quiet-button" on:click={() => navigate('/settings/sequence-definitions')}>Cancel</button></div>
  </div>
{/if}
