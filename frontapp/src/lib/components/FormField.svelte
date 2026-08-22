<script>
  let {
    label,
    inputId,
    description = null,
    labelClass = 'block text-sm font-medium text-gray-700',
    children,
    class: className = ''
  } = $props();
</script>

<div class={className}>
  <label for={inputId} class={labelClass}>{label}</label>
  {@render children?.()}
  {#if description}
    <p class="mt-1 text-xs text-gray-500">{description}</p>
  {/if}
</div>
