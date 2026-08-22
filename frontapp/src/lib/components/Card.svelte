<script>
  let { variant = 'default', children, class: className = '' } = $props();

  const variantClass = $derived(
    variant === 'featured' ? 'border-indigo-100' : 'border-gray-200'
  );
</script>

<div class={`rounded-lg border bg-white shadow-sm ${variantClass} ${className}`}>
  {@render children?.()}
</div>
