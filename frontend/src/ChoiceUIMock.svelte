<script lang="ts">
  const components = {
    prefix: 'PR',
    family: 'FAMILY',
    standard: ['STD', 'LOCK'],
    compact: ['CMP', 'SEAL'],
    adapter: 'ADP',
    service: ['SRV', 'LOCK'],
    nestedA: 'A1',
    nestedB: 'B2',
    parentAfterNested: 'SIZE',
    suffix: 'REV',
  }
</script>

<div class="choice-mock">
  <p class="eyebrow">Sequence definition prototype</p>
  <h1>Choice UI</h1>
  <p class="choice-mock-caption">A static example of how alternatives can branch, nest, and rejoin while keeping sequence order visible.</p>

  <section class="mock-sequence" aria-labelledby="mock-sequence-heading">
    <div class="mock-sequence-heading">
      <div>
        <h2 id="mock-sequence-heading">Connector sequence</h2>
        <p class="muted">Example: connector identifier</p>
      </div>
      <span class="mock-note">Prototype only</span>
    </div>

    <div class="flow-row before-choice">
      <span class="flow-label">Before choice</span>
      <span class="canvas-component type-literal"><strong>{components.prefix}</strong><small class="component-type">LITERAL</small></span><span class="flow-arrow">→</span><span class="canvas-component type-literal"><strong>{components.family}</strong><small class="component-type">LITERAL</small></span>
    </div>

    <div class="choice-block">
      <div class="choice-title"><span class="choice-kicker">Choice</span><strong>Mounting style</strong><span class="choice-rule">pick one alternative</span></div>

      <div class="alternative-row">
        <div class="branch-label"><span class="branch-marker">A</span><span>Standard</span></div>
        <div class="branch-flow"><span class="canvas-component type-literal"><strong>{components.standard[0]}</strong><small class="component-type">LITERAL</small></span><span class="flow-arrow">→</span><span class="canvas-component type-choice"><strong>{components.standard[1]}</strong><small class="component-type">CHOICE</small></span></div>
      </div>

      <div class="alternative-row nested-alternative">
        <div class="branch-label"><span class="branch-marker">B</span><span>Compact</span></div>
        <div class="branch-flow"><span class="canvas-component type-literal"><strong>{components.compact[0]}</strong><small class="component-type">LITERAL</small></span><span class="flow-arrow">→</span><span class="canvas-component type-choice"><strong>{components.compact[1]}</strong><small class="component-type">CHOICE</small></span><span class="flow-arrow">→</span><span class="canvas-component type-places"><strong>{components.adapter}</strong><small class="component-type">PLACES</small></span></div>

        <div class="nested-choice">
          <div class="choice-title"><span class="choice-kicker">Nested choice</span><strong>Seal type</strong><span class="choice-rule">within Compact</span></div>
          <div class="alternative-row">
            <div class="branch-label"><span class="branch-marker">B1</span><span>Weatherproof</span></div>
            <div class="branch-flow"><span class="canvas-component type-choice"><strong>{components.nestedA}</strong><small class="component-type">CHOICE</small></span></div>
          </div>
          <div class="alternative-row">
            <div class="branch-label"><span class="branch-marker">B2</span><span>Indoor</span></div>
            <div class="branch-flow"><span class="canvas-component type-choice"><strong>{components.nestedB}</strong><small class="component-type">CHOICE</small></span></div>
          </div>
        </div>

        <div class="continuation-row"><span class="continuation-label">Parent continues</span><span class="flow-arrow">→</span><span class="canvas-component type-range"><strong>{components.parentAfterNested}</strong><small class="component-type">RANGE</small></span></div>
      </div>

      <div class="alternative-row">
        <div class="branch-label"><span class="branch-marker">C</span><span>Service</span></div>
        <div class="branch-flow"><span class="canvas-component type-literal"><strong>{components.service[0]}</strong><small class="component-type">LITERAL</small></span><span class="flow-arrow">→</span><span class="canvas-component type-choice"><strong>{components.service[1]}</strong><small class="component-type">CHOICE</small></span></div>
      </div>
    </div>

    <div class="flow-row after-choice">
      <span class="flow-label">After whole choice</span>
      <span class="flow-arrow">→</span><span class="canvas-component type-places"><strong>{components.suffix}</strong><small class="component-type">PLACES</small></span>
    </div>
  </section>
