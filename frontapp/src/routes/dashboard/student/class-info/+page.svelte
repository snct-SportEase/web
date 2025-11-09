<script>
  import { page } from '$app/stores';

  let { data } = $page;
  const classInfo = data.classInfo;
  const className = data.className;
  const progressEntries = data.progress ?? [];

  const formatter = new Intl.NumberFormat('ja-JP');

  function formatCount(value) {
    return formatter.format(value ?? 0);
  }

  function calcAttendanceRate(item) {
    if (!item) return 0;
    const studentCount = item.student_count ?? 0;
    if (studentCount === 0) return 0;
    return ((item.attend_count ?? 0) / studentCount) * 100;
  }
</script>

<div class="space-y-8">
  <div class="space-y-2">
    <h1 class="text-3xl font-bold text-gray-900">クラス情報</h1>
    <p class="text-sm text-gray-600">
      自分のクラスの人数や出席状況を確認できます。クラス代表ロール（クラス名_rep）を持つユーザーのみ閲覧可能です。
    </p>
  </div>

  {#if !data.isClassRep}
    <div class="rounded-md border border-yellow-200 bg-yellow-50 px-4 py-3 text-sm text-yellow-800">
      クラス代表ロールが割り当てられていないため、クラス情報を表示できません。
    </div>
  {:else}
    {#if data.error}
      <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        データの取得に失敗しました: {data.error}
      </div>
    {:else}
      {#if classInfo}
        <section class="grid gap-4 sm:grid-cols-3">
          <div class="rounded-lg border border-indigo-100 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-indigo-600">クラス名</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">{classInfo.name ?? className}</p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-gray-500">登録学生数</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">
              {formatCount(classInfo.student_count)} <span class="text-sm font-normal text-gray-500">名</span>
            </p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-gray-500">出席数</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">
              {formatCount(classInfo.attend_count)} <span class="text-sm font-normal text-gray-500">名</span>
            </p>
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm space-y-6">
          <div>
            <h2 class="text-xl font-semibold text-gray-800">出席状況</h2>
            <p class="mt-1 text-sm text-gray-600">
              最新の集計に基づく、クラス全体の参加状況です。
            </p>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between text-sm text-gray-600">
              <span>出席率</span>
              <span class="font-semibold text-gray-900">{calcAttendanceRate(classInfo).toFixed(1)}%</span>
            </div>
            <div class="h-3 w-full overflow-hidden rounded-full bg-gray-100">
              <div
                class="h-full rounded-full bg-indigo-500 transition-all"
                style={`width: ${Math.min(calcAttendanceRate(classInfo), 100)}%;`}
                aria-hidden="true"
              ></div>
            </div>
          </div>

          <dl class="grid gap-4 sm:grid-cols-2">
            <div class="rounded-md border border-gray-100 bg-gray-50 p-4">
              <dt class="text-xs font-semibold uppercase tracking-wide text-gray-500">出席人数</dt>
              <dd class="mt-1 text-lg font-semibold text-gray-900">
                {formatCount(classInfo.attend_count)} 名
              </dd>
              <p class="mt-1 text-xs text-gray-500">
                出席登録済みの学生の数です。
              </p>
            </div>
            <div class="rounded-md border border-gray-100 bg-gray-50 p-4">
              <dt class="text-xs font-semibold uppercase tracking-wide text-gray-500">未出席人数</dt>
              <dd class="mt-1 text-lg font-semibold text-gray-900">
                {formatCount((classInfo.student_count ?? 0) - (classInfo.attend_count ?? 0))} 名
              </dd>
              <p class="mt-1 text-xs text-gray-500">
                未だ出席が確認されていない学生の数です。
              </p>
            </div>
          </dl>
        </section>
      {:else}
        <div class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
          クラス「{className}」の基本情報がまだ登録されていません。
        </div>
      {/if}

      <section class="space-y-4">
        <div>
          <h2 class="text-xl font-semibold text-gray-800">勝ち進み状況</h2>
          <p class="mt-1 text-sm text-gray-600">
            所属チームのトーナメント進行状況を確認できます。
          </p>
        </div>

        {#if progressEntries.length === 0}
          <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
            試合情報が登録されていません。
          </p>
        {:else}
          <div class="grid gap-4 lg:grid-cols-2">
            {#each progressEntries as item (item.team_name + item.sport_name)}
              <article class="rounded-lg border border-indigo-100 bg-white p-5 shadow-sm space-y-4">
                <header class="space-y-1">
                  <p class="text-sm font-medium text-indigo-600">{item.sport_name}</p>
                  <h3 class="text-lg font-semibold text-gray-900">{item.team_name}</h3>
                  <p class="text-xs text-gray-500">{item.tournament_name}</p>
                </header>

                <div class="rounded-md bg-indigo-50 px-4 py-3 text-sm text-indigo-800">
                  <div class="flex items-center justify-between">
                    <span class="font-semibold">{item.status}</span>
                    <span>{item.current_round}</span>
                  </div>
                </div>

                {#if item.next_match}
                  <div class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm space-y-1">
                    <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">次の試合</p>
                    <p class="text-gray-800">
                      {item.next_match.round_label}
                      {#if item.next_match.opponent_name}
                        ・対 {item.next_match.opponent_name}
                      {/if}
                    </p>
                    {#if item.next_match.start_time}
                      <p class="text-xs text-gray-500">開始予定: {item.next_match.start_time}</p>
                    {/if}
                    <p class="text-xs text-gray-500">ステータス: {item.next_match.match_status || '未定'}</p>
                  </div>
                {/if}

                {#if item.last_match}
                  <div class="rounded-md border border-gray-200 bg-white px-4 py-3 text-sm space-y-1">
                    <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">前の試合</p>
                    <p class="text-gray-800">
                      {item.last_match.round_label}
                      {#if item.last_match.opponent_name}
                        ・対 {item.last_match.opponent_name}
                      {/if}
                    </p>
                    <p class="text-xs text-gray-500">
                      結果: {item.last_match.result}
                      {#if item.last_match.score}
                        （{item.last_match.score}）
                      {/if}
                    </p>
                  </div>
                {/if}
              </article>
            {/each}
          </div>
        {/if}
      </section>
    {/if}
  {/if}
</div>