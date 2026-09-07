<script>
  import { browser } from '$app/environment';
  import { isPWAInstalled, isPWAInstallable } from '$lib/utils/pwa.js';

  let { show = false, onClose = () => {} } = $props();

  let isVisible = $state(false);

  $effect(() => {
    if (!browser || !show) {
      isVisible = false;
      return;
    }
    const timer = setTimeout(() => {
      isVisible = true;
    }, 500);
    return () => clearTimeout(timer);
  });

  function handleClose() {
    isVisible = false;
    setTimeout(() => {
      onClose();
    }, 300);
  }

  let installed = $derived(browser ? isPWAInstalled() : false);
  let installable = $derived(browser ? isPWAInstallable() : false);
</script>

{#if show && (installed || installable)}
  <div
    class="pointer-events-none fixed top-20 left-1/2 z-[110] w-[calc(100%-2rem)] max-w-md -translate-x-1/2 transform transition-all duration-300 {isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-4'}"
    role="alert"
  >
    <div class="mx-auto flex items-start space-x-3 rounded-lg bg-indigo-600 p-4 text-white shadow-lg">
      <div class="flex-shrink-0">
        {#if installed}
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        {:else}
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
          </svg>
        {/if}
      </div>
      <div class="flex-1">
        {#if installed}
          <h3 class="font-semibold text-sm mb-1">PWAが有効です</h3>
          <p class="text-xs text-indigo-100">
            SportEaseはPWAアプリとして正常に動作しています。
          </p>
        {:else if installable}
          <h3 class="font-semibold text-sm mb-1">PWAとしてインストール可能</h3>
          <p class="text-xs text-indigo-100">
            SportEaseをアプリとしてインストールすると、より快適にご利用いただけます。
          </p>
        {/if}
      </div>
      <button
        type="button"
        onclick={handleClose}
        aria-label="通知を閉じる"
        class="pointer-events-auto flex-shrink-0 text-indigo-200 transition-colors hover:text-white"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
        </svg>
      </button>
    </div>
  </div>
{/if}
