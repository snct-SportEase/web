<script>
  import { onMount } from 'svelte';
  import { error } from '@sveltejs/kit';

  let eventSports = $state([]);

  onMount(async () => {
    try {
      // Fetch the active event
      const eventRes = await fetch(`/api/events/active`, { credentials: 'include' });
      if (!eventRes.ok) {
        const errorBody = await eventRes.text();
        console.error(`Failed to load active event: ${eventRes.status} ${errorBody}`);
        throw error(eventRes.status, 'Failed to load active event');
      }
      const eventData = await eventRes.json();
      const activeEventId = eventData.event_id;

      if (!activeEventId) {
        eventSports = [];
        return;
      }

      // Fetch sports for the active event
      const sportsRes = await fetch(`/api/events/${activeEventId}/sports`, { credentials: 'include' });
      if (!sportsRes.ok) {
        const errorBody = await sportsRes.text();
        console.error(`Failed to load sports: ${sportsRes.status} ${errorBody}`);
        throw error(sportsRes.status, 'Failed to load sports');
      }
      eventSports = await sportsRes.json();

    } catch (err) {
      console.error('Error loading sport info:', err);
      // Handle error display to user if needed
    }
  });

  function getSportName(sport) {
    return sport?.sport_name || '不明な競技';
  }

  function displayLocation(location) {
    if (typeof location === 'string' && location.startsWith('other:')) {
      return location.slice('other:'.length);
    }

    const labels = {
      gym1: '第一体育館',
      gym2: '第二体育館',
      ground: 'グラウンド',
      noon_game: 'グラウンド',
      other: 'その他'
    };
    return labels[location] || location;
  }

  function openRulesPdf(pdfUrl) {
    window.open(pdfUrl, '_blank', 'noopener,noreferrer');
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-2xl md:text-3xl font-bold text-gray-800 border-b pb-2">競技一覧・詳細閲覧</h1>

  {#if eventSports.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each eventSports as sport (sport.sport_id)}
        <div class="bg-white p-6 rounded-lg shadow-lg flex flex-col">
          <h2 class="text-xl font-semibold text-gray-800 mb-2">{getSportName(sport)}</h2>
          {#if sport.description}
            <p class="text-gray-600 mb-4 flex-grow">{sport.description || '説明がありません'}</p>
          {/if}
          <div class="mt-auto pt-4 border-t border-gray-200">
            <p class="text-gray-700 text-sm mb-2"><strong>場所:</strong>
              <span>{displayLocation(sport.location)}</span>
            </p>
            {#if sport.rules_pdf_url}
              <div>
                <button
                  class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  onclick={() => openRulesPdf(sport.rules_pdf_url)}
                >
                  ルールPDFを見る
                </button>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="bg-blue-100 border-l-4 border-blue-500 text-blue-700 p-4" role="alert">
      <p class="font-bold">情報</p>
      <p>現在開催中の競技はありません。</p>
    </div>
  {/if}
</div>
