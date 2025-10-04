<script>
  import { page } from '$app/stores';

  let { data } = $page;
  $: whitelist = data.whitelist;

  let newEmailLocal = '';
	let newEmailDomain = '@sendai-nct.jp'; // デフォルト値
  const allowedDomains = ['@sendai-nct.jp', '@sendai-nct.ac.jp'];

  let newRole = 'student';
  let csvFile = null;
  let message = '';
  let errorMessage = '';

  async function addEmail() {
    errorMessage = '';
    message = '';

    // ローカル部とドメイン部を結合して完全なメールアドレスを作成
    const fullEmail = newEmailLocal.trim() + newEmailDomain;

    try {
      const response = await fetch('/api/root/whitelist', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email: fullEmail, role: newRole })
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to add email');
      }
      message = 'Email added successfully!';
      newEmailLocal = '';
      // Refresh whitelist
      const res = await fetch('/api/root/whitelist');
      whitelist = await res.json();
    } catch (error) {
      errorMessage = error.message;
    }
  }

  async function uploadCsv() {
    errorMessage = '';
    message = '';
    if (!csvFile) {
      errorMessage = 'Please select a CSV file.';
      return;
    }

    const formData = new FormData();
    formData.append('csvfile', csvFile);

    try {
      const response = await fetch('/api/root/whitelist/csv', {
        method: 'POST',
        body: formData,
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to upload CSV');
      }
      message = 'CSV uploaded and processed successfully!';
      // Refresh whitelist
      const res = await fetch('/api/root/whitelist');
      whitelist = await res.json();
    } catch (error) {
      errorMessage = error.message;
    }
  }
</script>

<div class="space-y-8">
  <h1 class="text-3xl font-bold">ホワイトリスト管理</h1>

  <!-- Messages -->
  {#if message}
    <div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative" role="alert">
      <span class="block sm:inline">{message}</span>
    </div>
  {/if}
  {#if errorMessage}
    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">
      <span class="block sm:inline">{errorMessage}</span>
    </div>
  {/if}

  <!-- Add Single Email -->
  <div class="bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-semibold mb-4">ホワイトリストに追加</h2>
    <form on:submit|preventDefault={addEmail} class="flex items-end space-x-4">
      <div class="flex-grow">
        <label for="email_local" class="block text-sm font-medium text-gray-700">メールアドレス</label>
				<div class="flex mt-1">
					<!-- ローカル部入力 -->
					<input
						type="text"
						id="email_local"
						bind:value={newEmailLocal}
						required
						class="block w-2/3 rounded-l-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm placeholder-gray-400"
						placeholder="taro.yamada"
					/>
					<!-- ドメイン部選択 -->
					<select
						id="email_domain"
						bind:value={newEmailDomain}
						class="block w-1/3 rounded-r-lg border-l-0 border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm bg-gray-50 text-gray-700 font-medium"
					>
						{#each allowedDomains as domain}
							<option value={domain}>{domain}</option>
						{/each}
					</select>
      </div>
      <div>
        <label for="role" class="block text-sm font-medium text-gray-700">Role</label>
        <select id="role" bind:value={newRole} class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm">
          <option value="student">Student</option>
          <option value="admin">Admin</option>
          <option value="root">Root</option>
        </select>
      </div>
      <button type="submit" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">Add</button>
    </form>
  </div>

  <!-- Bulk Upload CSV -->
  <div class="bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-semibold mb-4">CSVで一括追加</h2>
    <form on:submit|preventDefault={uploadCsv} class="flex items-end space-x-4">
      <div class="flex-grow">
        <label for="csvfile" class="block text-sm font-medium text-gray-700">CSV File (email,role)</label>
        <input type="file" id="csvfile" on:change={(e) => csvFile = e.target.files[0]} accept=".csv" class="mt-1 block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"/>
      </div>
      <button type="submit" class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">Upload</button>
    </form>
  </div>

  <!-- Whitelist Table -->
  <div class="bg-white p-6 rounded-lg shadow">
    <h2 class="text-xl font-semibold mb-4">現在のホワイトリスト</h2>
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Email</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Role</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          {#if whitelist && whitelist.length > 0}
            {#each whitelist as entry}
              <tr>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{entry.email}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{entry.role}</td>
              </tr>
            {/each}
          {:else}
            <tr>
              <td colspan="2" class="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">No whitelisted emails found.</td>
            </tr>
          {/if}
        </tbody>
      </table>
    </div>
  </div>
</div>
