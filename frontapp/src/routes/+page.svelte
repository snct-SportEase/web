<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';

  // URLパラメータからエラーメッセージをチェック
  onMount(() => {
    if (browser) {
      const urlParams = new URLSearchParams(window.location.search);
      const error = urlParams.get('error');
      
      if (error) {
        // エラーメッセージに基づいて適切なメッセージを表示
        let errorMessage = '';
        switch (error) {
          case 'access_denied':
            errorMessage = 'アクセスが拒否されました。@sendai-nct.jpまたは@sendai-nct.ac.jpのメールアドレスを使用してください。';
            break;
          case 'invalid_domain':
            errorMessage = '許可されていないドメインです。@sendai-nct.jpまたは@sendai-nct.ac.jpのメールアドレスを使用してください。';
            break;
          default:
            errorMessage = 'ログインに失敗しました。@sendai-nct.jpまたは@sendai-nct.ac.jpのメールアドレスを使用してください。';
        }
        
        // アラートポップアップを表示
        alert(errorMessage);
        
        // URLからエラーパラメータを削除
        const newUrl = new URL(window.location);
        newUrl.searchParams.delete('error');
        window.history.replaceState({}, '', newUrl);
      }
    }
  });
</script>

<div class="min-h-screen bg-gray-100 flex flex-col justify-center items-center px-4 sm:px-0">
    <div class="max-w-md w-full bg-white shadow-md rounded-lg p-6 sm:p-8">
        <div class="flex justify-center mb-6">
            <img src="/icon.png" alt="SportEase Logo" class="h-16 w-16 sm:h-20 sm:w-20">
        </div>
        <h1 class="text-xl sm:text-2xl font-bold text-center text-gray-800 mb-4">SportEaseへようこそ</h1>

        <div class="text-center">
            <p class="text-gray-600 mb-6">Googleアカウントでサインインしてください。</p>
            <a 
                href="/api/auth/google/login"
                class="w-full flex items-center justify-center bg-white border border-gray-300 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 transition duration-200"
            >
                <svg class="w-5 h-5 mr-2" aria-hidden="true" focusable="false" data-prefix="fab" data-icon="google" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 488 512"><path fill="currentColor" d="M488 261.8C488 403.3 381.5 512 244 512 109.8 512 0 402.2 0 261.8S109.8 11.6 244 11.6C318.3 11.6 382.8 45 427.3 99.9L353.5 168.3C327.2 145.3 289.3 129.8 244 129.8c-66.8 0-121.5 54.9-121.5 122.1s54.7 122.1 121.5 122.1c76.3 0 104.5-54.7 108.8-82.9H244v-66.8h236.1c2.4 12.6 3.9 26.1 3.9 40.9z"></path></svg>
                Googleでサインイン
            </a>
        </div>
    </div>
</div>