<script>
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import EditDisplayNameModal from '$lib/components/EditDisplayNameModal.svelte';
  import PWANotificationBanner from '$lib/components/PWANotificationBanner.svelte';
  import { isPWAInstalled, isPWAInstallable } from '$lib/utils/pwa.js';
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';
  import { pushSubscriptionStatus } from '$lib/stores/pushSubscriptionStore.js';
  import { openPWAInstallDialog } from '$lib/stores/pwaInstallStore.js';

  let { children } = $props();
  let { data } = $page;
  let user = $derived(data.user);

  let showEditDisplayNameModal = $state(false);
  let showPWANotification = $state(false);
  let isMobile = $state(false);
  let isPWA = $state(true);
  let canSeeNotifications = $derived(user?.roles?.some(role => ['student', 'admin', 'root'].includes(role.name)));
  let shouldShowPWASetupBadge = $derived(canSeeNotifications && !isPWA);
  let shouldShowPushSetupBadge = $derived(
    canSeeNotifications &&
    $pushSubscriptionStatus.loaded &&
    $pushSubscriptionStatus.isSupported &&
    $pushSubscriptionStatus.vapidKeySet &&
    !$pushSubscriptionStatus.isSubscribed
  );
  
  onMount(() => {
    if (browser) {
      isPWA = isPWAInstalled();

      // 初回ログイン時のみ通知を表示（localStorageで管理）
      const hasSeenNotification = localStorage.getItem('pwa-notification-seen');
      if (!hasSeenNotification && (isPWAInstalled() || isPWAInstallable())) {
        showPWANotification = true;
      }
      
      // 画面サイズを判定
      const checkMobile = () => {
        isMobile = window.innerWidth < 768;
      };
      checkMobile();
      window.addEventListener('resize', checkMobile);
      
      return () => {
        window.removeEventListener('resize', checkMobile);
      };
    }
  });

  function handleClosePWANotification() {
    showPWANotification = false;
    if (browser) {
      localStorage.setItem('pwa-notification-seen', 'true');
    }
  }

  function handleDisplayNameClick(e) {
    e?.preventDefault?.();
    e?.stopPropagation?.();
    console.log('handleDisplayNameClick called');
    showEditDisplayNameModal = true;
  }

  function handleCloseEditDisplayNameModal() {
    showEditDisplayNameModal = false;
  }

  async function handleSaveDisplayName(newDisplayName) {
    const response = await fetch('/api/user/profile', {
      method: 'PUT',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        display_name: newDisplayName,
        class_id: user?.class_id || 0
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || '表示名の更新に失敗しました。');
    }

    window.location.reload();
  }

  async function handleLogout() {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });
    } catch {
      // Clear local data even if the server is temporarily unreachable.
    } finally {
      // Remove legacy caches that may contain responses from the previous
      // account. The current Service Worker will repopulate public assets only.
      if (browser && 'caches' in window) {
        try {
          const cacheNames = await window.caches.keys();
          await Promise.all(cacheNames.map((cacheName) => window.caches.delete(cacheName)));
        } catch {
          // A cache failure must not prevent the logout redirect.
        }
      }
      window.location.href = '/';
    }
  }

  function openSidebar(e) {
    e?.preventDefault?.();
    e?.stopPropagation?.();
    console.log('openSidebar called');
    isSidebarOpen.set(true);
  }
</script>

