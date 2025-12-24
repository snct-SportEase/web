<script>
  import { onMount } from 'svelte';
  import PWAInstallGuideModal from '$lib/components/PWAInstallGuideModal.svelte';

  let showPWAInstallGuide = true;
  let activeEvent = null;
  let competitionGuidelinesUrl = null;

  onMount(async () => {
    try {
      const response = await fetch('/api/events/active');
      if (response.ok) {
        const data = await response.json();
        if (data.event_id) {
          activeEvent = {
            id: data.event_id,
            name: data.event_name
          };
          if (data.competition_guidelines_pdf_url) {
            competitionGuidelinesUrl = data.competition_guidelines_pdf_url;
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch active event:', error);
    }
  });

  function openCompetitionGuidelines() {
    if (competitionGuidelinesUrl) {
      window.open(competitionGuidelinesUrl, '_blank');
    }
  }
</script>

<div class="max-w-4xl mx-auto">
  <section class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">資料</h1>
        <p class="mt-2 text-sm text-gray-600">
          SportEaseの使い方や各種設定方法について説明します
        </p>
      </div>
    </div>

    <div class="grid gap-6 md:grid-cols-2">
      <button
        type="button"
        on:click={() => showPWAInstallGuide = true}
        class="group block rounded-lg border border-indigo-100 bg-white p-6 shadow-sm transition hover:border-indigo-300 hover:shadow text-left"
      >
        <div class="flex items-center mb-3">
          <svg class="w-8 h-8 text-indigo-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
          </svg>
          <h2 class="text-xl font-semibold text-indigo-700 group-hover:text-indigo-800">
            PWAインストール方法
          </h2>
        </div>
        <p class="text-sm text-gray-600 mb-4">
          iOS、Android、Windows、macOSなど、お使いのOS別のPWAインストール手順をご確認いただけます。
        </p>
        <div class="flex items-center text-sm font-medium text-indigo-600 group-hover:text-indigo-700">
          詳細を見る
          <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
          </svg>
        </div>
      </button>

      {#if competitionGuidelinesUrl}
        <button
          type="button"
          on:click={openCompetitionGuidelines}
          class="group block rounded-lg border border-indigo-100 bg-white p-6 shadow-sm transition hover:border-indigo-300 hover:shadow text-left w-full"
        >
          <div class="flex items-center mb-3">
            <svg class="w-8 h-8 text-indigo-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
            </svg>
            <h2 class="text-xl font-semibold text-indigo-700 group-hover:text-indigo-800">
              競技要項
            </h2>
          </div>
          <p class="text-sm text-gray-600 mb-4">
            {#if activeEvent}
              {activeEvent.name}の競技要項を確認できます。
            {:else}
              大会の競技要項を確認できます。
            {/if}
          </p>
          <div class="flex items-center text-sm font-medium text-indigo-600 group-hover:text-indigo-700">
            競技要項を見る
            <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
            </svg>
          </div>
        </button>
      {:else}
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-6">
          <div class="flex items-center mb-3">
            <svg class="w-8 h-8 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"></path>
            </svg>
            <h2 class="text-xl font-semibold text-gray-700">
              その他の資料
            </h2>
          </div>
          <p class="text-sm text-gray-600">
            今後、使い方ガイドやFAQなどの資料を追加予定です。
          </p>
        </div>
      {/if}
    </div>
  </section>

  <PWAInstallGuideModal
    isOpen={showPWAInstallGuide}
    onClose={() => showPWAInstallGuide = false}
  />
</div>
