<script>
    import { onMount } from 'svelte';
    import { activeEvent } from '$lib/stores/eventStore.js';

    let allSports = [];
    let eventSports = [];
    let newSportName = '';

    let newAssignment = {
        sport_id: null,
        description: '',
        rules: '',
        location: 'other',
    };

    let currentActiveEvent = null;

    const allLocations = ['gym1', 'gym2', 'ground', 'noon_game', 'other'];

    $: usedLocations = eventSports ? eventSports.map(es => es.location).filter(loc => loc !== 'other') : [];

    onMount(async () => {
        // 競技名解決のために、先に全競技マスタを読み込んでおく
        await fetchAllSports();

        // Subscribe to the active event store
        const unsubscribe = activeEvent.subscribe(value => {
            currentActiveEvent = value;
            if (value) {
                // アクティブイベントが変わったら、割り当て済み競技を再取得
                fetchEventSports(value.id);
            } else {
                eventSports = []; // クリア
            }
        });

        // Initialize and fetch the active event data
        await activeEvent.init();

        return unsubscribe; // Cleanup subscription on destroy
    });

    // $: {
    //     if (newAssignment.sport_id && allSports.length > 0) {
    //         const sportName = getSportName(newAssignment.sport_id);
    //         if (sportName !== '不明な競技') {
    //             const isRuleEmpty = newAssignment.rules.trim() === '';
                
    //             // Find if the current rule is a default for ANY sport
    //             const isRuleADefault = allSports.some(s => `# ${s.name}` === newAssignment.rules.trim());

    //             if (isRuleEmpty || isRuleADefault) {
    //                 newAssignment.rules = `# ${sportName}`;
    //             }
    //         }
    //     }
    // }
    
    // --- Data Fetching Functions ---

    async function fetchAllSports() {
        try {
            // /api/root/sports は管理者権限が必要
            const response = await fetch('/api/root/sports'); 
            if (!response.ok) throw new Error('Failed to fetch all sports');
            allSports = await response.json();
        } catch (error) {
            console.error(error);
            alert(`競技マスタの取得に失敗しました: ${error.message}`);
        }
    }

    async function fetchEventSports(eventId) {
        try {
            // /api/events/:id/sports は全ユーザー許可(GET)されている前提
            const response = await fetch(`/api/events/${eventId}/sports`); 
            if (!response.ok) throw new Error('Failed to fetch event sports');
            eventSports = await response.json();
        } catch (error) {
            console.error(error);
            alert(`割り当て済み競技の取得に失敗しました: ${error.message}`);
        }
    }

    // --- Action Functions ---

    async function createSport() {
        if (!newSportName.trim()) {
            alert('競技名を入力してください。');
            return;
        }
        try {
            const response = await fetch('/api/root/sports', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: newSportName }),
            });
            if (!response.ok) {
                 const errorData = await response.json();
                 throw new Error(errorData.error || 'Failed to create sport');
            }
            newSportName = '';
            await fetchAllSports(); // Refresh the list
            alert('新しい競技を登録しました。');
        } catch (error) {
            console.error(error);
            alert(`登録エラー: ${error.message}`);
        }
    }

    async function assignSport() {
        if (!currentActiveEvent) {
            alert('アクティブな大会が設定されていません。');
            return;
        }
        if (!newAssignment.sport_id) {
            alert('割り当てる競技を選択してください。');
            return;
        }

        try {
            // /api/admin/events/:id/sports は管理者権限が必要
            const response = await fetch(`/api/admin/events/${currentActiveEvent.id}/sports`, { 
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    ...newAssignment,
                    sport_id: parseInt(newAssignment.sport_id, 10),
                }),
            });
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to assign sport to event');
            }
            
            // Reset form
            newAssignment = { sport_id: null, description: '', rules: '', location: 'other' };

            await fetchEventSports(currentActiveEvent.id); // Refresh the list
            alert('競技を大会に割り当てました。');
        } catch (error) {
            console.error(error);
            alert(`割り当てエラー: ${error.message}`);
        }
    }

    async function deleteAssignedSport(sportId) {
        if (!currentActiveEvent) {
            alert('アクティブな大会が設定されていません。');
            return;
        }

        const sportName = getSportName(sportId);
        if (!confirm(`本当に「${sportName}」の割り当てを解除しますか？この操作は元に戻せません。`)) {
            return;
        }

        try {
            const response = await fetch(`/api/admin/events/${currentActiveEvent.id}/sports/${sportId}`, {
                method: 'DELETE',
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to delete sport assignment');
            }

            await fetchEventSports(currentActiveEvent.id); // Refresh the list
            alert(`「${sportName}」の割り当てを解除しました。`);
        } catch (error) {
            console.error(error);
            alert(`削除エラー: ${error.message}`);
        }
    }

    // Helper to get sport name from ID
    function getSportName(sportId) {
        const sport = allSports.find(s => s.id === sportId);
        return sport ? sport.name : '不明な競技';
    }
</script>

<div class="container mx-auto p-6 lg:p-10 space-y-12">
    <h1 class="text-3xl font-extrabold text-gray-800 border-b pb-2">大会競技管理ダッシュボード</h1>

    <!-- Active Event Banner -->
    <div class="p-4 rounded-xl shadow-lg {currentActiveEvent ? 'bg-indigo-100 border-l-4 border-indigo-600' : 'bg-red-100 border-l-4 border-red-600'} transition-all duration-300">
        <h2 class="text-lg font-bold {currentActiveEvent ? 'text-indigo-800' : 'text-red-800'}">
            アクティブな大会
        </h2>
        {#if currentActiveEvent}
            <p class="text-2xl font-extrabold {currentActiveEvent ? 'text-indigo-600' : 'text-gray-900'}">
                {currentActiveEvent.name}
            </p>
            <p class="text-sm {currentActiveEvent ? 'text-indigo-700' : 'text-red-700'} mt-1">
                開催期間: {currentActiveEvent.start_date ? new Date(currentActiveEvent.start_date).toLocaleDateString() : '未定'}
            </p>
        {:else}
            <p class="text-lg text-red-700 mt-2 font-semibold">
                現在、アクティブな大会が設定されていません。大会管理ページで設定してください。
            </p>
        {/if}
    </div>

    <div class="lg:grid lg:grid-cols-3 gap-8">
        <!-- Column 1: Sport Master Management -->
        <div class="lg:col-span-1 p-6 border rounded-xl bg-gray-50 shadow-lg h-full">
            <h2 class="text-xl font-bold text-gray-800 mb-4 border-b pb-2 border-gray-300">
                競技マスタ管理 <span class="text-sm text-red-500">(root権限)</span>
            </h2>
            
            <!-- 新規競技登録フォーム -->
            <div class="mb-6 p-4 bg-white rounded-lg shadow-inner">
                <h3 class="font-semibold mb-3 text-lg text-indigo-700">新規競技登録</h3>
                <div class="space-y-2">
                    <input type="text" bind:value={newSportName} placeholder="新しい競技名を入力" class="input-field" />
                    <button on:click={createSport} class="w-full bg-indigo-600 text-white p-3 rounded-lg hover:bg-indigo-700 font-semibold transition duration-150">
                        競技をマスタに登録
                    </button>
                </div>
            </div>

            <!-- 登録済み競技一覧 -->
            <div>
                <h3 class="font-semibold mb-3 text-lg text-gray-700">登録済み競技一覧 ({allSports.length}件)</h3>
                <div class="max-h-72 overflow-y-auto bg-white p-3 rounded-lg border">
                    <ul class="space-y-2">
                        {#each allSports as sport (sport.id)}
                            <li class="flex justify-between items-center text-sm p-2 bg-gray-50 rounded-md">
                                <span class="font-medium text-gray-800">{sport.name}</span>
                                <span class="text-xs text-gray-500">ID: {sport.id}</span>
                            </li>
                        {:else}
                            <li class="text-gray-500 text-sm italic">登録されている競技はありません。</li>
                        {/each}
                    </ul>
                </div>
            </div>
        </div>

        <!-- Column 2 & 3: Assign Sports to Active Event -->
        <div class="lg:col-span-2 p-8 border rounded-xl bg-white shadow-2xl space-y-6">
            <h2 class="text-xl font-bold text-gray-800 mb-4 border-b pb-2 border-gray-300">
                アクティブな大会への競技割り当て <span class="text-sm text-red-500">(admin/root権限)</span>
            </h2>

            {#if currentActiveEvent}
                <div class="space-y-8">
                    <!-- Left side: Assignment Form -->
                    <div class="space-y-4 p-4 border rounded-lg bg-blue-50">
                        <h3 class="font-bold text-blue-700 text-lg">大会への競技割り当てフォーム</h3>
                        
                        <!-- 競技選択 -->
                        <div>
                            <label for="sport-select" class="block text-sm font-medium text-gray-700 mb-1">割り当てる競技</label>
                            <select id="sport-select" bind:value={newAssignment.sport_id} class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm">
                                <option value={null} disabled>競技を選択...</option>
                                {#each allSports as sport}
                                    <option value={sport.id}>{sport.name}</option>
                                {/each}
                            </select>
                        </div>
                        
                        <!-- 場所選択 -->
                        <div>
                            <label for="location-select" class="block text-sm font-medium text-gray-700 mb-1">場所</label>
                            <select id="location-select" bind:value={newAssignment.location} class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm">
                                {#each allLocations as loc}
                                    <option value={loc} disabled={usedLocations.includes(loc)}>
                                        {loc} {usedLocations.includes(loc) ? '(使用中)' : ''}
                                    </option>
                                {/each}
                            </select>
                        </div>
                        
                        <!-- 概要 -->
                        <div>
                            <label for="description" class="block text-sm font-medium text-gray-700 mb-1">概要 (任意)</label>
                            <textarea id="description" bind:value={newAssignment.description} class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm h-20" placeholder="競技の簡単な説明や備考"></textarea>
                        </div>
                        
                        <!-- ルール -->
                        <div>
                            <label for="rules" class="block text-sm font-medium text-gray-700 mb-1">ルール詳細 (任意)</label>
                            <textarea id="rules" bind:value={newAssignment.rules} class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm h-32" placeholder="競技のルール詳細"></textarea>
                        </div>
                        
                        <button on:click={assignSport} class="bg-blue-600 text-white p-3 rounded-lg hover:bg-blue-700 font-semibold w-full mt-4 transition duration-150" disabled={!newAssignment.sport_id}>
                            大会に競技を割り当てる
                        </button>
                    </div>

                    <!-- Right side: Assigned Sports List -->
                    <div>
                        <h3 class="font-medium text-lg mb-4 text-gray-700">割り当て済み競技一覧 ({eventSports.length}件)</h3>
                        <div class="max-h-96 overflow-y-auto border rounded-lg shadow-inner">
                            <table class="w-full text-sm">
                                <thead class="sticky top-0 bg-gray-200">
                                    <tr>
                                        <th class="px-4 py-2 text-left font-semibold text-gray-700">競技名</th>
                                        <th class="px-4 py-2 text-left font-semibold text-gray-700 w-1/4">場所</th>
                                        <th class="px-4 py-2 text-left font-semibold text-gray-700 w-1/2">概要</th>
                                        <th class="px-4 py-2 text-left font-semibold text-gray-700">アクション</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each eventSports as es (es.sport_id)}
                                        <tr class="border-t hover:bg-gray-50">
                                            <td class="px-4 py-3 font-medium text-gray-900">{getSportName(es.sport_id)}</td>
                                            <td class="px-4 py-3 text-gray-600">{es.location}</td>
                                            <td class="px-4 py-3 text-gray-600 truncate max-w-xs">{es.description || '概要なし'}</td>
                                            <td class="px-4 py-3">
                                                <button on:click={() => deleteAssignedSport(es.sport_id)} class="text-red-600 hover:text-red-800 font-semibold text-xs py-1 px-3 rounded-full bg-red-100 hover:bg-red-200 transition-all duration-150">
                                                    解除
                                                </button>
                                            </td>
                                        </tr>
                                    {:else}
                                        <tr>
                                            <td colspan="4" class="text-center py-6 text-gray-500 italic">
                                                この大会に割り当てられた競技はありません。
                                            </td>
                                        </tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

            {:else}
                <div class="text-center py-10">
                    <p class="text-xl text-red-500 font-bold">アクティブな大会が未設定です。</p>
                    <p class="text-gray-600 mt-2">競技を割り当てるには、まず大会情報管理ページでアクティブな大会を設定してください。</p>
                </div>
            {/if}
        </div>
    </div>
</div>
