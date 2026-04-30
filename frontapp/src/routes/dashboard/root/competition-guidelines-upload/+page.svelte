<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let events = $state([]);
  let selectedEventId = $state(null);
  let selectedEvent = $state(null);
  let pdfFile = $state(null);
  let pdfPreviewUrl = $state(null);
  let existingPdfUrl = $state(null);

  let guideDocuments = $state([]);
  let guideTitle = $state('');
  let guideDescription = $state('');
  let guidePdfFile = $state(null);
  let guidePdfPreviewUrl = $state(null);

  let isUploading = $state(false);
  let message = $state('');
  let errorMessage = $state('');

  onMount(async () => {
    await Promise.all([fetchEvents(), fetchGuideDocuments()]);
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

  async function fetchGuideDocuments() {
    try {
      const response = await fetch('/api/root/guide-documents');
      if (!response.ok) {
        throw new Error('Failed to fetch guide documents');
      }
      const data = await response.json();
      guideDocuments = data.documents ?? [];
    } catch (error) {
      console.error('Failed to fetch guide documents:', error);
      errorMessage = '資料一覧の取得に失敗しました';
    }
  }

  async function loadEventDetails(eventId) {
    try {
      const response = await fetch(`/api/root/events`);
      if (!response.ok) {
        throw new Error('Failed to fetch event details');
      }
      const allEvents = await response.json();
      selectedEvent = allEvents.find((event) => event.id === eventId);
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

  function resetMessages() {
    message = '';
    errorMessage = '';
  }

  async function handleEventChange() {
    resetMessages();
    pdfFile = null;
    pdfPreviewUrl = null;
    existingPdfUrl = null;
    if (selectedEventId) {
      await loadEventDetails(selectedEventId);
    } else {
      selectedEvent = null;
    }
  }

  function validatePdfFile(file) {
    if (!file) return 'PDFファイルを選択してください';
    if (file.type !== 'application/pdf') return 'PDFファイルのみアップロードできます';
    if (file.size > 10 * 1024 * 1024) return 'ファイルサイズは10MB以下にしてください';
    return '';
  }

  function handleFileSelect(event) {
    const file = event.target.files[0];
    if (!file) return;

    const validationError = validatePdfFile(file);
    if (validationError) {
      errorMessage = validationError;
      event.target.value = '';
      return;
    }

    pdfFile = file;
    errorMessage = '';
    if (browser) {
      pdfPreviewUrl = URL.createObjectURL(file);
    }
  }

  function handleGuideFileSelect(event) {
    const file = event.target.files[0];
    if (!file) return;

    const validationError = validatePdfFile(file);
    if (validationError) {
      errorMessage = validationError;
      event.target.value = '';
      return;
    }

    guidePdfFile = file;
    errorMessage = '';
    if (browser) {
      guidePdfPreviewUrl = URL.createObjectURL(file);
    }
  }

  async function uploadPdf(file) {
    const formData = new FormData();
    formData.append('pdf', file);

    const uploadResponse = await fetch('/api/admin/pdfs', {
      method: 'POST',
      body: formData
    });

    if (!uploadResponse.ok) {
      const errorData = await uploadResponse.json();
      throw new Error(errorData.error || 'PDFのアップロードに失敗しました');
    }

    const uploadData = await uploadResponse.json();
    return uploadData.url;
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
    resetMessages();

    try {
      const pdfUrl = await uploadPdf(pdfFile);
      const updateResponse = await fetch(`/api/root/events/${selectedEventId}/competition-guidelines`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pdf_url: pdfUrl })
      });

      if (!updateResponse.ok) {
        const errorData = await updateResponse.json();
        throw new Error(errorData.error || '大会要項の更新に失敗しました');
      }

      message = '大会要項をアップロードしました';
      existingPdfUrl = pdfUrl;
      pdfFile = null;
      if (browser) {
        const fileInput = document.getElementById('pdf-file-input');
        if (fileInput) fileInput.value = '';
      }
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

    if (!confirm('大会要項を削除しますか？')) return;

    isUploading = true;
    resetMessages();

    try {
      const updateResponse = await fetch(`/api/root/events/${selectedEventId}/competition-guidelines`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ pdf_url: null })
      });

      if (!updateResponse.ok) {
        const errorData = await updateResponse.json();
        throw new Error(errorData.error || '大会要項の削除に失敗しました');
      }

      message = '大会要項を削除しました';
      existingPdfUrl = null;
      pdfPreviewUrl = null;
      await loadEventDetails(selectedEventId);
    } catch (error) {
      console.error('Delete error:', error);
      errorMessage = error.message || '削除に失敗しました';
    } finally {
      isUploading = false;
    }
  }

  async function handleGuideDocumentCreate() {
    if (!guideTitle.trim()) {
      errorMessage = '資料タイトルを入力してください';
      return;
    }
    if (!guidePdfFile) {
      errorMessage = '資料PDFを選択してください';
      return;
    }

    isUploading = true;
    resetMessages();

    try {
      const pdfUrl = await uploadPdf(guidePdfFile);
      const response = await fetch('/api/root/guide-documents', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: guideTitle.trim(),
          description: guideDescription.trim() || null,
          pdf_url: pdfUrl
        })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '資料の登録に失敗しました');
      }

      guideTitle = '';
      guideDescription = '';
      guidePdfFile = null;
      guidePdfPreviewUrl = null;
      if (browser) {
        const fileInput = document.getElementById('guide-pdf-file-input');
        if (fileInput) fileInput.value = '';
      }
      await fetchGuideDocuments();
      message = '資料を登録しました';
    } catch (error) {
      console.error('Create guide document error:', error);
      errorMessage = error.message || '資料の登録に失敗しました';
    } finally {
      isUploading = false;
    }
  }

  async function handleGuideDocumentDelete(id) {
    if (!confirm('この資料を削除しますか？')) return;

    isUploading = true;
    resetMessages();

    try {
      const response = await fetch(`/api/root/guide-documents/${id}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || '資料の削除に失敗しました');
      }

      await fetchGuideDocuments();
      message = '資料を削除しました';
    } catch (error) {
      console.error('Delete guide document error:', error);
      errorMessage = error.message || '資料の削除に失敗しました';
    } finally {
      isUploading = false;
    }
  }
</script>

<svelte:head>
  <title>資料管理 | Dashboard</title>
</svelte:head>

<div class="space-y-6">
  <header class="space-y-2">
    <h1 class="text-2xl font-semibold text-gray-900">資料管理</h1>
    <p class="text-sm text-gray-600">
      大会要項PDFと、資料ページに掲載する任意のPDF資料を登録・削除できます。
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
        <h2 class="text-lg font-semibold text-gray-900">大会要項PDF</h2>
        <p class="mt-1 text-sm text-gray-600">
          資料ページの「大会要項」として表示されるPDFです。
        </p>
      </div>

      <div>
        <label for="event-select" class="block text-sm font-medium text-gray-700 mb-2">
          大会選択
        </label>
        <select
          id="event-select"
          bind:value={selectedEventId}
          onchange={handleEventChange}
          class="mt-1 block w-full rounded-md border-gray-300 py-2 pl-3 pr-10 text-base focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm"
        >
          <option value={null}>大会を選択してください</option>
          {#each events as event (event.id)}
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
            onchange={handleFileSelect}
            class="block w-full text-sm text-gray-500 file:mr-4 file:rounded-md file:border-0 file:bg-indigo-50 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-indigo-700 hover:file:bg-indigo-100"
            disabled={isUploading}
          />
          <p class="mt-1 text-xs text-gray-500">PDFファイルのみ（最大10MB）</p>
        </div>

        {#if pdfPreviewUrl || existingPdfUrl}
          <div class="border rounded-md h-96 overflow-hidden">
            <embed src={pdfPreviewUrl || existingPdfUrl} type="application/pdf" width="100%" height="100%" />
          </div>
        {/if}

        <div class="flex gap-4">
          <button
            onclick={handleUpload}
            disabled={!pdfFile || isUploading}
            class="rounded-md bg-indigo-600 px-4 py-2 font-semibold text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:bg-gray-400"
          >
            {isUploading ? 'アップロード中...' : 'アップロード'}
          </button>
          {#if existingPdfUrl}
            <button
              onclick={handleDelete}
              disabled={isUploading}
              class="rounded-md bg-red-600 px-4 py-2 font-semibold text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-gray-400"
            >
              {isUploading ? '削除中...' : '削除'}
            </button>
          {/if}
        </div>
      {/if}
    </div>
  </section>

  <section class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">任意資料PDF</h2>
        <p class="mt-1 text-sm text-gray-600">
          ガイド、会場案内、運営資料などを資料ページに掲載できます。
        </p>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label for="guide-title" class="block text-sm font-medium text-gray-700 mb-2">
            資料タイトル
          </label>
          <input
            id="guide-title"
            type="text"
            bind:value={guideTitle}
            placeholder="例: 会場案内"
            class="block w-full rounded-md border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
          />
        </div>
        <div>
          <label for="guide-pdf-file-input" class="block text-sm font-medium text-gray-700 mb-2">
            資料PDF
          </label>
          <input
            id="guide-pdf-file-input"
            type="file"
            accept=".pdf"
            onchange={handleGuideFileSelect}
            class="block w-full text-sm text-gray-500 file:mr-4 file:rounded-md file:border-0 file:bg-indigo-50 file:px-4 file:py-2 file:text-sm file:font-semibold file:text-indigo-700 hover:file:bg-indigo-100"
            disabled={isUploading}
          />
        </div>
      </div>

      <div>
        <label for="guide-description" class="block text-sm font-medium text-gray-700 mb-2">
          説明
        </label>
        <textarea
          id="guide-description"
          rows="3"
          bind:value={guideDescription}
          placeholder="資料の用途や補足を入力"
          class="block w-full rounded-md border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
        ></textarea>
      </div>

      {#if guidePdfPreviewUrl}
        <div class="border rounded-md h-96 overflow-hidden">
          <embed src={guidePdfPreviewUrl} type="application/pdf" width="100%" height="100%" />
        </div>
      {/if}

      <div class="flex justify-end">
        <button
          onclick={handleGuideDocumentCreate}
          disabled={!guideTitle.trim() || !guidePdfFile || isUploading}
          class="rounded-md bg-indigo-600 px-4 py-2 font-semibold text-white hover:bg-indigo-700 disabled:cursor-not-allowed disabled:bg-gray-400"
        >
          {isUploading ? '登録中...' : '資料を登録'}
        </button>
      </div>

      <div class="space-y-3 border-t border-gray-200 pt-6">
        <h3 class="text-md font-semibold text-gray-900">登録済み資料</h3>
        {#if guideDocuments.length === 0}
          <p class="text-sm text-gray-500">まだ資料は登録されていません。</p>
        {:else}
          <div class="space-y-3">
            {#each guideDocuments as document (document.id)}
              <div class="flex flex-col gap-3 rounded-lg border border-gray-200 p-4 md:flex-row md:items-start md:justify-between">
                <div class="space-y-1">
                  <p class="font-semibold text-gray-900">{document.title}</p>
                  {#if document.description}
                    <p class="text-sm text-gray-600">{document.description}</p>
                  {/if}
                  <a
                    href={document.pdf_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-sm font-medium text-indigo-600 hover:text-indigo-700"
                  >
                    PDFを開く
                  </a>
                </div>
                <button
                  onclick={() => handleGuideDocumentDelete(document.id)}
                  disabled={isUploading}
                  class="rounded-md bg-red-50 px-3 py-2 text-sm font-semibold text-red-700 hover:bg-red-100 disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-400"
                >
                  削除
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </section>
</div>
