<script>
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';
  import EditDisplayNameModal from '$lib/components/EditDisplayNameModal.svelte';
  import PWANotificationBanner from '$lib/components/PWANotificationBanner.svelte';
  import { isPWAInstalled, isPWAInstallable } from '$lib/utils/pwa.js';
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';

  let { children } = $props();
  let { data } = $page;
  let user = $derived(data.user);

  let showEditDisplayNameModal = $state(false);
  let showPWANotification = $state(false);
  let isMobile = $state(false);
  
  onMount(() => {
    if (browser) {
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
      const response = await fetch('/api/auth/logout', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        window.location.href = '/';
      }
    } catch {
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
    <header class="bg-white shadow-sm p-4 sticky top-0 z-[100] pointer-events-auto">
      <div class="flex justify-between items-center pointer-events-auto">
        <div class="flex items-center pointer-events-auto">
          <button type="button" onclick={openSidebar} class="mr-4 p-2 rounded-md hover:bg-gray-100 pointer-events-auto" aria-label="サイドバーを開く">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
          </button>
          <a href="/dashboard" data-sveltekit-preload-data="hover" class="flex items-center"><h1 class="text-2xl font-bold text-gray-800">Dashboard</h1></a>
        </div>
        <div class="flex items-center pointer-events-auto">
          <button 
            type="button"
            onclick={handleDisplayNameClick}
            class="mr-4 flex items-center {isMobile ? 'px-2 space-x-0' : 'space-x-2 px-3'} py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200 border border-gray-200 rounded-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 pointer-events-auto"
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
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 transition-colors duration-200 pointer-events-auto"
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
      <p>SportEase @ 仙台高専行事委員会 佐藤佑作 2301059</p>
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
