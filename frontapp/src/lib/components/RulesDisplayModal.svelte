<script>
  import { createEventDispatcher } from 'svelte';
  import { marked } from 'marked';
  import SafeHtml from '$lib/components/SafeHtml.svelte';

  export let showModal = false;
  export let rulesType;
  export let rulesContent = ''; // For markdown or text
  export let rulesPdfUrl = ''; // For pdf
  export let sportName = '競技ルール';

  const dispatch = createEventDispatcher();

  let renderedRules = '';

  $: renderedRules = rulesType === 'markdown' ? marked.parse(rulesContent || '') : '';

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
      class="relative bg-white rounded-lg shadow-xl transform transition-all my-auto w-[80vw] h-[80vh] max-w-none flex flex-col"
    >
      <!-- Header -->
      <div class="px-6 pt-6 pb-4 border-b border-gray-200 flex-shrink-0">
        <h3 class="text-lg leading-6 font-medium text-gray-900" id="modal-title">
          {sportName} ルール
        </h3>
      </div>

      <!-- Content Area (Scrollable) -->
      <div class="flex-1 overflow-y-auto px-6 py-4">
        {#if rulesType === 'markdown'}
          <div class="bg-gray-50 rounded-lg border border-gray-200">
            <SafeHtml
              class="prose prose-slate max-w-none px-6 py-4 mx-auto"
              html={renderedRules}
            />
          </div>
        {:else if rulesType === 'pdf' && rulesPdfUrl}
          <iframe src={rulesPdfUrl} class="w-full h-full border rounded-md" title="{sportName} PDFルール"></iframe>
        {:else}
          <p class="text-gray-600">ルールは準備中です。</p>
        {/if}
      </div>

      <!-- Footer (Fixed) -->
      <div class="bg-gray-50 px-6 py-4 border-t border-gray-200 flex-shrink-0 flex justify-end">
        <button
          type="button"
          class="inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:text-sm"
          on:click={closeModal}
        >
          閉じる
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.prose) {
    color: #334155;
    line-height: 1.75;
  }

  :global(.prose h1) {
    font-size: 2em;
    font-weight: 700;
    margin-top: 1.5em;
    margin-bottom: 0.75em;
    color: #1e293b;
    border-bottom: 2px solid #e2e8f0;
    padding-bottom: 0.5em;
  }

  :global(.prose h2) {
    font-size: 1.5em;
    font-weight: 600;
    margin-top: 1.25em;
    margin-bottom: 0.75em;
    color: #1e293b;
    border-bottom: 1px solid #e2e8f0;
    padding-bottom: 0.4em;
  }

  :global(.prose h3) {
    font-size: 1.25em;
    font-weight: 600;
    margin-top: 1em;
    margin-bottom: 0.5em;
    color: #334155;
  }

  :global(.prose p) {
    margin-top: 1em;
    margin-bottom: 1em;
  }

  :global(.prose ul),
  :global(.prose ol) {
    margin-top: 1em;
    margin-bottom: 1em;
    padding-left: 1.75em;
  }

  :global(.prose li) {
    margin-top: 0.5em;
    margin-bottom: 0.5em;
  }

  :global(.prose code) {
    background-color: #f1f5f9;
    color: #e11d48;
    padding: 0.125em 0.375em;
    border-radius: 0.25em;
    font-size: 0.9em;
    font-family: 'Courier New', monospace;
  }

  :global(.prose pre) {
    background-color: #1e293b;
    color: #f1f5f9;
    padding: 1em;
    border-radius: 0.5em;
    overflow-x: auto;
    margin-top: 1.5em;
    margin-bottom: 1.5em;
  }

  :global(.prose pre code) {
    background-color: transparent;
    color: inherit;
    padding: 0;
  }

  :global(.prose blockquote) {
    border-left: 4px solid #3b82f6;
    padding-left: 1em;
    margin-left: 0;
    margin-top: 1.5em;
    margin-bottom: 1.5em;
    color: #64748b;
    font-style: italic;
  }

  :global(.prose table) {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1.5em;
    margin-bottom: 1.5em;
  }

  :global(.prose th),
  :global(.prose td) {
    border: 1px solid #e2e8f0;
    padding: 0.75em;
    text-align: left;
  }

  :global(.prose th) {
    background-color: #f8fafc;
    font-weight: 600;
    color: #1e293b;
  }

  :global(.prose a) {
    color: #3b82f6;
    text-decoration: underline;
    text-underline-offset: 2px;
  }

  :global(.prose a:hover) {
    color: #2563eb;
  }

  :global(.prose img) {
    max-width: 100%;
    height: auto;
    border-radius: 0.5em;
    margin-top: 1.5em;
    margin-bottom: 1.5em;
  }

  :global(.prose strong) {
    font-weight: 600;
    color: #1e293b;
  }

  :global(.prose hr) {
    border: none;
    border-top: 2px solid #e2e8f0;
    margin: 2em 0;
  }

  /* スクロールバーのスタイル */
  :global(.overflow-y-auto)::-webkit-scrollbar {
    width: 8px;
  }

  :global(.overflow-y-auto)::-webkit-scrollbar-track {
    background: #f1f5f9;
    border-radius: 4px;
  }

  :global(.overflow-y-auto)::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 4px;
  }

  :global(.overflow-y-auto)::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
  }
</style>
