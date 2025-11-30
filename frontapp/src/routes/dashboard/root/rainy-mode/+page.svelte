<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';

  let selectedEventId = null;
  let currentEvent = null;
  let isRainyMode = false;
  let loading = false;

  onMount(async () => {
    console.log('rainy-mode page: onMount started');
    try {
      console.log('rainy-mode page: initializing activeEvent...');
      const active = await activeEvent.init();
      console.log('rainy-mode page: activeEvent initialized', active);
      
      if (active) {
        selectedEventId = active.id;
        currentEvent = active;
        isRainyMode = active.is_rainy_mode || false;
        console.log('rainy-mode page: active event loaded', selectedEventId);
      } else {
        console.log('rainy-mode page: no active event found');
      }
      console.log('rainy-mode page: onMount completed');
    } catch (error) {
      console.error('初期化エラー:', error);
      alert('ページの初期化に失敗しました: ' + error.message);
      loading = false;
    }
  });

  async function toggleRainyMode() {
    if (!selectedEventId) {
      alert('大会を選択してください');
      return;
    }

    loading = true;
    try {
      const response = await fetch(`/api/root/events/${selectedEventId}/rainy-mode`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ is_rainy_mode: !isRainyMode })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || '雨天時モードの切り替えに失敗しました');
      }

      const result = await response.json();
      isRainyMode = result.is_rainy_mode;
      
      // アクティブなイベント情報を再取得して更新
      const active = await activeEvent.init();
      if (active) {
        currentEvent = active;
        isRainyMode = active.is_rainy_mode || false;
      }
      
      alert(isRainyMode ? '雨天時モードを有効にしました' : '雨天時モードを無効にしました');
    } catch (error) {
      console.error(error);
      alert(error.message);
    } finally {
      loading = false;
    }
  }

</script>

<div class="container mx-auto p-4 space-y-6">
  <div class="flex justify-between items-center">
    <h1 class="text-2xl font-bold">雨天時モード管理</h1>
  </div>

  <!-- アクティブな大会の表示 -->
  {#if !selectedEventId || !currentEvent}
    <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-yellow-800">アクティブな大会が設定されていません</h3>
          <div class="mt-2 text-sm text-yellow-700">
            <p>雨天時モードを管理するには、まずアクティブな大会を設定してください。</p>
          </div>
        </div>
      </div>
    </div>
  {:else}
    <div class="bg-white shadow-md rounded-lg p-6 mb-6">
      <h2 class="text-xl font-semibold mb-2">対象大会</h2>
      <p class="text-gray-700">{currentEvent.name}</p>
      <p class="text-sm text-gray-500 mt-1">{currentEvent.year}年 {currentEvent.season === 'spring' ? '春季' : '秋季'}</p>
    </div>
  {/if}

  {#if selectedEventId && currentEvent}
    <!-- 雨天時モード切り替え -->
    <div class="bg-white shadow-md rounded-lg p-6">
      <h2 class="text-xl font-semibold mb-4">雨天時モード</h2>
      <div class="flex items-center justify-between">
        <div>
          <p class="text-gray-700 mb-2">
            現在の状態: 
            <span class="font-bold {isRainyMode ? 'text-red-600' : 'text-gray-600'}">
              {isRainyMode ? '有効' : '無効'}
            </span>
          </p>
          <p class="text-sm text-gray-500">
            雨天時モードを有効にすると、昼競技とグラウンド競技が中止となり、gym2のトーナメントに敗者戦が追加されます。
          </p>
        </div>
        <button
          on:click={toggleRainyMode}
          disabled={loading}
          class="px-6 py-2 rounded-md font-medium {isRainyMode 
            ? 'bg-gray-600 text-white hover:bg-gray-700' 
            : 'bg-red-600 text-white hover:bg-red-700'} disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isRainyMode ? '無効にする' : '有効にする'}
        </button>
      </div>
    </div>

    <!-- 雨天時設定についての説明 -->
    <div class="bg-blue-50 border border-blue-200 rounded-lg p-6">
      <div class="flex items-start">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-blue-800">雨天時設定について</h3>
          <div class="mt-2 text-sm text-blue-700">
            <p>雨天時の各競技・クラスごとの登録可能人数の上限・下限と試合開始時間は、<strong>競技詳細情報登録</strong>ページの「雨天時定員設定」セクションで設定できます。</p>
          </div>
        </div>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white p-4 rounded-md">
        <p class="text-gray-700">処理中...</p>
      </div>
    </div>
  {/if}
</div>


