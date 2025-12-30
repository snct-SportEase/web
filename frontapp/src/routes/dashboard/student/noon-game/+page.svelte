<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';
  import { get } from 'svelte/store';

  let session = null;
  let matches = [];
  let pointsSummary = [];
  let loading = false;
  let errorMessage = '';

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
      const res = await fetch(`/api/student/events/${eventId}/noon-game/session`);
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '昼競技データの取得に失敗しました');
      }
      const data = await res.json();
      session = data.session;
      matches = data.matches || [];
      pointsSummary = data.points_summary || [];
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
    } finally {
      loading = false;
    }
  }

  async function safeJson(res) {
    try {
      return await res.json();
    } catch {
      return null;
    }
  }

  function displayWinner(match) {
    if (!match.result) return '未登録';
    if (match.result.winner === 'home') {
      return match.home_display_name || 'ホーム';
    } else if (match.result.winner === 'away') {
      return match.away_display_name || 'アウェイ';
    } else if (match.result.winner === 'draw') {
      return '引き分け';
    }
    return '未登録';
  }

  function detectTemplateFromMatch(match) {
    const title = match.title || '';
    if (title.includes('学年対抗リレー')) {
      if (title.includes('Aブロック')) {
        return { type: 'year-relay', block: 'A' };
      } else if (title.includes('Bブロック')) {
        return { type: 'year-relay', block: 'B' };
      } else if (title.includes('総合ボーナス')) {
        return { type: 'year-relay', block: 'BONUS' };
      }
    } else if (title.includes('コース対抗リレー')) {
      return { type: 'course-relay' };
    } else if (title === '綱引き') {
      return { type: 'tug-of-war' };
    }
    return null;
  }

  function formatResultDetails(details) {
    if (!details || details.length === 0) return '未登録';
    const sorted = [...details].sort((a, b) => {
      const rankA = a.rank || 999;
      const rankB = b.rank || 999;
      return rankA - rankB;
    });
    return sorted.map(d => {
      const rank = d.rank ? `${d.rank}位` : '-';
      const points = d.points !== null && d.points !== undefined ? ` (${d.points}点)` : '';
      const name = d.resolved_name || d.display_name || '-';
      return `${rank}: ${name}${points}`;
    }).join(', ');
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-3xl font-bold text-gray-800 border-b pb-2">昼競技結果</h1>
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
    </div>
  {:else}
    <!-- テンプレートラン用の結果表示 -->
    {#if matches.filter(m => detectTemplateFromMatch(m)).length > 0}
      <section class="bg-white shadow rounded-lg p-6 space-y-6">
        <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">テンプレート結果</h2>
        <div class="space-y-6">
          {#each matches.filter(m => detectTemplateFromMatch(m)) as match}
            {@const template = detectTemplateFromMatch(match)}
            {#if template}
              <div class="border rounded-lg p-4 bg-blue-50 space-y-3">
                <h3 class="text-lg font-semibold text-gray-800">{match.title}</h3>
                {#if match.result}
                  <div class="bg-white rounded p-4 space-y-2">
                    {#if match.result.details && match.result.details.length > 0}
                      <div class="space-y-2">
                        <p class="text-sm font-semibold text-gray-700">順位・得点</p>
                        <div class="space-y-2">
                          {#each match.result.details.sort((a, b) => (a.rank || 999) - (b.rank || 999)) as detail}
                            <div class="border rounded px-3 py-2 bg-gray-50">
                              <div class="flex justify-between items-center">
                                <span class="text-sm font-semibold text-gray-800">
                                  {detail.rank ? `${detail.rank}位` : '-'}
                                </span>
                                <span class="text-sm text-gray-700">
                                  {detail.resolved_name || detail.display_name || '-'}
                                </span>
                                {#if detail.points !== null && detail.points !== undefined}
                                  <span class="text-sm font-semibold text-indigo-600">
                                    {detail.points}点
                                  </span>
                                {/if}
                              </div>
                            </div>
                          {/each}
                        </div>
                      </div>
                    {:else}
                      <p class="text-sm text-gray-500">結果がまだ登録されていません。</p>
                    {/if}
                    {#if match.result.note}
                      <div class="mt-2">
                        <p class="text-sm font-semibold text-gray-700">備考</p>
                        <p class="text-sm text-gray-600">{match.result.note}</p>
                      </div>
                    {/if}
                  </div>
                {:else}
                  <p class="text-sm text-gray-500">結果がまだ登録されていません。</p>
                {/if}
              </div>
            {/if}
          {/each}
        </div>
      </section>
    {/if}

    <!-- 通常の試合一覧 -->
    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">試合一覧</h2>
      {#if matches.filter(m => !detectTemplateFromMatch(m)).length === 0}
        <p class="text-gray-500">登録された試合がありません。</p>
      {:else}
        <div class="space-y-4">
          {#each matches.filter(m => !detectTemplateFromMatch(m)) as match}
            <div class="border rounded-lg p-4 bg-gray-50 space-y-3">
              <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
                <div>
                  <p class="text-lg font-semibold text-gray-800">{match.title ?? `試合 #${match.id}`}</p>
                  <p class="text-sm text-gray-600">{match.home_display_name} vs {match.away_display_name}</p>
                  {#if match.scheduled_at}
                    <p class="text-xs text-gray-500">予定日時: {new Date(match.scheduled_at).toLocaleString('ja-JP')}</p>
                  {/if}
                  <p class="text-xs text-gray-500">ステータス: {match.status}</p>
                </div>
              </div>
              {#if match.result}
                <div class="bg-white rounded p-4 space-y-2">
                  <div>
                    <p class="text-sm font-semibold text-gray-700">結果</p>
                    <p class="text-sm text-gray-800">{displayWinner(match)}</p>
                  </div>
                  {#if match.result.details && match.result.details.length > 0}
                    <div class="space-y-2">
                      <p class="text-sm font-semibold text-gray-700">順位・得点</p>
                      <div class="space-y-2">
                        {#each match.result.details.sort((a, b) => (a.rank || 999) - (b.rank || 999)) as detail}
                          <div class="border rounded px-3 py-2 bg-gray-50">
                            <div class="flex justify-between items-center">
                              <span class="text-sm font-semibold text-gray-800">
                                {detail.rank ? `${detail.rank}位` : '-'}
                              </span>
                              <span class="text-sm text-gray-700">
                                {detail.resolved_name || detail.display_name || '-'}
                              </span>
                              {#if detail.points !== null && detail.points !== undefined}
                                <span class="text-sm font-semibold text-indigo-600">
                                  {detail.points}点
                                </span>
                              {/if}
                            </div>
                          </div>
                        {/each}
                      </div>
                    </div>
                  {/if}
                  {#if match.result.note}
                    <div>
                      <p class="text-sm font-semibold text-gray-700">備考</p>
                      <p class="text-sm text-gray-600">{match.result.note}</p>
                    </div>
                  {/if}
                </div>
              {:else}
                <p class="text-sm text-gray-500">結果がまだ登録されていません。</p>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <!-- ポイント集計 -->
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
              {#each pointsSummary.sort((a, b) => b.points - a.points) as item}
                <tr class="hover:bg-gray-50">
                  <td class="px-4 py-2">{item.class_name}</td>
                  <td class="px-4 py-2 text-right font-semibold">{item.points}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  {/if}
</div>

