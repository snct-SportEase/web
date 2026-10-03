<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let checking = $state(false);

  async function checkStatus() {
    if (checking) return;
    checking = true;
    try {
      const response = await fetch('/api/events/active', {
        headers: { 'Cache-Control': 'no-cache, no-store, must-revalidate' }
      });
      if (!response.ok) return;
      const data = await response.json();
      if (!['starting', 'testing', 'restoring', 'failed'].includes(data.test_run_state)) {
        await goto('/dashboard', { invalidateAll: true });
      }
    } finally {
      checking = false;
    }
  }

  onMount(() => {
    const interval = window.setInterval(checkStatus, 15_000);
    return () => window.clearInterval(interval);
  });
</script>

<svelte:head>
  <title>メンテナンス中 | SportEase</title>
</svelte:head>

<div class="mx-auto flex min-h-[55vh] max-w-2xl items-center justify-center">
  <section class="w-full rounded-xl border border-amber-200 bg-amber-50 p-8 text-center shadow-sm" aria-labelledby="maintenance-title">
    <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-amber-100 text-2xl" aria-hidden="true">🛠️</div>
    <h1 id="maintenance-title" class="mt-4 text-2xl font-bold text-amber-950">大会テストのためメンテナンス中です</h1>
    <p class="mt-3 text-sm leading-6 text-amber-900">
      現在、運営担当者が大会のテスト試行を行っています。学生向け機能はテスト終了まで一時停止しています。
    </p>
    <p class="mt-2 text-sm text-amber-800">画面は15秒ごとに自動確認されます。</p>
    <button
      type="button"
      class="mt-6 rounded-md bg-amber-700 px-4 py-2 text-sm font-medium text-white hover:bg-amber-800 disabled:opacity-60"
      onclick={checkStatus}
      disabled={checking}
    >
      {checking ? '確認中…' : '終了したか確認'}
    </button>
  </section>
</div>
