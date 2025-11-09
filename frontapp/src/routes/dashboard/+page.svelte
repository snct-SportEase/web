<script>
  import { page } from '$app/stores';
  import ProfileSetupModal from '$lib/components/ProfileSetupModal.svelte';
  import EventSetupModal from '$lib/components/EventSetupModal.svelte';

  let { data } = $page;
  $: user = data.user;
  $: classes = data.classes;
  $: events = data.events;
  $: form = data.form;

  $: isRoot = user?.roles?.some(role => role.name === 'root');
  $: isAdmin = user?.roles?.some(role => role.name === 'admin' || role.name === 'root');
  $: isStudent = user?.roles?.some(role => role.name === 'student' || role.name === 'admin' || role.name === 'root');
  $: showEventSetup = isRoot && user?.is_init_root_first_login && events?.length === 0;

  const rootShortcuts = [
    { title: '通知管理', description: '通知作成と配信先の管理', href: '/dashboard/root/notification' },
    { title: '通知申請管理', description: '学生からの申請を確認・承認', href: '/dashboard/root/notification-requests' },
    { title: '大会設定', description: '大会情報と開催設定', href: '/dashboard/root/event-management' },
    { title: '競技管理', description: '競技情報とトーナメント生成', href: '/dashboard/root/sport-management' },
    { title: 'ホワイトリスト', description: 'ログイン許可メールの登録', href: '/dashboard/root/whitelist-management' },
    { title: 'ユーザー管理', description: 'ユーザー名やクラス人数を調整', href: '/dashboard/root/change-username' }
  ];

  const adminShortcuts = [
    { title: 'クラス・チーム割り当て', description: 'メンバー配置と確認', href: '/dashboard/admin/class-management' },
    { title: 'ロール管理', description: 'ユーザーロールの付与・削除', href: '/dashboard/admin/role-management' },
    { title: '出席登録', description: '参加者の出席を記録', href: '/dashboard/admin/attendance-management' },
    { title: '試合結果入力', description: 'スコアと勝敗を登録', href: '/dashboard/admin/insert-matche-result' },
    { title: 'QRコードツール', description: '大会QRの発行と読み取り', href: '/dashboard/admin/qr-code-reader' },
    { title: '競技詳細登録', description: '競技ルールや資料の管理', href: '/dashboard/admin/sport-details-registration' }
  ];

  const studentShortcuts = [
    { title: 'マイページ', description: '自分の出場情報を確認', href: '/dashboard/student/my-page' },
    { title: 'クラス情報', description: 'クラスの出席と試合状況', href: '/dashboard/student/class-info' },
    { title: '通知一覧', description: '配信された通知を確認', href: '/dashboard/student/notification' },
    { title: '通知申請', description: 'rootへの通知依頼を送信', href: '/dashboard/student/notification-request' },
    { title: '競技詳細', description: '競技のルール・日程を確認', href: '/dashboard/student/sport-info' },
    { title: 'QRコード発行', description: '参加証QRコードを生成', href: '/dashboard/student/issueqr-code' }
  ];

  let shortcutSections = [];
  $: {
    shortcutSections = [];
    if (isRoot) {
      shortcutSections.push({ title: 'Rootメニュー', shortcuts: rootShortcuts });
    }
    if (isAdmin) {
      shortcutSections.push({ title: 'Adminメニュー', shortcuts: adminShortcuts });
    }
    if (isStudent) {
      shortcutSections.push({ title: 'Studentメニュー', shortcuts: studentShortcuts });
    }
  }

  const classMembers = data.members ?? [];
  const progressEntries = data.progress ?? [];
  const classInfo = data.classInfo;

  function memberDisplayName(member) {
    if (member?.display_name) return member.display_name;
    return member?.email ?? '（名称未設定）';
  }

  function formatAssignments(assignments = []) {
    if (!assignments || assignments.length === 0) {
      return '未割り当て';
    }
    return assignments
      .map((assignment) => `${assignment.sport_name}（${assignment.team_name}）`)
      .join(' / ');
  }
</script>

