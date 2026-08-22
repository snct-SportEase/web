<script>
  import { onMount } from 'svelte';

  let eventId = $state(null);
  let isMicVotingEnabled = $state(false);
  let micResult = $state(null);
  let error = $state(null);
  let isLoading = $state(true);
  let isUpdating = $state(false);

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    isLoading = true;
    error = null;
    micResult = null;

    try {
      const eventRes = await fetch('/api/events/active');
      if (!eventRes.ok) {
        throw new Error('開催中のイベント情報の取得に失敗しました。');
      }
      const eventData = await eventRes.json();
      if (!eventData?.event_id) {
        throw new Error('開催中のイベントがありません。');
      }
      eventId = eventData.event_id;

      const settingsRes = await fetch(`/api/root/events/${eventId}/mic/settings`);
      if (!settingsRes.ok) {
        throw new Error('行事委員会賞投票の設定取得に失敗しました。');
      }
      const settingsData = await settingsRes.json();
      isMicVotingEnabled = Boolean(settingsData.is_mic_voting_enabled);

      const resultRes = await fetch(`/api/root/mic/class?event_id=${eventId}`);
      if (!resultRes.ok) {
        throw new Error('行事委員会賞の結果取得に失敗しました。');
      }
      const resultData = await resultRes.json();
      if (resultData?.message) {
        micResult = null;
      } else {
        micResult = resultData;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : '予期せぬエラーが発生しました。';
    } finally {
      isLoading = false;
    }
  }

  async function toggleMicVoting() {
    if (!eventId || isUpdating) return;

    const next = !isMicVotingEnabled;
    isUpdating = true;
    try {
      const res = await fetch(`/api/root/events/${eventId}/mic/settings`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ is_mic_voting_enabled: next }),
      });

      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload?.error ?? '設定の更新に失敗しました。');
      }

      const data = await res.json();
      isMicVotingEnabled = Boolean(data.is_mic_voting_enabled);
    } catch (err) {
      error = err instanceof Error ? err.message : '予期せぬエラーが発生しました。';
      await loadData();
    } finally {
      isUpdating = false;
    }
  }
</script>

<h1 class="text-2xl font-bold mb-4">行事委員会賞確認</h1>

{#if isLoading}
  <p class="text-gray-600">読み込み中...</p>
{:else if error}
  <p class="text-red-500">{error}</p>
{:else}
  <div class="space-y-4">
    <div class="bg-white rounded-lg shadow p-6 space-y-4">
      <div class="flex items-center justify-between gap-4">
        <p>
          現在の投票状態:
          <span class={isMicVotingEnabled ? 'text-green-700 font-semibold' : 'text-red-700 font-semibold'}>
            {isMicVotingEnabled ? '有効' : '無効'}
          </span>
        </p>
        <button
          class="px-4 py-2 rounded-md text-white font-semibold bg-blue-600 hover:bg-blue-700 disabled:opacity-50"
          onclick={toggleMicVoting}
          disabled={isUpdating}
        >
          {#if isUpdating}
            更新中...
          {:else}
            {isMicVotingEnabled ? '無効化する' : '有効化する'}
          {/if}
        </button>
      </div>
      <p class="text-gray-600 text-sm">イベントID: {eventId}</p>
    </div>

    {#if micResult}
      <div class="bg-white shadow-md rounded-lg p-6">
        <h2 class="text-xl font-semibold mb-2">行事委員会賞クラス</h2>
        <p><strong>クラス:</strong> {micResult.class_name}</p>
        <p><strong>得票数:</strong> {micResult.vote_count}</p>
        <p><strong>合計ポイント:</strong> {micResult.total_points}</p>
        <p><strong>シーズン:</strong> {micResult.season}</p>
      </div>
    {:else}
      <div class="bg-white rounded-lg shadow p-6">
        <p>まだ行事委員会賞の開票結果がありません。</p>
      </div>
    {/if}
  </div>
{/if}
