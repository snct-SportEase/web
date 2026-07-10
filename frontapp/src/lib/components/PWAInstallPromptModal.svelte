<script>
  import PWAInstallGuideModal from '$lib/components/PWAInstallGuideModal.svelte';
  import {
    closePWAInstallDialog,
    promptPWAInstall,
    pwaInstallDialogOpen,
    pwaInstallPromptAvailable
  } from '$lib/stores/pwaInstallStore.js';

  let isInstalling = $state(false);
  let installMessage = $state('');
  let showInstallGuide = $state(false);

  function closeDialog() {
    installMessage = '';
    closePWAInstallDialog();
  }

  function openInstallGuide() {
    closeDialog();
    showInstallGuide = true;
  }

  async function installOnDevice() {
    isInstalling = true;
    installMessage = '';

    try {
      const result = await promptPWAInstall();
      if (result?.outcome === 'dismissed') {
        installMessage = 'インストールはキャンセルされました。もう一度試す場合はページを再読み込みしてください。';
      } else if (result?.outcome === 'unavailable') {
        installMessage = 'このブラウザでは自動インストールを開始できません。手動のインストール方法をご確認ください。';
      }
    } catch (error) {
      console.error('[pwa] Failed to show install prompt:', error);
      installMessage = 'インストール画面を開けませんでした。手動のインストール方法をご確認ください。';
    } finally {
      isInstalling = false;
    }
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') closeDialog();
  }
</script>

{#if $pwaInstallDialogOpen}
  <div
    class="fixed inset-0 z-[9998] flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
    role="presentation"
    onclick={closeDialog}
    onkeydown={handleKeydown}
  >
    <div
      class="w-full max-w-md rounded-xl bg-white p-6 shadow-2xl"
      role="dialog"
      aria-modal="true"
      aria-labelledby="pwa-install-title"
      tabindex="-1"
      onclick={(event) => event.stopPropagation()}
      onkeydown={handleKeydown}
    >
      <div class="flex items-start justify-between gap-4">
        <div>
          <p class="text-sm font-semibold text-sky-700">SportEase PWA</p>
          <h2 id="pwa-install-title" class="mt-1 text-2xl font-bold text-gray-900">デバイスにインストール</h2>
        </div>
        <button
          type="button"
          class="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
          aria-label="インストール画面を閉じる"
          onclick={closeDialog}
        >
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <p class="mt-4 text-sm leading-6 text-gray-600">
        ホーム画面からすぐに起動でき、対応端末ではバックグラウンドでも通知を受け取りやすくなります。
      </p>

      <ul class="mt-4 space-y-2 rounded-lg bg-sky-50 p-4 text-sm text-sky-950">
        <li>・ホーム画面やアプリ一覧から直接起動</li>
        <li>・ブラウザの画面を省いたアプリ表示</li>
        <li>・プッシュ通知に適した利用環境</li>
      </ul>

      {#if installMessage}
        <p class="mt-4 rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-900" role="status">
          {installMessage}
        </p>
      {/if}

      <div class="mt-6 flex flex-col gap-3 sm:flex-row sm:justify-end">
        <button
          type="button"
          class="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          onclick={openInstallGuide}
        >
          インストール方法を見る
        </button>
        {#if $pwaInstallPromptAvailable}
          <button
            type="button"
            class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isInstalling}
            onclick={installOnDevice}
          >
            {isInstalling ? '確認中...' : 'デバイスにインストール'}
          </button>
        {:else}
          <button
            type="button"
            class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700"
            onclick={openInstallGuide}
          >
            手動でインストール
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<PWAInstallGuideModal
  isOpen={showInstallGuide}
  onClose={() => showInstallGuide = false}
/>
