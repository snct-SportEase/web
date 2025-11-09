<script>
  import { page } from '$app/stores';

  const roleLabels = {
    student: '学生',
    admin: '管理者',
    root: 'ルート'
  };

  let { data } = $page;
  let notifications = data.notifications ? [...data.notifications] : [];

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
  <header>
    <h1 class="text-3xl font-bold text-gray-900">通知一覧</h1>
    <p class="mt-2 text-sm text-gray-600">
      重要なお知らせを受け取るために、ブラウザの通知を許可してください。通知は最新100件まで表示されます。
    </p>
  </header>

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