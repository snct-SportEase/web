<script>
  import { onMount } from 'svelte';
  import { activeEventId } from '$lib/stores/eventStore.js';

  let mvpResult = null;
  let error = null;
  let eventId;

  activeEventId.subscribe(value => {
    eventId = value;
  });

  onMount(async () => {
    if (!eventId) {
      error = 'No active event selected.';
      return;
    }

    try {
      const response = await fetch(`/api/mvp/class?event_id=${eventId}`);
      if (!response.ok) {
        throw new Error('Failed to fetch MVP data');
      }
      const data = await response.json();
      if (data.message) {
        error = data.message;
      } else {
        mvpResult = data;
      }
    } catch (err) {
      error = err.message;
    }
  });
</script>

<h1 class="text-2xl font-bold mb-4">MVP確認</h1>

{#if error}
  <p class="text-red-500">{error}</p>
{:else if mvpResult}
  <div class="bg-white shadow-md rounded-lg p-6">
    <h2 class="text-xl font-semibold mb-2">MVP Class</h2>
    <p><strong>Class:</strong> {mvpResult.class_name}</p>
    <p><strong>Total Points:</strong> {mvpResult.total_points}</p>
    <p><strong>Season:</strong> {mvpResult.season}</p>
  </div>
{:else}
  <p>Loading MVP data...</p>
{/if}
