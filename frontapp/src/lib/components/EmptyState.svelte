<script>
  let {
    title,
    description = null,
    children,
    class: className = ''
  } = $props();
</script>

<div class={`rounded-lg border border-gray-200 bg-gray-50 p-8 text-center text-gray-500 ${className}`}>
  {#if children}
    <div class="mb-3 flex justify-center">
      {@render children()}
    </div>
  {/if}
  <p class="font-medium text-gray-700">{title}</p>
  {#if description}
    <p class="mt-1 text-sm">{description}</p>
  {/if}
</div>
