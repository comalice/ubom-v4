<script lang="ts">
  import ControlButton from './ControlButton.svelte'
  import PrimitiveLink from './PrimitiveLink.svelte'
  import PrimitiveButton from './PrimitiveButton.svelte'
  import TextInput from './TextInput.svelte'
  import SelectInput from './SelectInput.svelte'
  import TextAreaInput from './TextAreaInput.svelte'
  import ChoiceGroup from './ChoiceGroup.svelte'
  let submitted = false
</script>

<section class="html-primitives" aria-labelledby="html-primitives-heading">
  <p class="eyebrow">HTML primitives</p>
  <h2 id="html-primitives-heading">HTML primitives</h2>
  <p class="muted html-primitives-intro">Native elements wrapped in the first shared primitives for links, controls, and forms.</p>

  <section class="primitive-group" aria-labelledby="links-heading">
    <p class="mock-section-label">Navigation</p>
    <h3 id="links-heading">Links</h3>
    <div class="primitive-card primitive-link-list">
      <PrimitiveLink href="/parts">Default link <span aria-hidden="true">→</span></PrimitiveLink>
      <PrimitiveLink href="/settings" variant="muted">Muted link <span aria-hidden="true">→</span></PrimitiveLink>
      <PrimitiveLink href="/playground/html-primitives" current>Current-page link</PrimitiveLink>
      <PrimitiveLink href="/parts" disabled>Unavailable link</PrimitiveLink>
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="buttons-heading">
    <p class="mock-section-label">Actions</p>
    <h3 id="buttons-heading">Buttons</h3>
    <div class="primitive-card primitive-button-row">
      <PrimitiveButton>Primary</PrimitiveButton>
      <PrimitiveButton variant="secondary">Secondary</PrimitiveButton>
      <PrimitiveButton variant="quiet">Quiet</PrimitiveButton>
      <PrimitiveButton disabled>Disabled</PrimitiveButton>
    </div>
    <div class="primitive-card primitive-control-button-card">
      <div><strong>ControlButton</strong><small>Icon-only button primitive used by expandable and removable controls.</small></div>
      <div class="primitive-control-button-row"><ControlButton kind="chevron" label="Expand example" expanded={false} /><ControlButton kind="chevron" label="Collapse example" expanded={true} /><ControlButton kind="close" label="Close example" /></div>
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="form-controls-heading">
    <p class="mock-section-label">Input controls</p>
    <h3 id="form-controls-heading">Forms and controls</h3>
    <div class="primitive-control-grid">
      <div class="primitive-card primitive-field-card">
        <TextInput id="primitive-text" label="Text input" value="PN-100" help="Single-line text with a visible label." />
      </div>
      <div class="primitive-card primitive-field-card">
        <SelectInput id="primitive-select" label="Select" value="Steel" options={[{ value: 'Steel', label: 'Steel' }, { value: 'Aluminium', label: 'Aluminium' }, { value: 'Polymer', label: 'Polymer' }]} help="Native option selection." />
      </div>
      <div class="primitive-card primitive-field-card">
        <TextInput id="primitive-number" label="Number input" type="number" min={0} value="10" help="Numeric input with a constraint." />
      </div>
      <div class="primitive-card primitive-field-card">
        <TextAreaInput id="primitive-textarea" label="Textarea" value="A short description." help="Multi-line text entry." />
      </div>
    </div>
    <div class="primitive-card primitive-choice-card">
      <ChoiceGroup legend="Display preference" choices={[{ kind: 'radio', value: 'compact', label: 'Compact', checked: true }, { kind: 'radio', value: 'comfortable', label: 'Comfortable' }, { kind: 'checkbox', value: 'identifiers', label: 'Show identifiers', checked: true }]} />
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="form-heading">
    <p class="mock-section-label">Composition</p>
    <h3 id="form-heading">Form</h3>
    <form class="primitive-card primitive-form" on:submit|preventDefault={() => submitted = true}>
      <div class="primitive-form-fields">
        <div class="primitive-form-field">
          <TextInput id="primitive-name" name="name" label="Definition name" required placeholder="Part number" help="Required field for the definition." />
        </div>
        <div class="primitive-form-field">
          <TextAreaInput id="primitive-notes" name="notes" label="Notes" placeholder="Optional notes" help="Additional context is optional." />
        </div>
      </div>
      <div class="primitive-form-actions">
        <PrimitiveButton type="submit">Submit example</PrimitiveButton>
        {#if submitted}<span class="primitive-form-status" role="status" aria-live="polite">Submitted locally.</span>{/if}
      </div>
    </form>
  </section>
</section>
