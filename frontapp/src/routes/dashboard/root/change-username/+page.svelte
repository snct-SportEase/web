<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { fade, fly } from 'svelte/transition';

  let users = [];
  let classes = []; // クラス一覧
  let searchQuery = '';
  let searchType = 'email'; // 'email' or 'display_name'
  let selectedUser = null;
  let newDisplayName = '';
  // ロール管理用
  let newRoleName = '';
  let selectedClassRep = ''; // クラス代表変更用

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

  async function fetchClasses() {
    try {
      const response = await fetch('/api/classes');
      if (response.ok) {
        classes = await response.json();
      } else {
        console.error('Failed to fetch classes');
      }
    } catch (error) {
      console.error('Error fetching classes:', error);
    }
  }

  async function searchUsers() {
    await fetchUsers(searchQuery, searchType);
  }

  function openEditModal(user) {
    selectedUser = user;
    newDisplayName = user.display_name || '';
    newRoleName = '';
    
    // 現在のクラス代表ロールを探す
    const repRole = user.roles?.find(r => r.name.endsWith('_rep'));
    if (repRole) {
      // "3A_rep" -> "3A" を抽出して初期値にする
      // クラス名が可変長の場合も考慮して _rep を除去
      const className = repRole.name.slice(0, -4);
      // クラスIDではなくクラス名をバインドする必要があるため、クラス一覧から探す
      // API仕様上、classesは{id, name, ...}を返すはず
      const cls = classes.find(c => c.name === className);
      selectedClassRep = cls ? cls.id : '';
    } else {
      selectedClassRep = '';
    }

    showModal = true;
  }

  function closeEditModal() {
    showModal = false;
    selectedUser = null;
    newDisplayName = '';
    newRoleName = '';
    selectedClassRep = '';
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
        updateLocalUser({ ...selectedUser, display_name: newDisplayName });
        alert('表示名を更新しました');
      } else {
        console.error('Failed to update display name:', response.statusText);
        alert('表示名の更新に失敗しました');
      }
    } catch (error) {
      console.error('Error updating display name:', error);
      alert('エラーが発生しました');
    }
  }

  // クラス代表ロールの付け替え
  async function handleClassRepChange() {
    if (!selectedUser || !selectedClassRep) return;

    const targetClass = classes.find(c => c.id == selectedClassRep);
    if (!targetClass) return;

    const newRole = `${targetClass.name}_rep`;
    
    // 現在の代表ロールを取得（削除用）
    const currentRepRole = selectedUser.roles?.find(r => r.name.endsWith('_rep'));

    // 1. 新しいロールを追加
    try {
      const addRes = await fetchWithBackoff('/api/admin/users/role', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: selectedUser.id,
          role: newRole
        }),
      });

      if (!addRes.ok) {
        const err = await addRes.json();
        alert(`ロール追加失敗: ${err.error}`);
        return;
      }

      // 2. 成功したら古いロールを削除（もしあれば、かつ新しいロールと違う場合）
      if (currentRepRole && currentRepRole.name !== newRole) {
        await fetchWithBackoff('/api/admin/users/role', {
          method: 'DELETE',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            user_id: selectedUser.id,
            role: currentRepRole.name
          }),
        });
      }

      alert('クラス代表を変更しました');
      // ユーザー情報を再取得して更新
      // 簡易的にローカル更新したいが、ロールIDなどが不明なので再取得が無難
      await fetchUsers(searchQuery, searchType);
      closeEditModal();

    } catch (error) {
      console.error('Error changing class rep:', error);
      alert('エラーが発生しました');
    }
  }

  // その他のロール追加
  async function handleRoleAdd() {
    if (!newRoleName.trim()) return;
    
    // _rep制限
    if (newRoleName.endsWith('_rep')) {
      alert('クラス代表ロール（_rep）はここからは追加できません。「クラス代表の変更」を使用してください。');
      return;
    }

    try {
      const response = await fetchWithBackoff('/api/admin/users/role', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: selectedUser.id,
          role: newRoleName
        }),
      });

      if (response.ok) {
        alert('ロールを追加しました');
        newRoleName = '';
        await fetchUsers(searchQuery, searchType);
        // モーダルは閉じずに更新された情報を再反映させたいが、
        // fetchUsersでusersが更新されるとselectedUserの参照が切れる可能性があるため
        // ここでは簡易的に閉じるか、selectedUserを再検索して更新する
        const updatedUser = users.find(u => u.id === selectedUser.id);
        if (updatedUser) selectedUser = updatedUser; 
      } else {
        const err = await response.json();
        alert(`追加失敗: ${err.error}`);
      }
    } catch (error) {
      console.error('Error adding role:', error);
      alert('エラーが発生しました');
    }
  }

  // その他のロール削除
  async function handleRoleDelete(roleName) {
    if (!confirm(`ロール "${roleName}" を削除しますか？`)) return;

    try {
      const response = await fetchWithBackoff('/api/admin/users/role', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_id: selectedUser.id,
          role: roleName
        }),
      });

      if (response.ok) {
        alert('ロールを削除しました');
        await fetchUsers(searchQuery, searchType);
        const updatedUser = users.find(u => u.id === selectedUser.id);
        if (updatedUser) selectedUser = updatedUser;
      } else {
        const err = await response.json();
        alert(`削除失敗: ${err.error}`);
      }
    } catch (error) {
      console.error('Error deleting role:', error);
      alert('エラーが発生しました');
    }
  }

  function updateLocalUser(updatedUser) {
    const index = users.findIndex(u => u.id === updatedUser.id);
    if (index !== -1) {
      users[index] = updatedUser;
      users = [...users];
    }
    // selectedUserも更新
    selectedUser = updatedUser;
  }

  onMount(async () => {
    await Promise.all([fetchUsers(), fetchClasses()]);
  });
