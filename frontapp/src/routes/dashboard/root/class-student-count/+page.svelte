<script>
  import { onMount } from 'svelte';

  /** @type {import('./$types').PageData} */
  export let data;

  let classes = [];
  let errorMessage = '';
  let successMessage = '';
  let isLoading = false;
  let csvFile = null;

  onMount(() => {
    if (data.classes) {
      // Create a deep copy for editing
      classes = JSON.parse(JSON.stringify(data.classes));
    }
    if (data.error) {
      errorMessage = data.error;
    }
  });

  async function handleSave() {
    isLoading = true;
    errorMessage = '';
    successMessage = '';

    const payload = classes.map(c => ({
      class_id: c.id,
      student_count: Number(c.student_count) || 0,
    }));

    try {
      const response = await fetch(`/api/root/classes/student-counts`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const res = await response.json();
        throw new Error(res.error || '更新に失敗しました。');
      }

      successMessage = '生徒数を更新しました。';
      // Optionally, re-fetch data to confirm changes
      const freshData = await fetch(`/api/classes`).then(res => res.json());
      classes = JSON.parse(JSON.stringify(freshData));

    } catch (error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  function handleFileSelect(e) {
    csvFile = e.target.files[0];
  }

  async function handleCSVUpload() {
    if (!csvFile) {
      errorMessage = 'CSVファイルを選択してください。';
      return;
    }

    isLoading = true;
    errorMessage = '';
    successMessage = '';

    const formData = new FormData();
    formData.append('csv', csvFile);

    try {
      const response = await fetch(`/api/root/classes/student-counts/csv`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const res = await response.json();
        throw new Error(res.error || 'CSVでの更新に失敗しました。');
      }
      
      successMessage = 'CSVで生徒数を更新しました。ページをリロードして確認してください。';
      // After successful upload, refresh the class list
      const freshData = await fetch(`/api/classes`).then(res => res.json());
      classes = JSON.parse(JSON.stringify(freshData));
      csvFile = null; // Reset file input

    } catch (error) {
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="container mx-auto p-8">
  <h1 class="text-2xl font-bold mb-6">各クラス人数設定</h1>

  {#if errorMessage}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">
      <span class="block sm:inline">{errorMessage}</span>
    </div>
  {/if}

  {#if successMessage}
    <div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative mb-4" role="alert">
      <span class="block sm:inline">{successMessage}</span>
    </div>
  {/if}

  <div class="bg-white shadow-md rounded-lg p-6">
    <div class="mb-8">
      <h2 class="text-xl font-semibold mb-4">CSVで一括更新</h2>
      <div class="flex items-center space-x-4">
        <input
          type="file"
          accept=".csv"
          on:change={handleFileSelect}
          class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
        />
        <button
          on:click={handleCSVUpload}
          disabled={isLoading || !csvFile}
          class="btn btn-primary"
        >
          {#if isLoading}
            <span class="loading loading-spinner"></span>
          {/if}
          CSVで更新
        </button>
      </div>
       <p class="text-sm text-gray-500 mt-2">CSVフォーマット: 1列目にクラス名, 2列目に生徒数を入力してください。(ヘッダー行あり)</p>
    </div>

    <div class="divider">OR</div>

    <form on:submit|preventDefault={handleSave}>
      <h2 class="text-xl font-semibold mb-4">手動で更新</h2>
      <div class="overflow-x-auto">
        <table class="table w-full">
          <thead>
            <tr>
              <th>クラス名</th>
              <th>現在の生徒数</th>
              <th>新しい生徒数</th>
            </tr>
          </thead>
          <tbody>
            {#each classes as classItem, i}
              <tr>
                <td>{classItem.name}</td>
                <td>{data.classes[i].student_count}</td>
                <td>
                  <input
                    type="number"
                    bind:value={classItem.student_count}
                    class="input input-bordered w-full max-w-xs"
                    min="0"
                  />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="mt-6 text-right">
        <button type="submit" disabled={isLoading} class="btn btn-primary">
          {#if isLoading}
            <span class="loading loading-spinner"></span>
          {/if}
          保存
        </button>
      </div>
    </form>
  </div>
</div>
