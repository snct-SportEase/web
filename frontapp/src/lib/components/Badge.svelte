<script>
  let { variant = 'neutral', children, class: className = '' } = $props();

  const variantClass = $derived({
    neutral: 'bg-gray-100 text-gray-800',
    info: 'bg-blue-100 text-blue-800',
    success: 'bg-green-100 text-green-800',
    warning: 'bg-amber-100 text-amber-800',
    danger: 'bg-red-100 text-red-800'
  }[variant] ?? 'bg-gray-100 text-gray-800');
</script>

<span class={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold leading-5 ${variantClass} ${className}`}>
  {@render children?.()}
</span>
