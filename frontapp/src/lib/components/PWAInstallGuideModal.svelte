<script>
  import { browser } from '$app/environment';
  import { getDeviceType, getBrowserType } from '$lib/utils/pwa.js';

  export let isOpen = false;
  export let onClose = () => {};

  let deviceType = 'unknown';
  let browserType = 'unknown';

  $: if (isOpen && browser) {
    deviceType = getDeviceType();
    browserType = getBrowserType();
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      onClose();
    }
  }

  function handleOverlayKeydown(event) {
    if (event.key === 'Escape') {
      event.preventDefault();
      onClose();
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onClose();
    }
  }
</script>

<!-- モーダルの背景 -->
{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm overflow-y-auto"
    role="presentation"
    tabindex="-1"
    on:click={onClose}
    on:keydown={handleOverlayKeydown}
  >
    <!-- モーダルの本体 -->
    <div
      class="w-full max-w-3xl m-4 p-6 space-y-6 bg-white rounded-lg shadow-xl"
      role="dialog"
      aria-modal="true"
      aria-labelledby="pwa-install-guide-title"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown={handleKeydown}
    >
      <div class="flex justify-between items-center">
        <h2 id="pwa-install-guide-title" class="text-2xl font-bold text-gray-800">PWAインストール方法</h2>
        <button
          type="button"
          on:click={onClose}
          aria-label="モーダルを閉じる"
          class="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <div class="space-y-6">
        <!-- iOS Safari -->
        <section class="border border-gray-200 rounded-lg p-5">
          <h3 class="text-lg font-semibold text-gray-800 mb-3 flex items-center">
            <svg class="w-6 h-6 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
            </svg>
            iOS (Safari)
          </h3>
          <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
            <li>SafariでSportEaseを開きます</li>
            <li>画面下部の共有ボタン（<svg class="w-4 h-4 inline-block" fill="currentColor" viewBox="0 0 24 24"><path d="M17.5 4.5c-1.95 0-4.05.4-5.5 1.5-1.45-1.1-3.55-1.5-5.5-1.5-1.45 0-2.99.22-4.28.79C1.49 5.62 1 6.33 1 7.14v11.28c0 1.3 1.22 2.26 2.48 1.94.98-.25 2.02-.36 3.02-.36 1.56 0 3.22.26 4.56.92.6.3 1.28.3 1.88 0 1.34-.67 3-.92 4.56-.92 1 0 2.04.11 3.02.36 1.26.33 2.48-.63 2.48-1.94V7.14c0-.81-.49-1.52-1.22-1.85-1.29-.57-2.83-.79-4.28-.79zM21 17.23c0 .63-.58 1.09-1.2.98-.75-.14-1.53-.2-2.3-.2-1.7 0-4.15.65-5.5 1.5V8c1.35-.85 3.8-1.5 5.5-1.5.77 0 1.55.06 2.3.2.62.11 1.2.35 1.2.98v10.55z"/></svg>）をタップします</li>
            <li>「ホーム画面に追加」を選択します</li>
            <li>「追加」をタップして完了です</li>
          </ol>
        </section>

        <!-- Android Chrome -->
        <section class="border border-gray-200 rounded-lg p-5">
          <h3 class="text-lg font-semibold text-gray-800 mb-3 flex items-center">
            <svg class="w-6 h-6 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0C8.21 0 4.831 1.757 2.632 4.5l1.363 1.464C6.295 3.64 9.076 2 12 2c2.925 0 5.706 1.64 7.005 3.964l1.363-1.464C18.169 1.757 14.79 0 12 0zm0 4C9.239 4 6.81 5.29 5.246 7.345l1.363 1.464C7.96 7.19 9.89 6 12 6c2.11 0 4.04 1.19 5.391 2.809l1.363-1.464C17.19 5.29 14.761 4 12 4zm0 4c-2.374 0-4.543 1.181-5.857 3.135l1.363 1.464C8.694 10.88 10.258 10 12 10c1.742 0 3.306.88 4.494 2.599l1.363-1.464C16.543 9.181 14.374 8 12 8zm0 4c-1.105 0-2 .895-2 2s.895 2 2 2 2-.895 2-2-.895-2-2-2z"/>
            </svg>
            Android (Chrome)
          </h3>
          <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
            <li>ChromeでSportEaseを開きます</li>
            <li>画面右上のメニューボタン（⋮）をタップします</li>
            <li>「ホーム画面に追加」または「インストール」を選択します</li>
            <li>確認ダイアログで「インストール」をタップします</li>
          </ol>
          <div class="mt-3 p-3 bg-blue-50 rounded-md text-xs text-blue-800">
            <strong>ヒント:</strong> 一部のブラウザでは、ページ下部に表示されるインストールバナーからもインストールできます。
          </div>
        </section>

        <!-- Windows Chrome/Edge -->
        <section class="border border-gray-200 rounded-lg p-5">
          <h3 class="text-lg font-semibold text-gray-800 mb-3 flex items-center">
            <svg class="w-6 h-6 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M3 12V6.75l6-1.32v6.48L3 12zm17-9v8.75l-10 .15V5.21L20 3zM3 13l6 .09v6.81l-6-1.15V13zm17 .25V22l-10-1.8v-7.15l10 .15z"/>
            </svg>
            Windows (Chrome/Edge)
          </h3>
          <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
            <li>ChromeまたはEdgeでSportEaseを開きます</li>
            <li>アドレスバーの右側にあるインストールアイコン（<svg class="w-4 h-4 inline-block" fill="currentColor" viewBox="0 0 24 24"><path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/></svg>）をクリックします</li>
            <li>「インストール」をクリックします</li>
            <li>確認ダイアログで「インストール」をクリックして完了です</li>
          </ol>
        </section>

        <!-- macOS Safari -->
        <section class="border border-gray-200 rounded-lg p-5">
          <h3 class="text-lg font-semibold text-gray-800 mb-3 flex items-center">
            <svg class="w-6 h-6 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm0 2c5.514 0 10 4.486 10 10s-4.486 10-10 10S2 17.514 2 12 6.486 2 12 2z"/>
              <path d="M12 5v7l5 3-1 2-6-4V5h2z"/>
            </svg>
            macOS (Safari)
          </h3>
          <ol class="list-decimal list-inside space-y-2 text-sm text-gray-700">
            <li>SafariでSportEaseを開きます</li>
            <li>メニューバーの「ファイル」→「ホーム画面に追加...」を選択します</li>
            <li>アプリ名を確認して「追加」をクリックします</li>
            <li>DockまたはLaunchpadからアプリを起動できます</li>
          </ol>
        </section>

        <!-- 共通のメリット -->
        <section class="border border-indigo-200 bg-indigo-50 rounded-lg p-5">
          <h3 class="text-lg font-semibold text-indigo-800 mb-3">PWAとしてインストールするメリット</h3>
          <ul class="list-disc list-inside space-y-1 text-sm text-indigo-700">
            <li>オフラインでも基本的な機能が利用できます</li>
            <li>プッシュ通知を受け取れます</li>
            <li>ホーム画面から直接起動できます</li>
            <li>アプリのように快適に操作できます</li>
            <li>データ使用量を節約できます</li>
          </ul>
        </section>
      </div>

      <!-- 閉じるボタン -->
      <div class="flex justify-end pt-4 border-t border-gray-200">
        <button
          type="button"
          on:click={onClose}
          class="px-6 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          閉じる
        </button>
      </div>
    </div>
  </div>
{/if}
