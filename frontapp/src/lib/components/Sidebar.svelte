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

  function closeSidebar() {
    isSidebarOpen.set(false);
  }
</script>

<aside class="w-64 bg-gray-800 text-white flex flex-col transition-all duration-300" class:closed={!$isSidebarOpen}>
  <div class="h-16 flex items-center justify-between px-4">
    <span class="text-2xl font-bold">SportEase</span>
    <button on:click={closeSidebar} class="p-2 rounded-md hover:bg-gray-700">
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
    </button>
  </div>
  <nav class="flex-1 px-2 py-4 space-y-1">
    <!-- Admin/Root Menu -->
    {#if isAdmin || isRoot}
      <div class="pt-4">
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Admin</h3>
        <a href="/dashboard/user-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          ユーザー管理
        </a>
        <a href="/dashboard/class-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          クラス管理
        </a>
        <a href="/dashboard/event-management" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          大会管理
        </a>
      </div>
    {/if}
    
    <!-- Student Menu -->
    {#if isStudent || isAdmin || isRoot}
      <div>
        <h3 class="px-4 mb-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">Student</h3>
        <a href="/dashboard/my-page" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          マイページ
        </a>
        <a href="/dashboard/class-info" class="flex items-center px-4 py-2 text-sm font-medium rounded-md hover:bg-gray-700">
          クラス情報
        </a>
      </div>
    {/if}
  </nav>
</aside>
