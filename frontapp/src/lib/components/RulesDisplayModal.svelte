<script>
  import { createEventDispatcher } from 'svelte';
  import { marked } from 'marked';
    import { log } from 'console';

  export let showModal = false;
  export let rulesType;
  export let rulesContent = ''; // For markdown or text
  export let rulesPdfUrl = ''; // For pdf
  export let sportName = '競技ルール';

  const dispatch = createEventDispatcher();

  function closeModal() {
    showModal = false;
    dispatch('close');
  }

  // Close modal on escape key
  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if showModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center"
    aria-labelledby="modal-title"
    role="dialog"
    aria-modal="true"
  >
    <!-- Modal Panel -->
    <div
      class="relative bg-white rounded-lg shadow-xl transform transition-all my-auto w-[80vw] h-[80vh] max-w-none"
    >
      <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4 h-full flex flex-col">
        <div class="flex-grow flex flex-col">
          <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4" id="modal-title">
            {sportName} ルール
          </h3>
            {#if rulesType === 'markdown'}
              <div class="prose mt-2 overflow-y-auto p-2 border rounded-md h-full">
                {@html marked(rulesContent)}
              </div>
            {:else if rulesType === 'pdf' && rulesPdfUrl}
              <iframe src={rulesPdfUrl} class="mt-2 flex-grow overflow-y-auto p-2 border rounded-md h-full" title="{sportName} PDFルール"></iframe>
            {:else}
              <p>ルールは準備中です。</p>
            {/if}
        </div>
      </div>
      <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
        <button
          type="button"
          class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
          on:click={closeModal}
        >
          閉じる
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  /* No custom styles needed for centering and no overlay */
</style>
