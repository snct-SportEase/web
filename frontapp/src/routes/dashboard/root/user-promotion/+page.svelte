<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let users = $state([]);
  let searchQuery = $state('');
  let searchType = $state('email');
  let isLoading = $state(false);
  let errorMessage = $state('');

  const privilegeRoles = ['admin', 'root'];

  async function fetchUsers(query = '', type = '') {
    if (!browser) return;
    isLoading = true;
    errorMessage = '';
    try {
      const url = `/api/root/users?query=${encodeURIComponent(query)}&searchType=${encodeURIComponent(type)}`;
      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();
        users = Array.isArray(data) ? data : [];
      } else {
        errorMessage = `ユーザーの取得に失敗しました: ${res.status}`;
        users = [];
      }
    } catch {
      errorMessage = 'ユーザーの取得中にエラーが発生しました';
      users = [];
    } finally {
      isLoading = false;
    }
  }

  async function searchUsers() {
    await fetchUsers(searchQuery, searchType);
  }

  function hasRole(user, roleName) {
    return user.roles?.some(r => r.name === roleName) ?? false;
  }

  async function promote(user, role) {
    if (!confirm(`${user.email} を ${role} に昇格しますか？`)) return;
    try {
      const res = await fetch('/api/root/users/promote', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: user.id, role }),
      });
      if (res.ok) {
        await fetchUsers(searchQuery, searchType);
      } else {
        const err = await res.json();
        alert(`昇格失敗: ${err.error}`);
      }
    } catch {
      alert('エラーが発生しました');
    }
  }

  async function demote(user, role) {
    if (!confirm(`${user.email} から ${role} ロールを剥奪しますか？`)) return;
    try {
      const res = await fetch('/api/root/users/promote', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: user.id, role }),
      });
      if (res.ok) {
        await fetchUsers(searchQuery, searchType);
      } else {
        const err = await res.json();
        alert(`降格失敗: ${err.error}`);
      }
    } catch {
      alert('エラーが発生しました');
    }
  }

  onMount(() => fetchUsers());
</script>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">権限管理</h1>
    <p class="text-sm text-gray-600">
      ユーザーに admin・root 権限を付与または剥奪できます。
    </p>
  </header>

  <!-- 検索カード -->
  <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
    <h2 class="text-lg font-semibold text-gray-800 mb-4">ユーザー検索</h2>
    <div class="flex flex-col md:flex-row items-start md:items-center gap-4">
      <div class="flex items-center gap-2">
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="radio"
            name="searchType"
            value="email"
            bind:group={searchType}
            class="w-4 h-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
          />
          <span class="text-sm font-medium text-gray-700">Email</span>
        </label>
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="radio"
            name="searchType"
            value="display_name"
            bind:group={searchType}
            class="w-4 h-4 text-indigo-600 focus:ring-indigo-500 border-gray-300"
          />
          <span class="text-sm font-medium text-gray-700">表示名</span>
        </label>
      </div>

      <div class="flex-1 max-w-md">
        <div class="relative">
          <input
            type="text"
            class="w-full rounded-md border border-gray-300 px-4 py-2 pl-10 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            placeholder="検索キーワードを入力..."
            bind:value={searchQuery}
            onkeydown={(e) => e.key === 'Enter' && searchUsers()}
          />
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400">
            <path fill-rule="evenodd" d="M9.965 11.026a5 5 0 1 1 1.06-1.06l2.755 2.754a.75.75 0 1 1-1.06 1.06l-2.755-2.754ZM10.5 7a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z" clip-rule="evenodd" />
          </svg>
        </div>
      </div>

      <button
        class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow hover:bg-indigo-700 transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:bg-indigo-300 disabled:cursor-not-allowed"
        onclick={searchUsers}
        disabled={isLoading}
      >
        {isLoading ? '検索中...' : '検索'}
      </button>

      <button
        class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-semibold text-gray-700 shadow-sm hover:bg-gray-50 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
        onclick={() => { searchQuery = ''; fetchUsers('', ''); }}
        disabled={isLoading}
      >
        すべて表示
      </button>
    </div>
  </div>

  {#if errorMessage}
    <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {errorMessage}
    </div>
  {/if}

  <!-- ユーザーテーブル -->
  <div class="rounded-lg border border-gray-200 bg-white shadow-sm overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50">
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Email</th>
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">表示名</th>
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">現在のロール</th>
            <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">admin</th>
            <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">root</th>
          </tr>
        </thead>
        <tbody>
          {#each users as user (user.id)}
            <tr class="border-b border-gray-100 hover:bg-gray-50 transition-colors">
              <td class="px-4 py-3 text-sm font-medium text-gray-900">{user.email}</td>
              <td class="px-4 py-3 text-sm text-gray-600">
                {#if user.display_name}{user.display_name}{:else}<span class="text-gray-400 italic">未設定</span>{/if}
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-1">
                  {#each (user.roles ?? []) as role (role.id)}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium
                      {role.name === 'root' ? 'bg-red-100 text-red-800' :
                       role.name === 'admin' ? 'bg-amber-100 text-amber-800' :
                       'bg-blue-100 text-blue-800'}">
                      {role.name}
                    </span>
                  {/each}
                  {#if !user.roles || user.roles.length === 0}
                    <span class="text-xs text-gray-400">なし</span>
                  {/if}
                </div>
              </td>
              {#each privilegeRoles as role (role)}
                <td class="px-4 py-3 text-center">
                  {#if hasRole(user, role)}
                    <button
                      class="rounded px-3 py-1 text-xs font-semibold bg-red-100 text-red-700 border border-red-200 hover:bg-red-200 transition-colors"
                      onclick={() => demote(user, role)}
                    >
                      剥奪
                    </button>
                  {:else}
                    <button
                      class="rounded px-3 py-1 text-xs font-semibold bg-amber-500 text-white hover:bg-amber-600 transition-colors"
                      onclick={() => promote(user, role)}
                    >
                      付与
                    </button>
                  {/if}
                </td>
              {/each}
            </tr>
          {/each}

          {#if isLoading}
            <tr>
              <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">読み込み中...</td>
            </tr>
          {:else if users.length === 0}
            <tr>
              <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">ユーザーが見つかりませんでした。</td>
            </tr>
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>
