<!--
  This file is for the Sport Details Registration page.
  It allows admins to register and update sport details, including overview, rules, and match start times.
-->
<script>
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { marked } from 'marked';
  import SafeHtml from '$lib/components/SafeHtml.svelte';

  const colorExtension = {
    name: 'colorText',
    level: 'inline',
    start(src) {
      return src.match(/##|#color\(/)?.index;
    },
    tokenizer(src) {
      const redRule = /^##(.*?)##/;
      let match = redRule.exec(src);
      if (match) {
        return {
          type: 'colorText',
          raw: match[0],
          text: this.lexer.inlineTokens(match[1]),
          color: 'red'
        };
      }

      const customColorRule = /^#color\((#[0-9a-fA-F]{3,6}),\s*(.*?)\)/;
      match = customColorRule.exec(src);
      if (match) {
        return {
          type: 'colorText',
          raw: match[0],
          text: this.lexer.inlineTokens(match[2]),
          color: match[1]
        };
      }
    },
    renderer(token) {
      return `<span style="color: ${token.color};">${this.parser.parseInline(token.text)}</span>`;
    }
  };

  marked.use({ extensions: [colorExtension] });

  let sports = [];
  let selectedEventId = null;
  let selectedSportId = null;
  let sportDetails = { description: '', rules: '', rules_type: 'markdown', rules_pdf_url: null };
  let capacityMode = 'bulk'; // 'bulk' or 'per-class'
  let minCapacity = null;
  let maxCapacity = null;
  let classes = [];
  let classCapacities = {}; // { classId: { min: number | null, max: number | null } }
  
  // Rainy mode capacity settings
  let rainyModeCapacityMode = 'bulk'; // 'bulk' or 'per-class'
  let rainyModeMinCapacity = null;
  let rainyModeMaxCapacity = null;
  let rainyModeClassCapacities = {}; // { classId: { min: number | null, max: number | null } }
  let tournaments = [];
  let selectedTournamentId = null;
  let selectedTournament = null;
  let allMatchesForSport = []; // 選択されたスポーツの全てのトーナメント（本戦+敗者戦）の試合
  let bulkStartTime = '';
  let matchStartTimes = {};
  let bulkRainyModeStartTime = '';
  let matchRainyModeStartTimes = {};
  let rulesTextarea;
  let previewDiv;
  let markdownPreviewHtml = '';
  $: markdownPreviewHtml = sportDetails.rules_type === 'markdown'
    ? marked.parse(sportDetails.rules || '')
    : '';
  let selectedPdfFile = null;
  let pdfPreviewUrl = null;
  let activeEventName = '';
  let customColor = '#000000';

  onMount(async () => {
    // Fetch active event
    const res = await fetch('/api/events/active');
    if (res.ok) {
      const data = await res.json();
      if (data.event_id) {
        selectedEventId = data.event_id;
        activeEventName = data.event_name;
        await fetchSports();
        await fetchClasses();
        await fetchTournaments(selectedEventId);
      }
    }
    await renderBracket();
  });

  async function fetchSports() {
    const res = await fetch('/api/admin/allsports');
    if (res.ok) {
      sports = await res.json();
    }
  }

  async function fetchClasses() {
    if (!selectedEventId) return;
    const res = await fetch('/api/admin/class-team/managed-class');
    if (res.ok) {
      classes = await res.json();
      // Initialize class capacities
      if (selectedSportId) {
        await fetchClassCapacities();
      }
    }
  }

  async function fetchClassCapacities() {
    if (!selectedEventId || !selectedSportId) return;
    
    // Fetch teams for this sport to get class information and capacities
    const teamsRes = await fetch(`/api/root/sports/${selectedSportId}/teams`);
    if (teamsRes.ok) {
      const teams = await teamsRes.json();
      // Initialize class capacities from teams
      classCapacities = {};
      classes.forEach(cls => {
        const team = teams.find(t => t.class_id === cls.id && t.event_id === selectedEventId);
        if (team) {
          classCapacities[cls.id] = {
            min: team.min_capacity ?? null,
            max: team.max_capacity ?? null
          };
        } else {
          classCapacities[cls.id] = {
            min: null,
            max: null
          };
        }
      });
      classCapacities = { ...classCapacities }; // Trigger reactivity
    }
  }

  async function fetchSportDetails(eventId, sportId) {
    const res = await fetch(`/api/admin/events/${eventId}/sports/${sportId}/details`);
    if (res.ok) {
      const details = await res.json();
      sportDetails = {
        description: details.description || '',
        rules: details.rules || '',
        rules_type: details.rules_type || 'markdown',
        rules_pdf_url: details.rules_pdf_url || null
      };
      minCapacity = details.min_capacity ?? null;
      maxCapacity = details.max_capacity ?? null;
    } else {
      sportDetails = { description: '', rules: '', rules_type: 'markdown', rules_pdf_url: null };
      minCapacity = null;
      maxCapacity = null;
    }
    selectedPdfFile = null;
    pdfPreviewUrl = null;
    await fetchRainyModeSettings();
  }

  async function fetchTournaments(eventId) {
    const res = await fetch(`/api/admin/events/${eventId}/tournaments`);
    if (res.ok) {
      const fetched = await res.json();
      tournaments = fetched.map(t => {
        let data = t.data;
        if (typeof data === 'string') {
          try {
            data = JSON.parse(data);
          } catch {
            data = null;
          }
        }
        return { ...t, data };
      });
      selectedTournamentId = null;
      selectedTournament = null;
      // スポーツが選択されている場合は、全ての試合を更新
      if (selectedSportId) {
        updateAllMatchesForSport();
      }
    }
  }

  async function updateSelectedTournament() {
    selectedTournament = tournaments.find(t => t.id == selectedTournamentId) || null;
    updateAllMatchesForSport();
    await renderBracket();
  }

  // 選択されたスポーツの全てのトーナメント（本戦+敗者戦）の試合をまとめる
  function updateAllMatchesForSport() {
    if (!selectedSportId) {
      allMatchesForSport = [];
      return;
    }

    // 同じスポーツIDの全てのトーナメントを取得
    const sportTournaments = tournaments.filter(t => t.sport_id == selectedSportId);
    
    // 全てのトーナメントの試合をまとめる
    allMatchesForSport = [];
    sportTournaments.forEach(tournament => {
      if (tournament.data && tournament.data.matches) {
        tournament.data.matches.forEach(match => {
          // トーナメント名を追加して試合を識別しやすくする
          allMatchesForSport.push({
            ...match,
            tournamentName: tournament.name,
            tournamentId: tournament.id
          });
        });
      }
    });
  }

  async function renderBracket() {
    if (!browser) return;
    setTimeout(async () => {
      const wrapper = document.getElementById('bracket-container');
      if (wrapper) {
        wrapper.innerHTML = '';
        if (selectedTournament && selectedTournament.data) {
          try {
            const { createBracket } = await import('bracketry');
            createBracket(selectedTournament.data, wrapper);
          } catch (error) {
            console.error('Failed to load createBracket:', error);
            wrapper.innerHTML = '<p>ブラケットの読み込みに失敗しました。</p>';
          }
        } else {
          wrapper.innerHTML = '<p>このトーナメント情報はありません。</p>';
        }
      }
    }, 0);
  }

  async function handleSportChange(e) {
    selectedSportId = e.target.value;
    if (selectedSportId) {
      await fetchSportDetails(selectedEventId, selectedSportId);
      const selectedSport = sports.find(s => s.id == selectedSportId);
      const sportName = selectedSport ? selectedSport.name : '';

      if (classes.length === 0 && selectedEventId) {
        await fetchClasses();
      } else if (selectedEventId) {
        await fetchClassCapacities();
      }
      await fetchRainyModeSettings();

      if (tournaments.length === 0 && selectedEventId) {
        await fetchTournaments(selectedEventId);
      }
      // 本戦トーナメントを探す（" Tournament"が含まれ、敗者戦ではないもの）
      const mainTournament = tournaments.find(t => 
        t.sport_id == selectedSportId && 
        t.name.includes(' Tournament') && 
        !t.name.includes('敗者戦')
      );
      selectedTournamentId = mainTournament ? mainTournament.id : null;
      if (mainTournament) {
        mainTournament.display_name = `${sportName} Tournament`;
      }
      updateSelectedTournament();
    } else {
      selectedTournamentId = null;
      selectedTournament = null;
      if (browser) {
        const wrapper = document.getElementById('bracket-container');
        if(wrapper) wrapper.innerHTML = '';
      }
    }
  }

  function handlePdfFileSelect(e) {
    selectedPdfFile = e.target.files[0];
    if (selectedPdfFile) {
      pdfPreviewUrl = URL.createObjectURL(selectedPdfFile);
    } else {
      pdfPreviewUrl = null;
    }
  }

  async function uploadPdf() {
    if (!selectedPdfFile) return null;

    const formData = new FormData();
    formData.append('pdf', selectedPdfFile);

    const res = await fetch('/api/admin/pdfs', {
      method: 'POST',
      body: formData
    });

    if (res.ok) {
      const data = await res.json();
      return data.url;
    } else {
      alert('PDF upload failed');
      return null;
    }
  }

  async function handleSave() {
    if (!selectedEventId || !selectedSportId) {
      alert('Please select an event and a sport.');
      return;
    }

    let detailsToSave = {
      description: sportDetails.description,
      rules_type: sportDetails.rules_type,
      rules: sportDetails.rules_type === 'markdown' ? sportDetails.rules : null,
      rules_pdf_url: sportDetails.rules_type === 'pdf' ? sportDetails.rules_pdf_url : null,
    };

    if (detailsToSave.rules_type === 'pdf' && selectedPdfFile) {
      const newPdfUrl = await uploadPdf();
      if (newPdfUrl) {
        detailsToSave.rules_pdf_url = newPdfUrl;
      } else {
        return; // PDF upload failed, so we stop saving.
      }
    }

    const res = await fetch(`/api/admin/events/${selectedEventId}/sports/${selectedSportId}/details`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(detailsToSave)
    });

    if (res.ok) {
      alert('Sport details saved successfully');
      selectedPdfFile = null;
      await fetchSportDetails(selectedEventId, selectedSportId);
    } else {
      alert('Failed to save sport details');
    }
  }

  function updateMatchTimeLocally(matchId, newTime) {
    matchStartTimes[matchId] = newTime;
  }

  function updateMatchRainyModeTimeLocally(matchId, newTime) {
    matchRainyModeStartTimes[matchId] = newTime;
  }

  async function handleSaveAllRainyModeMatchTimes() {
    const updates = Object.entries(matchRainyModeStartTimes);
    if (updates.length === 0) {
        alert('変更された雨天時試合開始時間がありません。');
        return;
    }

    const confirmation = confirm(`${updates.length}件の雨天時試合開始時間を更新します。よろしいですか？`);
    if (!confirmation) {
        return;
    }

    try {
        const updatePromises = updates.map(([matchId, startTime]) => {
            return fetch(`/api/admin/matches/${matchId}/rainy-mode-start-time`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ rainy_mode_start_time: formatTime(startTime) })
            });
        });

        const results = await Promise.all(updatePromises);
        const failedUpdates = results.filter(res => !res.ok);

        if (failedUpdates.length > 0) {
            alert(`${failedUpdates.length}件の試合の更新に失敗しました。`);
        } else {
            alert('すべての雨天時試合開始時間を正常に更新しました。');
            matchRainyModeStartTimes = {}; // Reset after successful save
        }
    } catch (error) {
        console.error('Error during bulk rainy mode match time save:', error);
        alert('雨天時試合開始時間の一括保存中にエラーが発生しました。');
    } finally {
        if (selectedEventId) {
            await fetchTournaments(selectedEventId);
            updateSelectedTournament();
        }
    }
  }

  async function handleBulkRainyModeTimeUpdate() {
    if (!bulkRainyModeStartTime) {
      alert('一括設定する雨天時開始時間を入力してください。');
      return;
    }
    if (!allMatchesForSport || allMatchesForSport.length === 0) {
      alert('試合がありません。');
      return;
    }

    const formattedTime = new Date(bulkRainyModeStartTime).toLocaleString('ja-JP');
    const confirmation = confirm(`すべての試合（本戦+敗者戦）の雨天時開始時間を ${formattedTime} に設定します。よろしいですか？`);
    if (!confirmation) {
      return;
    }

    try {
      const updatePromises = allMatchesForSport.map(match => {
        return fetch(`/api/admin/matches/${match.id}/rainy-mode-start-time`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ rainy_mode_start_time: formatTime(bulkRainyModeStartTime) })
        });
      });

      const results = await Promise.all(updatePromises);

      const failedUpdates = results.filter(res => !res.ok);

      if (failedUpdates.length > 0) {
        alert(`${failedUpdates.length}件の試合の更新に失敗しました。`);
      } else {
        alert('すべての試合の雨天時開始時間を更新しました。');
      }

    } catch (error) {
      console.error('Error during bulk rainy mode update:', error);
      alert('雨天時一括更新中にエラーが発生しました。');
    } finally {
      // Refetch data regardless of success or failure to get latest state
      if (selectedEventId) {
        await fetchTournaments(selectedEventId);
        updateSelectedTournament();
      }
    }
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
    } catch {
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

  function syncScroll() {
    if (!rulesTextarea || !previewDiv) return;
    const percentage = rulesTextarea.scrollTop / (rulesTextarea.scrollHeight - rulesTextarea.clientHeight);
    previewDiv.scrollTop = percentage * (previewDiv.scrollHeight - previewDiv.clientHeight);
  }

  function applyMarkdown(prefix, suffix = '') {
    if (!rulesTextarea) return;

    const start = rulesTextarea.selectionStart;
    const end = rulesTextarea.selectionEnd;
    const selectedText = sportDetails.rules.substring(start, end);
    const newText = `${prefix}${selectedText}${suffix}`;
    
    sportDetails.rules = 
      sportDetails.rules.substring(0, start) + 
      newText + 
      sportDetails.rules.substring(end);

    // Wait for Svelte to update the textarea value
    setTimeout(() => {
      rulesTextarea.focus();
      if (selectedText) {
        // If text was selected, select the newly formatted text
        rulesTextarea.selectionStart = start;
        rulesTextarea.selectionEnd = start + newText.length;
      } else {
        // If no text was selected, place cursor in the middle
        rulesTextarea.selectionStart = start + prefix.length;
        rulesTextarea.selectionEnd = start + prefix.length;
      }
    }, 0);
  }

  function addLink() {
    if (!rulesTextarea) return;
    const url = prompt('Enter the URL:');
    if (url) {
      const start = rulesTextarea.selectionStart;
      const end = rulesTextarea.selectionEnd;
      const selectedText = sportDetails.rules.substring(start, end) || 'link text';
      const newText = `[${selectedText}](${url})`;
      
      sportDetails.rules = 
        sportDetails.rules.substring(0, start) + 
        newText + 
        sportDetails.rules.substring(end);

      setTimeout(() => {
        rulesTextarea.focus();
        if (selectedText === 'link text') {
          rulesTextarea.selectionStart = start + 1;
          rulesTextarea.selectionEnd = start + 1 + 'link text'.length;
        } else {
          rulesTextarea.selectionStart = start + newText.length;
          rulesTextarea.selectionEnd = start + newText.length;
        }
      }, 0);
    }
  }

  function addList() {
    if (!rulesTextarea) return;

    const start = rulesTextarea.selectionStart;
    const end = rulesTextarea.selectionEnd;
    const selectedText = sportDetails.rules.substring(start, end);

    const lines = selectedText.split('\n');
    const newText = lines.map(line => `- ${line}`).join('\n');

    sportDetails.rules = 
      sportDetails.rules.substring(0, start) + 
      newText + 
      sportDetails.rules.substring(end);
    
    setTimeout(() => {
      rulesTextarea.focus();
      rulesTextarea.selectionStart = start;
      rulesTextarea.selectionEnd = start + newText.length;
    }, 0);
  }

  function applyHeading(level) {
    if (!rulesTextarea) return;

    const prefix = '#'.repeat(level) + ' ';
    const cursorPos = rulesTextarea.selectionStart;
    const text = sportDetails.rules;
    
    let lineStart = text.lastIndexOf('\n', cursorPos - 1) + 1;
    let lineEnd = text.indexOf('\n', cursorPos);
    if (lineEnd === -1) lineEnd = text.length;

    const originalLine = text.substring(lineStart, lineEnd);
    let newLine;
    let change;

    // Remove any existing heading
    const lineWithoutHeading = originalLine.replace(/^#+\s*/, '');
    const currentPrefixLength = originalLine.length - lineWithoutHeading.length;

    if (originalLine.startsWith(prefix)) {
      // Toggle off
      newLine = lineWithoutHeading;
      change = -prefix.length;
    } else {
      // Apply new heading
      newLine = prefix + lineWithoutHeading;
      change = prefix.length - currentPrefixLength;
    }

    sportDetails.rules = text.substring(0, lineStart) + newLine + text.substring(lineEnd);

    setTimeout(() => {
      rulesTextarea.focus();
      rulesTextarea.selectionStart = rulesTextarea.selectionEnd = cursorPos + change;
    }, 0);
  }

  function addTable() {
    if (!rulesTextarea) return;
    const tableTemplate = '\n| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |\n| Cell 3   | Cell 4   |\n';
    
    const start = rulesTextarea.selectionStart;
    const end = rulesTextarea.selectionEnd;
    
    sportDetails.rules = 
      sportDetails.rules.substring(0, start) + 
      tableTemplate + 
      sportDetails.rules.substring(end);

    setTimeout(() => {
      rulesTextarea.focus();
      rulesTextarea.selectionStart = rulesTextarea.selectionEnd = start + tableTemplate.length;
    }, 0);
  }

  function applyCustomColor() {
    if (!rulesTextarea) return;
    const prefix = `#color(${customColor}, `;
    const suffix = `)`
    applyMarkdown(prefix, suffix);
  }

  function handleKeydown(event) {
    if (event.key === 'Enter' && !event.shiftKey) {
      const textarea = event.target;
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const text = textarea.value;

      const lineStart = text.lastIndexOf('\n', start - 1) + 1;
      const currentLineToCursor = text.substring(lineStart, start);
      
      const listPrefixRegex = /^(\s*-\s)/;
      const listMatch = currentLineToCursor.match(listPrefixRegex);

      if (listMatch) {
        const lineEnd = text.indexOf('\n', start);
        const currentLine = text.substring(lineStart, lineEnd === -1 ? text.length : lineEnd);

        if (currentLine.trim() === '-') {
          event.preventDefault();
          // Remove the list item line
          const before = text.substring(0, lineStart);
          const after = text.substring(lineEnd === -1 ? text.length : lineEnd + 1);
          sportDetails.rules = before + after;
          setTimeout(() => {
            textarea.selectionStart = textarea.selectionEnd = before.length;
          }, 0);
        } else {
          // If list item has content, create a new list item on new line
          event.preventDefault();
          const prefix = listMatch[0]; // e.g., "  - "
          const newText = '\n' + prefix;
          
          sportDetails.rules = text.substring(0, start) + newText + text.substring(end);
          setTimeout(() => {
            textarea.selectionStart = textarea.selectionEnd = start + newText.length;
          }, 0);
        }
      }
    }
  }

  async function handleSaveCapacity() {
    if (!selectedEventId || !selectedSportId) {
      alert('競技を選択してください。');
      return;
    }

    if (capacityMode === 'bulk') {
      // Validate
      if (minCapacity !== null && maxCapacity !== null && minCapacity > maxCapacity) {
        alert('最低定員は最高定員以下である必要があります。');
        return;
      }

      try {
        const response = await fetch(`/api/admin/events/${selectedEventId}/sports/${selectedSportId}/capacity`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            min_capacity: minCapacity,
            max_capacity: maxCapacity,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to update capacity');
        }

        alert('定員設定を更新しました。');
        await fetchSportDetails(selectedEventId, selectedSportId);
      } catch (error) {
        console.error(error);
        alert(`更新エラー: ${error.message}`);
      }
    } else {
      // Per-class capacity saving
      await handleSaveClassCapacities();
    }
  }

  async function handleSaveClassCapacities() {
    if (!selectedEventId || !selectedSportId) {
      alert('競技を選択してください。');
      return;
    }

    // Validate all class capacities
    for (const [classId, capacity] of Object.entries(classCapacities)) {
      if (capacity.min !== null && capacity.max !== null && capacity.min > capacity.max) {
        const className = classes.find(c => c.id == classId)?.name || '不明なクラス';
        alert(`${className}: 最低定員は最高定員以下である必要があります。`);
        return;
      }
    }

    try {
      // Save each class capacity
      const updatePromises = Object.entries(classCapacities).map(async ([classId, capacity]) => {
        // Note: This API endpoint may need to be created in the backend
        // For now, we'll use a placeholder endpoint
        const response = await fetch(`/api/admin/events/${selectedEventId}/sports/${selectedSportId}/classes/${classId}/capacity`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            min_capacity: capacity.min,
            max_capacity: capacity.max,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || `Failed to update capacity for class ${classId}`);
        }
        return response;
      });

      await Promise.all(updatePromises);
      alert('すべてのクラスの定員設定を更新しました。');
      await fetchClassCapacities();
    } catch (error) {
      console.error(error);
      alert(`更新エラー: ${error.message}`);
    }
  }

  function updateClassCapacity(classId, field, value) {
    if (!classCapacities[classId]) {
      classCapacities[classId] = { min: null, max: null };
    }
    classCapacities[classId][field] = value === '' ? null : parseInt(value, 10);
    classCapacities = { ...classCapacities }; // Trigger reactivity
  }

  // Rainy mode capacity functions
  async function fetchRainyModeSettings() {
    if (!selectedEventId || !selectedSportId) return;
    
    try {
      const response = await fetch(`/api/root/events/${selectedEventId}/rainy-mode/settings`);
      if (response.ok) {
        const allSettings = await response.json();
        // Filter settings for current sport
        const sportSettings = allSettings.filter(s => s.sport_id == selectedSportId);
        
        // Initialize rainy mode capacities
        if (sportSettings.length > 0 && classes.length > 0) {
          // Check if all classes have the same values (bulk mode)
          const firstSetting = sportSettings[0];
          const allSame = classes.length === sportSettings.length && 
            sportSettings.every(s => 
              s.min_capacity === firstSetting.min_capacity && 
              s.max_capacity === firstSetting.max_capacity
            );
          
          if (allSame) {
            // All classes have the same values, use bulk mode
            rainyModeCapacityMode = 'bulk';
            rainyModeMinCapacity = firstSetting.min_capacity ?? null;
            rainyModeMaxCapacity = firstSetting.max_capacity ?? null;
          } else {
            // Different values or not all classes, use per-class mode
            rainyModeCapacityMode = 'per-class';
            rainyModeClassCapacities = {};
            // Initialize all classes
            classes.forEach(cls => {
              const setting = sportSettings.find(s => s.class_id == cls.id);
              rainyModeClassCapacities[cls.id] = {
                min: setting?.min_capacity ?? null,
                max: setting?.max_capacity ?? null
              };
            });
            rainyModeClassCapacities = { ...rainyModeClassCapacities };
          }
        } else if (sportSettings.length > 0 && classes.length === 0) {
          // Settings exist but classes not loaded yet, use first setting as bulk
          const firstSetting = sportSettings[0];
          rainyModeCapacityMode = 'bulk';
          rainyModeMinCapacity = firstSetting.min_capacity ?? null;
          rainyModeMaxCapacity = firstSetting.max_capacity ?? null;
        } else {
          // No settings found, reset to defaults
          rainyModeCapacityMode = 'bulk';
          rainyModeMinCapacity = null;
          rainyModeMaxCapacity = null;
          rainyModeClassCapacities = {};
        }
      }
    } catch (error) {
      console.error('Failed to fetch rainy mode settings:', error);
      // Reset to defaults on error
      rainyModeCapacityMode = 'bulk';
      rainyModeMinCapacity = null;
      rainyModeMaxCapacity = null;
      rainyModeClassCapacities = {};
    }
  }

  async function handleSaveRainyModeCapacity() {
    if (!selectedEventId || !selectedSportId) {
      alert('競技を選択してください。');
      return;
    }

    if (rainyModeCapacityMode === 'bulk') {
      // Validate
      if (rainyModeMinCapacity !== null && rainyModeMaxCapacity !== null && rainyModeMinCapacity > rainyModeMaxCapacity) {
        alert('最低定員は最高定員以下である必要があります。');
        return;
      }

      // For bulk mode, we need to save for all classes
      if (classes.length === 0) {
        alert('クラス情報が取得できませんでした。');
        return;
      }

      try {
        const updatePromises = classes.map(async (cls) => {
          const response = await fetch(`/api/root/events/${selectedEventId}/rainy-mode/settings`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              sport_id: selectedSportId,
              class_id: cls.id,
              min_capacity: rainyModeMinCapacity,
              max_capacity: rainyModeMaxCapacity,
            }),
          });

          if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to update rainy mode capacity');
          }
          return response;
        });

        await Promise.all(updatePromises);
        alert('雨天時定員設定を更新しました。');
        await fetchRainyModeSettings();
      } catch (error) {
        console.error(error);
        alert(`更新エラー: ${error.message}`);
      }
    } else {
      // Per-class capacity saving
      await handleSaveRainyModeClassCapacities();
    }
  }

  async function handleSaveRainyModeClassCapacities() {
    if (!selectedEventId || !selectedSportId) {
      alert('競技を選択してください。');
      return;
    }

    // Validate all class capacities
    for (const [classId, capacity] of Object.entries(rainyModeClassCapacities)) {
      if (capacity.min !== null && capacity.max !== null && capacity.min > capacity.max) {
        const className = classes.find(c => c.id == classId)?.name || '不明なクラス';
        alert(`${className}: 最低定員は最高定員以下である必要があります。`);
        return;
      }
    }

    try {
      // Save each class capacity
      const updatePromises = Object.entries(rainyModeClassCapacities).map(async ([classId, capacity]) => {
          const response = await fetch(`/api/root/events/${selectedEventId}/rainy-mode/settings`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              sport_id: selectedSportId,
              class_id: classId,
              min_capacity: capacity.min,
              max_capacity: capacity.max,
            }),
          });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || `Failed to update rainy mode capacity for class ${classId}`);
        }
        return response;
      });

      await Promise.all(updatePromises);
      alert('すべてのクラスの雨天時定員設定を更新しました。');
      await fetchRainyModeSettings();
    } catch (error) {
      console.error(error);
      alert(`更新エラー: ${error.message}`);
    }
  }

  function updateRainyModeClassCapacity(classId, field, value) {
    if (!rainyModeClassCapacities[classId]) {
      rainyModeClassCapacities[classId] = { min: null, max: null };
    }
    rainyModeClassCapacities[classId][field] = value === '' ? null : parseInt(value, 10);
    rainyModeClassCapacities = { ...rainyModeClassCapacities }; // Trigger reactivity
  }
