<script>
  import { isSidebarOpen } from '$lib/stores/sidebarStore.js';

  /** @type {import('@sveltejs/kit').MaybePromise<import('../../routes/$types').LayoutData>} */
  export let user;

  const hasRole = (roleName) => {
    return user?.roles?.some(role => role.name === roleName);
  };

  const isStudent = hasRole('student');
  const isAdmin = hasRole('admin');
  const isRoot = hasRole('root');

  function closeSidebar(event) {
    event.preventDefault();
    event.stopPropagation();
    isSidebarOpen.set(false);
  }
</script>

<aside class="w-64 bg-gray-800 text-white flex flex-col transition-all duration-300" class:closed={!$isSidebarOpen}>
  <div class="h-16 flex items-center justify-between px-4">
    <a href="/dashboard" class="flex items-center"><span class="text-2xl font-bold">SportEase</span></a>
    <button 
      type="button" 
      on:click={closeSidebar} 
      class="flex items-center justify-center min-w-[48px] min-h-[48px] p-3 rounded-md hover:bg-gray-700 active:bg-gray-600 touch-manipulation transition-colors" 
      aria-label="サイドバーを閉じる"
      style="touch-action: manipulation;"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
    </button>
  </div>
  <nav class="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
    <!-- Root Menu -->
    {#if isRoot}
      <div class="pt-4">
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Root</h3>
        <a href="/dashboard/root/event-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          大会情報登録・管理
        </a>
        <a href="/dashboard/root/rainy-mode" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          雨天時モード管理
        </a>
        <a href="/dashboard/root/sport-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          競技情報登録・管理
        </a>
        <a href="/dashboard/root/notification" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          通知管理
        </a>
        <a href="/dashboard/root/notification-requests" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          通知申請管理
        </a>
        <a href="/dashboard/root/tournament-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          トーナメント生成・管理
        </a>
        <a href="/dashboard/root/noon-game" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          昼競技管理
        </a>
        <a href="/dashboard/root/whitelist-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          ホワイトリスト
        </a>
        <a href="/dashboard/root/class-student-count" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          各クラス人数設定
        </a>
        <a href="/dashboard/root/change-username" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          ユーザー名変更
        </a>
        <a href="/dashboard/root/identify-mvp" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          MVP確認
        </a>
      </div>
    {/if}
    
    <!-- Admin/Root Menu -->
    {#if isAdmin || isRoot}
      <div class="pt-4">
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Admin</h3>
        <a href="/dashboard/admin/class-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          クラス・チーム割り当て
        </a>
        <a href="/dashboard/admin/role-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          ロール割り当て・管理
        </a>
        <a href="/dashboard/admin/qr-code-reader" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          QRコード読み取り
        </a>
        <a href="/dashboard/admin/insert-matche-result" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          試合結果入力
        </a>
        <a href="/dashboard/admin/attendance-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          出席登録
        </a>
        <!-- MVP投票の被対象者は1~2年生のみ。投票はadminとrootしか行えない -->
        <a href="/dashboard/admin/vorting-mvp" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          MVP投票
        </a>
        <a href="/dashboard/admin/sport-details-registration" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          競技詳細情報登録
        </a>
        <a href="/dashboard/admin/noon-game-results" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          昼競技結果入力
        </a>
      </div>
    {/if}
    
    <!-- Student Menu -->
    {#if isStudent || isAdmin || isRoot}
      <div>
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Student</h3>
        <a href="/dashboard/student/my-page" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          マイページ
        </a>
        <a href="/dashboard/student/class-info" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          クラス情報
        </a>
        <a href="/dashboard/student/sport-info" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          競技一覧・詳細閲覧
        </a>
        <a href="/dashboard/student/issueqr-code" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          QRコード発行
        </a>
        <a href="/dashboard/student/score-list" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          点数一覧
        </a>
        <a href="/dashboard/student/notification" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          通知
        </a>
        <a href="/dashboard/student/notification-request" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          通知申請
        </a>
      </div>
    {/if}
    
    <!-- 資料 -->
    <div class="pt-4 border-t border-gray-700">
      <a href="/dashboard/guide" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
        <svg class="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
        </svg>
        資料
      </a>
    </div>
  </nav>
</aside>