<div class="space-y-12">
  <section class="space-y-2">
    <h1 class="text-3xl font-bold text-gray-900">
      ようこそ、{user?.display_name || user?.email || 'User'} さん
    </h1>
    <p class="text-sm text-gray-600">
      権限に合わせたメニューとクラス情報をここで確認できます。
    </p>
  </section>

  {#if user && !user.is_profile_complete}
    <ProfileSetupModal classes={classes} form={form} />
  {/if}

  {#if showEventSetup}
    <EventSetupModal />
  {/if}

  <section class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-semibold text-gray-900">ショートカット</h2>
    </div>

    {#if shortcutSections.length === 0}
      <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
        現在アクセス可能なショートカットはありません。
      </p>
    {:else}
      <div class="space-y-6">
        {#each shortcutSections as section (section.title)}
          <div class="space-y-3">
            <h3 class="text-lg font-semibold text-gray-800">{section.title}</h3>
            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {#each section.shortcuts as shortcut (shortcut.href)}
                <a href={shortcut.href} class="group block rounded-lg border border-indigo-100 bg-white p-5 shadow-sm transition hover:border-indigo-300 hover:shadow">
                  <h4 class="text-base font-semibold text-indigo-700 group-hover:text-indigo-800">
                    {shortcut.title}
                  </h4>
                  <p class="mt-1 text-sm text-gray-600">{shortcut.description}</p>
                  <span class="mt-3 inline-flex items-center text-sm font-medium text-indigo-600 group-hover:text-indigo-700">
                    詳細を見る →
                  </span>
                </a>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  {#if data.isClassRep}
    <section class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-2xl font-semibold text-gray-900">クラス概要</h2>
        <a href="/dashboard/student/class-info" class="text-sm font-medium text-indigo-600 hover:text-indigo-700">
          詳細ページへ →
        </a>
      </div>

      {#if classInfo}
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="rounded-lg border border-indigo-100 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-indigo-600">クラス名</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">{classInfo.name}</p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-gray-500">登録学生数</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">{classInfo.student_count} 名</p>
          </div>
          <div class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
            <p class="text-sm font-medium text-gray-500">出席数</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900">{classInfo.attend_count} 名</p>
          </div>
        </div>
      {:else}
        <p class="rounded-md border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-600">
          クラスの基本情報がまだ登録されていません。
        </p>
      {/if}

      <div class="grid gap-6 lg:grid-cols-2">
        <div class="space-y-3 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h3 class="text-lg font-semibold text-gray-800">メンバー一覧</h3>
          {#if classMembers.length === 0}
            <p class="text-sm text-gray-600">クラスメンバーがまだ登録されていません。</p>
          {:else}
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 text-sm">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-4 py-3 text-left font-semibold text-gray-600">氏名</th>
                    <th class="px-4 py-3 text-left font-semibold text-gray-600">メール</th>
                    <th class="px-4 py-3 text-left font-semibold text-gray-600">担当競技</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 bg-white">
                  {#each classMembers as member (member.id)}
                    <tr class="hover:bg-gray-50">
                      <td class="px-4 py-3 font-medium text-gray-900">{memberDisplayName(member)}</td>
                      <td class="px-4 py-3 text-gray-700">{member.email}</td>
                      <td class="px-4 py-3 text-gray-700">{formatAssignments(member.assignments)}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>

        <div class="space-y-3 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h3 class="text-lg font-semibold text-gray-800">勝ち進み状況</h3>

          {#if progressEntries.length === 0}
            <p class="text-sm text-gray-600">現在進行中の競技はありません。</p>
          {:else}
            <div class="space-y-4">
              {#each progressEntries as item (item.team_name + item.sport_name)}
                <article class="rounded-md border border-indigo-100 bg-indigo-50 px-4 py-3 text-sm text-indigo-900">
                  <header class="mb-2">
                    <p class="text-xs font-semibold uppercase tracking-wide text-indigo-600">{item.sport_name}</p>
                    <p class="text-base font-semibold text-gray-900">{item.team_name}</p>
                    <p class="text-xs text-gray-600">{item.tournament_name}</p>
                  </header>
                  <p class="text-sm text-gray-800">
                    現在の状況: <span class="font-semibold">{item.status}</span>
                  </p>
                  <p class="text-xs text-gray-600">現ラウンド: {item.current_round}</p>

                  {#if item.next_match}
                    <div class="mt-2 rounded border border-indigo-200 bg-white px-3 py-2 text-xs text-gray-700">
                      <p class="font-semibold text-gray-800">次の試合</p>
                      <p>
                        {item.next_match.round_label}
                        {#if item.next_match.opponent_name}
                          ・対 {item.next_match.opponent_name}
                        {/if}
                      </p>
                      <p>ステータス: {item.next_match.match_status || '未定'}</p>
                      {#if item.next_match.start_time}
                        <p>開始予定: {item.next_match.start_time}</p>
                      {/if}
                    </div>
                  {/if}

                  {#if item.last_match}
                    <div class="mt-2 rounded border border-indigo-200 bg-white px-3 py-2 text-xs text-gray-700">
                      <p class="font-semibold text-gray-800">前の試合</p>
                      <p>
                        {item.last_match.round_label}
                        {#if item.last_match.opponent_name}
                          ・対 {item.last_match.opponent_name}
                        {/if}
                      </p>
                      <p>
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
        </div>
      </div>
    </section>
  {/if}
</div>
