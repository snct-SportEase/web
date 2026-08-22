<script>
  let { header, children, class: className = '' } = $props();
</script>

<div class={`overflow-x-auto ${className}`}>
  <table class="min-w-full divide-y divide-gray-200 text-sm">
    {#if header}
      <thead class="bg-gray-50">
        {@render header()}
      </thead>
    {/if}
    <tbody class="divide-y divide-gray-100 bg-white">
      {@render children?.()}
    </tbody>
  </table>
</div>
