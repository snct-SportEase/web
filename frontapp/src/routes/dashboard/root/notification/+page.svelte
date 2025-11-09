<script>
  import { page } from '$app/stores';

  let { data } = $page;
  let notifications = data.notifications ? [...data.notifications] : [];

  const roleLabelMap = {
    student: '学生',
    admin: '管理者',
    root: 'ルート'
  };

  const availableRoles = (data.roles ?? []).map((role) => {
    const name = role.name ?? role.Name ?? '';
    const id = role.id ?? role.ID;
    return {
      id,
      name,
      label: roleLabelMap[name] ?? name
    };
  }).filter((role) => role.name);

  function createDefaultSelections(roles) {
    const defaults = {};
    let hasSelected = false;
    for (const role of roles) {
      const shouldSelect = role.name === 'student';
      defaults[role.name] = shouldSelect;
      if (shouldSelect) {
        hasSelected = true;
      }
    }
    if (!hasSelected && roles.length > 0) {
      defaults[roles[0].name] = true;
    }
    return defaults;
  }

  let title = '';
  let body = '';
  let selectedRoles = createDefaultSelections(availableRoles);

  let message = '';
  let errorMessage = '';
  let isSubmitting = false;
  let isReloading = false;

  function toggleRole(roleName) {
    selectedRoles = { ...selectedRoles, [roleName]: !selectedRoles[roleName] };
  }

  function resetSelectedRoles() {
    selectedRoles = createDefaultSelections(availableRoles);
  }

  function getSelectedRoles() {
    return availableRoles
      .map((role) => role.name)
      .filter((roleName) => selectedRoles[roleName]);
  }

  async function handleSubmit() {
    message = '';
    errorMessage = '';

    const targetRoles = getSelectedRoles();
    if (!title.trim() || !body.trim()) {
      errorMessage = 'タイトルと本文を入力してください。';
      return;
    }

    if (targetRoles.length === 0) {
      errorMessage = '少なくとも1つの宛先ロールを選択してください。';
      return;
    }

    isSubmitting = true;
    try {
      const response = await fetch('/api/root/notifications', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          title,
          body,
          target_roles: targetRoles
        })
      });

      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        throw new Error(err.error || '通知の作成に失敗しました。');
      }

      message = '通知を送信しました。';
      title = '';
      body = '';
      resetSelectedRoles();

      await refreshNotifications();
    } catch (error) {
      errorMessage = error.message;
    } finally {
      isSubmitting = false;
    }
  }

  async function refreshNotifications() {
    isReloading = true;
    try {
      const params = new URLSearchParams({
        include_authored: 'true',
        limit: '100'
      });
      const response = await fetch(`/api/notifications?${params.toString()}`);
      if (!response.ok) {
        const errText = await response.text();
        throw new Error(errText || '通知の再取得に失敗しました。');
      }
      const result = await response.json();
      notifications = result.notifications ? [...result.notifications] : [];
    } catch (error) {
      console.error('通知の再取得に失敗しました:', error);
    } finally {
      isReloading = false;
    }
  }

  function formatDate(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

<div class="space-y-8">
  <div class="flex items-center justify-between">
    <h1 class="text-3xl font-bold text-gray-900">通知管理</h1>
    <button
      type="button"
      class="inline-flex items-center px-4 py-2 text-sm font-medium text-indigo-600 border border-indigo-600 rounded-md hover:bg-indigo-50 disabled:opacity-50"
      on:click={refreshNotifications}
      disabled={isReloading}
    >
      {#if isReloading}
        再読込中...
      {:else}
        最新の状態に更新
      {/if}
    </button>
  </div>

  {#if message}
    <div class="rounded-md bg-green-50 p-4 text-green-800 border border-green-200">
      {message}
    </div>
  {/if}

  {#if errorMessage}
    <div class="rounded-md bg-red-50 p-4 text-red-800 border border-red-200">
      {errorMessage}
    </div>
  {/if}

  <section class="bg-white shadow rounded-lg p-6 space-y-6">
    <h2 class="text-xl font-semibold text-gray-800">新しい通知を作成</h2>

    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1" for="title">タイトル</label>
        <input
          id="title"
          type="text"
          class="block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          bind:value={title}
          placeholder="例）競技開始時刻変更のお知らせ"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1" for="body">本文</label>
        <textarea
          id="body"
          class="block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          rows="5"
          bind:value={body}
          placeholder="通知の内容を入力してください。"
        ></textarea>
      </div>

      <div>
        <span class="block text-sm font-medium text-gray-700 mb-2">宛先ロール</span>
        {#if availableRoles.length === 0}
          <p class="text-sm text-gray-500">選択可能なロールが登録されていません。</p>
        {:else}
          <div class="flex flex-wrap gap-4">
            {#each availableRoles as role (role.id ?? role.name)}
              <label class="inline-flex items-center space-x-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={!!selectedRoles[role.name]}
                  on:change={() => toggleRole(role.name)}
                />
                <span>{role.label}</span>
              </label>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <button
      type="button"
      class="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 disabled:opacity-50"
      on:click|preventDefault={handleSubmit}
      disabled={isSubmitting || availableRoles.length === 0}
    >
      {#if isSubmitting}
        送信中...
      {:else}
        通知を送信
      {/if}
    </button>
  </section>

  <section class="bg-white shadow rounded-lg p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl font-semibold text-gray-800">送信済み通知</h2>
      <p class="text-sm text-gray-500">最新100件まで表示</p>
    </div>

    {#if notifications.length === 0}
      <p class="text-gray-500">まだ通知はありません。</p>
    {:else}
      <ul class="space-y-4">
        {#each notifications as notification (notification.id ?? `${notification.title}-${notification.created_at}`)}
          <li class="border border-gray-200 rounded-lg p-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold text-gray-900">{notification.title}</h3>
              <span class="text-sm text-gray-500">{formatDate(notification.created_at)}</span>
            </div>
            <p class="mt-2 text-gray-700 whitespace-pre-wrap">{notification.body}</p>
            {#if notification.target_roles?.length}
              <div class="mt-3 flex flex-wrap gap-2">
                {#each notification.target_roles as role (role)}
                  <span class="inline-flex items-center rounded-full bg-indigo-50 px-3 py-1 text-xs font-medium text-indigo-700">
                    {roleLabelMap[role] ?? role}
                  </span>
                {/each}
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</div>