<script>
  import { page } from '$app/stores';
  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';
  import EventSetupModal from '$lib/components/EventSetupModal.svelte';

  let { data } = $page;
  $: user = data.user;
  $: classes = data.classes;
  $: events = data.events;
  $: form = data.form;

  $: isRoot = user?.roles?.some(role => role.name === 'root');
  $: showEventSetup = user?.is_profile_complete && isRoot && events?.length === 0;
</script>

<h1 class="text-3xl font-bold">Welcome, {user?.display_name || user?.email || 'User'}!</h1>
<p class="text-gray-600">This is your dashboard.</p>

{#if user && !user.is_profile_complete}
  <ProfileSetupModal classes={classes} form={form} />
{/if}

{#if showEventSetup}
  <EventSetupModal />
{/if}
