<script lang="ts">
  export type Choice = { value: string; label: string; kind: 'radio' | 'checkbox'; checked?: boolean }
  export let legend: string
  export let choices: Choice[] = []
</script>

<fieldset class="primitive-choice-group">
  <legend>{legend}</legend>
  {#each choices as choice}
    <label><input type={choice.kind} name={choice.kind === 'radio' ? legend.toLowerCase().replaceAll(' ', '-') : undefined} value={choice.value} checked={choice.checked ?? false} /> {choice.label}</label>
  {/each}
</fieldset>
