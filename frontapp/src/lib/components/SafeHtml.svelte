<!-- eslint-disable svelte/no-at-html-tags -->
<script>
  import { sanitizeHtml } from '$lib/utils/sanitizeHtml.js';

  export let html = '';
  export let tag = 'div';
  export let sanitizeContent = true;
  let element;
  export { element };

  $: safeHtml = sanitizeContent ? sanitizeHtml(html) : html;
</script>

<svelte:element this={tag} bind:this={element} {...$$restProps}>
  <!-- svelte-ignore svelte/no-at-html-tags -->
  {@html safeHtml}
</svelte:element>

