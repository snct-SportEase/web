<script lang="ts">
  import type { Snippet } from 'svelte';
  let {
    title,
    description = null,
    children,
    class: className = ''
  }: { title: string; description?: string | null; children?: Snippet; class?: string } = $props();
</script>

<header class={`mb-6 flex flex-wrap items-start justify-between gap-4 ${className}`}>
  <div>
    <h1 class="text-2xl font-bold text-gray-900">{title}</h1>
    {#if description}
      <p class="mt-2 text-sm text-gray-600">{description}</p>
    {/if}
  </div>
  {#if children}
    <div class="flex flex-wrap items-center gap-2">
      {@render children()}
    </div>
  {/if}
</header>
