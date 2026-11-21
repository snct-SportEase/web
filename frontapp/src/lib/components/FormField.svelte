<script lang="ts">
  import type { Snippet } from 'svelte';
  let {
    label,
    inputId,
    description = null,
    labelClass = 'block text-sm font-medium text-gray-700',
    children,
    class: className = ''
  }: { label: string; inputId?: string; description?: string | null; labelClass?: string; children?: Snippet; class?: string } = $props();
</script>

<div class={className}>
  <label for={inputId} class={labelClass}>{label}</label>
  {@render children?.()}
  {#if description}
    <p class="mt-1 text-xs text-gray-500">{description}</p>
  {/if}
</div>
