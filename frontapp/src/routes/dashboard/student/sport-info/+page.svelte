<script>
  import { onMount } from 'svelte';
  import { error } from '@sveltejs/kit';
  import RulesDisplayModal from '$lib/components/RulesDisplayModal.svelte';

  let eventSports = [];
  let allSports = [];

  let showRulesModal = false;
  let selectedRulesType = '';
  let selectedRulesContent = '';
  let selectedRulesPdfUrl = '';
  let selectedSportName = '';

  onMount(async () => {
    const sessionToken = document.cookie.split('; ').find(row => row.startsWith('session_token='))?.split('=')[1];
    const headers = {
      'Content-Type': 'application/json',
      'Cookie': `session_token=${sessionToken}`,
    };

    try {
      // Fetch the active event
      const eventRes = await fetch(`/api/events/active`, { headers });
      if (!eventRes.ok) {
        const errorBody = await eventRes.text();
        console.error(`Failed to load active event: ${eventRes.status} ${errorBody}`);
        throw error(eventRes.status, 'Failed to load active event');
      }
      const eventData = await eventRes.json();
      const activeEventId = eventData.event_id;
      selectedRulesType = eventData.rules_type;

      if (!activeEventId) {
        eventSports = [];
        allSports = [];
        return;
      }

      // Fetch sports for the active event
      const sportsRes = await fetch(`/api/events/${activeEventId}/sports`, { headers });
      if (!sportsRes.ok) {
        const errorBody = await sportsRes.text();
        console.error(`Failed to load sports: ${sportsRes.status} ${errorBody}`);
        throw error(sportsRes.status, 'Failed to load sports');
      }
      eventSports = await sportsRes.json();

      // Fetch all sports to get their names
      const allSportsRes = await fetch(`/api/admin/allsports`, { headers });
      if (!allSportsRes.ok) {
        const errorBody = await allSportsRes.text();
        console.error(`Failed to load all sports: ${allSportsRes.status} ${errorBody}`);
        throw error(allSportsRes.status, 'Failed to load all sports');
      }
      allSports = await allSportsRes.json();

    } catch (err) {
      console.error('Error loading sport info:', err);
      // Handle error display to user if needed
    }
  });

  function getSportName(sportId) {
    const sport = allSports.find(s => s.id === sportId);
    return sport ? sport.name : '不明な競技';
  }

  function openRulesModal(sport) {
    // PDFの場合は別タブで開く
    if (sport.rules_type === 'pdf' && sport.rules_pdf_url) {
      window.open(sport.rules_pdf_url, '_blank');
      return;
    }
    
    // マークダウンの場合はモーダルを開く
    selectedSportName = getSportName(sport.sport_id);
    selectedRulesType = sport.rules_type;
    selectedRulesContent = sport.rules;
    selectedRulesPdfUrl = sport.rules_pdf_url;
    showRulesModal = true;
  }

  function closeRulesModal() {
    showRulesModal = false;
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-2xl md:text-3xl font-bold text-gray-800 border-b pb-2">競技一覧・詳細閲覧</h1>

  {#if eventSports.length > 0 && allSports.length > 0}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each eventSports as sport}
        <div class="bg-white p-6 rounded-lg shadow-lg flex flex-col">
          <h2 class="text-xl font-semibold text-gray-800 mb-2">{getSportName(sport.sport_id)}</h2>
          {#if sport.description}
            <p class="text-gray-600 mb-4 flex-grow">{sport.description || '説明がありません'}</p>
          {/if}
          <div class="mt-auto pt-4 border-t border-gray-200">
            <p class="text-gray-700 text-sm mb-2"><strong>場所:</strong>
              {#if sport.location === 'gym1'}
                <span>第一体育館</span>
              {:else if sport.location === 'gym2'}
                <span>第二体育館</span>
              {:else if sport.location === 'ground'}
                <span>グラウンド</span>
              {:else if sport.location === 'noon_game'}
                <span>グラウンド</span>
              {:else}
                <span>{sport.location}</span>
              {/if}
            </p>
            <div>
              <button
                class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                on:click={() => openRulesModal(sport)}
              >
                ルールを見る
              </button>
            </div>
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

<RulesDisplayModal
  bind:showModal={showRulesModal}
  rulesType={selectedRulesType}
  rulesContent={selectedRulesContent}
  rulesPdfUrl={selectedRulesPdfUrl}
  sportName={selectedSportName}
  on:close={closeRulesModal}
/>