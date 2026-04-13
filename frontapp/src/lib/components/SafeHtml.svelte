<!-- eslint-disable svelte/no-at-html-tags -->
<script>
  import { sanitizeHtml } from '$lib/utils/sanitizeHtml.js';

  let { html = '', tag = 'div', sanitizeContent = true, element = $bindable(), ...restProps } = $props();

  let safeHtml = $derived(sanitizeContent ? sanitizeHtml(html) : html);

  // data-mk-color 属性を持つ要素に JS 経由で色を適用する。
  // style="" 属性を使わないため CSP の unsafe-inline が不要になる。
  // 値はマークダウンパーサー側で hex (#rrggbb) または "red" に制限済み。
  $effect(() => {
    safeHtml; // depend on safeHtml
    if (!element) return;
    element.querySelectorAll('[data-mk-color]').forEach((el) => {
      const color = el.dataset.mkColor;
      if (/^#[0-9a-fA-F]{3,6}$/.test(color) || color === 'red') {
        el.style.color = color;
      }
    });
  });
</script>

<svelte:element this={tag} bind:this={element} {...restProps}>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
  {@html safeHtml}
</svelte:element>
