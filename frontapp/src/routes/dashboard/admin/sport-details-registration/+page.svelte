<!--
  This file is for the Sport Details Registration page.
  It allows admins to register and update sport details, including overview, rules, and match start times.
-->
<script>
  import { onMount } from 'svelte';
  import { createBracket } from 'bracketry';
  import { marked } from 'marked';

  let events = [];
  let sports = [];
  let selectedEventId = null;
  let selectedSportId = null;
  let sportDetails = { description: '', rules: '' };
  let tournaments = [];
  let currentTournamentData = null;
  let selectedTournamentId = null;
  let selectedTournament = null;
  let bulkStartTime = '';
  let matchStartTimes = {};
  let rulesTextarea;

  onMount(async () => {
    // Fetch events
    const res = await fetch('/api/admin/events');
    if (res.ok) {
      events = await res.json();
    }
    renderBracket();
  });

  async function fetchSports(eventId) {
    const res = await fetch('/api/admin/allsports');
    if (res.ok) {
      sports = await res.json();
    }
  }

  async function fetchSportDetails(eventId, sportId) {
    const res = await fetch(`/api/admin/events/${eventId}/sports/${sportId}/details`);
    if (res.ok) {
      const details = await res.json();
      sportDetails = { description: details.description, rules: details.rules };
    } else {
      sportDetails = { description: '', rules: '' };
    }
  }

  async function fetchTournaments(eventId) {
    const res = await fetch(`/api/admin/events/${eventId}/tournaments`);
    if (res.ok) {
      const fetched = await res.json();
      console.log("fetched:", fetched);
      tournaments = fetched.map(t => {
        let data = t.data;
        if (typeof data === 'string') {
          try {
            data = JSON.parse(data);
          } catch (e) {
            data = null;
          }
        }
        return { ...t, data };
      });
      selectedTournamentId = null;
      selectedTournament = null;
    }
  }

  function updateSelectedTournament() {
    selectedTournament = tournaments.find(t => t.id == selectedTournamentId) || null;
    renderBracket();
  }

  function renderBracket() {
    setTimeout(() => {
      const wrapper = document.getElementById('bracket-container');
      if (wrapper) {
        wrapper.innerHTML = '';
        if (selectedTournament && selectedTournament.data) {
          createBracket(selectedTournament.data, wrapper);
        } else {
          wrapper.innerHTML = '<p>このトーナメント情報はありません。</p>';
        }
      }
    }, 0);
  }

  async function handleEventChange(e) {
    selectedEventId = e.target.value;
    selectedSportId = null;
    sports = [];
    tournaments = [];
    currentTournamentData = null;
    selectedTournamentId = null;
    selectedTournament = null;
    const wrapper = document.getElementById('bracket-container');
    if(wrapper) wrapper.innerHTML = '';

    if (selectedEventId) {
      await fetchSports(selectedEventId);
      await fetchTournaments(selectedEventId);
    }
  }

  async function handleSportChange(e) {
    selectedSportId = e.target.value;
    if (selectedSportId) {
      await fetchSportDetails(selectedEventId, selectedSportId);
      // 選択された競技の名前を取得
      const selectedSport = sports.find(s => s.id == selectedSportId);
      const sportName = selectedSport ? selectedSport.name : '';

      // Ensure tournaments are loaded before updating
      if (tournaments.length === 0 && selectedEventId) {
        await fetchTournaments(selectedEventId);
      }
      // 競技選択時に該当トーナメントを選択
      const t = tournaments.find(t => t.sport_id == selectedSportId);
      selectedTournamentId = t ? t.id : null;
      if (t) {
        t.display_name = `${sportName} Tournament`;
      }
      updateSelectedTournament();
    } else {
      selectedTournamentId = null;
      selectedTournament = null;
      const wrapper = document.getElementById('bracket-container');
      if(wrapper) wrapper.innerHTML = '';
    }
  }

  async function handleSave() {
    if (!selectedEventId || !selectedSportId) {
      alert('Please select an event and a sport.');
      return;
    }
    const res = await fetch(`/api/admin/events/${selectedEventId}/sports/${selectedSportId}/details`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(sportDetails)
    });

    if (res.ok) {
      alert('Sport details saved successfully');
    }
    else {
      alert('Failed to save sport details');
    }
  }

  function updateMatchTimeLocally(matchId, newTime) {
    matchStartTimes[matchId] = newTime;
  }

  async function handleSaveAllMatchTimes() {
    const updates = Object.entries(matchStartTimes);
    if (updates.length === 0) {
        alert('変更された試合時間がありません。');
        return;
    }

    const confirmation = confirm(`${updates.length}件の試合開始時間を更新します。よろしいですか？`);
    if (!confirmation) {
        return;
    }

    try {
        const updatePromises = updates.map(([matchId, startTime]) => {
            return fetch(`/api/admin/matches/${matchId}/start-time`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ start_time: formatTime(startTime) })
            });
        });

        const results = await Promise.all(updatePromises);
        const failedUpdates = results.filter(res => !res.ok);

        if (failedUpdates.length > 0) {
            alert(`${failedUpdates.length}件の試合の更新に失敗しました。`);
        } else {
            alert('すべての試合開始時間を正常に更新しました。');
            matchStartTimes = {}; // Reset after successful save
        }
    } catch (error) {
        console.error('Error during bulk match time save:', error);
        alert('試合開始時間の一括保存中にエラーが発生しました。');
    } finally {
        if (selectedEventId) {
            await fetchTournaments(selectedEventId);
            updateSelectedTournament();
        }
    }
  }

  async function handleBulkTimeUpdate() {
    if (!bulkStartTime) {
      alert('一括設定する開始時間を入力してください。');
      return;
    }
    if (!selectedTournament?.data?.matches) {
      alert('トーナメントが選択されていません。');
      return;
    }

    const formattedTime = new Date(bulkStartTime).toLocaleString('ja-JP');
    const confirmation = confirm(`すべての試合の開始時間を ${formattedTime} に設定します。よろしいですか？`);
    if (!confirmation) {
      return;
    }

    try {
      const updatePromises = selectedTournament.data.matches.map(match => {
        return fetch(`/api/admin/matches/${match.id}/start-time`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ start_time: formatTime(bulkStartTime) })
        });
      });

      const results = await Promise.all(updatePromises);

      const failedUpdates = results.filter(res => !res.ok);

      if (failedUpdates.length > 0) {
        alert(`${failedUpdates.length}件の試合の更新に失敗しました。`);
      } else {
        alert('すべての試合の開始時間を更新しました。');
      }

    } catch (error) {
      console.error('Error during bulk update:', error);
      alert('一括更新中にエラーが発生しました。');
    } finally {
      // Refetch data regardless of success or failure to get latest state
      if (selectedEventId) {
        await fetchTournaments(selectedEventId);
        updateSelectedTournament();
      }
    }
  }

  function formatTime(isoString) {
    if (!isoString) return '';
    try {
        const date = new Date(isoString);
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        return `${day}日 ${hours}:${minutes}`;
    } catch (e) {
        console.error("Invalid date format for startTime:", isoString);
        return '';
    }
  }

  async function handlePaste(event) {
    const items = (event.clipboardData || event.originalEvent.clipboardData).items;
    for (const item of items) {
      if (item.kind === 'file' && item.type.startsWith('image/')) {
        event.preventDefault();
        const blob = item.getAsFile();
        const formData = new FormData();
        formData.append('image', blob);

        try {
          const res = await fetch('/api/admin/images', {
            method: 'POST',
            body: formData
          });

          if (res.ok) {
            const data = await res.json();
            const imageUrl = data.url;
            const textarea = event.target;
            const start = textarea.selectionStart;
            const end = textarea.selectionEnd;
            const text = textarea.value;
            const before = text.substring(0, start);
            const after = text.substring(end, text.length);
            const markdownImage = `\n![pasted-image](${imageUrl})\n`;
            sportDetails.rules = before + markdownImage + after;
            // Move cursor after the inserted image
            setTimeout(() => {
              textarea.selectionStart = textarea.selectionEnd = start + markdownImage.length;
            }, 0);
          } else {
            const error = await res.json();
            alert(`Image upload failed: ${error.error}`);
          }
        } catch (error) {
          console.error('Error uploading image:', error);
          alert('Image upload failed.');
        }
      }
    }
  }

  async function handleFileSelect(event) {
    const file = event.target.files[0];
    if (!file || !file.type.startsWith('image/')) {
      return;
    }

    const formData = new FormData();
    formData.append('image', file);

    try {
      const res = await fetch('/api/admin/images', {
        method: 'POST',
        body: formData
      });

      if (res.ok) {
        const data = await res.json();
        const imageUrl = data.url;
        const altText = file.name.split('.')[0]; // use filename without extension as alt text
        const markdownImage = `\n![${altText}](${imageUrl})\n`;

        const start = rulesTextarea.selectionStart;
        const end = rulesTextarea.selectionEnd;
        const text = rulesTextarea.value;
        const before = text.substring(0, start);
        const after = text.substring(end, text.length);
        
        sportDetails.rules = before + markdownImage + after;

        // Move cursor after the inserted image
        setTimeout(() => {
          rulesTextarea.selectionStart = rulesTextarea.selectionEnd = start + markdownImage.length;
          rulesTextarea.focus();
        }, 0);

      } else {
        const error = await res.json();
        alert(`Image upload failed: ${error.error}`);
      }
    } catch (error) {
      console.error('Error uploading image:', error);
      alert('Image upload failed.');
    }

    // Reset file input
    event.target.value = '';
  }
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-4">競技詳細情報登録</h1>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
    <div>
      <label for="event-select" class="block text-sm font-medium text-gray-700">大会選択</label>
      <select id="event-select" on:change={handleEventChange} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
        <option value="">大会を選択してください</option>
        {#each events as event}
          <option value={event.id}>{event.name}</option>
        {/each}
      </select>
    </div>
    <div>
      <label for="sport-select" class="block text-sm font-medium text-gray-700">競技選択</label>
      <select id="sport-select" on:change={handleSportChange} disabled={!selectedEventId} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
        <option value="">競技を選択してください</option>
        {#each sports as sport}
          <option value={sport.id}>{sport.name}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if selectedSportId}
    <div class="mb-4">
      <h2 class="text-xl font-semibold mb-2">競技概要</h2>
      <textarea bind:value={sportDetails.description} rows="4" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 mt-1 block w-full sm:text-sm border border-gray-300 rounded-md"></textarea>
    </div>

    <div class="mb-4">
      <h2 class="text-xl font-semibold mb-2">ルール詳細 (Markdown)</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <textarea bind:value={sportDetails.rules} bind:this={rulesTextarea} on:paste={handlePaste} rows="10" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 mt-1 block w-full sm:text-sm border border-gray-300 rounded-md"></textarea>
        <div class="prose border p-4 rounded-md">
          {@html marked(sportDetails.rules || '')}
        </div>
      </div>
      <input type="file" id="image-upload" accept="image/*" class="hidden" on:change={handleFileSelect}>
      <button on:click={() => document.getElementById('image-upload').click()} class="mb-2 px-3 py-1 bg-gray-200 text-gray-800 rounded-md text-sm">
        画像アップロード
      </button>
    </div>

    <div class="flex justify-end mb-4">
      <button on:click={handleSave} class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">保存</button>
    </div>

    <div class="mb-4">
      <h2 class="text-xl font-semibold mb-2">トーナメント情報</h2>
      <div id="bracket-container">
        {#if !selectedSportId}
          <p>競技を選択してください。</p>
        {:else if !selectedTournament}
          <p>この競技のトーナメント情報はありません。</p>
        {/if}
      </div>
    </div>

    <div class="mb-4">
        <h2 class="text-xl font-semibold mb-2">試合開始時間</h2>
        {#if selectedTournament?.data?.matches}
            <div class="flex items-center gap-4 mb-4">
              <input type="datetime-local" bind:value={bulkStartTime} class="border rounded px-2 py-1">
              <button on:click={handleBulkTimeUpdate} class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700">全試合に適用</button>
              <button on:click={handleSaveAllMatchTimes} class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">設定した時間をすべて保存</button>
            </div>
            <div class="space-y-2">
                {#each selectedTournament.data.matches as match}
                    <div class="flex items-center justify-between p-2 border rounded">
                        <span class="font-medium">Round {match.roundIndex + 1}, Match {match.order + 1}</span>
                        <div class="flex items-center gap-4">
                            <span class="text-sm text-gray-600">
                                {#if match.matchStatus}
                                    開始日時: {match.matchStatus}
                                {/if}
                            </span>
                            <input 
                                type="datetime-local" 
                                class="border rounded px-2 py-1"
                                value={match.startTime ? match.startTime.slice(0, 16) : ''} 
                                on:change={(e) => updateMatchTimeLocally(match.id, e.target.value)}
                            />
                        </div>
                    </div>
                {/each}
            </div>
        {:else}
            <p class="text-gray-600">トーナメントを選択してください。</p>
        {/if}
    </div>
  {/if}
</div>