</script>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">ユーザー管理</h1>
    <p class="text-sm text-gray-600">
      ユーザーの表示名の変更や、ロール（権限）の管理を行うことができます。
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
            <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">ロール</th>
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
              <td class="px-4 py-3 text-sm text-gray-700">
                {#if user.roles && user.roles.length > 0}
                  <div class="flex flex-wrap gap-1">
                    {#each user.roles as role}
                      <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                        {role.name}
                      </span>
                    {/each}
                  </div>
                {:else}
                  <span class="text-gray-400 text-xs">なし</span>
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
                  管理
                </button>
              </td>
            </tr>
          {/each}
          {#if !isLoading && users.length === 0}
            <tr>
              <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">
                ユーザーが見つかりませんでした。
              </td>
            </tr>
          {/if}
          {#if isLoading}
            <tr>
              <td colspan="5" class="px-4 py-8 text-center text-sm text-gray-500">
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
      class="w-full max-w-2xl rounded-lg border border-gray-200 bg-white shadow-xl flex flex-col max-h-[90vh]"
      transition:fly={{ y: -50, duration: 300 }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      tabindex="-1"
    >
      <!-- モーダルヘッダー -->
      <div class="px-6 py-4 border-b border-gray-200">
        <h3 id="modal-title" class="text-xl font-semibold text-gray-900">ユーザー管理</h3>
        <p class="text-sm text-gray-500 mt-1">
          {selectedUser.email} の情報を編集しています
        </p>
      </div>

      <!-- モーダルボディ（スクロール可能） -->
      <div class="p-6 overflow-y-auto space-y-8">
        
        <!-- セクション1: 表示名設定 -->
        <section>
          <h4 class="text-md font-bold text-gray-800 mb-3 pb-1 border-b">表示名設定</h4>
          <div class="flex gap-4 items-end">
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700 mb-1" for="displayNameInput">
                表示名
              </label>
              <input 
                id="displayNameInput" 
                type="text" 
                class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500" 
                bind:value={newDisplayName}
              />
            </div>
            <button 
              class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-700 disabled:bg-indigo-300" 
              on:click={handleDisplayNameUpdate} 
              disabled={newDisplayName.trim() === ''}
            >
              更新
            </button>
          </div>
        </section>

        <!-- セクション2: クラス代表ロール変更 -->
        <section class="bg-blue-50 p-4 rounded-md border border-blue-100">
          <h4 class="text-md font-bold text-blue-900 mb-2">クラス代表の変更</h4>
          <p class="text-xs text-blue-700 mb-3">
            誤ったクラスの代表権限を持ってしまった場合、ここで正しいクラスに付け替えることができます。
          </p>
          <div class="flex gap-4 items-end">
            <div class="flex-1">
              <label class="block text-sm font-medium text-blue-800 mb-1" for="classRepSelect">
                担当クラスを選択
              </label>
              <select 
                id="classRepSelect"
                class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                bind:value={selectedClassRep}
              >
                <option value="">（選択してください）</option>
                {#each classes as cls}
                  <option value={cls.id}>{cls.name}</option>
                {/each}
              </select>
            </div>
            <button 
              class="rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 disabled:bg-blue-300" 
              on:click={handleClassRepChange}
              disabled={!selectedClassRep}
            >
              変更・保存
            </button>
          </div>
        </section>

        <!-- セクション3: その他のロール管理 -->
        <section>
          <h4 class="text-md font-bold text-gray-800 mb-3 pb-1 border-b">その他のロール管理</h4>
          
          <!-- 現在のロールリスト -->
          <div class="mb-4">
            <p class="text-sm font-medium text-gray-700 mb-2">現在のロール:</p>
            <div class="flex flex-wrap gap-2">
              {#if selectedUser.roles && selectedUser.roles.length > 0}
                {#each selectedUser.roles as role}
                  <div class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800 border border-gray-200">
                    <span>{role.name}</span>
                    <!-- _repロールは削除不可（上のセクションで変更） -->
                    {#if !role.name.endsWith('_rep')}
                      <button 
                        class="text-gray-400 hover:text-red-500 focus:outline-none ml-1"
                        on:click={() => handleRoleDelete(role.name)}
                        title="ロールを削除"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                        </svg>
                      </button>
                    {/if}
                  </div>
                {/each}
              {:else}
                <span class="text-sm text-gray-500">ロールなし</span>
              {/if}
            </div>
          </div>

          <!-- ロール追加フォーム -->
          <div class="flex gap-4 items-end bg-gray-50 p-3 rounded-md">
            <div class="flex-1">
              <label class="block text-sm font-medium text-gray-700 mb-1" for="newRoleInput">
                新規ロール追加
              </label>
              <input 
                id="newRoleInput" 
                type="text" 
                placeholder="admin, root など"
                class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500" 
                bind:value={newRoleName}
              />
              <p class="text-xs text-gray-500 mt-1">※ _rep で終わるロールはここでは追加できません。</p>
            </div>
            <button 
              class="rounded-md bg-gray-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-gray-700 disabled:bg-gray-300" 
              on:click={handleRoleAdd}
              disabled={!newRoleName.trim()}
            >
              追加
            </button>
          </div>
        </section>

      </div>

      <!-- モーダルフッター -->
      <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 flex justify-end">
        <button 
          class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-semibold text-gray-700 shadow-sm hover:bg-gray-100 transition-colors" 
          on:click={closeEditModal}
        >
          閉じる
        </button>
      </div>
    </div>
  </div>
{/if}
