<script>
  import { browser } from '$app/environment';

  /** @type {Array<{id: number, name: string}>} */
  export let classes = [];

  let isLoading = false;
  let errorMessage = '';

  async function handleSubmit(event) {
    event.preventDefault();
    isLoading = true;
    errorMessage = '';

    const formData = new FormData(event.target);
    const displayName = formData.get('displayName');
    const classId = formData.get('classId');
    
    if (!displayName || !classId) {
      throw new Error('表示名とクラスを入力してください。');
    }

    try {
      // セッションクッキーを取得
      let sessionToken = null;
      if (browser) {
        const cookies = document.cookie.split('; ');
        const sessionCookie = cookies.find(row => row.startsWith('session_token='));
        sessionToken = sessionCookie ? sessionCookie.split('=')[1] : null;
      }

      if (!sessionToken) {
        throw new Error('セッションが見つかりません。再度ログインしてください。');
      }

      const response = await fetch('/api/user/profile', {
        method: 'PUT',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          'Cookie': `session_token=${sessionToken}`,
        },
        body: JSON.stringify({ 
          display_name: displayName, 
          class_id: parseInt(classId) 
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'プロフィールの更新に失敗しました。');
      }

      // 成功時はページをリロードしてユーザー情報を更新
      window.location.reload();
      
    } catch (error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }
</script>

<!-- モーダルの背景 -->
<div class="fixed top-0 left-0 right-0 bottom-0 z-50 flex items-center justify-center bg-black min-h-screen w-full overflow-hidden">
  <!-- モーダルの本体 -->
  <div class="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-xl">
    <h2 class="text-2xl font-bold text-center text-gray-800">プロフィールを設定してください</h2>
    <p class="text-center text-gray-600">初回ログインありがとうございます。サービスを利用する前に、表示名とクラスを設定してください。</p>

    <!-- プロフィール更新フォーム -->
    <form on:submit={handleSubmit}>
      <div class="space-y-4">
        <!-- 表示名 -->
        <div>
          <label for="displayName" class="block text-sm font-medium text-gray-700">表示名</label>
          <input
            type="text"
            id="displayName"
            name="displayName"
            required
            class="w-full px-3 py-2 mt-1 text-gray-900 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="例: 山田 太郎"
          />
        </div>

        <!-- クラス選択 -->
        <div>
          <label for="classId" class="block text-sm font-medium text-gray-700">クラス</label>
          {#if classes.length === 0}
            <div class="w-full px-3 py-2 mt-1 text-sm text-gray-500 bg-gray-100 border border-gray-300 rounded-md">
              クラスがまだ設定されていません。管理者に連絡してください。
            </div>
          {:else}
            <select
              id="classId"
              name="classId"
              required
              class="w-full px-3 py-2 mt-1 text-gray-900 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
            >
              <option value="" disabled selected>クラスを選択してください</option>
              {#each classes as cls}
                <option value={cls.id}>{cls.name}</option>
              {/each}
            </select>
          {/if}
        </div>
      </div>

      <!-- エラーメッセージ -->
      {#if errorMessage}
        <p class="mt-4 text-sm text-center text-red-600">{errorMessage}</p>
      {/if}

      <!-- 送信ボタン -->
      <div class="mt-6">
        <button
          type="submit"
          disabled={isLoading || classes.length === 0}
          class="w-full px-4 py-2 font-bold text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? '保存中...' : '保存して開始する'}
        </button>
      </div>
    </form>
  </div>
</div>
