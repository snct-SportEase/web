<script>
  import { browser } from '$app/environment';
  import { tick } from 'svelte';

  export let isOpen = false;
  export let currentDisplayName = '';
  export let userRoles = [];
  export let onClose = () => {};
export let onSave = async () => {};

  let newDisplayName = currentDisplayName;
  let isLoading = false;
  let errorMessage = '';
  let displayNameInput;

  // モーダルが開かれるたびに現在の表示名をリセット
  $: if (isOpen) {
    newDisplayName = currentDisplayName;
    errorMessage = '';
  }

  $: if (isOpen && browser) {
    tick().then(() => {
      displayNameInput?.focus();
    });
  }

  async function handleSave() {
    if (!newDisplayName.trim()) {
      errorMessage = '表示名を入力してください。';
      return;
    }

    if (newDisplayName.trim() === currentDisplayName) {
      onClose();
      return;
    }

    isLoading = true;
    errorMessage = '';

    try {
      await onSave(newDisplayName.trim());
      onClose();
    } catch (error) {
      errorMessage = error.message || '表示名の更新に失敗しました。';
    } finally {
      isLoading = false;
    }
  }

  function handleCancel() {
    newDisplayName = currentDisplayName;
    errorMessage = '';
    onClose();
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      handleCancel();
    }
  }

  function handleOverlayKeydown(event) {
    if (event.key === 'Escape') {
      event.preventDefault();
      handleCancel();
    }
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleCancel();
    }
  }
</script>

<!-- モーダルの背景 -->
{#if isOpen}
  <div
    class="fixed top-0 left-0 right-0 bottom-0 z-[9999] flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm min-h-screen overflow-hidden"
    role="presentation"
    tabindex="-1"
    on:click={handleCancel}
    on:keydown={handleOverlayKeydown}
  >
    <!-- モーダルの本体 -->
    <div
      class="w-full max-w-md p-6 space-y-4 bg-white rounded-lg shadow-xl"
      role="dialog"
      aria-modal="true"
      aria-labelledby="edit-display-name-title"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown={handleKeydown}
    >
      <div class="flex justify-between items-center">
        <h2 id="edit-display-name-title" class="text-xl font-bold text-gray-800">プロフィール管理</h2>
        <button
          type="button"
          on:click={handleCancel}
          aria-label="プロフィール管理を閉じる"
          class="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>
      
      <!-- ロール情報表示 -->
      {#if userRoles && userRoles.length > 0}
        <div class="bg-gray-50 p-3 rounded-md">
          <h3 class="text-sm font-medium text-gray-700 mb-2">現在のロール</h3>
          <div class="flex flex-wrap gap-2">
            {#each userRoles as role (role.name)}
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
                {role.name}
              </span>
            {/each}
          </div>
        </div>
      {/if}
      
      <div>
        <label for="displayName" class="block text-sm font-medium text-gray-700 mb-2">
          新しい表示名
        </label>
        <input
          type="text"
          id="displayName"
          bind:value={newDisplayName}
          disabled={isLoading}
          bind:this={displayNameInput}
          class="w-full px-3 py-2 text-gray-900 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          placeholder="表示名を入力してください"
        />
      </div>

      <!-- エラーメッセージ -->
      {#if errorMessage}
        <p class="text-sm text-center text-red-600">{errorMessage}</p>
      {/if}

      <!-- ボタン -->
      <div class="flex space-x-3 pt-2">
        <button
          type="button"
          on:click={handleCancel}
          disabled={isLoading}
          class="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          キャンセル
        </button>
        <button
          type="button"
          on:click={handleSave}
          disabled={isLoading || !newDisplayName.trim()}
          class="flex-1 px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? '保存中...' : '保存'}
        </button>
      </div>
    </div>
  </div>
{/if}