</div>

<style>
  .choice-mock { max-width: 74rem; }
  .choice-mock-caption { max-width: 42rem; margin: 0 0 1.75rem; color: var(--muted); }
  .mock-sequence { background: var(--surface); border: 1px solid var(--hairline); border-radius: .35rem; overflow: hidden; }
  .mock-sequence-heading { display: flex; align-items: end; justify-content: space-between; gap: 1rem; padding: 1.1rem 1.25rem; border-bottom: 1px solid var(--hairline); }
  .mock-sequence h2 { margin: 0; font-size: 1.3rem; }
  .mock-sequence-heading p { margin: .2rem 0 0; font-size: .8rem; }
  .mock-note, .choice-rule, .flow-label, .continuation-label { color: var(--muted); font: .68rem ui-monospace, monospace; letter-spacing: .04em; }
  .flow-row { display: flex; align-items: center; flex-wrap: wrap; gap: .45rem; min-height: 4.5rem; padding: .85rem 1.25rem; }
  .before-choice { border-bottom: 1px solid var(--hairline-light); }
  .after-choice { border-top: 1px solid var(--hairline-light); }
  .flow-label { width: 7.5rem; text-transform: uppercase; }
  .mock-sequence .canvas-component { width: 5.25rem; min-height: 4.25rem; margin: 0; padding: .85rem .45rem .8rem; font-size: 1.25rem; cursor: default; }
  .mock-sequence .canvas-component .component-type { bottom: .25rem; font-size: .5rem; }
  .flow-arrow { color: var(--accent); font-weight: 600; }
  .choice-block { margin: 1rem 1.25rem; border-left: 3px solid var(--accent); background: var(--ground); }
  .choice-title { display: flex; align-items: baseline; flex-wrap: wrap; gap: .55rem; padding: .75rem 1rem; background: var(--accent-light); border-bottom: 1px solid var(--hairline); }
  .choice-kicker { color: var(--accent); font: .65rem ui-monospace, monospace; font-weight: 700; letter-spacing: .1em; text-transform: uppercase; }
  .choice-title strong { font-family: Georgia, serif; font-weight: 400; }
  .choice-rule { margin-left: auto; }
  .alternative-row { position: relative; display: flex; align-items: center; flex-wrap: wrap; gap: .7rem; padding: .8rem 1rem .8rem 1.25rem; border-bottom: 1px solid var(--hairline-light); }
  .alternative-row:last-child { border-bottom: 0; }
  .alternative-row::before { position: absolute; top: 0; bottom: 0; left: .55rem; width: 1px; background: var(--hairline); content: ''; }
  .branch-label { display: inline-flex; align-items: center; gap: .45rem; width: 9rem; color: var(--muted); font-size: .78rem; }
  .branch-marker { z-index: 1; display: inline-grid; place-items: center; width: 1.5rem; height: 1.5rem; color: var(--surface); background: var(--accent); border-radius: 50%; font: .62rem ui-monospace, monospace; }
  .branch-flow { display: flex; align-items: center; flex-wrap: wrap; gap: .45rem; }
  .nested-alternative { display: block; padding-bottom: .9rem; }
  .nested-alternative > .branch-label, .nested-alternative > .branch-flow { display: inline-flex; vertical-align: middle; }
  .nested-choice { margin: .9rem 0 .8rem 9.7rem; border: 1px solid var(--hairline); border-left: 2px solid var(--accent); background: var(--surface); }
  .nested-choice .choice-title { padding: .55rem .75rem; }
  .nested-choice .alternative-row { padding-top: .55rem; padding-bottom: .55rem; }
  .continuation-row { display: flex; align-items: center; flex-wrap: wrap; gap: .45rem; margin-left: 9.7rem; padding-top: .15rem; }

  @media (max-width: 640px) {
    .mock-sequence-heading { align-items: start; flex-direction: column; }
    .mock-note { margin-top: -.4rem; }
    .flow-label { width: 100%; }
    .choice-block { margin: .75rem .7rem; }
    .branch-label { width: 100%; }
    .mock-sequence .canvas-component { width: 4.5rem; min-height: 3.75rem; font-size: 1rem; }
    .nested-choice, .continuation-row { margin-left: 1.25rem; }
    .choice-rule { width: 100%; margin-left: 0; }
  }
</style>
