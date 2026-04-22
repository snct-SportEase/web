<script>
  import { page } from '$app/stores';
  import { enhance } from '$app/forms';
  import { markNotificationsSeen } from '$lib/stores/notificationBadgeStore.js';

  const roleLabels = {
    student: '学生',
    admin: '管理者',
    root: 'ルート'
  };

  const filterLabels = {
    general: '一般通知',
    match_my_class: '自分のクラスの試合',
    finals: '決勝戦',
    all_matches: '全ての試合'
  };

  let data = $derived($page.data);
  let form = $derived($page.form);
  let notifications = $derived(data.notifications ? [...data.notifications] : []);
  let initialSelectedFilters = $derived(data.user?.notification_filters || ['general']);
  let selectedFilters = $state([]);
  let selectedFiltersInitialized = $state(false);

  $effect(() => {
    if (selectedFiltersInitialized) return;
    selectedFilters = [...initialSelectedFilters];
    selectedFiltersInitialized = true;
  });

  $effect(() => {
    markNotificationsSeen(data.user, notifications);
  });

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

  function toggleFilter(filter) {
    if (filter === 'general') return; // general is always required
    if (selectedFilters.includes(filter)) {
      selectedFilters = selectedFilters.filter(f => f !== filter);
    } else {
      selectedFilters = [...selectedFilters, filter];
    }
  }
</script>

<div class="space-y-8">
  <header>
    <h1 class="text-3xl font-bold text-gray-900">通知一覧</h1>
    <p class="mt-2 text-sm text-gray-600">
      重要なお知らせを受け取るために、ブラウザの通知を許可してください。通知は最新100件まで表示されます。
    </p>
  </header>

  <section class="bg-white shadow rounded-lg p-6">
    <h2 class="text-xl font-semibold text-gray-900 mb-4">通知フィルタ設定</h2>
    <p class="text-sm text-gray-600 mb-4">
      受け取りたい通知の種類を選択してください。最低でも「一般通知」は選択されます。
    </p>

    <form method="POST" action="?/updateFilters" use:enhance>
      <div class="space-y-3">
        {#each Object.entries(filterLabels) as [key, label] (key)}
          <label class="flex items-center">
            <input
              type="checkbox"
              name="filters"
              value={key}
              checked={selectedFilters.includes(key)}
              disabled={key === 'general'}
              onchange={() => toggleFilter(key)}
              class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
            />
            <span class="ml-2 text-sm text-gray-700">{label}</span>
            {#if key === 'general'}
              <span class="ml-1 text-xs text-gray-500">(必須)</span>
            {/if}
          </label>
        {/each}
      </div>

      <div class="mt-4">
        <button
          type="submit"
          class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          設定を保存
        </button>
      </div>

      {#if form?.message}
        <p class="mt-2 text-sm text-green-600">{form.message}</p>
      {/if}
      {#if form?.error}
        <p class="mt-2 text-sm text-red-600">{form.error}</p>
      {/if}
    </form>
  </section>

  <section class="bg-white shadow rounded-lg p-6">
    {#if notifications.length === 0}
      <p class="text-gray-500">現在表示できる通知はありません。</p>
    {:else}
      <ul class="space-y-4">
        {#each notifications as notification (notification.id ?? `${notification.title}-${notification.created_at}`)}
          <li class="border border-gray-200 rounded-lg p-4">
            <div class="flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900">{notification.title}</h2>
              <span class="text-sm text-gray-500">{formatDate(notification.created_at)}</span>
            </div>
            <p class="mt-2 text-gray-700 whitespace-pre-wrap">{notification.body}</p>
            {#if notification.target_roles?.length}
              <div class="mt-3 flex flex-wrap gap-2">
                {#each notification.target_roles as role (role)}
                  <span class="inline-flex items-center rounded-full bg-indigo-50 px-3 py-1 text-xs font-medium text-indigo-700">
                    {roleLabels[role] ?? role}
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