</script>

<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-4">競技詳細情報登録</h1>

  {#if selectedEventId}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div>
        <label for="event-select" class="block text-sm font-medium text-gray-700">アクティブな大会</label>
        <div class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 bg-gray-100 rounded-md">
          {activeEventName}
        </div>
      </div>
      <div>
        <label for="sport-select" class="block text-sm font-medium text-gray-700">競技選択</label>
        <select id="sport-select" on:change={handleSportChange} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
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
        <h2 class="text-xl font-semibold mb-2">定員設定</h2>
        
        <!-- Mode Selection -->
        <div class="flex gap-4 mb-4">
          <label class="flex items-center">
            <input type="radio" bind:group={capacityMode} value={'bulk'} class="mr-1">
            一括設定
          </label>
          <label class="flex items-center">
            <input type="radio" bind:group={capacityMode} value={'per-class'} class="mr-1">
            クラスごと設定
          </label>
        </div>

        {#if capacityMode === 'bulk'}
          <!-- Bulk Capacity Setting -->
          <div class="flex items-center gap-4 mb-2">
            <div class="flex items-center gap-2">
              <label for="min-capacity" class="text-sm font-medium text-gray-700">最低定員</label>
              <input 
                type="number" 
                id="min-capacity"
                bind:value={minCapacity}
                placeholder="未設定"
                min="0"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
            <span class="text-gray-500">〜</span>
            <div class="flex items-center gap-2">
              <label for="max-capacity" class="text-sm font-medium text-gray-700">最高定員</label>
              <input 
                type="number" 
                id="max-capacity"
                bind:value={maxCapacity}
                placeholder="未設定"
                min="0"
                class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
            <button 
              on:click={handleSaveCapacity}
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
            >
              定員設定を保存
            </button>
          </div>
          <p class="text-sm text-gray-600">現在の設定: {minCapacity ?? '未設定'} 〜 {maxCapacity ?? '未設定'}</p>
        {:else}
          <!-- Per-Class Capacity Setting -->
          {#if classes.length > 0}
            <div class="space-y-2 mb-4">
              <div class="overflow-x-auto">
                <table class="min-w-full border border-gray-300 rounded-md">
                  <thead class="bg-gray-100">
                    <tr>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">クラス名</th>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">最低定員</th>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">最高定員</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each classes as cls}
                      <tr class="border-b hover:bg-gray-50">
                        <td class="px-4 py-2 text-sm text-gray-900">{cls.name}</td>
                        <td class="px-4 py-2">
                          <input 
                            type="number" 
                            placeholder="未設定"
                            min="0"
                            value={classCapacities[cls.id]?.min ?? ''}
                            on:input={(e) => updateClassCapacity(cls.id, 'min', e.target.value)}
                            class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </td>
                        <td class="px-4 py-2">
                          <input 
                            type="number" 
                            placeholder="未設定"
                            min="0"
                            value={classCapacities[cls.id]?.max ?? ''}
                            on:input={(e) => updateClassCapacity(cls.id, 'max', e.target.value)}
                            class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
              <button 
                on:click={handleSaveCapacity}
                class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
              >
                すべてのクラスの定員設定を保存
              </button>
            </div>
          {:else}
            <p class="text-sm text-gray-600">クラス情報を読み込み中...</p>
          {/if}
        {/if}
      </div>

      <!-- Rainy Mode Capacity Settings -->
      <div class="mb-4 border-t pt-4">
        <h2 class="text-xl font-semibold mb-2">雨天時定員設定</h2>
        <p class="text-sm text-gray-600 mb-4">雨天時モード時の各競技・クラスごとの登録可能人数の上限・下限を設定できます。試合開始時間は「試合開始時間」セクションで設定できます。</p>
        
        <!-- Mode Selection -->
        <div class="flex gap-4 mb-4">
          <label class="flex items-center">
            <input type="radio" bind:group={rainyModeCapacityMode} value={'bulk'} class="mr-1">
            一括設定
          </label>
          <label class="flex items-center">
            <input type="radio" bind:group={rainyModeCapacityMode} value={'per-class'} class="mr-1">
            クラスごと設定
          </label>
        </div>

        {#if rainyModeCapacityMode === 'bulk'}
          <!-- Bulk Rainy Mode Capacity Setting -->
          <div class="space-y-4">
            <div class="flex items-center gap-4">
              <div class="flex items-center gap-2">
                <label for="rainy-min-capacity" class="text-sm font-medium text-gray-700">最低定員</label>
                <input 
                  type="number" 
                  id="rainy-min-capacity"
                  bind:value={rainyModeMinCapacity}
                  placeholder="未設定"
                  min="0"
                  class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>
              <span class="text-gray-500">〜</span>
              <div class="flex items-center gap-2">
                <label for="rainy-max-capacity" class="text-sm font-medium text-gray-700">最高定員</label>
                <input 
                  type="number" 
                  id="rainy-max-capacity"
                  bind:value={rainyModeMaxCapacity}
                  placeholder="未設定"
                  min="0"
                  class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                />
              </div>
            </div>
            <button 
              on:click={handleSaveRainyModeCapacity}
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
            >
              雨天時定員設定を保存
            </button>
          </div>
          <p class="text-sm text-gray-600 mt-2">
            現在の設定: 定員 {rainyModeMinCapacity ?? '未設定'} 〜 {rainyModeMaxCapacity ?? '未設定'}
          </p>
        {:else}
          <!-- Per-Class Rainy Mode Capacity Setting -->
          {#if classes.length > 0}
            <div class="space-y-2 mb-4">
              <div class="overflow-x-auto">
                <table class="min-w-full border border-gray-300 rounded-md">
                  <thead class="bg-gray-100">
                    <tr>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">クラス名</th>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">最低定員</th>
                      <th class="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">最高定員</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each classes as cls}
                      <tr class="border-b hover:bg-gray-50">
                        <td class="px-4 py-2 text-sm text-gray-900">{cls.name}</td>
                        <td class="px-4 py-2">
                          <input 
                            type="number" 
                            placeholder="未設定"
                            min="0"
                            value={rainyModeClassCapacities[cls.id]?.min ?? ''}
                            on:input={(e) => updateRainyModeClassCapacity(cls.id, 'min', e.target.value)}
                            class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </td>
                        <td class="px-4 py-2">
                          <input 
                            type="number" 
                            placeholder="未設定"
                            min="0"
                            value={rainyModeClassCapacities[cls.id]?.max ?? ''}
                            on:input={(e) => updateRainyModeClassCapacity(cls.id, 'max', e.target.value)}
                            class="w-24 px-2 py-1 text-sm border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500"
                          />
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
              <button 
                on:click={handleSaveRainyModeCapacity}
                class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm"
              >
                すべてのクラスの雨天時定員設定を保存
              </button>
            </div>
          {:else}
            <p class="text-sm text-gray-600">クラス情報を読み込み中...</p>
          {/if}
        {/if}
      </div>

      <div class="mb-4">
        <h2 class="text-xl font-semibold mb-2">ルール詳細</h2>
        <div class="flex gap-4 mb-2">
          <label class="flex items-center">
            <input type="radio" bind:group={sportDetails.rules_type} value={'markdown'} class="mr-1">
            Markdown
          </label>
          <label class="flex items-center">
            <input type="radio" bind:group={sportDetails.rules_type} value={'pdf'} class="mr-1">
            PDF
          </label>
        </div>

        {#if sportDetails.rules_type === 'markdown'}
          <div>
            <div class="flex items-center gap-2 mb-2 p-2 bg-gray-100 rounded-md">
              <button on:click={() => applyMarkdown('**', '**')} class="px-3 py-1 font-bold">B</button>
              <button on:click={() => applyMarkdown('*', '*')} class="px-3 py-1 italic">I</button>
              <button on:click={() => applyHeading(1)} class="px-3 py-1 font-bold">H1</button>
              <button on:click={() => applyHeading(2)} class="px-3 py-1 font-bold">H2</button>
              <button on:click={() => applyHeading(3)} class="px-3 py-1 font-bold">H3</button>
              <button on:click={addLink} class="px-3 py-1">Link</button>
              <button on:click={addList} class="px-3 py-1">List</button>
              <button on:click={addTable} class="px-3 py-1">Table</button>
              <button on:click={() => applyMarkdown('##', '##')} class="px-3 py-1 text-red-600 font-bold">Red</button>
              <div class="flex items-center gap-2 border border-gray-300 rounded-md p-1">
                <input type="color" bind:value={customColor} class="w-8 h-7 p-0 border-none cursor-pointer" title="Select a color">
                <button on:click={applyCustomColor} class="px-3 py-1 bg-gray-200 text-gray-800 rounded-md text-sm hover:bg-gray-300">Apply</button>
              </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <textarea bind:value={sportDetails.rules} bind:this={rulesTextarea} on:scroll={syncScroll} on:paste={handlePaste} on:keydown={handleKeydown} rows="10" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 mt-1 block w-full sm:text-sm border border-gray-300 rounded-md h-96 overflow-y-scroll"></textarea>
              <SafeHtml
                bind:element={previewDiv}
                class="prose border p-4 rounded-md h-96 overflow-y-scroll"
                html={markdownPreviewHtml}
              />
            </div>
            <input type="file" id="image-upload" accept="image/*" class="hidden" on:change={handleFileSelect}>
            <button on:click={() => { if (browser) { const el = document.getElementById('image-upload'); if (el) el.click(); } }} class="mt-2 px-3 py-1 bg-gray-200 text-gray-800 rounded-md text-sm">
              画像アップロード
            </button>
          </div>
        {:else}
          <div class="flex flex-col gap-4">
            <input type="file" accept=".pdf" on:change={handlePdfFileSelect} class="file-input file-input-bordered w-full max-w-xs">
            {#if pdfPreviewUrl || sportDetails.rules_pdf_url}
              <div class="border rounded-md h-96">
                <embed src={pdfPreviewUrl || sportDetails.rules_pdf_url} type="application/pdf" width="100%" height="100%">
              </div>
            {/if}
          </div>
        {/if}
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
              <!-- 通常モードの試合開始時間 -->
              <div class="mb-6">
                <h3 class="text-lg font-medium mb-2">通常モード</h3>
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
              </div>
              
              <!-- 雨天時モードの試合開始時間（本戦+敗者戦） -->
              <div class="mb-6 border-t pt-4">
                <h3 class="text-lg font-medium mb-2">雨天時モード（本戦+敗者戦）</h3>
                <p class="text-sm text-gray-600 mb-4">雨天時モード時の本戦と敗者戦の試合開始時間を設定できます。雨天時モードが有効になる前に設定しておくことができます。</p>
                {#if allMatchesForSport && allMatchesForSport.length > 0}
                  <div class="flex items-center gap-4 mb-4">
                    <input type="datetime-local" bind:value={bulkRainyModeStartTime} class="border rounded px-2 py-1">
                    <button on:click={handleBulkRainyModeTimeUpdate} class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700">全試合に適用</button>
                    <button on:click={handleSaveAllRainyModeMatchTimes} class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">設定した時間をすべて保存</button>
                  </div>
                  <div class="space-y-2">
                      {#each allMatchesForSport as match}
                          <div class="flex items-center justify-between p-2 border rounded {match.isLoserBracketMatch ? 'bg-yellow-50' : ''}">
                              <div class="flex items-center gap-2">
                                <span class="font-medium">
                                  {#if match.isLoserBracketMatch}
                                    敗者戦{match.loserBracketBlock ? match.loserBracketBlock + 'ブロック' : ''} Round {match.roundIndex + 1}, Match {match.order + 1}
                                    {#if match.loserBracketRound}
                                      (敗者戦ラウンド {match.loserBracketRound})
                                    {/if}
                                  {:else}
                                    本戦 Round {match.roundIndex + 1}, Match {match.order + 1}
                                  {/if}
                                </span>
                                <span class="text-xs text-gray-500">({match.tournamentName})</span>
                              </div>
                              <div class="flex items-center gap-4">
                                  <span class="text-sm text-gray-600">
                                      {#if match.rainyModeStartTime}
                                          雨天時開始日時: {match.rainyModeStartTime}
                                      {:else}
                                          未設定
                                      {/if}
                                  </span>
                                  <input 
                                      type="datetime-local" 
                                      class="border rounded px-2 py-1"
                                      value={match.rainyModeStartTime ? match.rainyModeStartTime.slice(0, 16) : ''} 
                                      on:change={(e) => updateMatchRainyModeTimeLocally(match.id, e.target.value)}
                                  />
                              </div>
                          </div>
                      {/each}
                  </div>
                {:else}
                  <p class="text-sm text-gray-500 italic">試合がありません。トーナメントが生成された後に試合が表示されます。</p>
                {/if}
              </div>
          {:else}
              <p class="text-gray-600">トーナメントを選択してください。</p>
          {/if}
      </div>
    {/if}
  {:else}
    <p>アクティブな大会が設定されていません。rootユーザーでアクティブな大会を設定してください。</p>
  {/if}
</div>