<script>
  let showModal = true;
  let currentEvent = {
    name: '',
    year: new Date().getFullYear(),
    season: 'spring',
    start_date: '',
    end_date: '',
  };

  async function handleSave() {
    try {
      const body = {
        ...currentEvent,
        year: parseInt(currentEvent.year, 10),
      };

      const response = await fetch('/api/root/events', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to create event');
      }

      // Reload the page to reflect the changes or navigate
      showModal = false;
      //window.location.reload(); 

    } catch (error) {
      console.error(error);
      alert(error.message);
    }
  }
</script>

{#if showModal}
  <div class="fixed z-10 inset-0 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity" aria-hidden="true">
        <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
      </div>

      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

      <div class="relative z-30 inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900 mb-2">最初の大会情報を設定してください</h3>
          <p class="text-sm text-gray-500 mb-4">SportEaseを始めるには、まず最初の大会情報を登録する必要があります。</p>
          <div class="space-y-4">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700">大会名</label>
              <input type="text" id="name" bind:value={currentEvent.name} placeholder="例：令和6年度 春季球技大会" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
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
        </div>
      </div>
    </div>
  </div>
{/if}
