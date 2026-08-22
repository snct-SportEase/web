<script>
  import Modal from '$lib/components/Modal.svelte';
  import FormField from '$lib/components/FormField.svelte';
  let showModal = $state(true);
  let currentEvent = $state({
    name: '',
    year: new Date().getFullYear(),
    season: 'spring',
    start_date: '',
    end_date: '',
  });

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

<Modal bind:open={showModal} title="最初の大会情報を設定してください">
  <p class="mb-4 text-sm text-gray-500">SportEaseを始めるには、まず最初の大会情報を登録する必要があります。</p>
  <div class="space-y-4">
            <FormField label="大会名" inputId="name">
              <input type="text" id="name" bind:value={currentEvent.name} placeholder="例：令和6年度 春季球技大会" class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </FormField>
            <FormField label="年度" inputId="year">
              <input type="number" id="year" bind:value={currentEvent.year} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </FormField>
            <FormField label="シーズン" inputId="season">
              <select id="season" bind:value={currentEvent.season} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                <option value="spring">春</option>
                <option value="autumn">秋</option>
              </select>
            </FormField>
            <FormField label="開始日" inputId="start_date">
              <input type="date" id="start_date" bind:value={currentEvent.start_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </FormField>
            <FormField label="終了日" inputId="end_date">
              <input type="date" id="end_date" bind:value={currentEvent.end_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </FormField>
  </div>
  {#snippet footer()}
    <button onclick={handleSave} type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
      保存
    </button>
  {/snippet}
</Modal>
