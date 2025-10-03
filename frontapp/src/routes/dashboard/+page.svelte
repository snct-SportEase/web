<script>
  import { page } from '$app/stores';
  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';

  let { data } = $page;
  $: user = data.user;
  $: classes = data.classes;
  $: form = data.form;

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
          <span class="mr-4">{user?.email}</span>
          <form action="?/logout" method="POST">
            <button type="submit" class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700">Logout</button>
          </form>
        </div>
      </div>
    </div>
  </nav>

  <main class="p-8">
    <h1 class="text-3xl font-bold">Welcome, {user?.displayName || 'User'}!</h1>
    <p class="text-gray-600">This is your dashboard.</p>
  </main>

  {#if user && !user.is_profile_complete}
    <ProfileSetupModal classes={classes} form={form} />
  {/if}
</div>
