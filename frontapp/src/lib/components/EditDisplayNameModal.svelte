<script>
  import { browser } from '$app/environment';

  export let isOpen = false;
  export let currentDisplayName = '';
  export let onClose = () => {};
  export let onSave = async (newDisplayName) => {};

  let newDisplayName = currentDisplayName;
  let isLoading = false;
  let errorMessage = '';

  // モーダルが開かれるたびに現在の表示名をリセット
  $: if (isOpen) {
    newDisplayName = currentDisplayName;
    errorMessage = '';
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
      console.error('Display name update error:', error);
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
</script>

<!-- モーダルの背景 -->
{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm" on:click={handleCancel}>
    <!-- モーダルの本体 -->
    <div class="w-full max-w-md p-6 space-y-4 bg-white rounded-lg shadow-xl" on:click|stopPropagation on:keydown={handleKeydown}>
      <h2 class="text-xl font-bold text-gray-800">表示名を変更</h2>
      
      <div>
        <label for="displayName" class="block text-sm font-medium text-gray-700 mb-2">
          新しい表示名
        </label>
        <input
          type="text"
          id="displayName"
          bind:value={newDisplayName}
          disabled={isLoading}
          class="w-full px-3 py-2 text-gray-900 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          placeholder="表示名を入力してください"
          autofocus
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
