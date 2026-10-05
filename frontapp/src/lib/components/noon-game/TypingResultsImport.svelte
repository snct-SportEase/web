<script>
 let { sessionId, typingImportStatus, onImported } = $props();
 let typingFile = $state(null);
 let importingTypingResults = $state(false);
 let typingImportError = $state('');
  async function importTypingResults(replace = false) {
    if (!sessionId || !typingFile) { alert('JSON ファイルを選択してください。'); return; }
    importingTypingResults = true;
    typingImportError = '';
    try {
      const form = new FormData();
      form.append('file', typingFile);
      const response = await fetch(`/api/root/noon-game/sessions/${sessionId}/typing-system/import${replace ? '?replace=true' : ''}`, { method: 'POST', body: form });
      const detail = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(detail?.error || '結果のインポートに失敗しました');
      alert(detail?.status === 'already_imported' ? 'この結果はすでにインポート済みです。' : '競技タイピング結果を確定しました。');
      typingFile = null;
      await onImported();
    } catch (err) {
      typingImportError = err.message;
      alert(err.message);
    }
    finally { importingTypingResults = false; }
  }
</script>

      <section class="bg-white shadow rounded-lg p-6 space-y-4">
        <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">競技タイピング結果のインポート</h2>
        <p class="text-sm text-gray-600">typing-results-v1 の JSON を読み込み、順位点を大会得点へ反映します。</p>
        {#if typingImportError}
          <p class="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-700">検証失敗: {typingImportError}</p>
        {:else if typingImportStatus === 'finalized'}
          <p class="rounded border border-green-300 bg-green-50 px-3 py-2 text-sm text-green-700">確定済みです。訂正時は置換インポートを使用してください。</p>
        {:else}
          <p class="rounded border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-800">結果は未インポートです。</p>
        {/if}
        <div class="flex flex-wrap items-center gap-3">
          <input type="file" accept="application/json,.json" onchange={(event) => { typingFile = event.currentTarget.files?.[0] ?? null; }} />
          <button class="px-4 py-2 bg-indigo-600 text-white rounded disabled:opacity-50" onclick={() => importTypingResults(false)} disabled={!typingFile || importingTypingResults}>{importingTypingResults ? 'インポート中…' : '結果をインポート'}</button>
          <button class="px-4 py-2 border border-orange-500 text-orange-700 rounded disabled:opacity-50" onclick={() => importTypingResults(true)} disabled={!typingFile || importingTypingResults}>置換インポート</button>
        </div>
      </section>
