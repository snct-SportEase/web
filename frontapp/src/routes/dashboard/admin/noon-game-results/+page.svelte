<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';
  import { get } from 'svelte/store';

  let session = null;
  let matches = [];
  let groups = [];
  let classes = [];
  let pointsSummary = [];
  let loading = false;
  let saving = {};
  let errorMessage = '';

  let resultForms = {};

  onMount(async () => {
    await activeEvent.init();
    const current = get(activeEvent);
    if (current) {
      await fetchSession(current.id);
    }
  });

  async function fetchSession(eventId) {
    loading = true;
    errorMessage = '';
    try {
      const res = await fetch(`/api/admin/events/${eventId}/noon-game/session`);
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '昼競技データの取得に失敗しました');
      }
      const data = await res.json();
      session = data.session;
      matches = data.matches || [];
      groups = data.groups || [];
      classes = data.classes || [];
      pointsSummary = data.points_summary || [];
      initializeResultForms(matches);
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
    } finally {
      loading = false;
    }
  }

  function initializeResultForms(matchList) {
    const forms = {};
    for (const match of matchList) {
      const hasEntries = Array.isArray(match.entries) && match.entries.length > 0;
      const detailMap = new Map();
      if (match.result?.details) {
        for (const detail of match.result.details) {
          detailMap.set(detail.entry_id, detail);
        }
      }
      const participants = hasEntries
        ? match.entries.map((entry, index) => {
            const detail = detailMap.get(entry.id);
            const rankValue =
              detail && detail.rank !== undefined && detail.rank !== null ? String(detail.rank) : '';
            const pointsValue =
              detail && detail.points !== undefined && detail.points !== null
                ? String(detail.points)
                : '';
            return {
              entry_id: entry.id,
              name:
                entry.resolved_name ||
                entry.display_name ||
                entry.home_display_name ||
                entry.away_display_name ||
                `参加者 ${index + 1}`,
              rank: rankValue,
              points: pointsValue
            };
          })
        : [];

      forms[match.id] = {
        useRankings: hasEntries,
        participants,
        winner: match.result?.winner ?? '',
        note: match.result?.note ?? ''
      };
    }
    resultForms = forms;
  }

  async function submitResult(match) {
    const form = resultForms[match.id];
    if (!form) {
      alert('フォーム情報がありません。');
      return;
    }
    if (!form.useRankings && !form.winner) {
      alert('勝者を選択してください。');
      return;
    }
    saving = { ...saving, [match.id]: true };
    errorMessage = '';
    try {
      let rankings = [];
      if (form.useRankings) {
        rankings = form.participants
          .map((participant) => {
            const rankValue = participant.rank === '' ? null : Number(participant.rank);
            const pointsValue =
              participant.points === '' || participant.points === null
                ? 0
                : Number(participant.points);
            return {
              entry_id: participant.entry_id,
              rank: rankValue,
              points: Number.isNaN(pointsValue) ? 0 : pointsValue
            };
          })
          .filter((row) => row.rank !== null || row.points !== 0);

        if (rankings.length === 0) {
          throw new Error('順位・得点を入力してください。');
        }
      }

      const payload = {
        winner: form.useRankings ? '' : form.winner,
        note: form.note || null,
        rankings
      };

      const res = await fetch(`/api/admin/noon-game/matches/${match.id}/result`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '試合結果の登録に失敗しました');
      }
      const current = get(activeEvent);
      if (current) {
        await fetchSession(current.id);
      }
      alert('試合結果を登録しました。');
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
      alert(err.message);
    } finally {
      saving = { ...saving, [match.id]: false };
    }
  }

  function displayWinner(match) {
    if (!match.result) return '未登録';
    if (match.result.winner === 'home') return match.home_display_name;
    if (match.result.winner === 'away') return match.away_display_name;
    if (match.result.winner === 'draw') return '引き分け';
    return '未登録';
  }

  function updateResultForm(matchId, patch) {
    const current = resultForms[matchId] ?? {};
    resultForms = {
      ...resultForms,
      [matchId]: {
        ...current,
        ...patch
      }
    };
  }

  function updateResultParticipantField(matchId, index, field, value) {
    const form = resultForms[matchId];
    if (!form || !Array.isArray(form.participants)) return;
    const participants = [...form.participants];
    participants[index] = {
      ...participants[index],
      [field]: value
    };
    updateResultForm(matchId, { participants });
  }

  function toggleResultUseRankings(matchId, value) {
    updateResultForm(matchId, { useRankings: value });
  }

  async function safeJson(response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-3xl font-bold text-gray-800 border-b pb-2">昼競技結果入力</h1>
  {#if errorMessage}
    <div class="bg-red-100 border-l-4 border-red-500 text-red-700 p-4">
      <p class="font-semibold">エラー</p>
      <p>{errorMessage}</p>
    </div>
  {/if}

  {#if loading}
    <div class="text-gray-600">読み込み中...</div>
  {:else if !session}
    <div class="bg-yellow-100 border-l-4 border-yellow-400 text-yellow-700 p-4">
      <p class="font-semibold">昼競技セッションが設定されていません。</p>
      <p>root 権限ユーザーにてセッションの作成を依頼してください。</p>
    </div>
  {:else}
    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">試合一覧</h2>
      {#if matches.length === 0}
        <p class="text-gray-500">登録された試合がありません。</p>
      {:else}
        <div class="space-y-4">
          {#each matches as match}
            <div class="border rounded-lg p-4 bg-gray-50 space-y-3">
              <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
                <div>
                  <p class="text-lg font-semibold text-gray-800">{match.title ?? `試合 #${match.id}`}</p>
                  <p class="text-sm text-gray-600">{match.home_display_name} vs {match.away_display_name}</p>
                  {#if match.scheduled_at}
                    <p class="text-xs text-gray-500">予定日時: {new Date(match.scheduled_at).toLocaleString()}</p>
                  {/if}
                  <p class="text-xs text-gray-500">ステータス: {match.status}</p>
                  <p class="text-xs text-gray-500">登録済み結果: {displayWinner(match)}</p>
                </div>
                <div class="text-sm text-gray-500">
                  {#if match.allow_draw}
                    引き分け入力可能
                  {:else}
                    引き分け不可
                  {/if}
                </div>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-5 gap-3">
                <div class="space-y-2 md:col-span-3">
                  <label class="flex items-center space-x-2 text-sm font-semibold text-gray-700">
                    <input
                      type="checkbox"
                      checked={resultForms[match.id]?.useRankings}
                      on:change={(e) => toggleResultUseRankings(match.id, e.target.checked)}
                    />
                    <span>参加者ごとの手動得点入力</span>
                  </label>
                  {#if resultForms[match.id]?.useRankings}
                    <p class="text-sm font-semibold text-gray-700">順位・得点</p>
                    <div class="space-y-3">
                      {#each resultForms[match.id].participants as participant, index}
                        <div class="border rounded px-3 py-3 space-y-2 bg-white">
                          <p class="text-sm font-semibold text-gray-800">{participant.name}</p>
                          <div class="grid grid-cols-2 gap-2">
                            <div class="space-y-1">
                              <label for={`match-${match.id}-participant-${index}-rank`} class="text-xs font-semibold text-gray-600">順位</label>
                              <input
                                id={`match-${match.id}-participant-${index}-rank`}
                                type="number"
                                min="1"
                                class="border rounded px-2 py-1 text-sm w-full"
                                value={participant.rank}
                                on:input={(e) => updateResultParticipantField(match.id, index, 'rank', e.target.value)}
                              />
                            </div>
                            <div class="space-y-1">
                              <label for={`match-${match.id}-participant-${index}-points`} class="text-xs font-semibold text-gray-600">得点</label>
                              <input
                                id={`match-${match.id}-participant-${index}-points`}
                                type="number"
                                class="border rounded px-2 py-1 text-sm w-full"
                                value={participant.points}
                                on:input={(e) => updateResultParticipantField(match.id, index, 'points', e.target.value)}
                              />
                            </div>
                          </div>
                        </div>
                      {/each}
                    </div>
                  {:else}
                    <p class="text-sm font-semibold text-gray-700">勝者</p>
                    <label class="flex items-center space-x-2 text-sm text-gray-700">
                      <input type="radio" name={`winner-${match.id}`} value="home" bind:group={resultForms[match.id].winner} />
                      <span>{match.home_display_name}</span>
                    </label>
                    <label class="flex items-center space-x-2 text-sm text-gray-700">
                      <input type="radio" name={`winner-${match.id}`} value="away" bind:group={resultForms[match.id].winner} />
                      <span>{match.away_display_name}</span>
                    </label>
                    {#if match.allow_draw}
                      <label class="flex items-center space-x-2 text-sm text-gray-700">
                        <input type="radio" name={`winner-${match.id}`} value="draw" bind:group={resultForms[match.id].winner} />
                        <span>引き分け</span>
                      </label>
                    {/if}
                  {/if}
                </div>
                <div class="space-y-2 md:col-span-2">
                  <p class="text-sm font-semibold text-gray-700">備考</p>
                  <textarea rows="3" class="border rounded px-3 py-2 text-sm"
                    bind:value={resultForms[match.id].note}></textarea>
                </div>
              </div>
              <div class="flex justify-end">
                <button class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                  on:click={() => submitResult(match)}
                  disabled={saving[match.id]}>
                  {saving[match.id] ? '送信中...' : '結果を登録'}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">昼競技ポイント集計</h2>
      {#if pointsSummary.length === 0}
        <p class="text-gray-500">まだ集計結果がありません。</p>
      {:else}
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left font-semibold text-gray-600">クラス</th>
                <th class="px-4 py-2 text-right font-semibold text-gray-600">昼競技ポイント</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              {#each pointsSummary as item}
                <tr>
                  <td class="px-4 py-2">{item.class_name}</td>
                  <td class="px-4 py-2 text-right">{item.points}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</div>

