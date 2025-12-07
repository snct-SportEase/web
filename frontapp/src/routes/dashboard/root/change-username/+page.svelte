<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { fade, fly } from 'svelte/transition';

  let users = [];
  let searchQuery = '';
  let searchType = 'email'; // 'email' or 'display_name'
  let selectedUser = null;
  let newDisplayName = '';
  let showModal = false;
  let isLoading = false;
  let errorMessage = '';

  // 指数バックオフ付きのフェッチ関数
  async function fetchWithBackoff(url, options = {}, maxRetries = 5) {
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        const response = await fetch(url, options);
        if (response.ok) {
          return response;
        }
        // 4xx/5xx エラーの場合は再試行
        console.warn(`Fetch failed (Status: ${response.status}). Retrying...`);
      } catch (error) {
        console.error(`Fetch error on attempt ${attempt + 1}:`, error);
      }

      // バックオフ遅延
      const delay = Math.pow(2, attempt) * 1000;
      if (attempt < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }
    throw new Error("API call failed after multiple retries.");
  }

  async function fetchUsers(query = '', type = '') {
    if (!browser) return;
    isLoading = true;
    errorMessage = '';
    try {
      const url = `/api/root/users?query=${encodeURIComponent(query)}&searchType=${encodeURIComponent(type)}`;
      const response = await fetch(url);
      if (response.ok) {
        const data = await response.json();
        // レスポンスが配列であることを確認
        if (Array.isArray(data)) {
          users = data;
        } else {
          console.error('Unexpected response format:', data);
          users = [];
          errorMessage = '予期しないレスポンス形式です';
        }
      } else {
        const errorText = await response.text();
        console.error('Failed to fetch users:', response.status, errorText);
        users = [];
        errorMessage = `ユーザーの取得に失敗しました: ${response.status}`;
      }
    } catch (error) {
      console.error('Error fetching users:', error);
      users = [];
      errorMessage = 'ユーザーの取得中にエラーが発生しました';
    } finally {
      isLoading = false;
    }
  }

  async function searchUsers() {
    await fetchUsers(searchQuery, searchType);
  }

  function openEditModal(user) {
    selectedUser = user;
    newDisplayName = user.display_name || '';
    showModal = true;
  }

  function closeEditModal() {
    showModal = false;
    selectedUser = null;
    newDisplayName = '';
  }

  function handleOverlayClick(event) {
    if (event.target === event.currentTarget) {
      closeEditModal();
    }
  }

  async function handleDisplayNameUpdate() {
    if (!selectedUser || !browser) return;

    try {
      const response = await fetchWithBackoff('/api/root/users/display-name', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: selectedUser.id,
          display_name: newDisplayName,
        }),
      });

      if (response.ok) {
        // 1. ローカルリスト内のユーザーを更新
        const updatedUser = { ...selectedUser, display_name: newDisplayName };
        const index = users.findIndex(u => u.id === selectedUser.id);
        if (index !== -1) {
          users[index] = updatedUser;
          users = [...users]; // 2. リアクティビティをトリガー
        }
        // 3. モーダルを閉じる
        closeEditModal();
        location.reload();
      } else {
        console.error('Failed to update display name:', response.statusText);
      }
    } catch (error) {
      console.error('Error updating display name:', error);
    }
  }

  onMount(async () => {
    await fetchUsers();
  });
