<script>
  let {
    open = $bindable(false),
    title,
    onclose,
    children,
    footer
  } = $props();

  function close() {
    open = false;
    onclose?.();
  }
</script>

{#if open}
  <div class="fixed z-50 inset-0 overflow-y-auto" role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <div class="flex items-end justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 bg-gray-500 opacity-75" aria-hidden="true"></div>
      <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>

      <div class="relative z-50 inline-block w-full overflow-hidden rounded-lg bg-white text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-lg sm:align-middle">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          {#if title}
            <h3 id="modal-title" class="text-lg leading-6 font-medium text-gray-900">{title}</h3>
          {/if}
          {@render children?.()}
        </div>
        {#if footer}
          <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            {@render footer({ close })}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
