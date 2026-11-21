<script lang="ts">
  import type { Snippet } from 'svelte';
  let { header, children, class: className = '', label = '一覧表' }: { header?: Snippet; children?: Snippet; class?: string; label?: string } = $props();
</script>

<!-- 横スクロール領域をキーボードでも操作できるようにする。 -->
<!-- eslint-disable-next-line svelte/no-unused-svelte-ignore -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div class={`table-scroll ${className}`} role="region" aria-label={label} tabindex="0">
  <table class="min-w-full divide-y divide-gray-200 text-sm">
    {#if header}
      <thead class="bg-gray-50">
        {@render header()}
      </thead>
    {/if}
    <tbody class="divide-y divide-gray-100 bg-white">
      {@render children?.()}
    </tbody>
  </table>
</div>
