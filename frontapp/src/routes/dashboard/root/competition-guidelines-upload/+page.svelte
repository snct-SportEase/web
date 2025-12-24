<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let events = [];
  let selectedEventId = null;
  let selectedEvent = null;
  let pdfFile = null;
  let pdfPreviewUrl = null;
  let existingPdfUrl = null;
  let isUploading = false;
  let message = '';
  let errorMessage = '';

  onMount(async () => {
    await fetchEvents();
    // アクティブなイベントを取得
    try {
      const response = await fetch('/api/events/active');
      if (response.ok) {
        const data = await response.json();
        if (data.event_id) {
          selectedEventId = data.event_id;
          await loadEventDetails(data.event_id);
        }
      }
    } catch (error) {
      console.error('Failed to fetch active event:', error);
    }
  });

  async function fetchEvents() {
    try {
      const response = await fetch('/api/root/events');
      if (!response.ok) {
        throw new Error('Failed to fetch events');
      }
      events = await response.json();
    } catch (error) {
      console.error('Failed to fetch events:', error);
      errorMessage = '大会情報の取得に失敗しました';
    }
  }

  async function loadEventDetails(eventId) {
    try {
      const response = await fetch(`/api/root/events`);
      if (!response.ok) {
        throw new Error('Failed to fetch event details');
      }
      const allEvents = await response.json();
      selectedEvent = allEvents.find(e => e.id === eventId);
      if (selectedEvent && selectedEvent.competition_guidelines_pdf_url) {
        existingPdfUrl = selectedEvent.competition_guidelines_pdf_url;
        pdfPreviewUrl = existingPdfUrl;
      } else {
        existingPdfUrl = null;
        pdfPreviewUrl = null;
      }
    } catch (error) {
      console.error('Failed to load event details:', error);
      errorMessage = '大会情報の読み込みに失敗しました';
    }
  }

  async function handleEventChange() {
    message = '';
    errorMessage = '';
    pdfFile = null;
    pdfPreviewUrl = null;
    existingPdfUrl = null;
    if (selectedEventId) {
      await loadEventDetails(selectedEventId);
    } else {
      selectedEvent = null;
    }
  }

  function handleFileSelect(event) {
    const file = event.target.files[0];
    if (!file) {
      return;
    }

    if (file.type !== 'application/pdf') {
      errorMessage = 'PDFファイルのみアップロードできます';
      event.target.value = '';
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      errorMessage = 'ファイルサイズは10MB以下にしてください';
      event.target.value = '';
      return;
    }

    pdfFile = file;
    errorMessage = '';

    // プレビュー用のURLを生成
    if (browser) {
      pdfPreviewUrl = URL.createObjectURL(file);
    }
  }

  async function handleUpload() {
    if (!selectedEventId) {
      errorMessage = '大会を選択してください';
      return;
    }

    if (!pdfFile) {
      errorMessage = 'PDFファイルを選択してください';
      return;
    }

    isUploading = true;
    message = '';
    errorMessage = '';

    try {
      // まずPDFをアップロード
      const formData = new FormData();
      formData.append('pdf', pdfFile);

      const uploadResponse = await fetch('/api/admin/pdfs', {
        method: 'POST',
        body: formData
      });

      if (!uploadResponse.ok) {
        const errorData = await uploadResponse.json();
        throw new Error(errorData.error || 'PDFのアップロードに失敗しました');
      }

      const uploadData = await uploadResponse.json();
      const pdfUrl = uploadData.url;

      // 次にイベントの競技要項URLを更新
      const updateResponse = await fetch(`/api/root/events/${selectedEventId}/competition-guidelines`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          pdf_url: pdfUrl
        })
      });

      if (!updateResponse.ok) {
        const errorData = await updateResponse.json();
        throw new Error(errorData.error || '競技要項の更新に失敗しました');
      }

      message = '競技要項をアップロードしました';
      existingPdfUrl = pdfUrl;
      pdfFile = null;
      
      // ファイル入力をリセット
      if (browser) {
        const fileInput = document.getElementById('pdf-file-input');
        if (fileInput) {
          fileInput.value = '';
        }
      }

      // イベント情報を再読み込み
      await loadEventDetails(selectedEventId);
    } catch (error) {
      console.error('Upload error:', error);
      errorMessage = error.message || 'アップロードに失敗しました';
    } finally {
      isUploading = false;
    }
  }

  async function handleDelete() {
    if (!selectedEventId) {
      errorMessage = '大会を選択してください';
      return;
    }

    if (!confirm('競技要項を削除しますか？')) {
      return;
    }

    isUploading = true;
    message = '';
    errorMessage = '';

    try {
      const updateResponse = await fetch(`/api/root/events/${selectedEventId}/competition-guidelines`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          pdf_url: null
        })
      });

      if (!updateResponse.ok) {
        const errorData = await updateResponse.json();
        throw new Error(errorData.error || '競技要項の削除に失敗しました');
      }

      message = '競技要項を削除しました';
      existingPdfUrl = null;
      pdfPreviewUrl = null;
      
      // イベント情報を再読み込み
      await loadEventDetails(selectedEventId);
    } catch (error) {
      console.error('Delete error:', error);
      errorMessage = error.message || '削除に失敗しました';
    } finally {
      isUploading = false;
    }
  }
</script>

<svelte:head>
  <title>競技要項アップロード | Dashboard</title>
</svelte:head>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">競技要項アップロード</h1>
    <p class="text-sm text-gray-600">
      大会の競技要項PDFをアップロードします。アップロードされた競技要項は資料ページで確認できます。
    </p>
  </header>

  {#if errorMessage}
    <div class="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
      {errorMessage}
    </div>
  {/if}

  {#if message}
    <div class="rounded-md border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
      {message}
    </div>
  {/if}

  <section class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
    <div class="space-y-6">
      <div>
        <label for="event-select" class="block text-sm font-medium text-gray-700 mb-2">
          大会選択
        </label>
        <select
          id="event-select"
          bind:value={selectedEventId}
          on:change={handleEventChange}
          class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
        >
          <option value={null}>大会を選択してください</option>
          {#each events as event}
            <option value={event.id}>{event.name}</option>
          {/each}
        </select>
      </div>

      {#if selectedEvent}
        <div>
          <label for="pdf-file-input" class="block text-sm font-medium text-gray-700 mb-2">
            PDFファイル選択
          </label>
          <input
            id="pdf-file-input"
            type="file"
            accept=".pdf"
            on:change={handleFileSelect}
            class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
            disabled={isUploading}
          />
          <p class="mt-1 text-xs text-gray-500">
            PDFファイルのみ（最大10MB）
          </p>
        </div>

        {#if pdfPreviewUrl || existingPdfUrl}
          <div class="border rounded-md h-96">
            <embed
              src={pdfPreviewUrl || existingPdfUrl}
              type="application/pdf"
              width="100%"
              height="100%"
            />
          </div>
        {/if}

        <div class="flex gap-4">
          <button
            on:click={handleUpload}
            disabled={!pdfFile || isUploading}
            class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:bg-gray-400 disabled:cursor-not-allowed font-semibold"
          >
            {isUploading ? 'アップロード中...' : 'アップロード'}
          </button>
          {#if existingPdfUrl}
            <button
              on:click={handleDelete}
              disabled={isUploading}
              class="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:bg-gray-400 disabled:cursor-not-allowed font-semibold"
            >
              {isUploading ? '削除中...' : '削除'}
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </section>
</div>

