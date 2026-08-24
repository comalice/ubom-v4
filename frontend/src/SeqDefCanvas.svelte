<script lang="ts">
  export type SequenceComponent = {
    type: 'literal' | 'range' | 'sequence' | 'select'
    value?: string
    start?: string
    end?: string
    padding?: number
    step?: number
    values?: string[]
    options?: string[]
    optionDescriptions?: string[]
    fieldName?: string
    description?: string
  }

  export let components: SequenceComponent[] = []
  export let inactiveComponents: SequenceComponent[] = []
  export let selectedIndex: number | null = null
  export let selectedOptions: Record<number, number> = {}
  export let onComponentsChange: (value: SequenceComponent[]) => void
  export let onInactiveChange: (value: SequenceComponent[]) => void
  export let onSelectedIndexChange: (value: number | null) => void
  export let onSelectedOptionsChange: (value: Record<number, number>) => void

  let editingIndex: number | null = null
  let editingInactive = false
  let draggedIndex: number | null = null
  let draggedInactive = false
  let dropPosition: number | null = null
  let menuDragging = false

  const templates: { type: SequenceComponent['type']; label: string; icon: string; description: string; example: string }[] = [
    { type: 'range', label: 'Range', icon: '#', description: 'Sequential values like A-Z or 1-100', example: '[A-Z]' },
    { type: 'literal', label: 'Literal', icon: 'T', description: 'Fixed text that never changes', example: "'USER_'" },
    { type: 'sequence', label: 'Sequence', icon: '[]', description: 'Cycle through specific values', example: '[POS, NEG, SKIP]' },
    { type: 'select', label: 'Select', icon: '{}', description: 'Pick one option from a set', example: '{small, large}' },
  ]

  function template(type: SequenceComponent['type']): SequenceComponent {
    if (type === 'literal') return { type, value: 'PREFIX_' }
    if (type === 'range') return { type, start: 'A', end: 'Z', padding: 0, step: 1 }
    if (type === 'sequence') return { type, values: ['A', 'B', 'C'], step: 1 }
    return { type, options: ['option1', 'option2'] }
  }

  function add(type: SequenceComponent['type']) {
    const next = [...components, template(type)]
    onComponentsChange(next)
    onSelectedIndexChange(next.length - 1)
    editingIndex = next.length - 1
    editingInactive = false
  }

  function update(index: number, value: SequenceComponent) {
    const next = [...(editingInactive ? inactiveComponents : components)]
    next[index] = value
    editingInactive ? onInactiveChange(next) : onComponentsChange(next)
  }

  function remove(index: number, inactive = false) {
    const next = (inactive ? inactiveComponents : components).filter((_, i) => i !== index)
    inactive ? onInactiveChange(next) : onComponentsChange(next)
    if (!inactive && selectedIndex === index) onSelectedIndexChange(null)
    if (editingIndex === index) editingIndex = null
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
      add('literal'); resetDrag(); return
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

  function resetDrag() { draggedIndex = null; draggedInactive = false; dropPosition = null; menuDragging = false }
  function display(component: SequenceComponent, index: number): string {
    if (component.type === 'literal') return component.value || ''
    if (component.type === 'select') return component.options?.[selectedOptions[index] ?? 0] || ''
    if (component.type === 'sequence') return component.values?.[0] || ''
    return component.start || ''
  }
  function edit(index: number, inactive = false) { editingIndex = index; editingInactive = inactive }
  function changeText(event: Event, key: string) {
    if (editingIndex === null) return
    const current = (editingInactive ? inactiveComponents : components)[editingIndex]
    update(editingIndex, { ...current, [key]: (event.currentTarget as HTMLInputElement).value })
  }
  function numberChange(event: Event, key: string) {
    if (editingIndex === null) return
    const current = (editingInactive ? inactiveComponents : components)[editingIndex]
    const value = (event.currentTarget as HTMLInputElement).value
    update(editingIndex, { ...current, [key]: value === '' ? undefined : Number(value) })
  }
</script>

<div class="generator-frame">
  <div class="component-menu" data-component-menu>
    <span class="menu-label">Components</span>
    {#each templates as item}
      <button class="component-tool" draggable="true" on:dragstart={(event) => { menuDragging = true; event.dataTransfer?.setData('application/json', JSON.stringify(template(item.type))) }} on:click={() => add(item.type)}>
        <b>{item.icon}</b><span>{item.label}</span>
      </button>
    {/each}
    <span class="menu-help">Drag to canvas or click to add</span>
  </div>

  <div class="builder-canvas" role="application" aria-label="Sequence builder canvas" on:dragover={(event) => dragOver(event)} on:drop={(event) => dropActive(event)} on:click={(event) => event.target === event.currentTarget && onSelectedIndexChange(null)}>
    <div class="builder-heading"><div><p class="eyebrow">Sequence canvas</p><h2>Builder</h2><p class="muted">Arrange components left to right. Click a component to select it.</p></div><span class="component-count">{components.length} active</span></div>
    {#if components.length === 0}
      <div class="canvas-empty">Drag components from the menu above to start building</div>
    {:else}
      <div class="component-strip">
        {#each components as component, index}
          <div class="canvas-component-wrap" role="listitem" on:dragover={(event) => dragOver(event, index)} on:drop={(event) => dropActive(event, dropPosition ?? index)}>
            {#if dropPosition === index && draggedIndex !== null}<span class="drop-marker"></span>{/if}
            <button class:selected-component={selectedIndex === index} class:dragged-component={draggedIndex === index && !draggedInactive} class="canvas-component type-{component.type}" draggable="true" on:dragstart={(event) => dragStart(event, index, false)} on:click|stopPropagation={() => onSelectedIndexChange(index)}>
              {#if component.fieldName}<small class="component-field">{component.fieldName}</small>{/if}
              <strong>{display(component, index)}</strong>{#if component.type === 'select'}<span class="select-cycle" role="button" tabindex="0" on:keydown={(event) => event.key === 'Enter' && onSelectedOptionsChange({ ...selectedOptions, [index]: ((selectedOptions[index] ?? 0) + 1) % (component.options?.length || 1) })} on:click|stopPropagation={() => onSelectedOptionsChange({ ...selectedOptions, [index]: ((selectedOptions[index] ?? 0) + 1) % (component.options?.length || 1) })} title="Choose option">⌄</span>{/if}<small class="component-type">{component.type.toUpperCase()}</small>
              <span class="component-buttons"><span role="button" tabindex="0" on:keydown={(event) => event.key === 'Enter' && edit(index)} on:click|stopPropagation={() => edit(index)} title="Configure">⚙</span><span role="button" tabindex="0" on:keydown={(event) => event.key === 'Enter' && remove(index)} on:click|stopPropagation={() => remove(index)} title="Remove">×</span></span>
            </button>
            {#if index === components.length - 1 && dropPosition === index + 1}<span class="drop-marker end-marker"></span>{/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <div class="inactive-zone" role="region" on:dragover={(event) => event.preventDefault()} on:drop={dropInactive}>
    <div class="zone-title">INACTIVE ZONE</div>
    {#if inactiveComponents.length === 0}<p class="muted">Drag components here to temporarily remove them</p>{:else}
      <div class="inactive-components">{#each inactiveComponents as component, index}<button class="inactive-component" draggable="true" on:dragstart={(event) => dragStart(event, index, true)} on:click={() => { const next = [...components, component]; onComponentsChange(next); onInactiveChange(inactiveComponents.filter((_, i) => i !== index)); onSelectedIndexChange(next.length - 1) }}>{display(component, index)} <span role="button" tabindex="0" on:keydown={(event) => event.key === 'Enter' && remove(index, true)} on:click|stopPropagation={() => remove(index, true)}>×</span></button>{/each}</div>
    {/if}
  </div>
</div>

{#if editingIndex !== null}
  {@const current = (editingInactive ? inactiveComponents : components)[editingIndex]}
  <div class="dialog-backdrop" role="presentation" on:click={(event) => event.target === event.currentTarget && (editingIndex = null)}>
    <div class="component-dialog" role="dialog" aria-modal="true" aria-labelledby="component-dialog-title">
      <div class="dialog-heading"><h2 id="component-dialog-title">Configure Component</h2><button class="close-button" on:click={() => editingIndex = null}>×</button></div>
      <label>Component type<select value={current.type} on:change={(event) => { const type = (event.currentTarget as HTMLSelectElement).value as SequenceComponent['type']; update(editingIndex!, { ...template(type), fieldName: current.fieldName, description: current.description }) }}><option value="literal">Literal</option><option value="range">Range</option><option value="sequence">Sequence</option><option value="select">Select</option></select></label>
      <label>Field name (optional)<input value={current.fieldName || ''} placeholder="e.g. prefix, year, id" on:input={(event) => changeText(event, 'fieldName')} /></label>
      {#if current.type === 'literal'}
        <label>Text value<input value={current.value || ''} on:input={(event) => changeText(event, 'value')} /></label>
      {:else if current.type === 'range'}
        <div class="field-grid"><label>Start<input value={current.start || ''} on:input={(event) => changeText(event, 'start')} /></label><label>End<input value={current.end || ''} on:input={(event) => changeText(event, 'end')} /></label><label>Padding<input type="number" min="0" value={current.padding ?? 0} on:input={(event) => numberChange(event, 'padding')} /></label><label>Step<input type="number" min="1" value={current.step ?? 1} on:input={(event) => numberChange(event, 'step')} /></label></div>
      {:else if current.type === 'sequence'}
        <label>Values <span class="muted">one value per line</span><textarea rows="5" value={(current.values || []).join('\n')} on:input={(event) => update(editingIndex!, { ...current, values: (event.currentTarget as HTMLTextAreaElement).value.split('\n') })}></textarea></label><label>Step<input type="number" min="1" value={current.step ?? 1} on:input={(event) => numberChange(event, 'step')} /></label>
        <p class="semantic-note">UBOM persists Sequence as a choice of literal values. Its reference cycling step is used by the local preview only.</p>
      {:else}
        <label>Options <span class="muted">one option per line</span><textarea rows="4" value={(current.options || []).join('\n')} on:input={(event) => update(editingIndex!, { ...current, options: (event.currentTarget as HTMLTextAreaElement).value.split('\n') })}></textarea></label>
        <label>Option descriptions <span class="muted">one description per line</span><textarea rows="3" value={(current.optionDescriptions || []).join('\n')} on:input={(event) => update(editingIndex!, { ...current, optionDescriptions: (event.currentTarget as HTMLTextAreaElement).value.split('\n') })}></textarea></label>
      {/if}
      <label>Description (optional)<input value={current.description || ''} on:input={(event) => changeText(event, 'description')} /></label>
      <button class="primary-button" on:click={() => editingIndex = null}>Confirm</button>
    </div>
  </div>
{/if}
