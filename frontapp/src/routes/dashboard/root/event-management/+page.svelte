<script>
  import { onMount } from 'svelte';
  import { get } from 'svelte/store';
  import { activeEvent } from '$lib/stores/eventStore.js';

  let events = [];
  let showModal = false;
  let selectedEvent = null;
  let isNameManuallyChanged = false;

  let currentEvent = {
    id: null,
    name: '',
    year: new Date().getFullYear(),
    season: 'spring',
    start_date: '',
    end_date: '',
  };

  $: {
    if (!isNameManuallyChanged && currentEvent.year && currentEvent.season) {
      const seasonText = currentEvent.season === 'spring' ? '春' : '秋';
      currentEvent.name = `${currentEvent.year}${seasonText}季スポーツ大会`;
    }
  }

  function onNameInput() {
    isNameManuallyChanged = true;
  }

  let selectedEventIdForActivation = null;

  onMount(async () => {
    await fetchEvents();
    // initialize activeEvent store from backend
    const active = await activeEvent.init();
    if (active) {
      selectedEventIdForActivation = active.id;
    } else if (events.length > 0) {
      // If no active event is set, default to the latest one
      selectedEventIdForActivation = events[0].id;
    }
  });

  async function fetchEvents() {
    try {
      const response = await fetch('/api/root/events');
      if (!response.ok) {
        throw new Error('Failed to fetch events');
      }
      events = await response.json();
    } catch (error) {
      console.error(error);
      alert(error.message);
    }
  }

  function setActiveEvent() {
    if (selectedEventIdForActivation) {
      activeEvent.setActiveEventById(selectedEventIdForActivation)
        .then(() => {
          const eventToActivate = events.find(e => e.id === parseInt(selectedEventIdForActivation));
          alert(`「${eventToActivate.name}」がアクティブな大会として設定されました。`);
        })
        .catch(err => {
          console.error(err);
          alert('アクティブ大会の設定に失敗しました。');
        });
    }
  }

  function openCreateModal() {
    selectedEvent = null;
    isNameManuallyChanged = false;
    currentEvent = {
      id: null,
      name: '',
      year: new Date().getFullYear(),
      season: 'spring',
      start_date: '',
      end_date: '',
    };
    showModal = true;
  }

  function openEditModal(event) {
    selectedEvent = event;
    isNameManuallyChanged = true; // 編集時は手動変更とみなし、自動更新しない
    currentEvent = {
      ...event,
      start_date: event.start_date ? new Date(event.start_date).toISOString().split('T')[0] : '',
      end_date: event.end_date ? new Date(event.end_date).toISOString().split('T')[0] : '',
    };
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }

  async function handleSave() {
    try {
      const method = selectedEvent ? 'PUT' : 'POST';
      const url = selectedEvent ? `/api/root/events/${selectedEvent.id}` : '/api/root/events';

      const body = {
        ...currentEvent,
        year: parseInt(currentEvent.year, 10),
      };

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save event');
      }

      await fetchEvents();
      closeModal();
    } catch (error) {
      console.error(error);
      alert(error.message);
    }
  }
</script>

<div class="container mx-auto p-4">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-2xl font-bold">大会情報登録・管理</h1>
    <button on:click={openCreateModal} class="btn btn-primary bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700">
      新規作成
    </button>
  </div>

  <div class="mb-8 p-4 border rounded-lg bg-gray-50">
    <h2 class="text-xl font-bold mb-4">アクティブな大会の設定</h2>
    <p class="text-sm text-gray-600 mb-4">ここで設定した大会が、他の管理ページでの操作対象となります。</p>
    <div class="flex items-center space-x-4">
        <select bind:value={selectedEventIdForActivation} class="block w-full max-w-xs border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            <option value={null} disabled>大会を選択...</option>
            {#each events as event}
                <option value={event.id}>{event.name}</option>
            {/each}
        </select>
        <button on:click={setActiveEvent} class="btn btn-secondary bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700" disabled={!selectedEventIdForActivation}>
            この大会をアクティブにする
        </button>
    </div>
    {#if $activeEvent}
        <p class="mt-4 text-sm text-gray-600">現在のアクティブな大会: <span class="font-semibold text-indigo-600">{$activeEvent.name}</span></p>
    {/if}
  </div>

  <div class="bg-white shadow-md rounded-lg overflow-hidden">
    <table class="min-w-full leading-normal">
      <thead>
        <tr>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">大会名</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">年度</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">シーズン</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">期間</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100"></th>
        </tr>
      </thead>
      <tbody>
        {#each events as event}
          <tr>
            <td class="px-5 py-5 border-b border-gray-200 bg-white text-sm">{event.name}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-white text-sm">{event.year}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-white text-sm">{event.season}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-white text-sm">
              {event.start_date ? new Date(event.start_date).toLocaleDateString() : ''} - 
              {event.end_date ? new Date(event.end_date).toLocaleDateString() : ''}
            </td>
            <td class="px-5 py-5 border-b border-gray-200 bg-white text-sm text-right">
              <button on:click={() => openEditModal(event)} class="text-indigo-600 hover:text-indigo-900">編集</button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

{#if showModal}
  <div class="fixed z-10 inset-0 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity" aria-hidden="true">
        <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
      </div>

      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

      <div class="relative z-30 inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">{selectedEvent ? '大会編集' : '大会作成'}</h3>
          <div class="space-y-4">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700">大会名</label>
              <input type="text" id="name" bind:value={currentEvent.name} on:input={onNameInput} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="year" class="block text-sm font-medium text-gray-700">年度</label>
              <input type="number" id="year" bind:value={currentEvent.year} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="season" class="block text-sm font-medium text-gray-700">シーズン</label>
              <select id="season" bind:value={currentEvent.season} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                <option value="spring">春</option>
                <option value="autumn">秋</option>
              </select>
            </div>
            <div>
              <label for="start_date" class="block text-sm font-medium text-gray-700">開始日</label>
              <input type="date" id="start_date" bind:value={currentEvent.start_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="end_date" class="block text-sm font-medium text-gray-700">終了日</label>
              <input type="date" id="end_date" bind:value={currentEvent.end_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
          </div>
        </div>
        <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
          <button on:click={handleSave} type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
            保存
          </button>
          <button on:click={closeModal} type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
            キャンセル
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
