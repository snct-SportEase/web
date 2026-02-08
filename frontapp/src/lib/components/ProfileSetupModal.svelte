<script>
  import { browser } from '$app/environment';

  /** @type {Array<{id: number, name: string}>} */
  export let classes = [];

  let isLoading = false;
  let errorMessage = '';

  let isConfirming = false;
  let confirmData = { displayName: '', className: '', classId: '' };

  function handleSubmit(event) {
    event.preventDefault();
    errorMessage = '';

    const formData = new FormData(event.target);
    const displayName = formData.get('displayName');
    const classId = formData.get('classId');
    
    if (!displayName || !classId) {
      errorMessage = '表示名とクラスを入力してください。';
      return;
    }

    const selectedClass = classes.find(c => c.id === parseInt(classId));
    if (!selectedClass) {
      errorMessage = '無効なクラスです。';
      return;
    }

    confirmData = {
      displayName: displayName.toString(),
      className: selectedClass.name,
      classId: classId.toString()
    };
    isConfirming = true;
  }

  function handleBack() {
    isConfirming = false;
    errorMessage = '';
  }

  async function handleConfirm() {
    isLoading = true;
    errorMessage = '';

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
          display_name: confirmData.displayName, 
          class_id: parseInt(confirmData.classId) 
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
      isConfirming = false; // エラー時は入力画面に戻るか、確認画面のままにするか。ここでは入力画面に戻す
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

    {#if !isConfirming}
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
              value={confirmData.displayName}
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
                value={confirmData.classId}
              >
                <option value="" disabled selected={!confirmData.classId}>クラスを選択してください</option>
                {#each classes as cls}
                  <option value={cls.id} selected={String(cls.id) === confirmData.classId}>{cls.name}</option>
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
            次へ
          </button>
        </div>
      </form>
    {:else}
      <!-- 確認画面 -->
      <div class="space-y-4">
        <div class="p-4 bg-gray-50 rounded-md">
          <dl class="space-y-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">表示名</dt>
              <dd class="mt-1 text-lg font-semibold text-gray-900">{confirmData.displayName}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">クラス</dt>
              <dd class="mt-1 text-lg font-semibold text-gray-900">{confirmData.className}</dd>
            </div>
          </dl>
        </div>
        
        <p class="text-sm text-center text-gray-600">この内容で登録してよろしいですか？</p>

        <!-- エラーメッセージ -->
        {#if errorMessage}
          <p class="mt-4 text-sm text-center text-red-600">{errorMessage}</p>
        {/if}

        <div class="mt-6 flex space-x-3">
          <button
            type="button"
            on:click={handleBack}
            disabled={isLoading}
            class="flex-1 px-4 py-2 font-bold text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            修正する
          </button>
          <button
            type="button"
            on:click={handleConfirm}
            disabled={isLoading}
            class="flex-1 px-4 py-2 font-bold text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? '保存中...' : '登録する'}
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>
