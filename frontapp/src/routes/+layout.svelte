<script>
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import '../app.css';
  import { onMount } from 'svelte';

  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';

  let { data, children } = $props();

  $effect(() => {
    const user = data.user;
    const userProfile = data.userProfile;
    const pathname = $page.url.pathname;

    // ログイン済みでプロフィールも完了しているユーザーを /dashboard へリダイレクト
    if (user && userProfile?.is_profile_complete) {
      if (pathname !== '/dashboard' && !pathname.startsWith('/dashboard/')) {
        goto('/dashboard', { replaceState: true });
      }
    }

    // 未ログインで保護されたページにアクセスしようとした場合、ルートにリダイレクト
    if (!user && pathname.startsWith('/dashboard')) {
      goto('/', { replaceState: true });
    }
  });

</script>

{#if data.user && data.userProfile && !data.userProfile.is_profile_complete}
  <ProfileSetupModal userProfile={data.userProfile} classes={data.classes} />
{/if}

<div class="app">
  <main>
    {@render children()}
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    min-height: 100vh;
  }

  main {
    flex: 1;
    display: flex;
    flex-direction: column;
  }
</style>
