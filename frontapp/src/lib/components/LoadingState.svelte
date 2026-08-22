<script>
  let {
    label = '読み込み中...',
    size = 'md',
    class: className = ''
  } = $props();

  const spinnerSize = $derived(size === 'sm' ? 'h-5 w-5' : 'h-10 w-10');
</script>

<div class={`flex justify-center p-12 ${className}`} role="status">
  <div class={`animate-spin rounded-full border-b-2 border-indigo-600 ${spinnerSize}`} aria-hidden="true"></div>
  <span class="sr-only">{label}</span>
</div>
