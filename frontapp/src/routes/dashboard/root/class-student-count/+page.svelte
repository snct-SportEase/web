<script>
  import { onMount } from 'svelte';

  /** @type {import('./$types').PageData} */
  export let data;

  let classes = [];
  let originalClasses = []; // 保存前の状態を保持
  let errorMessage = '';
  let successMessage = '';
  let isSaving = false;
  let isUploading = false;
  let csvFile = null;

  onMount(() => {
    if (data.classes) {
      // データをディープコピーして編集用と保存前用に保持
      classes = JSON.parse(JSON.stringify(data.classes));
      originalClasses = JSON.parse(JSON.stringify(data.classes));
    }
    if (data.error) {
      errorMessage = data.error;
    }
  });

  async function handleSave() {
    isSaving = true;
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
      // データを再取得して画面を更新
      const freshData = await fetch(`/api/classes`).then(res => res.json());
      classes = JSON.parse(JSON.stringify(freshData));
      originalClasses = JSON.parse(JSON.stringify(freshData));

    } catch (error) {
      errorMessage = error.message;
    } finally {
      isSaving = false;
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

    isUploading = true;
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
      
      successMessage = 'CSVで生徒数を更新しました。';
      // データを再取得して画面を更新
      const freshData = await fetch(`/api/classes`).then(res => res.json());
      classes = JSON.parse(JSON.stringify(freshData));
      originalClasses = JSON.parse(JSON.stringify(freshData));
      csvFile = null; // ファイル選択をリセット

    } catch (error) {
      errorMessage = error.message;
    } finally {
      isUploading = false;
    }
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-2xl md:text-3xl font-bold">各クラス人数設定</h1>

  <!-- Messages -->
  {#if successMessage}
    <div class="bg-green-100 border-l-4 border-green-500 text-green-700 p-4" role="alert">
      <p class="font-bold">成功</p>
      <p>{successMessage}</p>
    </div>
  {/if}
  {#if errorMessage}
    <div class="bg-red-100 border-l-4 border-red-500 text-red-700 p-4" role="alert">
      <p class="font-bold">エラー</p>
      <p>{errorMessage}</p>
    </div>
  {/if}

  <!-- CSV Upload -->
  <div class="bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-semibold mb-4">CSVで一括更新</h2>
    <form on:submit|preventDefault={handleCSVUpload} class="flex flex-col sm:flex-row items-start sm:items-end space-y-4 sm:space-y-0 sm:space-x-4">
      <div class="flex-grow w-full">
        <label for="csvfile" class="block text-sm font-medium text-gray-700">CSVファイル</label>
        <input 
          type="file" 
          id="csvfile" 
          accept=".csv" 
          on:change={handleFileSelect} 
          class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"/>
        <p class="text-xs text-gray-500 mt-1">フォーマット: 1列目にクラス名, 2列目に生徒数 (ヘッダー行あり)</p>
      </div>
      <button type="submit" disabled={isUploading || !csvFile} class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
        {#if isUploading}
          <span class="loading loading-spinner"></span>
        {/if}
        CSVで更新
      </button>
    </form>
  </div>

  <!-- Manual Update -->
  <div class="bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-semibold mb-4">手動で更新</h2>
    <form on:submit|preventDefault={handleSave}>
      <div class="overflow-x-auto rounded-lg border border-gray-200">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">クラス名</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">現在の生徒数</th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">新しい生徒数</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {#if classes.length > 0}
              {#each classes as classItem, i}
                <tr class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{classItem.name}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{originalClasses[i].student_count}</td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      bind:value={classItem.student_count}
                      class="input input-bordered w-full max-w-xs"
                      min="0"
                    />
                  </td>
                </tr>
              {/each}
            {:else}
              <tr>
                <td colspan="3" class="px-6 py-4 text-center text-sm text-gray-500">
                  クラスデータが見つかりません。
                </td>
              </tr>
            {/if}
          </tbody>
        </table>
      </div>
      <div class="mt-6 text-right">
        <button type="submit" disabled={isSaving} class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500" >
          {#if isSaving}
            <span class="loading loading-spinner"></span>
          {/if}
          保存
        </button>
      </div>
    </form>
  </div>
</div>