<script>
    import { onMount } from 'svelte';
    import { activeEvent } from '$lib/stores/eventStore.js';
    import { get } from 'svelte/store';

    let scheduleItems = [];
    let isLoading = true;

    // ヘルパー：hh:mm 形式に変換
    function formatTime(dateStr) {
        if (!dateStr) return '未定';
        try {
            const d = new Date(dateStr);
            if (isNaN(d.getTime())) return '未定';
            return d.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
        } catch {
            return '未定';
        }
    }

    onMount(async () => {
        await activeEvent.init();
        await fetchTimetableData();
    });

    async function fetchTimetableData() {
        const currentEvent = get(activeEvent);
        if (!currentEvent) {
            isLoading = false;
            return;
        }

        isLoading = true;
        try {
            const rawItems = [];

            // 1. トーナメント（一般競技）の取得
            const tourRes = await fetch(`/api/student/events/${currentEvent.id}/tournaments`);
            if (tourRes.ok) {
                const tournaments = await tourRes.json();
                for (const tour of tournaments) {
                    let data = tour.data;
                    if (typeof data === 'string') {
                        try {
                            data = JSON.parse(data);
                        } catch (e) {
                            continue;
                        }
                    }
                    if (data && data.matches) {
                        for (const match of data.matches) {
                            // 開始予定時間（通常 or 雨天時）を取得
                            // 雨天時モード有効なら rainyModeStartTime を優先
                            let timeStr = null;
                            if (currentEvent.is_rainy_mode && match.rainyModeStartTime) {
                                timeStr = match.rainyModeStartTime;
                            } else if (match.startTime) {
                                timeStr = match.startTime;
                            }

                            if (timeStr) {
                                // チーム名を抽出
                                let team1 = "未定";
                                let team2 = "未定";
                                if (match.sides && match.sides.length >= 2) {
                                    team1 = match.sides[0]?.title || "未定";
                                    team2 = match.sides[1]?.title || "未定";
                                }

                                rawItems.push({
                                    id: `match-${match.id || Math.random()}`,
                                    type: 'sport',
                                    title: tour.name,
                                    subtitle: `${team1} vs ${team2}`,
                                    timeStr: timeStr,
                                    timeObj: new Date(timeStr),
                                    status: match.matchStatus || 'SCHEDULED'
                                });
                            }
                        }
                    }
                }
            }

            // 2. 昼競技の取得
            const noonRes = await fetch(`/api/student/events/${currentEvent.id}/noon-game/session`);
            if (noonRes.ok) {
                const session = await noonRes.json();
                if (session && session.matches) {
                    for (const match of session.matches) {
                        if (match.scheduled_at) {
                            rawItems.push({
                                id: `noon-${match.id}`,
                                type: 'noon',
                                title: match.title || session.name,
                                subtitle: match.format ? `形式: ${match.format}` : (match.location ? `場所: ${match.location}` : '昼競技'),
                                timeStr: match.scheduled_at,
                                timeObj: new Date(match.scheduled_at),
                                status: match.status || 'scheduled'
                            });
                        }
                    }
                }
            }

            // 時間順にソート (timeObjが有効なものを前へ)
            scheduleItems = rawItems.sort((a, b) => {
                if (isNaN(a.timeObj.getTime()) && isNaN(b.timeObj.getTime())) return 0;
                if (isNaN(a.timeObj.getTime())) return 1;
                if (isNaN(b.timeObj.getTime())) return -1;
                return a.timeObj.getTime() - b.timeObj.getTime();
            });

        } catch (error) {
            console.error('Error fetching timetable:', error);
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="space-y-8 p-4 md:p-8">
    <div class="flex items-center justify-between border-b pb-2">
        <h1 class="text-2xl md:text-3xl font-bold text-gray-800">タイムテーブル</h1>
        <button 
            on:click={fetchTimetableData} 
            class="text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 py-1 px-3 rounded flex items-center transition-colors"
            disabled={isLoading}
        >
            <svg class="w-4 h-4 mr-1 {isLoading ? 'animate-spin' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            更新
        </button>
    </div>

    {#if isLoading}
        <div class="flex justify-center items-center py-12">
            <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
            <span class="ml-3 text-gray-600 font-medium">スケジュールを読み込み中...</span>
        </div>
    {:else if scheduleItems.length > 0}
        <div class="relative">
            <!-- タイムラインの縦線 -->
            <div class="absolute inset-0 flex items-center justify-center w-8 md:w-16">
                <div class="h-full w-0.5 bg-gray-200 pointer-events-none"></div>
            </div>

            <div class="space-y-6 md:space-y-8">
                {#each scheduleItems as item (item.id)}
                    <div class="relative flex items-center">
                        <!-- 時間表示 (左側) -->
                        <div class="w-16 md:w-24 flex-shrink-0 text-right pr-4 md:pr-6">
                            <span class="text-sm md:text-base font-bold text-gray-700">{formatTime(item.timeStr)}</span>
                        </div>
                        
                        <!-- タイムラインの丸アイコン -->
                        <div class="absolute left-16 md:left-24 -ml-3 md:-ml-4 flex items-center justify-center w-6 h-6 md:w-8 md:h-8 rounded-full border-4 border-white z-10 
                            {item.type === 'noon' ? 'bg-orange-400' : 'bg-indigo-500'}">
                        </div>

                        <!-- コンテンツカード (右側) -->
                        <div class="ml-4 md:ml-8 flex-grow">
                            <div class="bg-white p-4 md:p-5 rounded-lg shadow-sm border border-gray-100 hover:shadow-md transition-shadow relative overflow-hidden">
                                <!-- ステータス表示のバッジなどを追加する場合に備えた構成 -->
                                <div class="flex justify-between items-start mb-1">
                                    <h3 class="text-lg font-bold text-gray-800">
                                        {#if item.type === 'noon'}
                                            【昼競技】{item.title}
                                        {:else}
                                            {item.title}
                                        {/if}
                                    </h3>
                                    {#if item.status === 'COMPLETED' || item.status === 'completed'}
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                            終了
                                        </span>
                                    {:else if item.status === 'IN_PROGRESS' || item.status === 'in_progress'}
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                            進行中
                                        </span>
                                    {/if}
                                </div>
                                <p class="text-gray-600 text-sm md:text-base">{item.subtitle}</p>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {:else}
        <div class="bg-blue-50 border border-blue-200 text-blue-700 p-6 rounded-lg shadow-sm flex items-start">
            <svg class="w-6 h-6 text-blue-500 mr-3 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
                <h3 class="font-bold mb-1">スケジュールがありません</h3>
                <p class="text-sm">現在、開始時間が設定されている予定はありません。</p>
            </div>
        </div>
    {/if}
</div>

<style>
    /* 追加のスタイル調整があればここに */
</style>
