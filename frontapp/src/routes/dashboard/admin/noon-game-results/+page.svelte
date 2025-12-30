<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';
  import { get } from 'svelte/store';

  let session = null;
  let matches = [];
  let pointsSummary = [];
  let loading = false;
  let saving = {};
  let errorMessage = '';

  let resultForms = {};
  let templateRunForms = {};
  let templateRuns = [];

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
      pointsSummary = data.points_summary || [];
      initializeResultForms(matches);
      await initializeTemplateRuns(matches);
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

  // 試合のタイトルからテンプレートランを特定
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

  // 試合IDからrun_idを取得（試合のタイトルから推測）
  async function getRunIdFromMatch(match) {
    const template = detectTemplateFromMatch(match);
    if (!template) return null;

    // 同じセッション内で同じテンプレートタイプの試合を探す
    const sameTemplateMatches = matches.filter(m => {
      const t = detectTemplateFromMatch(m);
      return t && t.type === template.type;
    });

    // 学年対抗リレーの場合、A/B/BONUSの3試合が同じrun_idを持つ
    if (template.type === 'year-relay') {
      // 最初に見つかった試合のrun_idを取得（実際にはAPIから取得する必要がある）
      // ここでは仮実装として、試合IDからrun_idを推測する
      return null; // 後で実装
    }

    // コース対抗リレーと綱引きの場合、1試合が1run_idを持つ
    return null; // 後で実装
  }

  async function initializeTemplateRuns(matchList) {
    // テンプレートランをグループ化
    const runs = new Map();
    
    // 学年対抗リレーの場合、A/B/BONUSの3試合を同じrun_idとしてグループ化
    const yearRelayMatches = [];
    const courseRelayMatches = [];
    const tugOfWarMatches = [];
    
    for (const match of matchList) {
      const template = detectTemplateFromMatch(match);
      if (!template) continue;

      if (template.type === 'year-relay') {
        // 総合ボーナスは自動計算されるため、入力UIから除外
        if (template.block !== 'BONUS') {
          yearRelayMatches.push({ match, template });
        }
      } else if (template.type === 'course-relay') {
        courseRelayMatches.push({ match, template });
      } else if (template.type === 'tug-of-war') {
        tugOfWarMatches.push({ match, template });
      }
    }

    // 学年対抗リレーをグループ化（A/B/BONUSで同じrun_id）
    if (yearRelayMatches.length > 0) {
      // run_idを取得（最初の試合から）
      const runId = await getRunIdFromMatchId(yearRelayMatches[0].match.id);
      const runKey = runId ? `year-relay-${runId}` : 'year-relay-1';
      runs.set(runKey, {
        key: runKey,
        type: 'year-relay',
        matches: yearRelayMatches,
        runId: runId
      });
    }

    // コース対抗リレーと綱引きは各試合が1つのrun
    for (const { match, template } of courseRelayMatches) {
      const runId = await getRunIdFromMatchId(match.id);
      const runKey = runId ? `course-relay-${runId}` : `course-relay-${match.id}`;
      runs.set(runKey, {
        key: runKey,
        type: 'course-relay',
        matches: [{ match, template }],
        runId: runId
      });
    }

    for (const { match, template } of tugOfWarMatches) {
      const runId = await getRunIdFromMatchId(match.id);
      const runKey = runId ? `tug-of-war-${runId}` : `tug-of-war-${match.id}`;
      runs.set(runKey, {
        key: runKey,
        type: 'tug-of-war',
        matches: [{ match, template }],
        runId: runId
      });
    }

    templateRuns = Array.from(runs.values());

    // フォームを初期化
    const forms = {};
    for (const run of templateRuns) {
      for (const { match, template } of run.matches) {
        const formKey = `${run.key}-${match.id}`;
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
              return {
                entry_id: entry.id,
                name:
                  entry.resolved_name ||
                  entry.display_name ||
                  `参加者 ${index + 1}`,
                rank: rankValue,
                points: null // 同順位の場合のみ
              };
            })
          : [];

        forms[formKey] = {
          matchId: match.id,
          template: template,
          participants,
          note: match.result?.note ?? ''
        };
      }
    }
    templateRunForms = forms;
  }

  async function submitTemplateResult(run, match, template) {
    const formKey = `${run.key}-${match.id}`;
    const form = templateRunForms[formKey];
    if (!form) {
      alert('フォーム情報がありません。');
      return;
    }

    // run_idを取得（試合IDから）
    const runId = run.runId || await getRunIdFromMatchId(match.id);
    if (!runId) {
      alert('テンプレートランIDを取得できませんでした。');
      return;
    }

    saving = { ...saving, [formKey]: true };
    errorMessage = '';

    try {
      const rankings = form.participants
        .map((participant) => {
          const rankValue = participant.rank === '' ? null : Number(participant.rank);
          if (rankValue === null) return null;
          return {
            entry_id: participant.entry_id,
            rank: rankValue,
            points: participant.points !== null ? participant.points : null
          };
        })
        .filter((row) => row !== null);

      if (rankings.length === 0) {
        throw new Error('順位を入力してください。');
      }

      // 同順位チェック
      const rankCounts = {};
      for (const r of rankings) {
        rankCounts[r.rank] = (rankCounts[r.rank] || 0) + 1;
      }

      for (const r of rankings) {
        if (rankCounts[r.rank] > 1 && r.points === null) {
          throw new Error(`同順位${r.rank}位には点数を入力してください。`);
        }
      }

      let endpoint = '';
      let payload = {
        rankings,
        note: form.note || null
      };

      if (template.type === 'year-relay') {
        // 総合ボーナスは自動計算されるため、ここには来ない
        endpoint = `/api/admin/noon-game/template-runs/${runId}/year-relay/blocks/${template.block}/result`;
      } else if (template.type === 'course-relay') {
        endpoint = `/api/admin/noon-game/template-runs/${runId}/course-relay/result`;
      } else if (template.type === 'tug-of-war') {
        endpoint = `/api/admin/noon-game/template-runs/${runId}/tug-of-war/result`;
      }

      const res = await fetch(endpoint, {
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
      saving = { ...saving, [formKey]: false };
    }
  }

  // 試合IDからrun_idを取得
  async function getRunIdFromMatchId(matchId) {
    try {
      const res = await fetch(`/api/admin/noon-game/matches/${matchId}/template-run`);
      if (!res.ok) {
        if (res.status === 404) {
          return null; // テンプレートランに紐づいていない
        }
        const detail = await safeJson(res);
        throw new Error(detail?.error || 'テンプレートラン情報の取得に失敗しました');
      }
      const data = await res.json();
      return data.run?.id || null;
    } catch (err) {
      console.error(err);
      return null;
    }
  }

  function updateTemplateParticipantField(formKey, index, field, value) {
    const form = templateRunForms[formKey];
    if (!form || !Array.isArray(form.participants)) return;
    const participants = [...form.participants];
    participants[index] = {
      ...participants[index],
      [field]: value === '' ? null : (field === 'rank' ? Number(value) : Number(value))
    };
    templateRunForms = {
      ...templateRunForms,
      [formKey]: {
        ...form,
        participants
      }
    };
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
    <!-- テンプレートラン用の結果入力 -->
    {#if templateRuns.length > 0}
      <section class="bg-white shadow rounded-lg p-6 space-y-6">
        <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">テンプレート結果入力</h2>
        <div class="space-y-6">
          {#each templateRuns as run}
            <div class="border rounded-lg p-4 bg-blue-50 space-y-4">
              <h3 class="text-lg font-semibold text-gray-800">
                {#if run.type === 'year-relay'}
                  学年対抗リレー
                {:else if run.type === 'course-relay'}
                  コース対抗リレー
                {:else if run.type === 'tug-of-war'}
                  綱引き
                {/if}
              </h3>
              {#each run.matches as { match, template }}
                {@const formKey = `${run.key}-${match.id}`}
                {@const form = templateRunForms[formKey]}
                {#if form}
                  <div class="border rounded p-4 bg-white space-y-3">
                    <h4 class="font-semibold text-gray-700">{match.title}</h4>
                    <div class="space-y-3">
                      <p class="text-sm font-semibold text-gray-700">順位入力</p>
                      <div class="space-y-2">
                        {#each form.participants as participant, index}
                          <div class="border rounded px-3 py-2 space-y-2 bg-gray-50">
                            <p class="text-sm font-semibold text-gray-800">{participant.name}</p>
                            <div class="grid grid-cols-2 gap-2">
                              <div class="space-y-1">
                                <label class="text-xs font-semibold text-gray-600">順位</label>
                                <input
                                  type="number"
                                  min="1"
                                  class="border rounded px-2 py-1 text-sm w-full"
                                  value={participant.rank}
                                  on:input={(e) => updateTemplateParticipantField(formKey, index, 'rank', e.target.value)}
                                />
                              </div>
                              <div class="space-y-1">
                                <label class="text-xs font-semibold text-gray-600">点数（同順位の場合のみ必須）</label>
                                <input
                                  type="number"
                                  class="border rounded px-2 py-1 text-sm w-full"
                                  value={participant.points ?? ''}
                                  on:input={(e) => updateTemplateParticipantField(formKey, index, 'points', e.target.value)}
                                  placeholder="同順位時のみ"
                                />
                              </div>
                            </div>
                          </div>
                        {/each}
                      </div>
                      <div class="space-y-2">
                        <label class="text-sm font-semibold text-gray-700">備考</label>
                        <textarea
                          rows="2"
                          class="border rounded px-3 py-2 text-sm w-full"
                          bind:value={form.note}
                        ></textarea>
                      </div>
                      <div class="flex justify-end">
                        <button
                          class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                          on:click={() => submitTemplateResult(run, match, template)}
                          disabled={saving[formKey]}>
                          {saving[formKey] ? '送信中...' : '結果を登録'}
                        </button>
                      </div>
                    </div>
                  </div>
                {/if}
              {/each}
            </div>
          {/each}
        </div>
      </section>
    {/if}

    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">試合一覧</h2>
      {#if matches.length === 0}
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

