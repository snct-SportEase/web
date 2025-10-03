<script>
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';

  let { data } = $page;
  $: user = data.user;
  $: classes = data.classes;
  $: form = data.form;

  async function handleLogout() {
    try {
      // セッションクッキーを取得
      let sessionToken = null;
      if (browser) {
        const cookies = document.cookie.split('; ');
        const sessionCookie = cookies.find(row => row.startsWith('session_token='));
        sessionToken = sessionCookie ? sessionCookie.split('=')[1] : null;
      }

      if (sessionToken) {
        const response = await fetch('/api/auth/logout', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            'Cookie': `session_token=${sessionToken}`,
          },
        });

        if (response.ok) {
          // ログアウト成功時はページをリロードしてログインページに遷移
          window.location.href = '/';
        }
      } else {
        // セッションがない場合は直接ログインページに遷移
        window.location.href = '/';
      }
    } catch (error) {
      console.error('Logout error:', error);
      // エラーが発生してもログインページに遷移
      window.location.href = '/';
    }
  }
</script>

<div class="min-h-screen bg-gray-50">
  <nav class="bg-white shadow-sm">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center">
          <img class="h-8 w-auto" src="/icon.png" alt="SportEase Logo">
          <span class="font-semibold text-xl ml-2">SportEase</span>
        </div>
        <div class="flex items-center">
          <span class="mr-4">{user?.display_name || user?.email || 'User'}</span>
          <button 
            type="button" 
            on:click={handleLogout}
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  </nav>

  <main class="p-8">
    <h1 class="text-3xl font-bold">Welcome, {user?.display_name || user?.email || 'User'}!</h1>
    <p class="text-gray-600">This is your dashboard.</p>
  </main>

  {#if user && !user.is_profile_complete}
    <ProfileSetupModal classes={classes} form={form} />
  {/if}
</div>
