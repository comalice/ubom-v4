<script lang="ts">
  type SeqDefComponentFields = {
    kind: 'literal' | 'range' | 'choice' | 'places'
    text?: string
    min?: string
    max?: string
    width?: number
    options?: string[]
    alphabets?: string[]
  }
  type SeqDefComponentMetadata = {
    fieldName?: string
  }

  export type SeqDefComponent = SeqDefComponentFields & SeqDefComponentMetadata

  export let components: SeqDefComponent[] = []
  export let inactiveComponents: SeqDefComponent[] = []
  export let selectedIndex: number | null = null
  export let selectedOptions: Record<number, number> = {}
  export let onComponentsChange: (value: SeqDefComponent[]) => void
  export let onInactiveChange: (value: SeqDefComponent[]) => void
  export let onSelectedIndexChange: (value: number | null) => void
  export let onSelectedOptionsChange: (value: Record<number, number>) => void

  let editingIndex: number | null = null
  let editingInactive = false
  let editingNew = false
  let draft: SeqDefComponent | null = null
  let insertPosition: number | null = null
  let draggedIndex: number | null = null
  let draggedInactive = false
  let dropPosition: number | null = null
  let menuDragging = false

  const templates: { kind: SeqDefComponent['kind']; label: string; icon: string; description: string; example: string }[] = [
    { kind: 'range', label: 'Range', icon: '#', description: 'Sequential values like A-Z or 1-100', example: '[A-Z]' },
    { kind: 'literal', label: 'Literal', icon: 'T', description: 'Fixed text that never changes', example: "'USER_'" },
    { kind: 'choice', label: 'Select', icon: '{}', description: 'Pick one option from a set', example: '{small, large}' },
    { kind: 'places', label: 'Places', icon: 'ABC', description: 'One alphabet per character position', example: 'ABC / 012' },
  ]

  function template(kind: SeqDefComponent['kind']): SeqDefComponent {
    if (kind === 'literal') return { kind, text: 'PREFIX_' }
    if (kind === 'range') return { kind, min: '0', max: '9', width: 1 }
    if (kind === 'places') return { kind, alphabets: ['ABC'] }
    return { kind, options: ['option1', 'option2'] }
  }

  function add(kind: SeqDefComponent['kind']) {
    draft = template(kind)
    editingIndex = null
    editingNew = true
    insertPosition = components.length
    editingInactive = false
  }

  function openNew(value: SeqDefComponent, position = components.length) {
    draft = value
    editingIndex = null
    editingNew = true
    insertPosition = position
    editingInactive = false
  }

  function menuDrag(event: DragEvent, kind: SeqDefComponent['kind']) {
    menuDragging = true
    event.dataTransfer?.setData('application/json', JSON.stringify(template(kind)))
    if (event.dataTransfer) event.dataTransfer.effectAllowed = 'copy'
  }

  function draggedTemplate(event: DragEvent): SeqDefComponent | null {
    const data = event.dataTransfer?.getData('application/json')
    if (!data) return null
    try {
      const value = JSON.parse(data) as SeqDefComponent
      if (['literal', 'range', 'choice', 'places'].includes(value.kind)) return value
    } catch { /* Ignore invalid external drops. */ }
    return null
  }

  function clone(value: SeqDefComponent): SeqDefComponent {
    return { ...value, options: value.options ? [...value.options] : undefined, alphabets: value.alphabets ? [...value.alphabets] : undefined }
  }

  function edit(index: number, inactive = false) {
    editingIndex = index
    editingInactive = inactive
    editingNew = false
    insertPosition = null
    draft = clone((inactive ? inactiveComponents : components)[index])
  }

  function closeEditor() {
    draft = null
    editingIndex = null
    editingNew = false
    insertPosition = null
  }

  function remove(index: number, inactive = false) {
    const next = (inactive ? inactiveComponents : components).filter((_, i) => i !== index)
    inactive ? onInactiveChange(next) : onComponentsChange(next)
    if (!inactive && selectedIndex === index) onSelectedIndexChange(null)
    if (editingIndex === index) closeEditor()
  }

  function dragStart(event: DragEvent, index: number, inactive: boolean) {
    draggedIndex = index; draggedInactive = inactive; menuDragging = false
    event.dataTransfer?.setData('text/plain', 'component'); if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move'
  }

  function dragOver(event: DragEvent, index?: number) {
    event.preventDefault()
    if (index === undefined || draggedIndex === null || draggedInactive) return
    const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
    dropPosition = event.clientX < rect.left + rect.width / 2 ? index : index + 1
  }

  function dropActive(event: DragEvent, position?: number) {
    event.preventDefault()
    if (menuDragging) {
      const item = draggedTemplate(event)
      if (!item) { resetDrag(); return }
      const target = Math.max(0, Math.min(components.length, position ?? components.length))
      openNew(item, target)
      resetDrag()
      return
    }
    if (draggedIndex === null) return
    if (draggedInactive) {
      const item = inactiveComponents[draggedIndex]
      const inactive = inactiveComponents.filter((_, i) => i !== draggedIndex)
      const active = [...components]; active.splice(position ?? active.length, 0, item)
      onInactiveChange(inactive); onComponentsChange(active); onSelectedIndexChange(position ?? active.length - 1)
    } else {
      const next = [...components]; const [item] = next.splice(draggedIndex, 1)
      const target = Math.max(0, Math.min(next.length, (position ?? next.length) - (draggedIndex < (position ?? next.length) ? 1 : 0)))
      next.splice(target, 0, item); onComponentsChange(next); onSelectedIndexChange(target)
    }
    resetDrag()
  }

  function dropInactive(event: DragEvent) {
    event.preventDefault()
    if (draggedIndex !== null && !draggedInactive) {
      const item = components[draggedIndex]
      onComponentsChange(components.filter((_, i) => i !== draggedIndex)); onInactiveChange([...inactiveComponents, item]); onSelectedIndexChange(null)
    }
    resetDrag()
  }

  function commitDraft() {
    if (!draft) return
    if (editingNew) {
      const next = [...components]
      const target = Math.max(0, Math.min(next.length, insertPosition ?? next.length))
      next.splice(target, 0, draft)
      onComponentsChange(next)
      onSelectedIndexChange(target)
    } else if (editingIndex !== null) {
      const next = [...(editingInactive ? inactiveComponents : components)]
      next[editingIndex] = draft
      editingInactive ? onInactiveChange(next) : onComponentsChange(next)
    }
    closeEditor()
  }

  function resetDrag() { draggedIndex = null; draggedInactive = false; dropPosition = null; menuDragging = false }
  function display(component: SeqDefComponent, index: number): string {
    if (component.kind === 'literal') return component.text || ''
    if (component.kind === 'choice') return component.options?.[selectedOptions[index] ?? 0] || ''
    if (component.kind === 'places') return (component.alphabets || []).map(alphabet => alphabet[0] || '').join('')
    return component.min || ''
  }
  function changeText(event: Event, key: string) {
    if (!draft) return
    draft = { ...draft, [key]: (event.currentTarget as HTMLInputElement).value }
  }
  function numberChange(event: Event, key: string) {
    if (!draft) return
    const value = (event.currentTarget as HTMLInputElement).value
    draft = { ...draft, [key]: value === '' ? undefined : Number(value) }
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && draft) closeEditor()
  }
