<script>
  import { page } from '$app/stores';
  import { goto, invalidateAll } from '$app/navigation';

  let { data } = $page;

  $: requests = data.requests ?? [];
  $: activeRequest = data.activeRequest;
  $: hasRequests = requests.length > 0;

  let messageInput = '';
  let isPostingMessage = false;
  let isDeciding = false;
  let errorMessage = data.error;

  function formatStatus(status) {
    switch (status) {
      case 'approved':
        return { label: '承認済み', class: 'bg-green-100 text-green-700' };
      case 'rejected':
        return { label: '否認', class: 'bg-red-100 text-red-700' };
      default:
        return { label: '審査中', class: 'bg-yellow-100 text-yellow-700' };
    }
  }

  function getDisplayName(user) {
    if (user?.display_name) {
      return user.display_name;
    }
    return user?.email ?? '不明なユーザー';
  }

  function selectRequest(id) {
    goto(`/dashboard/root/notification-requests?request_id=${id}`);
  }

  async function refreshDetail(id) {
    try {
      const res = await fetch(`/api/root/notification-requests/${id}`);
      if (res.ok) {
        const { request } = await res.json();
        activeRequest = request ?? null;
        await invalidateAll();
      }
    } catch (error) {
      console.error('Failed to refresh request detail:', error);
    }
  }

  async function handleMessageSubmit(event) {
    event.preventDefault();
    if (!activeRequest || isPostingMessage) return;
    const message = messageInput.trim();
    if (!message) return;

    isPostingMessage = true;
    errorMessage = null;
    try {
      const res = await fetch(`/api/root/notification-requests/${activeRequest.id}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ message })
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error ?? 'メッセージの送信に失敗しました');
      }
      messageInput = '';
      await refreshDetail(activeRequest.id);
    } catch (error) {
      console.error(error);
      errorMessage = error.message ?? 'メッセージの送信に失敗しました';
    } finally {
      isPostingMessage = false;
    }
  }

  async function decide(status) {
    if (!activeRequest || isDeciding) return;
    isDeciding = true;
    errorMessage = null;

    try {
      const res = await fetch(`/api/root/notification-requests/${activeRequest.id}/decision`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ status })
      });
      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error ?? '処理に失敗しました');
      }
      await refreshDetail(activeRequest.id);
      await invalidateAll();
    } catch (error) {
      console.error(error);
      errorMessage = error.message ?? '処理に失敗しました';
    } finally {
      isDeciding = false;
    }
  }
</script>

<svelte:head>
  <title>通知申請管理 | Dashboard</title>
</svelte:head>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">通知申請管理</h1>
    <p class="text-sm text-gray-600">
      学生からの通知作成依頼を確認し、チャットでやり取りして承認・否認を決定します。
    </p>
  </header>

  {#if errorMessage}
    <p class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{errorMessage}</p>
  {/if}

  <section class="grid gap-6 lg:grid-cols-12">
    <div class="space-y-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm lg:col-span-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-800">申請一覧</h2>
        <span class="rounded-full bg-indigo-100 px-3 py-1 text-xs font-semibold text-indigo-700">{requests.length} 件</span>
      </div>

      {#if !hasRequests}
        <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
          申請はまだありません。
        </p>
      {:else}
        <div class="space-y-3">
          {#each requests as item (item.id)}
            {@const status = formatStatus(item.status)}
            <button
              class="w-full rounded-md border px-4 py-3 text-left text-sm transition hover:border-indigo-300 hover:bg-indigo-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 {activeRequest?.id === item.id ? 'border-indigo-300 bg-indigo-50' : 'border-gray-200 bg-white'}"
              on:click={() => selectRequest(item.id)}
            >
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold text-gray-900">{item.title}</span>
                <span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${status.class}`}>{status.label}</span>
              </div>
              <p class="mt-1 line-clamp-2 text-xs text-gray-600">{item.body}</p>
              <p class="mt-1 text-xs text-gray-500">申請者: {getDisplayName(item.requester)}</p>
              <p class="mt-2 text-xs text-gray-500">対象: {item.target_text}</p>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <div class="space-y-4 rounded-lg border border-gray-200 bg-white p-5 shadow-sm lg:col-span-8">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-800">チャット / 審査</h2>
      </div>

      {#if !activeRequest}
        <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
          申請を選択して詳細を確認してください。
        </p>
      {:else}
        {@const status = formatStatus(activeRequest.status)}
        <div class="space-y-4">
          <div class="rounded-md border border-indigo-100 bg-indigo-50 px-4 py-3 text-sm text-indigo-900">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-base font-semibold text-gray-900">{activeRequest.title}</p>
                <p class="text-xs text-gray-600">対象: {activeRequest.target_text}</p>
                <p class="text-xs text-gray-600 mt-1">
                  申請者: {getDisplayName(activeRequest.requester)}
                </p>
              </div>
              <span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${status.class}`}>{status.label}</span>
            </div>
            <p class="mt-2 whitespace-pre-wrap text-sm text-gray-800">{activeRequest.body}</p>
          </div>

          <div class="space-y-3">
            {#if activeRequest.messages?.length === 0}
              <p class="text-sm text-gray-600">まだメッセージはありません。</p>
            {:else}
              <div class="space-y-4">
                {#each activeRequest.messages as message (message.id)}
                  <div class="rounded-md border border-gray-200 bg-white px-4 py-3 text-sm shadow-sm">
                    <div class="flex items-center justify-between">
                      <span class="font-semibold text-gray-800">{getDisplayName(message.sender)}</span>
                      <span class="text-xs text-gray-500">{new Date(message.created_at).toLocaleString()}</span>
                    </div>
                    <p class="mt-2 whitespace-pre-wrap text-sm text-gray-700">{message.message}</p>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <form class="space-y-3" on:submit|preventDefault={handleMessageSubmit}>
            <label class="block text-sm font-medium text-gray-700" for="rootMessageInput">メッセージを送信</label>
            <textarea
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              id="rootMessageInput"
              rows="3"
              bind:value={messageInput}
              placeholder="追加の確認事項や回答を入力してください"
            ></textarea>
            <div class="flex justify-end">
              <button
                type="submit"
                class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow hover:bg-indigo-700 disabled:bg-indigo-300"
                disabled={isPostingMessage}
              >
                {isPostingMessage ? '送信中...' : 'メッセージを送信'}
              </button>
            </div>
          </form>

          <div class="flex items-center justify-end space-x-3">
            <button
              class="rounded-md border border-gray-300 px-4 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-100 disabled:opacity-50"
              on:click={() => decide('rejected')}
              disabled={isDeciding || activeRequest.status !== 'pending'}
            >
              否認する
            </button>
            <button
              class="rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow hover:bg-indigo-700 disabled:bg-indigo-300"
              on:click={() => decide('approved')}
              disabled={isDeciding || activeRequest.status !== 'pending'}
            >
              承認する
            </button>
          </div>
        </div>
      {/if}
    </div>
  </section>
</div>