</script>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">ユーザー名変更管理</h1>
    <p class="text-sm text-gray-600">
      ユーザーの表示名を検索し、変更することができます。
    </p>
  </header>

  <!-- ユーザー検索カード -->
  <div class="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
    <h2 class="text-lg font-semibold text-gray-800 mb-4">ユーザー検索</h2>
    <div class="flex flex-col md:flex-row items-start md:items-center gap-4">
      <!-- 検索タイプ選択 -->
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
      
      <!-- 検索入力 -->
      <div class="flex-1 max-w-md">
        <div class="relative">
          <input 
            type="text" 
            class="w-full rounded-md border border-gray-300 px-4 py-2 pl-10 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500" 
            placeholder="検索キーワードを入力..." 
            bind:value={searchQuery} 
            on:keydown={(e) => e.key === 'Enter' && searchUsers()}
          />
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400">
            <path fill-rule="evenodd" d="M9.965 11.026a5 5 0 1 1 1.06-1.06l2.755 2.754a.75.75 0 1 1-1.06 1.06l-2.755-2.754ZM10.5 7a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z" clip-rule="evenodd" />
          </svg>
        </div>
      </div>
      
      <!-- 検索ボタン -->
      <button 
        class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow hover:bg-indigo-700 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:bg-indigo-300 disabled:cursor-not-allowed" 
        on:click={searchUsers}
        disabled={isLoading}
      >
        {isLoading ? '検索中...' : '検索'}
      </button>
      
      <!-- すべて表示ボタン -->
      <button 
        class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-semibold text-gray-700 shadow-sm hover:bg-gray-50 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed" 
        on:click={() => { searchQuery = ''; fetchUsers('', ''); }}
        disabled={isLoading}
      >
        すべて表示
      </button>
    </div>
  </div>

  <!-- エラーメッセージ -->
  {#if errorMessage}
    <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {errorMessage}
    </div>
  {/if}

  <!-- ユーザーリストカード -->
  <div class="rounded-lg border border-gray-200 bg-white shadow-sm overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-200 bg-gray-50">
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Email</th>
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">表示名</th>
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">クラスID</th>
            <th class="px-4 py-3 text-right text-sm font-semibold text-gray-700">アクション</th>
          </tr>
        </thead>
        <tbody>
          {#each users as user (user.id)}
            <tr class="border-b border-gray-100 hover:bg-gray-50 transition-colors">
              <td class="px-4 py-3 text-sm font-medium text-gray-900">{user.email}</td>
              <td class="px-4 py-3 text-sm text-gray-700">
                {#if user.display_name}
                  {user.display_name}
                {:else}
                  <span class="text-gray-400 italic">未設定</span>
                {/if}
              </td>
              <td class="px-4 py-3 text-sm font-mono text-indigo-600">{user.class_id || '-'}</td>
              <td class="px-4 py-3 text-right">
                <button 
                  class="inline-flex items-center gap-2 rounded-md border border-indigo-600 bg-white px-3 py-1.5 text-sm font-semibold text-indigo-600 hover:bg-indigo-50 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-1" 
                  on:click={() => openEditModal(user)}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M17.414 2.586a2 2 0 00-2.828 0L7 10.172V13h2.828l7.586-7.586a2 2 0 000-2.828z" />
                    <path fill-rule="evenodd" d="M2 6a2 2 0 012-2h4a1 1 0 010 2H4v10h10v-4a1 1 0 112 0v4a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" clip-rule="evenodd" />
                  </svg>
                  編集
                </button>
              </td>
            </tr>
          {/each}
          {#if !isLoading && users.length === 0}
            <tr>
              <td colspan="4" class="px-4 py-8 text-center text-sm text-gray-500">
                ユーザーが見つかりませんでした。
              </td>
            </tr>
          {/if}
          {#if isLoading}
            <tr>
              <td colspan="4" class="px-4 py-8 text-center text-sm text-gray-500">
                読み込み中...
              </td>
            </tr>
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>

<!-- 編集モーダル -->
{#if showModal && selectedUser}
  <div
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
    transition:fade={{ duration: 150 }}
    on:introstart={() => (document.body.style.overflow = 'hidden')}
    on:outroend={() => (document.body.style.overflow = 'auto')}
    on:click={handleOverlayClick}
    on:keydown={(e) => e.key === 'Escape' && closeEditModal()}
    role="button"
    tabindex="-1"
    aria-label="モーダルを閉じる"
  >
    <div
      class="w-full max-w-lg rounded-lg border border-gray-200 bg-white p-6 shadow-xl"
      transition:fly={{ y: -50, duration: 300 }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      tabindex="-1"
    >
      <div class="mb-6">
        <h3 id="modal-title" class="text-xl font-semibold text-gray-900 mb-4">表示名を編集</h3>
        <div class="space-y-2 rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm">
          <p class="text-gray-600">
            <span class="font-semibold text-gray-700">ユーザーID:</span>
            <span class="ml-2 font-mono text-gray-900 break-all">{selectedUser.id}</span>
          </p>
          <p class="text-gray-600">
            <span class="font-semibold text-gray-700">Email:</span>
            <span class="ml-2 font-mono text-gray-900 break-all">{selectedUser.email}</span>
          </p>
        </div>
      </div>

      <div class="mb-6">
        <label class="block text-sm font-semibold text-gray-700 mb-2" for="displayNameInput">
          新しい表示名
        </label>
        <input 
          id="displayNameInput" 
          type="text" 
          placeholder="新しい表示名を入力" 
          class="w-full rounded-md border border-gray-300 px-4 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500" 
          bind:value={newDisplayName}
          on:keydown={(e) => e.key === 'Enter' && newDisplayName.trim() !== '' && handleDisplayNameUpdate()}
        />
      </div>

      <div class="flex justify-end gap-3">
        <button 
          class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-semibold text-gray-700 shadow-sm hover:bg-gray-50 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2" 
          on:click={closeEditModal}
        >
          キャンセル
        </button>
        <button 
          class="inline-flex items-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-700 transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:bg-indigo-300 disabled:cursor-not-allowed" 
          on:click={handleDisplayNameUpdate} 
          disabled={newDisplayName.trim() === ''}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
          保存
        </button>
      </div>
    </div>
  </div>
{/if}
