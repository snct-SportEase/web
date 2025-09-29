<script>
  import { goto } from '$app/navigation';
  import { applyAction, enhance } from '$app/forms';

  let { data } = $props();

  console.log('data.profile', data.userProfile);
</script>

<div class="min-h-screen bg-gray-100 p-4 sm:p-8">
  <div class="max-w-4xl mx-auto">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl sm:text-3xl font-bold text-gray-800">ダッシュボード</h1>
      <form 
        action="/api/logout" 
        method="POST" 
        use:enhance={() => {
          return async ({ result }) => {
            console.log('Form submission result:', result);
            // The result object appears to be the JSON body directly.
            if (result.type === 'success') {
              await goto('/', { invalidateAll: true });
            } else {
              console.error('Logout failed:', result);
              await applyAction(result);
            }
          };
        }}
      >
        <button 
          type="submit"
          class="bg-red-500 text-white py-2 px-4 rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-opacity-50 transition duration-200"
        >
          ログアウト
        </button>
      </form>
    </div>

    <div class="bg-white shadow-md rounded-lg p-6">
      <h2 class="text-xl font-semibold text-gray-700 mb-4">ようこそ</h2>
      {#if data.userProfile}
        <p class="text-gray-600">こんにちは、{data.userProfile.display_name}さん。</p>
        <p class="text-gray-600">クラス: {data.userProfile.class.name}</p>
      {/if}
    </div>
  </div>
</div>
