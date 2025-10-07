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

  // APIキーはここでは空文字列として扱い、実行環境から提供されることを想定
  const apiKey = "";

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
    try {
      const url = `/api/root/users?query=${encodeURIComponent(query)}&amp;searchType=${encodeURIComponent(type)}`;
      // Google Search Groundingは不要なため、そのままのfetchを使用
      const response = await fetch(url);
      if (response.ok) {
        users = await response.json();
      } else {
        console.error('Failed to fetch users:', response.statusText);
      }
    } catch (error) {
      console.error('Error fetching users:', error);
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

<div class="container mx-auto p-4 lg:p-8">
  <h1 class="text-3xl font-extrabold mb-8 text-center text-primary">ユーザー名変更管理</h1>

  <!-- ユーザー検索カード -->
  <div class="card bg-base-100 shadow-2xl mb-8 border border-base-200">
    <div class="card-body">
      <h2 class="card-title text-xl font-semibold text-secondary">ユーザー検索</h2>
      <div class="flex flex-col md:flex-row items-center space-y-4 md:space-y-0 md:space-x-4">
        <!-- 検索タイプ選択 -->
        <div class="join border border-primary/50 rounded-full overflow-hidden">
          <input class="join-item btn btn-sm transition-all duration-300" type="radio" name="searchType" value="email" bind:group={searchType} aria-label="Email" checked />
          <input class="join-item btn btn-sm transition-all duration-300" type="radio" name="searchType" value="display_name" bind:group={searchType} aria-label="表示名" />
        </div>
        <!-- 検索入力 -->
        <div class="form-control w-full md:max-w-xs">
          <label class="input input-bordered flex items-center gap-2 rounded-full shadow-inner bg-base-200">
            <input type="text" class="grow bg-transparent" placeholder="検索キーワードを入力..." bind:value={searchQuery} on:keydown.enter={searchUsers} />
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-4 h-4 opacity-70"><path fill-rule="evenodd" d="M9.965 11.026a5 5 0 1 1 1.06-1.06l2.755 2.754a.75.75 0 1 1-1.06 1.06l-2.755-2.754ZM10.5 7a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z" clip-rule="evenodd" /></svg>
          </label>
        </div>
        <!-- 検索ボタン -->
        <button class="btn btn-primary btn-sm md:btn-md shadow-lg hover:shadow-xl transition-all duration-300" on:click={searchUsers}>
          検索
        </button>
        <div class="w-full md:w-auto"></div>
      </div>
    </div>
  </div>

  <!-- ユーザーリストカード -->
  <div class="card bg-base-100 shadow-2xl">
    <div class="card-body p-0">
      <div class="overflow-x-auto rounded-xl">
        <table class="table table-lg w-full">
          <thead class="bg-base-200 text-base-content/80 sticky top-0">
            <tr>
              <th class="uppercase">Email</th>
              <th class="uppercase">表示名</th>
              <th class="uppercase">クラスID</th>
              <th class="uppercase">アクション</th>
            </tr>
          </thead>
          <tbody>
            {#each users as user (user.id)}
            <tr class="hover:bg-base-200/50 transition-colors duration-200">
              <td class="font-medium text-sm lg:text-base">{user.email}</td>
              <td>{user.display_name || '未設定'}</td>
              <td class="font-mono text-xs lg:text-sm text-info">{user.class_id || '-'}</td>
              <td>
                <button class="btn btn-sm btn-outline btn-primary rounded-full hover:shadow-md" on:click={() => openEditModal(user)}>
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M17.414 2.586a2 2 0 00-2.828 0L7 10.172V13h2.828l7.586-7.586a2 2 0 000-2.828z" />
                    <path fill-rule="evenodd" d="M2 6a2 2 0 012-2h4a1 1 0 010 2H4v10h10v-4a1 1 0 112 0v4a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" clip-rule="evenodd" />
                  </svg>
                  編集
                </button>
              </td>
            </tr>
            {/each}
            {#if users.length === 0}
              <tr>
                <td colspan="4" class="text-center py-8 text-lg text-gray-500">
                  ユーザーが見つかりませんでした。
                </td>
              </tr>
            {/if}
          </tbody>
        </table>
        <div class="w-full h-4"></div>
      </div>
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
  >
    <div
      class="modal-box rounded-xl shadow-2xl w-full max-w-lg mx-auto transform transition-all duration-300 ease-out"
      transition:fly={{ y: -50, duration: 300 }}
    >
      <h3 class="font-bold text-2xl mb-4 text-primary">表示名を編集</h3>
      <p class="py-2 text-sm text-base-content/80">ユーザーID: <span class="font-mono text-info break-all">{selectedUser.id}</span></p>
      <p class="py-2 text-sm text-base-content/80">Email: <span class="font-mono text-info break-all">{selectedUser.email}</span></p>

      <div class="form-control w-full mt-4">
        <label class="label" for="displayNameInput">
          <span class="label-text font-semibold">新しい表示名</span>
        </label>
        <input id="displayNameInput" type="text" placeholder="新しい表示名を入力" class="input input-bordered w-full rounded-lg" bind:value={newDisplayName} />
      </div>

      <div class="modal-action mt-6 flex justify-end">
        <button class="btn btn-primary shadow-lg hover:shadow-xl transition-all duration-300 rounded-full" on:click={handleDisplayNameUpdate} disabled={newDisplayName.trim() === ''}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
          保存
        </button>
        <button class="btn btn-ghost hover:bg-base-300 transition-all duration-300 rounded-full" on:click={closeEditModal}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
          キャンセル
        </button>
      </div>
    </div>
  </div>
{/if}
