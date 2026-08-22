<script>
  let {
    variant = 'info',
    title = null,
    children,
    class: className = ''
  } = $props();

  const variantClass = $derived({
    info: 'border-blue-200 bg-blue-50 text-blue-800',
    success: 'border-green-200 bg-green-50 text-green-800',
    warning: 'border-amber-200 bg-amber-50 text-amber-800',
    error: 'border-red-200 bg-red-50 text-red-800'
  }[variant] ?? 'border-blue-200 bg-blue-50 text-blue-800');
</script>

<div
  role={variant === 'error' ? 'alert' : 'status'}
  class={`rounded-md border px-4 py-3 text-sm ${variantClass} ${className}`}
>
  {#if title}
    <p class="font-semibold">{title}</p>
  {/if}
  <div class:mt-1={!!title}>
    {@render children?.()}
  </div>
</div>
