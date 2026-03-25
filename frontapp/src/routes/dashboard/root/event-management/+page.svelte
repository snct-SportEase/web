<script>
  import { onMount } from 'svelte';
  import { activeEvent } from '$lib/stores/eventStore.js';

  let events = [];
  let showModal = false;
  let selectedEvent = null;
  let isNameManuallyChanged = false;

  let currentEvent = {
    id: null,
    name: '',
    year: new Date().getFullYear(),
    season: 'spring',
    start_date: '',
    end_date: '',
    survey_url: null,
    status: 'upcoming',
    hide_scores: false,
  };

  $effect(() => {
    if (!isNameManuallyChanged && currentEvent.year && currentEvent.season) {
      const seasonText = currentEvent.season === 'spring' ? '春' : '秋';
      currentEvent.name = `${currentEvent.year}${seasonText}季スポーツ大会`;
    }
  });

  function onNameInput() {
    isNameManuallyChanged = true;
  }

  onMount(async () => {
    await fetchEvents();
    await activeEvent.init();
  });

  async function fetchEvents() {
    try {
      const response = await fetch('/api/root/events', {
        headers: {
          'Cache-Control': 'no-cache, no-store, must-revalidate',
          'Pragma': 'no-cache'
        }
      });
      if (!response.ok) {
        throw new Error('Failed to fetch events');
      }
      events = await response.json();
    } catch (error) {
      console.error(error);
      alert(error.message);
    }
  }



  function openCreateModal() {
    selectedEvent = null;
    isNameManuallyChanged = false;
    currentEvent = {
      id: null,
      name: '',
      year: new Date().getFullYear(),
      season: 'spring',
      start_date: '',
      end_date: '',
      survey_url: null,
      status: 'upcoming',
      hide_scores: false,
    };
    showModal = true;
  }

  function openEditModal(event) {
    selectedEvent = event;
    isNameManuallyChanged = true; // 編集時は手動変更とみなし、自動更新しない
    currentEvent = {
      ...event,
      start_date: event.start_date ? new Date(event.start_date).toISOString().split('T')[0] : '',
      end_date: event.end_date ? new Date(event.end_date).toISOString().split('T')[0] : '',
      survey_url: event.survey_url || '',
      status: event.status || 'upcoming',
      hide_scores: event.hide_scores || false,
    };
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }

  async function handleSave() {
    try {
      const method = selectedEvent ? 'PUT' : 'POST';
      const url = selectedEvent ? `/api/root/events/${selectedEvent.id}` : '/api/root/events';

      const body = {
        ...currentEvent,
        year: parseInt(currentEvent.year, 10),
      };

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save event');
      }

      await fetchEvents();
      await activeEvent.init();
      closeModal();
    } catch (error) {
      console.error(error);
      alert(error.message);
    }
  }

  async function notifySurvey(id) {
    if (!confirm('アンケート通知を全ユーザーに送信します。よろしいですか？')) return;
    try {
      const resp = await fetch(`/api/root/events/${id}/notify-survey`, { method: 'POST' });
      if (!resp.ok) {
        const d = await resp.json();
        throw new Error(d.error || 'Failed to send notification');
      }
      alert('アンケート通知を送信しました。');
    } catch (err) {
      alert(err.message);
    }
  }

  async function handleCsvUpload(id, e) {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const resp = await fetch(`/api/root/events/${id}/import-survey-scores`, {
        method: 'POST',
        body: formData,
      });

      if (!resp.ok) {
        const d = await resp.json();
        throw new Error(d.error || 'Failed to import survey scores');
      }
      const data = await resp.json();
      alert(`インポート成功: ${data.imported_classes_count} クラス分の点数が反映されました。`);
    } catch (err) {
      alert(err.message);
    }
    // Upload inputリセット
    e.target.value = '';
  }

  async function downloadExport(event, type) {
    try {
      if (type === 'pdf') {
        const res = await fetch(`/api/scores/class?event_id=${event.id}`);
        if (!res.ok) throw new Error('クラススコアの取得に失敗しました');
        const scores = await res.json();
        
        let htmlContent = `
          <div style="font-family: Arial, 'Hiragino Sans', 'Hiragino Kaku Gothic ProN', 'Noto Sans JP', Meiryo, sans-serif; padding: 20px;">
              <h1 style="text-align: center; font-size: 24px; margin-bottom: 20px;">${event.name} - 最終結果</h1>
              <table style="width: 100%; border-collapse: collapse; text-align: center; font-size: 14px;">
                  <thead>
                      <tr style="background-color: #f2f2f2;">
                          <th style="border: 1px solid #ddd; padding: 12px; font-weight: bold;">全体順位</th>
                          <th style="border: 1px solid #ddd; padding: 12px; font-weight: bold;">クラス名</th>
                          <th style="border: 1px solid #ddd; padding: 12px; font-weight: bold;">総合スコア</th>
                          <th style="border: 1px solid #ddd; padding: 12px; font-weight: bold;">今大会スコア</th>
                      </tr>
                  </thead>
                  <tbody>
                      ${scores.map(s => `
                      <tr>
                          <td style="border: 1px solid #ddd; padding: 12px;">${s.rank_overall}</td>
                          <td style="border: 1px solid #ddd; padding: 12px;">${s.class_name}</td>
                          <td style="border: 1px solid #ddd; padding: 12px;">${s.total_points_overall}</td>
                          <td style="border: 1px solid #ddd; padding: 12px;">${s.total_points_current_event}</td>
                      </tr>
                      `).join('')}
                  </tbody>
              </table>
          </div>
        `;
        
        const opt = {
          margin: 10,
          filename: `event_${event.id}_class_scores.pdf`,
          image: { type: 'jpeg', quality: 0.98 },
          html2canvas: { scale: 2 },
          jsPDF: { unit: 'mm', format: 'a4', orientation: 'portrait' }
        };
        const html2pdf = (await import('html2pdf.js')).default;
        html2pdf().set(opt).from(htmlContent).save();
      } else {
        const resp = await fetch(`/api/root/events/${event.id}/export/${type}`);
        if (!resp.ok) {
          const errorData = await resp.json().catch(() => null);
          throw new Error((errorData && errorData.error) || `${type.toUpperCase()}のダウンロードに失敗しました`);
        }
        
        const blob = await resp.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `event_${event.id}_class_scores.${type}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      }
    } catch (err) {
      alert(err.message);
    }
  }

  async function downloadDBDump() {
    try {
      const resp = await fetch('/api/root/db/export');
      if (!resp.ok) {
        const errorData = await resp.json().catch(() => null);
        throw new Error((errorData && errorData.error) || 'DBダンプのダウンロードに失敗しました');
      }
      
      const blob = await resp.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      // Get filename from Content-Disposition header if possible, else default
      const contentDisposition = resp.headers.get('Content-Disposition');
      let filename = 'database_dump.sql';
      if (contentDisposition && contentDisposition.includes('filename=')) {
        filename = contentDisposition.split('filename=')[1].replace(/"/g, '');
      }
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      alert(err.message);
    }
  }
</script>

<div class="container mx-auto p-4">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-2xl font-bold">大会情報登録・管理</h1>
    <div class="flex space-x-2">
      <button onclick={downloadDBDump} class="btn btn-secondary bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700">
        DBダンプ出力
      </button>
      <button onclick={openCreateModal} class="btn btn-primary bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700">
        新規作成
      </button>
    </div>
  </div>

  <div class="bg-white shadow-md rounded-lg overflow-x-auto">
    <table class="min-w-full leading-normal">
      <thead>
        <tr>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">大会名</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">年度</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">シーズン</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">期間</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">ステータス</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider min-w-[120px]">アンケート操作</th>
          <th class="px-5 py-3 border-b-2 border-gray-200 bg-gray-100 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider min-w-[100px]">結果出力</th>
        </tr>
      </thead>
      <tbody>
        {#each events as event}
          <tr onclick={() => openEditModal(event)} class="cursor-pointer hover:bg-gray-50 transition-colors">
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm">{event.name}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm">{event.year}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm">{event.season === 'spring' ? '春' : '秋'}</td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm whitespace-nowrap">
              {event.start_date ? new Date(event.start_date).toLocaleDateString() : ''} - 
              {event.end_date ? new Date(event.end_date).toLocaleDateString() : ''}
            </td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm">
              {#if event.status === 'active'}
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">開催中</span>
              {:else if event.status === 'archived'}
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-gray-100 text-gray-800">アーカイブ</span>
              {:else}
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">予定</span>
              {/if}
            </td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm" onclick={(e) => e.stopPropagation()}>
              <div class="flex flex-col space-y-2">
                {#if event.season === 'spring'}
                  <div class="flex flex-col">
                    <label class="text-xs text-gray-600 mb-1">点数インポート(CSV)
                      <input type="file" accept=".csv" class="text-xs mt-1 block" onchange={(e) => handleCsvUpload(event.id, e)} />
                    </label>
                  </div>
                {/if}
              </div>
            </td>
            <td class="px-5 py-5 border-b border-gray-200 bg-transparent text-sm" onclick={(e) => e.stopPropagation()}>
              <div class="flex flex-col space-y-2">
                <button type="button" onclick={(e) => { e.stopPropagation(); downloadExport(event, 'csv'); }} class="text-xs bg-green-100 text-green-700 px-2 py-1 rounded hover:bg-green-200 w-fit">
                  CSV出力
                </button>
                <button type="button" onclick={(e) => { e.stopPropagation(); downloadExport(event, 'pdf'); }} class="text-xs bg-red-100 text-red-700 px-2 py-1 rounded hover:bg-red-200 w-fit">
                  PDF出力
                </button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

{#if showModal}
  <div class="fixed z-10 inset-0 overflow-y-auto">
    <div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <div class="fixed inset-0 transition-opacity" aria-hidden="true">
        <div class="absolute inset-0 bg-gray-500 opacity-75"></div>
      </div>

      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

      <div class="relative z-30 inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
        <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">{selectedEvent ? '大会編集' : '大会作成'}</h3>
          <div class="space-y-4">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700">大会名</label>
              <input type="text" id="name" bind:value={currentEvent.name} oninput={onNameInput} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="year" class="block text-sm font-medium text-gray-700">年度</label>
              <input type="number" id="year" bind:value={currentEvent.year} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="season" class="block text-sm font-medium text-gray-700">シーズン</label>
              <select id="season" bind:value={currentEvent.season} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                <option value="spring">春</option>
                <option value="autumn">秋</option>
              </select>
            </div>
            <div>
              <label for="start_date" class="block text-sm font-medium text-gray-700">開始日</label>
              <input type="date" id="start_date" bind:value={currentEvent.start_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="end_date" class="block text-sm font-medium text-gray-700">終了日</label>
              <input type="date" id="end_date" bind:value={currentEvent.end_date} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
            </div>
            <div>
              <label for="survey_url" class="block text-sm font-medium text-gray-700">アンケートURL</label>
              <div class="flex flex-col sm:flex-row sm:items-center space-y-2 sm:space-y-0 sm:space-x-3 mt-1">
                <input type="url" id="survey_url" bind:value={currentEvent.survey_url} placeholder="https://forms.gle/..." class="block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                {#if selectedEvent && currentEvent.survey_url}
                  <button type="button" onclick={() => notifySurvey(selectedEvent.id)} class="whitespace-nowrap text-xs bg-blue-100 text-blue-700 px-3 py-2 rounded hover:bg-blue-200 border border-blue-200 font-medium">
                    通知を送信
                  </button>
                {/if}
              </div>
            </div>
            <div>
              <label for="status" class="block text-sm font-medium text-gray-700">ステータス</label>
              <select id="status" bind:value={currentEvent.status} class="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
                <option value="upcoming">予定 (Upcoming)</option>
                <option value="active">開催中 (Active)</option>
                <option value="archived">アーカイブ (Archived)</option>
              </select>
            </div>
            <div class="flex items-center">
              <label class="flex items-center cursor-pointer">
                <input type="checkbox" id="hide_scores" bind:checked={currentEvent.hide_scores} class="sr-only">
                <div class="relative">
                  <div class="block w-14 h-8 rounded-full transition-colors" class:bg-indigo-600={currentEvent.hide_scores} class:bg-gray-300={!currentEvent.hide_scores}></div>
                  <div class="dot absolute top-1 w-6 h-6 bg-white rounded-full shadow transition-transform" class:translate-x-7={currentEvent.hide_scores} class:translate-x-1={!currentEvent.hide_scores}></div>
                </div>
                <span class="ml-3 text-sm font-medium text-gray-700">スコアを非表示にする</span>
              </label>
            </div>
          </div>
        </div>
        <div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
          <button onclick={handleSave} type="button" class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm">
            保存
          </button>
          <button onclick={closeModal} type="button" class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm">
            キャンセル
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
