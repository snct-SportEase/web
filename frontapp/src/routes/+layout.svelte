<script>
  import '../app.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import { page } from '$app/stores';
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { ensurePushSubscription, userHasPushEligibleRole } from '$lib/utils/push.js';

  $: data = $page.data;
  let pushSetupTriggered = false;

  onMount(() => {
    if (browser && 'serviceWorker' in navigator) {
      window.addEventListener('load', async () => {
        try {
          const registration = await navigator.serviceWorker.register('/service-worker.js');
          
          // Check for updates periodically (every hour)
          setInterval(() => {
            registration.update();
          }, 60 * 60 * 1000);
          
          // Listen for service worker updates
          registration.addEventListener('updatefound', () => {
            const newWorker = registration.installing;
            
            if (newWorker) {
              newWorker.addEventListener('statechange', () => {
                // When the new service worker is activated, reload the page
                if (newWorker.state === 'activated') {
                  // Check if there's a controller change (new SW is controlling the page)
                  if (navigator.serviceWorker.controller) {
                    window.location.reload();
                  }
                }
              });
            }
          });
          
          // Also check for updates when the page becomes visible again
          document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
              registration.update();
            }
          });
        } catch (error) {
          console.error('Service Worker registration failed:', error);
        }
      });
    }
  });

  // 通知の自動設定は無効化（ユーザーが明示的に有効化するまで待つ）
  // $: if (browser && data?.user && !pushSetupTriggered && userHasPushEligibleRole(data.user)) {
  //   pushSetupTriggered = true;
  //   ensurePushSubscription().catch((error) => {
  //     console.error('[push] failed to ensure subscription', error);
  //     pushSetupTriggered = false;
  //   });
  // }

  function openSidebar() {
    isSidebarOpen.set(true);
  }

  function closeSidebar() {
    isSidebarOpen.set(false);
  }
</script>

<div class="app-container">
  {#if data.user}
    <!-- モバイル用オーバーレイ背景 -->
    {#if $isSidebarOpen}
      <button 
        type="button"
        class="sidebar-overlay md:hidden"
        on:click={closeSidebar}
        on:touchstart|stopPropagation={closeSidebar}
        on:keydown={(e) => e.key === 'Enter' || e.key === ' ' ? closeSidebar() : null}
        aria-label="メニューを閉じる"
      ></button>
    {/if}
    <Sidebar user={data.user} />
  {/if}
  <main class="main-content">
    {#if !$isSidebarOpen}
      <button type="button" on:click={openSidebar} class="open-sidebar-button" aria-label="サイドバーを開く">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
      </button>
    {/if}
    <slot />
  </main>
</div>
