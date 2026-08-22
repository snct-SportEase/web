<script>
  import FormField from '$lib/components/FormField.svelte';
  import Button from '$lib/components/Button.svelte';
  import { onMount } from 'svelte';

  let eligibleClasses = $state([]);
  let selectedClass = $state('');
  let reason = $state('');
  let eventId = $state(null);
  let isMicVotingEnabled = $state(false);
  let isLoading = $state(true);

  let hasVoted = $state(false);
  let votedForClassId = $state(null);
  let votedForClassName = $state('');

  onMount(async () => {
    try {
      const eventRes = await fetch('/api/events/active');
      if (!eventRes.ok) {
        alert('開催中のイベント情報の取得に失敗しました。');
        isLoading = false;
        return;
      }
      const eventData = await eventRes.json();
      if (!eventData.event_id) {
        alert('開催中のイベントがありません。');
        isLoading = false;
        return;
      }
      eventId = eventData.event_id;

      const classRes = await fetch(`/api/admin/mic/eligible-classes?event_id=${eventId}`);
      if (classRes.ok) {
        eligibleClasses = await classRes.json();
        isMicVotingEnabled = true;
      } else if (classRes.status === 403) {
        isMicVotingEnabled = false;
      } else {
        alert('投票対象クラスの取得に失敗しました。');
        isLoading = false;
        return;
      }

      if (isMicVotingEnabled) {
        const voteRes = await fetch(`/api/admin/mic/user-vote?event_id=${eventId}`);
        if (voteRes.ok) {
          const voteData = await voteRes.json();
          if (voteData.voted) {
            hasVoted = true;
            votedForClassId = voteData.vote.voted_for_class_id;
            const votedClass = eligibleClasses.find((c) => c.id === votedForClassId);
            if (votedClass) {
              votedForClassName = votedClass.name;
            }
          }
        } else if (voteRes.status !== 403) {
          alert('投票状態の取得に失敗しました。');
          isLoading = false;
          return;
        }
      }

      isLoading = false;
    } catch (error) {
      console.error('Error fetching MIC vote data:', error);
      alert('イベント情報の取得中にエラーが発生しました。');
      isLoading = false;
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

    if (!isMicVotingEnabled) {
      alert('現在、行事委員会賞投票は無効です。');
      return;
    }

    const res = await fetch('/api/admin/mic/vote', {
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
      const votedClass = eligibleClasses.find((c) => c.id === votedForClassId);
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

<h1 class="text-2xl font-bold mb-4">行事委員会賞投票</h1>

{#if isLoading}
  <p class="text-gray-600">読み込み中...</p>
{:else if !isMicVotingEnabled}
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 text-center">
    <h2 class="text-xl font-bold mb-2 text-yellow-700">現在無効</h2>
    <p>行事委員会賞投票は現在「無効」状態です。</p>
  </div>
{:else if hasVoted}
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 text-center">
    <h2 class="text-xl font-bold mb-2">投票済みです</h2>
    <p>あなたは <span class="font-bold">{votedForClassName}</span> に投票しました。</p>
    <p class="text-gray-600 mt-4">行事委員会賞投票は一人一票までです。</p>
  </div>
{:else}
  <form class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6 space-y-6" onsubmit={(e) => {
    e.preventDefault();
    vote(e);
  }}>
    <div class="text-sm text-gray-700 mb-2">
      投票状態: <span class="font-bold text-green-600">有効</span>
    </div>
    <FormField label="投票対象クラス" inputId="class-select" labelClass="mb-2 block font-bold text-gray-700">
      <select
        id="class-select"
        bind:value={selectedClass}
        required
        class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      >
        <option value="" disabled selected>クラスを選択してください</option>
        {#each eligibleClasses as c (c.id)}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
    </FormField>

    <FormField label="理由" inputId="reason" labelClass="mb-2 block font-bold text-gray-700">
      <textarea
        id="reason"
        bind:value={reason}
        required
        class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
        rows="4"
        placeholder="投票理由を入力してください"
      ></textarea>
    </FormField>

    <div class="flex items-center justify-between">
      <Button
        class="border-blue-500 bg-blue-500 font-bold hover:bg-blue-700 focus:ring-blue-500"
        type="submit"
      >
        投票する
      </Button>
    </div>
  </form>
{/if}
