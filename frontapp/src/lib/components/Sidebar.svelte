<script>
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';
  import { goto } from '$app/navigation';
  import { invalidateAll } from '$app/navigation';
  import { browser } from '$app/environment';

  /** @type {import('@sveltejs/kit').MaybePromise<import('../../routes/$types').LayoutData>} */
  export let user;

  const hasRole = (roleName) => {
    return user?.roles?.some(role => role.name === roleName);
  };

  const isStudent = hasRole('student');
  const isAdmin = hasRole('admin');
  const isRoot = hasRole('root');

  function closeSidebar(event) {
    if (event) {
      event.preventDefault();
      event.stopPropagation();
    }
    isSidebarOpen.set(false);
  }

  function handleTouchStart(event) {
    event.preventDefault();
    event.stopPropagation();
    closeSidebar(event);
  }

  function handleClick(event) {
    closeSidebar(event);
  }

  function handleTouchEnd(event) {
    event.preventDefault();
    event.stopPropagation();
    closeSidebar(event);
  }

  async function handleLinkClick(event, href) {
    if (!href) return;
    event.preventDefault();
    
    // URLからパスを抽出（完全なURLの場合はパス部分のみを取得）
    let path = href;
    try {
      const url = new URL(href, browser ? window.location.origin : 'http://localhost');
      path = url.pathname + url.search;
    } catch {
      // 相対パスの場合はそのまま使用
      path = href;
    }
    
    isSidebarOpen.set(false);
    
    // ナビゲーションとリロード
    await invalidateAll();
    await goto(path);
    if (browser) {
      window.location.reload();
    }
  }
</script>

<aside class="w-full md:w-64 bg-gray-800 text-white flex flex-col transition-all duration-300 fixed md:relative inset-y-0 left-0 z-50 md:z-auto" class:closed={!$isSidebarOpen}>
  <div class="h-16 flex items-center justify-between px-4 relative">
    <a href="/dashboard" class="flex items-center z-10" on:click={(e) => handleLinkClick(e, '/dashboard')}><span class="text-2xl font-bold">SportEase</span></a>
    <button 
      type="button" 
      on:click={handleClick}
      on:touchstart|stopPropagation={handleTouchStart}
      on:touchend|stopPropagation={handleTouchEnd}
      class="flex items-center justify-center w-12 h-12 rounded-md hover:bg-gray-700 active:bg-gray-600 transition-colors cursor-pointer" 
      aria-label="サイドバーを閉じる"
      style="touch-action: manipulation; -webkit-tap-highlight-color: rgba(0,0,0,0.1); pointer-events: auto; position: absolute; top: 0.5rem; right: 0.5rem; z-index: 10; -webkit-user-select: none; user-select: none; background-color: transparent; border: none;"
    >
      <svg class="w-6 h-6 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
    </button>
  </div>
  <nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
    <!-- Root Menu -->
    {#if isRoot}
      <div class="pt-4">
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Root</h3>
        <a href="/dashboard/root/event-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/event-management')}>
          大会情報登録・管理
        </a>
        <a href="/dashboard/root/rainy-mode" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/rainy-mode')}>
          雨天時モード管理
        </a>
        <a href="/dashboard/root/sport-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/sport-management')}>
          競技情報登録・管理
        </a>
        <a href="/dashboard/root/notification" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/notification')}>
          通知管理
        </a>
        <a href="/dashboard/root/notification-requests" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/notification-requests')}>
          通知申請管理
        </a>
        <a href="/dashboard/root/tournament-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/tournament-management')}>
          トーナメント生成・管理
        </a>
        <a href="/dashboard/root/noon-game" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/noon-game')}>
          昼競技管理
        </a>
        <a href="/dashboard/root/whitelist-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/whitelist-management')}>
          ホワイトリスト
        </a>
        <a href="/dashboard/root/class-student-count" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/class-student-count')}>
          各クラス人数設定
        </a>
        <a href="/dashboard/root/change-username" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/change-username')}>
          ユーザー名変更
        </a>
        <a href="/dashboard/root/identify-mvp" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/root/identify-mvp')}>
          MVP確認
        </a>
      </div>
    {/if}
    
    <!-- Admin/Root Menu -->
    {#if isAdmin || isRoot}
      <div class="pt-4">
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Admin</h3>
        <a href="/dashboard/admin/class-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/class-management')}>
          クラス・チーム割り当て
        </a>
        <a href="/dashboard/admin/role-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/role-management')}>
          ロール割り当て・管理
        </a>
        <a href="/dashboard/admin/qr-code-reader" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/qr-code-reader')}>
          QRコード読み取り
        </a>
        <a href="/dashboard/admin/confirmed-participants" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/confirmed-participants')}>
          QRコード参加本登録確認
        </a>
        <a href="/dashboard/admin/insert-matche-result" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/insert-matche-result')}>
          試合結果入力
        </a>
        <a href="/dashboard/admin/attendance-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/attendance-management')}>
          出席登録
        </a>
        <!-- MVP投票の被対象者は1~2年生のみ。投票はadminとrootしか行えない -->
        <a href="/dashboard/admin/vorting-mvp" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/vorting-mvp')}>
          MVP投票
        </a>
        <a href="/dashboard/admin/sport-details-registration" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/sport-details-registration')}>
          競技詳細情報登録
        </a>
        <a href="/dashboard/admin/noon-game-results" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/admin/noon-game-results')}>
          昼競技結果入力
        </a>
      </div>
    {/if}
    
    <!-- Student Menu -->
    {#if isStudent || isAdmin || isRoot}
      <div>
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Student</h3>
        <a href="/dashboard/student/my-page" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/my-page')}>
          マイページ
        </a>
        <a href="/dashboard/student/class-info" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/class-info')}>
          クラス情報
        </a>
        <a href="/dashboard/student/sport-info" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/sport-info')}>
          競技一覧・詳細閲覧
        </a>
        <a href="/dashboard/student/tournament" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/tournament')}>
          トーナメント
        </a>
        <a href="/dashboard/student/issueqr-code" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/issueqr-code')}>
          QRコード発行
        </a>
        <a href="/dashboard/student/score-list" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/score-list')}>
          点数一覧
        </a>
        <a href="/dashboard/student/notification" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/notification')}>
          通知
        </a>
        <a href="/dashboard/student/notification-request" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/student/notification-request')}>
          通知申請
        </a>
      </div>
    {/if}
    
    <!-- 資料 -->
    <div class="pt-4 border-t border-gray-700">
      <a href="/dashboard/guide" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700" on:click={(e) => handleLinkClick(e, '/dashboard/guide')}>
        <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
        </svg>
        資料
      </a>
    </div>
  </nav>
</aside>
