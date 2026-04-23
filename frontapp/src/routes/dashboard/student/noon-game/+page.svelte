<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';
  import { get } from 'svelte/store';

  let session = $state(null);
  let matches = $state([]);
  let pointsSummary = $state([]);
  let loading = $state(false);
  let errorMessage = $state('');
  let templateMatches = $derived(matches.filter((match) => detectTemplateFromMatch(match)));
  let sortedPointsSummary = $derived(orderByPointsDesc(pointsSummary));

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
      
      // デバッグ用: テンプレート結果のデータ構造を確認
      if (templateMatches.length > 0) {
        console.log('Template matches data:', templateMatches.map(m => ({
          id: m.id,
          title: m.title,
          entries: m.entries?.map(e => ({
            id: e.id,
            resolved_name: e.resolved_name,
            display_name: e.display_name
          })),
          result_details: m.result?.details?.map(d => ({
            entry_id: d.entry_id,
            entry_resolved_name: d.entry_resolved_name,
            resolved_name: d.resolved_name,
            display_name: d.display_name,
            rank: d.rank,
            points: d.points
          }))
        })));
      }
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

  function getEntryName(detail, entries) {
    // バックエンドから設定された entry_resolved_name を最優先
    if (detail.entry_resolved_name) {
      return detail.entry_resolved_name;
    }
    // entries から対応するエントリーを探す（型の不一致を考慮）
    if (detail.entry_id && entries && entries.length > 0) {
      const entryId = detail.entry_id;
      const entry = entries.find(e => {
        // 数値と文字列の両方に対応
        return e.id === entryId || 
               e.id === Number(entryId) || 
               String(e.id) === String(entryId);
      });
      if (entry) {
        if (entry.resolved_name) {
          return entry.resolved_name;
        }
        if (entry.display_name) {
          return entry.display_name;
        }
      }
      // デバッグ用: エントリーが見つからない場合
      if (!entry) {
        console.warn('Entry not found for detail:', {
          detail_entry_id: detail.entry_id,
          detail_entry_id_type: typeof detail.entry_id,
          detail_rank: detail.rank,
          available_entry_ids: entries.map(e => ({ id: e.id, id_type: typeof e.id, resolved_name: e.resolved_name, display_name: e.display_name }))
        });
      }
    }
    // フォールバック
    return detail.resolved_name || detail.display_name || '-';
  }

  function insertOrdered(items, item, compare) {
    let insertAt = items.length;
    for (let index = 0; index < items.length; index += 1) {
      if (compare(item, items[index]) < 0) {
        insertAt = index;
        break;
      }
    }
    return [
      ...items.slice(0, insertAt),
      item,
      ...items.slice(insertAt)
    ];
  }

  function orderByRank(details) {
    let ordered = [];
    for (const detail of details || []) {
      ordered = insertOrdered(
        ordered,
        detail,
        (left, right) => (left.rank || 999) - (right.rank || 999)
      );
    }
    return ordered;
  }

  function orderByPointsDesc(summary) {
    let ordered = [];
    for (const item of summary || []) {
      ordered = insertOrdered(
        ordered,
        item,
        (left, right) => (right.points || 0) - (left.points || 0)
      );
    }
    return ordered;
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
    {#if templateMatches.length > 0}
      <section class="bg-white shadow rounded-lg p-6 space-y-6">
        <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">テンプレート結果</h2>
        <div class="space-y-6">
          {#each templateMatches as match (match.id)}
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
                          {#each orderByRank(match.result.details) as detail (detail.class_id || detail.id)}
                            <div class="border rounded px-3 py-2 bg-gray-50">
                              <div class="flex justify-between items-center">
                                <span class="text-sm font-semibold text-gray-800">
                                  {detail.rank ? `${detail.rank}位` : '-'}
                                </span>
                                <span class="text-sm text-gray-700">
                                  {getEntryName(detail, match.entries)}
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
              {#each sortedPointsSummary as item (item.class_id || item.id)}
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
