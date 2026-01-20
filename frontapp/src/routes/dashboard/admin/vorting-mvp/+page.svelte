<script>
  import { onMount } from 'svelte';

  let eligibleClasses = [];
  let selectedClass = '';
  let reason = '';
  let eventId = null;
  let eventName = '';

  let hasVoted = false;
  let votedForClassId = null;
  let votedForClassName = '';

  onMount(async () => {
    // 1. Fetch active event first
    try {
      const eventRes = await fetch('/api/events/active');
      if (!eventRes.ok) {
        console.error('Failed to fetch active event');
        alert('開催中のイベント情報の取得に失敗しました。');
        return;
      }
      const eventData = await eventRes.json();
      if (!eventData.event_id) {
        console.error('No active event found');
        alert('開催中のイベントがありません。');
        return;
      }
      eventId = eventData.event_id;
      eventName = eventData.event_name;
    } catch (error) {
      console.error('Error fetching active event:', error);
      alert('イベント情報の取得中にエラーが発生しました。');
      return;
    }

    // 2. Fetch eligible classes using the active event ID
    const classRes = await fetch(`/api/admin/mvp/eligible-classes?event_id=${eventId}`);
    if (classRes.ok) {
      eligibleClasses = await classRes.json();
    } else {
      console.error('Failed to fetch eligible classes');
      // Handle error appropriately
    }

    // 3. Check if user has already voted for this event
    const voteRes = await fetch(`/api/admin/mvp/user-vote?event_id=${eventId}`);
    if (voteRes.ok) {
      const voteData = await voteRes.json();
      if (voteData.voted) {
        hasVoted = true;
        votedForClassId = voteData.vote.voted_for_class_id;
        const votedClass = eligibleClasses.find(c => c.id === votedForClassId);
        if (votedClass) {
          votedForClassName = votedClass.name;
        }
      }
    } else {
      console.error('Failed to fetch user vote status');
    }
  });

  async function vote(event) {
    event?.preventDefault?.();

    if (!selectedClass) {
      alert('投票するクラスを選択してください。');
      return;
    }

    if (!reason.trim()) {
      alert('投票理由を入力してください。');
      return;
    }

    const res = await fetch('/api/admin/mvp/vote', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        voted_for_class_id: parseInt(selectedClass),
        reason: reason,
        event_id: eventId,
      }),
    });

    if (res.ok) {
      alert('投票が完了しました。');
      hasVoted = true;
      votedForClassId = parseInt(selectedClass);
      const votedClass = eligibleClasses.find(c => c.id === votedForClassId);
      if (votedClass) {
        votedForClassName = votedClass.name;
      }
    } else {
      try {
        const error = await res.json();
        alert(`エラー: ${error.error}`);
      } catch {
        alert('不明なエラーが発生しました。');
      }
    }
  }
</script>

<h1 class="text-2xl font-bold mb-4">MVP投票</h1>

{#if hasVoted}
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 text-center">
    <h2 class="text-xl font-bold mb-2">投票済みです</h2>
    <p>あなたは <span class="font-bold">{votedForClassName}</span> に投票しました。</p>
    <p class="text-gray-600 mt-4">MVP投票は一人一票までです。</p>
  </div>
{:else}
  <form class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 space-y-6" on:submit|preventDefault={vote}>
    <div>
      <label for="class-select" class="block text-gray-700 font-bold mb-2">投票対象クラス</label>
      <select
        id="class-select"
        bind:value={selectedClass}
        required
        class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      >
        <option value="" disabled selected>クラスを選択してください</option>
        {#each eligibleClasses as c}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
    </div>

    <div>
      <label for="reason" class="block text-gray-700 font-bold mb-2">理由</label>
      <textarea
        id="reason"
        bind:value={reason}
        required
        class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
        rows="4"
        placeholder="投票理由を入力してください"
      ></textarea>
    </div>

    <div class="flex items-center justify-between">
      <button
        class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
        type="submit"
      >
        投票する
      </button>
    </div>
  </form>
{/if}
