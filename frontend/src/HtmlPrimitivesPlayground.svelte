<script lang="ts">
  let submitted = false
</script>

<section class="html-primitives" aria-labelledby="html-primitives-heading">
  <p class="eyebrow">HTML primitives</p>
  <h2 id="html-primitives-heading">HTML primitives</h2>
  <p class="muted html-primitives-intro">Hard-coded native elements that establish the baseline for links, controls, and forms before they become shared components.</p>

  <section class="primitive-group" aria-labelledby="links-heading">
    <p class="mock-section-label">Navigation</p>
    <h3 id="links-heading">Links</h3>
    <div class="primitive-card primitive-link-list">
      <a href="/parts">Default link <span aria-hidden="true">→</span></a>
      <a href="/settings" class="primitive-link-muted">Muted link <span aria-hidden="true">→</span></a>
      <a href="/playground/html-primitives" aria-current="page">Current-page link</a>
      <a href="/parts" class="primitive-link-disabled" aria-disabled="true" on:click|preventDefault>Unavailable link</a>
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="buttons-heading">
    <p class="mock-section-label">Actions</p>
    <h3 id="buttons-heading">Buttons</h3>
    <div class="primitive-card primitive-button-row">
      <button type="button" class="primitive-button-primary">Primary</button>
      <button type="button" class="primitive-button-secondary">Secondary</button>
      <button type="button" class="primitive-button-quiet">Quiet</button>
      <button type="button" class="primitive-button-primary" disabled>Disabled</button>
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="form-controls-heading">
    <p class="mock-section-label">Input controls</p>
    <h3 id="form-controls-heading">Forms and controls</h3>
    <div class="primitive-control-grid">
      <div class="primitive-card primitive-field-card">
        <label for="primitive-text">Text input</label>
        <input id="primitive-text" type="text" value="PN-100" />
        <small>Single-line text with a visible label.</small>
      </div>
      <div class="primitive-card primitive-field-card">
        <label for="primitive-select">Select</label>
        <select id="primitive-select"><option>Steel</option><option>Aluminium</option><option>Polymer</option></select>
        <small>Native option selection.</small>
      </div>
      <div class="primitive-card primitive-field-card">
        <label for="primitive-number">Number input</label>
        <input id="primitive-number" type="number" min="0" value="10" />
        <small>Numeric input with a constraint.</small>
      </div>
      <div class="primitive-card primitive-field-card">
        <label for="primitive-textarea">Textarea</label>
        <textarea id="primitive-textarea" rows="3">A short description.</textarea>
        <small>Multi-line text entry.</small>
      </div>
    </div>
    <div class="primitive-card primitive-choice-card">
      <fieldset>
        <legend>Display preference</legend>
        <label><input type="radio" name="display-preference" checked /> Compact</label>
        <label><input type="radio" name="display-preference" /> Comfortable</label>
        <label><input type="checkbox" checked /> Show identifiers</label>
      </fieldset>
    </div>
  </section>

  <section class="primitive-group" aria-labelledby="form-heading">
    <p class="mock-section-label">Composition</p>
    <h3 id="form-heading">Form</h3>
    <form class="primitive-card primitive-form" on:submit|preventDefault={() => submitted = true}>
      <div class="primitive-form-fields">
        <div class="primitive-form-field">
          <label for="primitive-name">Definition name <span aria-hidden="true">*</span></label>
          <input id="primitive-name" name="name" required placeholder="Part number" />
          <small>Required field for the definition.</small>
        </div>
        <div class="primitive-form-field">
          <label for="primitive-notes">Notes</label>
          <textarea id="primitive-notes" name="notes" rows="3" placeholder="Optional notes"></textarea>
          <small>Additional context is optional.</small>
        </div>
      </div>
      <div class="primitive-form-actions">
        <button type="submit" class="primitive-button-primary">Submit example</button>
        {#if submitted}<span class="primitive-form-status" role="status" aria-live="polite">Submitted locally.</span>{/if}
      </div>
    </form>
  </section>
</section>
