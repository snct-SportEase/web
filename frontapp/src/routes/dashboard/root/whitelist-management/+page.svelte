<script>
  import { page } from '$app/stores';

  let { data } = $page;
  $: whitelist = data.whitelist;

  // ロールのソート順序を定義 (数値が高いほど上位)
	const roleRank = {
		'root': 3,
		'admin': 2,
		'student': 1
	};

  // フィルタリング用の状態変数を追加
	let filterText = ''; // Email search term (メールアドレス検索用語)
	let filterRole = 'all'; // Role filter ('all', 'root', 'admin', 'student')

  // ソート用の状態変数を追加
	let sortColumn = 'email'; // デフォルトはメールアドレスでソート
	let sortDirection = 'asc'; // デフォルトは昇順

	// ソートロジック
	function handleSort(column) {
		if (sortColumn === column) {
			// 同じカラムがクリックされたら、ソート方向を切り替える
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			// 異なるカラムがクリックされたら、そのカラムで昇順ソートを開始
			sortColumn = column;
			sortDirection = 'asc';
		}
	}

  // フィルタリングされたリストをリアクティブに生成
	$: filteredWhitelist = (() => {
		if (!whitelist) return [];

		return whitelist.filter(entry => {
			// 1. Emailフィルタリング (大文字小文字を区別しない部分一致)
			// 日本語環境ではトリミングは必須ではありませんが、念のため残しています
			const matchesEmail = entry.email.toLowerCase().includes(filterText.toLowerCase().trim());

			// 2. Roleフィルタリング ('all'の場合はすべて一致)
			const matchesRole = filterRole === 'all' || entry.role === filterRole;

			return matchesEmail && matchesRole;
		});
	})();

	// ソートされたリストをリアクティブに生成
	// ソートされたリストをリアクティブに生成 (フィルタリングされたリストを基にソート)
	$: sortedWhitelist = (() => {
		const list = [...filteredWhitelist]; // フィルタリングされたリストを使用

		list.sort((a, b) => {
			let comparison = 0;

			if (sortColumn === 'role') {
				// ロール名ではなく、定義したランク（数値）で比較する
				const aRank = roleRank[a.role] || 0; // 未知のロールは最下位 (0)
				const bRank = roleRank[b.role] || 0;

				if (aRank > bRank) {
					comparison = 1;
				} else if (aRank < bRank) {
					comparison = -1;
				}
			} else {
				// メールアドレスなど、その他のカラムは標準の文字列比較
				const aValue = a[sortColumn];
				const bValue = b[sortColumn];
				
				if (aValue > bValue) {
					comparison = 1;
				} else if (aValue < bValue) {
					comparison = -1;
				}
			}

			// ソート方向に応じて比較結果を反転
			return sortDirection === 'desc' ? comparison * -1 : comparison;
		});

		return list;
	})();

  let newEmailLocal = '';
	let newEmailDomain = '@sendai-nct.jp'; // デフォルト値
  const allowedDomains = ['@sendai-nct.jp', '@sendai-nct.ac.jp'];

  let newRole = 'student';
  let csvFile = null;
  let message = '';
  let errorMessage = '';
  
  // 削除機能用の状態変数
  let selectedEmails = new Set();
  let isDeleting = false;

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

  // チェックボックスのトグル
  function toggleEmail(email) {
    const newSet = new Set(selectedEmails);
    if (newSet.has(email)) {
      newSet.delete(email);
    } else {
      newSet.add(email);
    }
    selectedEmails = newSet; // リアクティブ更新
  }

  // すべて選択/解除
  function toggleAll() {
    const newSet = new Set();
    if (selectedEmails.size !== sortedWhitelist.length) {
      sortedWhitelist.forEach(entry => newSet.add(entry.email));
    }
    selectedEmails = newSet; // リアクティブ更新
  }

  // 単一削除
  async function deleteEmail(email) {
    if (!confirm(`「${email}」をホワイトリストから削除しますか？`)) {
      return;
    }

    errorMessage = '';
    message = '';
    isDeleting = true;

    try {
      const response = await fetch('/api/root/whitelist', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email })
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to delete email');
      }
      message = 'Email deleted successfully!';
      // Refresh whitelist
      const res = await fetch('/api/root/whitelist');
      whitelist = await res.json();
      selectedEmails.clear();
    } catch (error) {
      errorMessage = error.message;
    } finally {
      isDeleting = false;
    }
  }

  // 複数削除
  async function deleteSelectedEmails() {
    if (selectedEmails.size === 0) {
      errorMessage = '削除するメールアドレスを選択してください。';
      return;
    }

    const emails = Array.from(selectedEmails);
    if (!confirm(`${emails.length}件のメールアドレスをホワイトリストから削除しますか？`)) {
      return;
    }

    errorMessage = '';
    message = '';
    isDeleting = true;

    try {
      const response = await fetch('/api/root/whitelist/bulk', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ emails })
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to delete emails');
      }
      message = `${emails.length}件のメールアドレスを削除しました！`;
      // Refresh whitelist
      const res = await fetch('/api/root/whitelist');
      whitelist = await res.json();
      selectedEmails.clear();
    } catch (error) {
      errorMessage = error.message;
    } finally {
      isDeleting = false;
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
    <!-- 💡 フィルタリング コントロール -->
		<div class="mb-6 flex flex-col md:flex-row space-y-4 md:space-y-0 md:space-x-4">
			<div class="flex-grow">
				<label for="email_search" class="block text-sm font-medium text-gray-700">Email検索</label>
				<input
					type="text"
					id="email_search"
					bind:value={filterText}
					placeholder="メールアドレスの一部を入力..."
					class="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
				/>
			</div>
			<div class="w-full md:w-1/4">
				<label for="role_filter" class="block text-sm font-medium text-gray-700">Roleでフィルタ</label>
				<select
					id="role_filter"
					bind:value={filterRole}
					class="mt-1 block w-full rounded-lg border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
				>
					<option value="all">すべて</option>
					<option value="root">Root</option>
					<option value="admin">Admin</option>
					<option value="student">Student</option>
				</select>
			</div>
		</div>

		<!-- 削除ボタン -->
		{#if sortedWhitelist && sortedWhitelist.length > 0}
			<div class="mb-4 flex justify-between items-center">
				<div class="text-sm text-gray-600">
					{selectedEmails.size}件選択中
				</div>
				<button
					on:click={deleteSelectedEmails}
					disabled={selectedEmails.size === 0 || isDeleting}
					class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:bg-gray-400 disabled:cursor-not-allowed"
				>
					{isDeleting ? '削除中...' : '選択した項目を削除'}
				</button>
			</div>
		{/if}

		<div class="overflow-x-auto rounded-lg border border-gray-200">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<!-- チェックボックス列 -->
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							<input
								type="checkbox"
								checked={sortedWhitelist && sortedWhitelist.length > 0 && selectedEmails.size === sortedWhitelist.length}
								on:change={toggleAll}
								class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
							/>
						</th>
						<!-- Email Sort Header -->
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 transition duration-100"
							on:click={() => handleSort('email')}
						>
							<div class="flex items-center">
								Email
								{#if sortColumn === 'email'}
									<span class="ml-1 text-gray-900">
										{#if sortDirection === 'asc'}
											&#9650; <!-- Up Arrow -->
										{:else}
											&#9660; <!-- Down Arrow -->
										{/if}
									</span>
								{/if}
							</div>
						</th>
						<!-- Role Sort Header -->
						<th
							scope="col"
							class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 transition duration-100"
							on:click={() => handleSort('role')}
						>
							<div class="flex items-center">
								Role
								{#if sortColumn === 'role'}
									<span class="ml-1 text-gray-900">
										{#if sortDirection === 'asc'}
											&#9650; <!-- Up Arrow -->
										{:else}
											&#9660; <!-- Down Arrow -->
										{/if}
									</span>
								{/if}
							</div>
						</th>
						<!-- 操作列 -->
						<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
							操作
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					<!-- 💡 sortedWhitelist を使用 -->
					{#if sortedWhitelist && sortedWhitelist.length > 0}
						{#each sortedWhitelist as entry}
							<tr class="hover:bg-gray-50 transition duration-100">
								<td class="px-6 py-4 whitespace-nowrap">
									<input
										type="checkbox"
										checked={selectedEmails.has(entry.email)}
										on:change={() => toggleEmail(entry.email)}
										class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
									/>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{entry.email}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{entry.role}</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button
										on:click={() => deleteEmail(entry.email)}
										disabled={isDeleting}
										class="text-red-600 hover:text-red-900 disabled:text-gray-400 disabled:cursor-not-allowed"
									>
										削除
									</button>
								</td>
							</tr>
						{/each}
					{:else}
						<tr>
							<td colspan="4" class="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">
                  {#if data.error}
                    データの読み込み中にエラーが発生しました: {data.error}
                  {:else if whitelist && whitelist.length > 0}
                    <!-- フィルタリングの結果、一致するメールアドレスが見つかりませんでした。 -->
                    <p class="text-gray-600">フィルタリング条件に一致するエントリーは見つかりませんでした。</p>
                  {:else}
                    No whitelisted emails found.
                  {/if}
                </td>
						</tr>
					{/if}
				</tbody>
			</table>
		</div>
  </div>
</div>
