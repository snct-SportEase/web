<script>
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import EditDisplayNameModal from '$lib/components/EditDisplayNameModal.svelte';

  let { data } = $page;
  $: user = data.user;

  let showEditDisplayNameModal = false;

  function handleDisplayNameClick() {
    showEditDisplayNameModal = true;
  }

  function handleCloseEditDisplayNameModal() {
    showEditDisplayNameModal = false;
  }

  async function handleSaveDisplayName(newDisplayName) {
    try {
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
          display_name: newDisplayName, 
          class_id: user?.class_id || 0 
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || '表示名の更新に失敗しました。');
      }

      window.location.reload();
    } catch (error) {
      throw error;
    }
  }

  async function handleLogout() {
    try {
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
          window.location.href = '/';
        }
      } else {
        window.location.href = '/';
      }
    } catch (error) {
      window.location.href = '/';
    }
  }
</script>

<div class="min-h-screen bg-gray-50">
  <header class="bg-white shadow-sm p-4">
    <div class="flex justify-between items-center">
      <a href="/dashboard" data-sveltekit-preload-data="hover" class="flex items-center"><h1 class="text-2xl font-bold text-gray-800 pl-12">Dashboard</h1></a>
      <div class="flex items-center">
        <button 
          type="button"
          on:click={handleDisplayNameClick}
          class="mr-4 flex items-center space-x-2 px-3 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200 border border-gray-200 rounded-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          title="表示名をクリックして変更"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
          </svg>
          <span>{user?.display_name || user?.email || 'User'}</span>
        </button>
        <button 
          type="button" 
          on:click={handleLogout}
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 transition-colors duration-200"
        >
          Logout
        </button>
      </div>
    </div>
  </header>

  <main class="p-8">
    <slot />
  </main>

  <EditDisplayNameModal
    isOpen={showEditDisplayNameModal}
    currentDisplayName={user?.display_name || ''}
    userRoles={user?.roles || []}
    onClose={handleCloseEditDisplayNameModal}
    onSave={handleSaveDisplayName}
  />
</div>
