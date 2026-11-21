<script lang="ts">
  import type { Snippet } from 'svelte';
  let { variant = 'default', children, class: className = '' }: { variant?: 'default' | 'featured'; children?: Snippet; class?: string } = $props();

  const variantClass = $derived(
    variant === 'featured' ? 'border-indigo-100' : 'border-gray-200'
  );
</script>

<div class={`rounded-lg border bg-white shadow-sm ${variantClass} ${className}`}>
  {@render children?.()}
</div>