<div class="min-h-screen bg-gray-50 flex flex-col">
  {#if (!$isSidebarOpen || (browser && !isMobile)) && user?.is_profile_complete}
    <header class="app-layer-header sticky top-0 z-30 bg-white p-2 shadow-sm pointer-events-auto md:p-4">
      <div class="flex justify-between items-center pointer-events-auto">
        <div class="flex items-center pointer-events-auto">
          <button type="button" onclick={openSidebar} class="mr-1 rounded-md p-2 hover:bg-gray-100 pointer-events-auto md:mr-4" aria-label="サイドバーを開く">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
          </button>
          <a href="/dashboard" data-sveltekit-preload-data="hover" class="flex items-center"><h1 class="text-xl font-bold text-gray-800 md:text-2xl">Dashboard</h1></a>
        </div>
        <div class="flex items-center pointer-events-auto">
          <div class="mr-3 hidden items-center gap-1 md:flex" aria-label="報告・提案">
            <a
              href="https://github.com/snct-SportEase/web/issues/new?template=bug_report.yml"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-md p-2 text-gray-600 hover:bg-red-50 hover:text-red-700"
              title="バグを報告"
              aria-label="バグを報告"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M5.07 19h13.86c1.54 0 2.5-1.67 1.73-3L13.73 4c-.77-1.33-2.69-1.33-3.46 0L3.34 16c-.77 1.33.19 3 1.73 3Z" /></svg>
            </a>
            <a
              href="https://github.com/snct-SportEase/web/security/advisories/new"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-md p-2 text-gray-600 hover:bg-amber-50 hover:text-amber-700"
              title="脆弱性を非公開で報告"
              aria-label="脆弱性を非公開で報告"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9c1.7 0 3-1.3 3-3s-1.3-3-3-3-3 1.3-3 3 1.3 3 3 3Zm0 0c-3.3 0-6 2.2-6 5v3h12v-3c0-2.8-2.7-5-6-5Zm0 4v4m-2-2h4" /></svg>
            </a>
            <a
              href="https://github.com/snct-SportEase/web/issues/new?template=feature_request.yml"
              target="_blank"
              rel="noopener noreferrer"
              class="rounded-md p-2 text-gray-600 hover:bg-indigo-50 hover:text-indigo-700"
              title="新機能を提案"
              aria-label="新機能を提案"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3a7 7 0 0 0-4 12.75V18a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-2.25A7 7 0 0 0 12 3Zm-2 19h4" /></svg>
            </a>
          </div>
          <details class="relative mr-1 md:hidden">
            <summary
              class="flex cursor-pointer list-none items-center rounded-md p-2 text-gray-600 hover:bg-gray-100 [&::-webkit-details-marker]:hidden"
              aria-label="報告・提案メニューを開く"
              title="報告・提案"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.75h.01M12 12h.01M12 17.25h.01" /></svg>
            </summary>
            <div class="absolute right-0 top-full z-40 mt-2 w-52 rounded-md border border-gray-200 bg-white py-1 shadow-lg" aria-label="報告・提案">
              <a
                href="https://github.com/snct-SportEase/web/issues/new?template=bug_report.yml"
                target="_blank"
                rel="noopener noreferrer"
                class="block px-4 py-2 text-sm text-gray-700 hover:bg-red-50 hover:text-red-700"
              >
                バグを報告
              </a>
              <a
                href="https://github.com/snct-SportEase/web/security/advisories/new"
                target="_blank"
                rel="noopener noreferrer"
                class="block px-4 py-2 text-sm text-gray-700 hover:bg-amber-50 hover:text-amber-700"
              >
                脆弱性を非公開で報告
              </a>
              <a
                href="https://github.com/snct-SportEase/web/issues/new?template=feature_request.yml"
                target="_blank"
                rel="noopener noreferrer"
                class="block px-4 py-2 text-sm text-gray-700 hover:bg-indigo-50 hover:text-indigo-600"
              >
                新機能を提案
              </a>
            </div>
          </details>
          {#if shouldShowPWASetupBadge}
            <button
              type="button"
              onclick={openPWAInstallDialog}
              class="mr-3 rounded-md border border-sky-300 bg-sky-100 px-3 py-2 text-sm font-semibold text-sky-900 hover:bg-sky-200"
            >
              PWA未設定
            </button>
          {/if}
          {#if shouldShowPushSetupBadge}
            <a
              href="/dashboard"
              class="mr-3 rounded-md border border-amber-300 bg-amber-100 px-3 py-2 text-sm font-semibold text-amber-900 hover:bg-amber-200"
            >
              {$pushSubscriptionStatus.permission === 'denied' ? '通知拒否中' : '通知未設定'}
            </a>
          {/if}
          <button 
            type="button"
            onclick={handleDisplayNameClick}
            class="mr-1 flex items-center {isMobile ? 'px-2 space-x-0' : 'space-x-2 px-3'} py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200 border border-gray-200 rounded-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 pointer-events-auto md:mr-4"
            title="表示名をクリックして変更"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
            </svg>
            {#if !isMobile}
              <span>{user?.display_name || user?.email || 'User'}</span>
            {/if}
          </button>
          <button 
            type="button" 
            onclick={handleLogout}
            class="rounded-md bg-indigo-600 px-2.5 py-2 text-sm font-medium text-white transition-colors duration-200 hover:bg-indigo-700 pointer-events-auto md:px-4"
          >
            Logout
          </button>
        </div>
      </div>
    </header>
  {/if}

  <main class="p-8 flex-1">
    {@render children?.()}
  </main>

  <footer class="border-t border-gray-200 bg-white px-6 py-4 text-center text-sm text-gray-600">
    <div class="flex flex-col items-center gap-2">
      <p>SportEase © 仙台高専行事委員会 佐藤佑作 2301059</p>
      <a
        href="https://github.com/snct-SportEase/web"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-2 text-gray-700 transition-colors duration-200 hover:text-indigo-600"
      >
        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M12 2C6.477 2 2 6.589 2 12.248c0 4.527 2.865 8.367 6.839 9.722.5.096.682-.222.682-.494 0-.244-.009-.89-.014-1.747-2.782.617-3.369-1.372-3.369-1.372-.454-1.184-1.11-1.499-1.11-1.499-.908-.638.069-.625.069-.625 1.004.072 1.532 1.056 1.532 1.056.892 1.573 2.341 1.118 2.91.855.091-.667.349-1.118.635-1.374-2.22-.259-4.555-1.14-4.555-5.074 0-1.121.39-2.038 1.029-2.756-.103-.26-.446-1.307.098-2.724 0 0 .84-.276 2.75 1.053A9.302 9.302 0 0 1 12 6.84a9.27 9.27 0 0 1 2.504.35c1.909-1.329 2.748-1.053 2.748-1.053.546 1.417.203 2.464.1 2.724.64.718 1.027 1.635 1.027 2.756 0 3.944-2.339 4.812-4.566 5.066.359.319.678.947.678 1.909 0 1.379-.012 2.491-.012 2.829 0 .274.18.594.688.493C19.138 20.612 22 16.773 22 12.248 22 6.589 17.523 2 12 2Z" />
        </svg>
        <span>GitHub</span>
      </a>
    </div>
  </footer>

  <EditDisplayNameModal
    isOpen={showEditDisplayNameModal}
    currentDisplayName={user?.display_name || ''}
    userRoles={user?.roles || []}
    onClose={handleCloseEditDisplayNameModal}
    onSave={handleSaveDisplayName}
  />
  
  <PWANotificationBanner
    show={showPWANotification}
    onClose={handleClosePWANotification}
  />
</div>
