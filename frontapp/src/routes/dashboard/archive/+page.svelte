<script>
    import { onMount } from 'svelte';
    
    let events = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            const res = await fetch('/api/events');
            if (!res.ok) {
                throw new Error('Failed to fetch events');
            }
            const allEvents = await res.json();
            events = allEvents.filter(e => e.status === 'archived');
            // Sort events by year (descending), season (autumn first)
            events.sort((a, b) => {
                if (a.year !== b.year) return b.year - a.year;
                if (a.season === 'autumn' && b.season === 'spring') return -1;
                if (a.season === 'spring' && b.season === 'autumn') return 1;
                return 0;
            });
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    });

    function formatDate(dateStr) {
        if (!dateStr) return '';
        const d = new Date(dateStr);
        return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`;
    }
</script>

<svelte:head>
    <title>過去の大会一覧 - SportEase</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
    <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900">過去の大会アーカイブ</h1>
        <p class="mt-2 text-sm text-gray-600">終了した大会の結果やスコアを振り返ることができます。</p>
    </div>

    {#if loading}
        <div class="flex justify-center p-12">
            <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
        </div>
    {:else if error}
        <div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded relative" role="alert">
            <span class="block sm:inline whitespace-pre-wrap">{error}</span>
        </div>
    {:else if events.length === 0}
        <div class="bg-white shadow rounded-lg p-10 text-center text-gray-500 border border-gray-100">
            <svg class="mx-auto h-12 w-12 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 002-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
            </svg>
            <p>過去の大会データがありません。</p>
        </div>
    {:else}
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {#each events as ev (ev.id)}
                <a href="/dashboard/archive/{ev.id}" class="block group h-full">
                    <div class="bg-white rounded-xl shadow-sm border border-indigo-50 p-6 h-full flex flex-col transition-all duration-300 hover:shadow-md hover:border-indigo-200 hover:-translate-y-1">
                        <div class="flex-1">
                            <div class="flex items-center justify-between mb-3">
                                <span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                                    {ev.year}年 {ev.season === 'spring' ? '春季' : '秋季'}
                                </span>
                                {#if ev.is_rainy_mode}
                                    <span class="px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 flex items-center">
                                        <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"></path></svg>
                                        雨天時
                                    </span>
                                {/if}
                            </div>
                            <h3 class="text-xl font-bold text-gray-900 group-hover:text-indigo-600 transition-colors mb-2">
                                {ev.name}
                            </h3>
                            <p class="text-sm text-gray-500 mb-4">
                                {formatDate(ev.start_date)} {#if ev.end_date}〜 {formatDate(ev.end_date)}{/if}
                            </p>
                        </div>
                        
                        <div class="mt-auto pt-4 border-t border-gray-100 flex items-center justify-between">
                            <span class="text-indigo-600 text-sm font-semibold group-hover:text-indigo-700 flex items-center">
                                詳細を見る
                                <svg class="ml-1 w-4 h-4 transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                                </svg>
                            </span>
                        </div>
                    </div>
                </a>
            {/each}
        </div>
    {/if}
</div>
