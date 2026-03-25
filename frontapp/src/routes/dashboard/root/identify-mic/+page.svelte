<script>
  import { onMount } from 'svelte';

  let micResult = null;
  let error = null;

  onMount(async () => {
    try {
      const eventRes = await fetch('/api/events/active');
      if (!eventRes.ok) {
        error = '開催中のイベント情報の取得に失敗しました。';
        return;
      }
      const eventData = await eventRes.json();
      if (!eventData.event_id) {
        error = '開催中のイベントがありません。';
        return;
      }

      const response = await fetch(`/api/mic/class?event_id=${eventData.event_id}`);
      if (!response.ok) {
        throw new Error('MICデータの取得に失敗しました。');
      }
      const data = await response.json();
      if (data.message) {
        error = data.message;
      } else {
        micResult = data;
      }
    } catch (err) {
      error = err.message;
    }
  });
</script>

<h1 class="text-2xl font-bold mb-4">MIC確認</h1>

{#if error}
  <p class="text-red-500">{error}</p>
{:else if micResult}
  <div class="bg-white shadow-md rounded-lg p-6">
    <h2 class="text-xl font-semibold mb-2">MIC Class</h2>
    <p><strong>Class:</strong> {micResult.class_name}</p>
    <p><strong>Total Points:</strong> {micResult.total_points}</p>
    <p><strong>Season:</strong> {micResult.season}</p>
  </div>
{:else}
  <p>Loading MIC data...</p>
{/if}
