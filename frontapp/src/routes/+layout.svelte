<script>
  import '../app.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import { page } from '$app/stores';
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  $: data = $page.data;

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

    // プルダウンリフレッシュ機能（モバイル用）
    if (browser) {
      let touchStartY = 0;
      let touchCurrentY = 0;
      let isPulling = false;
      const PULL_THRESHOLD = 80; // リロードをトリガーする距離（ピクセル）

      const handleTouchStart = (e) => {
        // スクロール位置が最上部の場合のみ有効
        if (window.scrollY === 0) {
          touchStartY = e.touches[0].clientY;
          isPulling = true;
        }
      };

      const handleTouchMove = (e) => {
        if (!isPulling) return;
        
        touchCurrentY = e.touches[0].clientY;
        const pullDistance = touchCurrentY - touchStartY;

        // 下方向へのスワイプのみ許可
        if (pullDistance > 0 && window.scrollY === 0) {
          // 視覚的フィードバック（オプション：必要に応じて実装）
          // ここではデフォルトのブラウザ動作に任せる
        } else {
          isPulling = false;
        }
      };

      const handleTouchEnd = () => {
        if (!isPulling) return;

        const pullDistance = touchCurrentY - touchStartY;
        
        // 一定距離以上下にスワイプした場合、リロード
        if (pullDistance >= PULL_THRESHOLD && window.scrollY === 0) {
          window.location.reload();
        }

        // リセット
        isPulling = false;
        touchStartY = 0;
        touchCurrentY = 0;
      };

      // タッチイベントリスナーを追加
      document.addEventListener('touchstart', handleTouchStart, { passive: true });
      document.addEventListener('touchmove', handleTouchMove, { passive: true });
      document.addEventListener('touchend', handleTouchEnd, { passive: true });

      // クリーンアップ
      return () => {
        document.removeEventListener('touchstart', handleTouchStart);
        document.removeEventListener('touchmove', handleTouchMove);
        document.removeEventListener('touchend', handleTouchEnd);
      };
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
    <slot />
  </main>
</div>
