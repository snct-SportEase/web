<script>
  import { enhance } from '$app/forms';
  import { goto } from '$app/navigation';

  /** @type {Array<{id: number, name: string}>} */
  export let classes = [];

  /** @type {import('./$types').ActionData} */
  export let form;

  let isLoading = false;
  let errorMessage = '';
</script>

<!-- モーダルの背景 -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 backdrop-blur-sm">
  <!-- モーダルの本体 -->
  <div class="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-xl">
    <h2 class="text-2xl font-bold text-center text-gray-800">プロフィールを設定してください</h2>
    <p class="text-center text-gray-600">初回ログインありがとうございます。サービスを利用する前に、表示名とクラスを設定してください。</p>

    <!-- プロフィール更新フォーム -->
    <form 
      method="POST" 
      action="/dashboard?/updateProfile"
      use:enhance={() => {
        isLoading = true;
        errorMessage = '';

        return async ({ result }) => {
          // 成功時はサーバーがリダイレクトを処理するため、クライアントでの遷移処理は不要。
          // use:enhanceがリダイレクトを自動的に追従します。

          // `result.type`が`failure`の場合のみ、クライアントでエラーメッセージを設定
          if (result.type === 'failure') {
            errorMessage = result.data?.message || 'プロフィールの更新に失敗しました。';
          }
          
          // フォーム送信が完了したら（成功・失敗問わず）ローディング状態を解除
          isLoading = false;
        };
      }}
    >
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
        </div>
      </div>

      <!-- エラーメッセージ -->
      {#if form?.message}
        <p class="mt-4 text-sm text-center text-red-600">{form.message}</p>
      {/if}

      <!-- 送信ボタン -->
      <div class="mt-6">
        <button
          type="submit"
          class="w-full px-4 py-2 font-bold text-white bg-indigo-600 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          保存して開始する
        </button>
      </div>
    </form>
  </div>
</div>