</script>

<svelte:window on:keydown={handleWindowKeydown} />

<div class="generator-frame">
  <div class="component-menu" data-component-menu>
    <span class="menu-label">Components</span>
    {#each templates as item}
      <button class="component-tool" draggable="true" on:dragstart={(event) => menuDrag(event, item.kind)} on:click={() => add(item.kind)}>
        <b>{item.icon}</b><span>{item.label}</span>
      </button>
    {/each}
    <span class="menu-help">Drag to canvas or click to add</span>
  </div>

  <div class="builder-canvas" role="application" aria-label="Sequence builder canvas" on:dragover={(event) => dragOver(event)} on:drop={(event) => dropActive(event)}>
    <div class="builder-heading"><div><p class="eyebrow">Sequence canvas</p><h2>Builder</h2><p class="muted">Arrange components left to right. Click a component to select it.</p></div><span class="component-count">{components.length} active</span></div>
    {#if components.length === 0}
      <div class="canvas-empty">Drag components from the menu above to start building</div>
    {:else}
      <div class="component-strip">
        {#each components as component, index}
          <div class="canvas-component-wrap" role="listitem" on:dragover={(event) => dragOver(event, index)} on:drop={(event) => dropActive(event, dropPosition ?? index)}>
            {#if dropPosition === index && draggedIndex !== null}<span class="drop-marker"></span>{/if}
            <div class:selected-component={selectedIndex === index} class:dragged-component={draggedIndex === index && !draggedInactive} class="canvas-component type-{component.kind}" draggable="true" role="button" tabindex="0" aria-label={`${component.kind} component`} on:dragstart={(event) => dragStart(event, index, false)} on:keydown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelectedIndexChange(index) } }} on:click|stopPropagation={() => onSelectedIndexChange(index)}>
              {#if component.fieldName}<small class="component-field">{component.fieldName}</small>{/if}
              <strong>{display(component, index)}</strong>{#if component.kind === 'choice'}<button type="button" class="select-cycle" aria-label="Choose option" on:click|stopPropagation={() => onSelectedOptionsChange({ ...selectedOptions, [index]: ((selectedOptions[index] ?? 0) + 1) % (component.options?.length || 1) })}>⌄</button>{/if}<small class="component-type">{component.kind.toUpperCase()}</small>
              <span class="component-buttons"><button type="button" aria-label="Configure component" on:click|stopPropagation={() => edit(index)} title="Configure">⚙</button><button type="button" aria-label="Remove component" on:click|stopPropagation={() => remove(index)} title="Remove">×</button></span>
            </div>
            {#if index === components.length - 1 && dropPosition === index + 1}<span class="drop-marker end-marker"></span>{/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div class="inactive-zone" role="region" on:dragover={(event) => event.preventDefault()} on:drop={dropInactive}>
    <div class="zone-title">INACTIVE ZONE</div>
    {#if inactiveComponents.length === 0}<p class="muted">Drag components here to temporarily remove them</p>{:else}
      <div class="inactive-components">{#each inactiveComponents as component, index}<div class="inactive-component" role="button" tabindex="0" draggable="true" aria-label={`${component.kind} inactive component`} on:dragstart={(event) => dragStart(event, index, true)} on:keydown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); const next = [...components, component]; onComponentsChange(next); onInactiveChange(inactiveComponents.filter((_, i) => i !== index)); onSelectedIndexChange(next.length - 1) } }} on:click={() => { const next = [...components, component]; onComponentsChange(next); onInactiveChange(inactiveComponents.filter((_, i) => i !== index)); onSelectedIndexChange(next.length - 1) }}>{display(component, index)} <button type="button" aria-label="Remove inactive component" on:click|stopPropagation={() => remove(index, true)}>×</button></div>{/each}</div>
    {/if}
  </div>
</div>

{#if draft}
    <div class="dialog-backdrop" role="presentation" on:click={(event) => event.target === event.currentTarget && closeEditor()}>
      <div class="component-dialog" role="dialog" aria-modal="true" aria-labelledby="component-dialog-title">
       <div class="dialog-heading"><h2 id="component-dialog-title">Configure Component</h2><button type="button" class="close-button" aria-label="Close component dialog" on:click={closeEditor}>×</button></div>
       <label>Component type<select value={draft.kind} on:change={(event) => { const kind = (event.currentTarget as HTMLSelectElement).value as SeqDefComponent['kind']; draft = { ...template(kind), fieldName: draft.fieldName } }}><option value="literal">Literal</option><option value="range">Range</option><option value="choice">Select</option><option value="places">Places</option></select></label>
       <label>Field name (optional)<input value={draft.fieldName || ''} placeholder="e.g. prefix, year, id" on:input={(event) => changeText(event, 'fieldName')} /></label>
       {#if draft.kind === 'literal'}
         <label>Text value<input value={draft.text || ''} on:input={(event) => changeText(event, 'text')} /></label>
       {:else if draft.kind === 'range'}
         <div class="field-grid"><label>Min<input type="number" min="0" value={draft.min || ''} on:input={(event) => changeText(event, 'min')} /></label><label>Max<input type="number" min="0" value={draft.max || ''} on:input={(event) => changeText(event, 'max')} /></label><label>Width<input type="number" min="1" value={draft.width ?? 1} on:input={(event) => numberChange(event, 'width')} /></label></div>
       {:else if draft.kind === 'places'}
         <label>Alphabets <span class="muted">one alphabet per line</span><textarea rows="4" value={(draft.alphabets || []).join('\n')} on:input={(event) => draft = { ...draft, alphabets: (event.currentTarget as HTMLTextAreaElement).value.split('\n') }}></textarea></label>
         <p class="semantic-note">Each position uses the first matching character during backend generation.</p>
       {:else}
         <label>Options <span class="muted">one option per line</span><textarea rows="4" value={(draft.options || []).join('\n')} on:input={(event) => draft = { ...draft, options: (event.currentTarget as HTMLTextAreaElement).value.split('\n') }}></textarea></label>
       {/if}
       <button class="primary-button" on:click={commitDraft}>Confirm</button>
    </div>
  </div>
{/if}
