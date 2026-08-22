<script>
  let {
    type = 'button',
    variant = 'primary',
    size = 'md',
    disabled = false,
    loading = false,
    loadingLabel = '処理中...',
    onclick,
    children,
    class: className = ''
  } = $props();

  const variantClass = $derived({
    primary: 'border-transparent bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500',
    secondary: 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 focus:ring-gray-500',
    danger: 'border-transparent bg-red-600 text-white hover:bg-red-700 focus:ring-red-500'
  }[variant] ?? 'border-transparent bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500');
  const sizeClass = $derived(size === 'sm' ? 'px-3 py-1.5 text-sm' : 'px-4 py-2 text-sm');

  function handleClick(event) {
    if (disabled || loading) {
      event.preventDefault();
      return;
    }
    onclick?.(event);
  }
</script>

<button
  {type}
  disabled={disabled || loading}
  aria-busy={loading}
  onclick={handleClick}
  class={`inline-flex items-center justify-center rounded-md border font-medium shadow-sm transition focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ${variantClass} ${sizeClass} ${className}`}
>
  {#if loading}
    <svg class="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
    </svg>
    {loadingLabel}
  {:else}
    {@render children?.()}
  {/if}
</button>
