<script>
  import { onMount, onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import { activeEvent } from '$lib/stores/eventStore.js';

  let session = null;
  let classes = [];
  let groups = [];
  let matches = [];
  let pointsSummary = [];
  let templateRuns = [];
  let loading = false;
  let savingSession = false;
  let savingGroup = false;
  let savingMatch = false;
  let savingManualPoint = false;
  let creatingTemplate = {};
  let selectedTemplateType = null;
  let templateConfigForm = {
    name: '',
    description: '',
    mode: 'group',
    win_points: 0,
    loss_points: 0,
    draw_points: 0,
    participation_points: 0,
    allow_manual_points: false,
    points_by_rank: {
      1: 40,
      2: 30,
      3: 20,
      4: 10
    },
    year_relay_points: {
      block_a: {1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5},
      block_b: {1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5},
      overall: {1: 30, 2: 20, 3: 10, 4: 0, 5: 0, 6: 0}
    }
  };

  let sessionForm = {
    name: '',
    description: '',
    mode: 'mixed',
    win_points: 0,
    loss_points: 0,
    draw_points: 0,
    participation_points: 0,
    allow_manual_points: true
  };

  let groupForm = {
    id: null,
    name: '',
    description: '',
    class_ids: []
  };

  let matchForm = {
    id: null,
    title: '',
    scheduled_at: '',
    location: '',
    format: '',
    memo: '',
    status: 'scheduled',
    allow_draw: false,
    participants: []
  };

  let manualPointForm = {
    class_id: null,
    points: 0,
    reason: ''
  };

  let errorMessage = '';

  let escapeHandler = null;

  onMount(async () => {
    await activeEvent.init();
    const current = get(activeEvent);
    if (current) {
      await fetchSession(current.id);
    }

    escapeHandler = (e) => {
      if (e.key === 'Escape' && selectedTemplateType) {
        closeTemplateConfig();
      }
    };
    window.addEventListener('keydown', escapeHandler);
  });

  onDestroy(() => {
    if (escapeHandler) {
      window.removeEventListener('keydown', escapeHandler);
    }
  });

  async function fetchSession(eventId) {
    loading = true;
    errorMessage = '';
    try {
      const res = await fetch(`/api/root/events/${eventId}/noon-game/session`);
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '昼競技情報の取得に失敗しました');
      }
      const data = await res.json();
      session = data.session;
      classes = data.classes || [];
      groups = data.groups || [];
      matches = data.matches || [];
      pointsSummary = data.points_summary || [];
      templateRuns = data.template_runs || [];
      if (session) {
        populateSessionForm(session);
      } else {
        resetSessionForm();
      }
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
    } finally {
      loading = false;
    }
  }

  function populateSessionForm(s) {
    sessionForm = {
      name: s.name ?? '',
      description: s.description ?? '',
      mode: s.mode ?? 'mixed',
      win_points: s.win_points ?? 0,
      loss_points: s.loss_points ?? 0,
      draw_points: s.draw_points ?? 0,
      participation_points: s.participation_points ?? 0,
      allow_manual_points: s.allow_manual_points ?? true
    };
  }

  function resetSessionForm() {
    sessionForm = {
      name: '',
      description: '',
      mode: 'mixed',
      win_points: 0,
      loss_points: 0,
      draw_points: 0,
      participation_points: 0,
      allow_manual_points: true
    };
  }

  function resetGroupForm() {
    groupForm = {
      id: null,
      name: '',
      description: '',
      class_ids: []
    };
  }

  function resetMatchForm() {
    matchForm = {
      id: null,
      title: '',
      scheduled_at: '',
      location: '',
      format: '',
      memo: '',
      status: 'scheduled',
      allow_draw: false,
      participants: [
        { id: null, type: 'class', class_id: '', group_id: '', display_name: '' },
        { id: null, type: 'class', class_id: '', group_id: '', display_name: '' }
      ]
    };
  }

  async function saveSession() {
    const current = get(activeEvent);
    if (!current) return;
    
    // 既にセッションが存在する場合は更新のみ
    if (session && session.id) {
      // 既存セッションの更新
      savingSession = true;
      errorMessage = '';
      try {
        const payload = {
          ...sessionForm,
          win_points: Number(sessionForm.win_points),
          loss_points: Number(sessionForm.loss_points),
          draw_points: Number(sessionForm.draw_points),
          participation_points: Number(sessionForm.participation_points),
          allow_manual_points: !!sessionForm.allow_manual_points
        };
        const res = await fetch(`/api/root/events/${current.id}/noon-game/session`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (!res.ok) {
          const detail = await safeJson(res);
          throw new Error(detail?.error || '昼競技セッションの更新に失敗しました');
        }
        const data = await res.json();
        session = data.session;
        groups = data.groups || [];
        matches = data.matches || [];
        pointsSummary = data.points_summary || [];
        templateRuns = data.template_runs || [];
        populateSessionForm(session);
        alert('昼競技セッションを更新しました。');
      } catch (err) {
        console.error(err);
        errorMessage = err.message;
        alert(err.message);
      } finally {
        savingSession = false;
      }
    } else {
      // 新規セッション作成
      savingSession = true;
      errorMessage = '';
      try {
        const payload = {
          ...sessionForm,
          win_points: Number(sessionForm.win_points),
          loss_points: Number(sessionForm.loss_points),
          draw_points: Number(sessionForm.draw_points),
          participation_points: Number(sessionForm.participation_points),
          allow_manual_points: !!sessionForm.allow_manual_points
        };
        const res = await fetch(`/api/root/events/${current.id}/noon-game/session`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (!res.ok) {
          const detail = await safeJson(res);
          throw new Error(detail?.error || '昼競技セッションの作成に失敗しました');
        }
        const data = await res.json();
        session = data.session;
        groups = data.groups || [];
        matches = data.matches || [];
        pointsSummary = data.points_summary || [];
        templateRuns = data.template_runs || [];
        populateSessionForm(session);
        alert('昼競技セッションを作成しました。');
      } catch (err) {
        console.error(err);
        errorMessage = err.message;
        alert(err.message);
      } finally {
        savingSession = false;
      }
    }
  }

  function startEditGroup(group) {
    groupForm = {
      id: group.id,
      name: group.name,
      description: group.description ?? '',
    class_ids: group.members?.map(m => String(m.class_id)) ?? []
    };
  }

  async function submitGroup() {
    if (!session) {
      alert('先に昼競技セッションを作成してください。');
      return;
    }
    savingGroup = true;
    errorMessage = '';
    try {
      const method = groupForm.id ? 'PUT' : 'POST';
      const endpoint = groupForm.id
        ? `/api/root/noon-game/sessions/${session.id}/groups/${groupForm.id}`
        : `/api/root/noon-game/sessions/${session.id}/groups`;
      const res = await fetch(endpoint, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: groupForm.name,
          description: groupForm.description,
          class_ids: groupForm.class_ids
            .filter((id) => id !== null && id !== undefined && id !== '')
            .map((id) => Number(id))
        })
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || 'グループの保存に失敗しました');
      }
      const updated = await res.json();
      if (updated?.group) {
        updateGroupsList(updated.group);
      } else {
        await refetchCurrentSession();
      }
      resetGroupForm();
      alert('グループを保存しました。');
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
      alert(err.message);
    } finally {
      savingGroup = false;
    }
  }

  async function deleteGroup(groupId) {
    if (!session) return;
    if (!confirm('このグループを削除しますか？関連する試合がある場合は削除できません。')) return;
    try {
      const res = await fetch(`/api/root/noon-game/sessions/${session.id}/groups/${groupId}`, {
        method: 'DELETE'
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || 'グループの削除に失敗しました');
      }
      groups = groups.filter(g => g.id !== groupId);
      alert('グループを削除しました。');
    } catch (err) {
      console.error(err);
      alert(err.message);
    }
  }

  function updateGroupsList(group) {
    const exists = groups.findIndex(g => g.id === group.id);
    if (exists >= 0) {
      groups = [
        ...groups.slice(0, exists),
        group,
        ...groups.slice(exists + 1)
      ];
    } else {
      groups = [...groups, group];
    }
  }

  function startEditMatch(match) {
    matchForm = {
      id: match.id,
      title: match.title ?? '',
      scheduled_at: match.scheduled_at ? toLocalDateTime(match.scheduled_at) : '',
      location: match.location ?? '',
      format: match.format ?? '',
      memo: match.memo ?? '',
      status: match.status ?? 'scheduled',
      allow_draw: match.allow_draw ?? false,
      participants: []
    };

    if (match.entries && match.entries.length > 0) {
      matchForm.participants = match.entries.map((entry) => ({
        id: entry.id,
        type: entry.side_type ?? 'class',
        class_id: entry.side_type === 'class' && entry.class_id != null ? String(entry.class_id) : '',
        group_id: entry.side_type === 'group' && entry.group_id != null ? String(entry.group_id) : '',
        display_name: entry.display_name ?? ''
      }));
    } else {
      matchForm.participants = [
        {
          id: null,
          type: match.home_side_type ?? 'class',
          class_id: match.home_side_type === 'class' && match.home_class_id != null ? String(match.home_class_id) : '',
          group_id: match.home_side_type === 'group' && match.home_group_id != null ? String(match.home_group_id) : '',
          display_name: ''
        },
        {
          id: null,
          type: match.away_side_type ?? 'class',
          class_id: match.away_side_type === 'class' && match.away_class_id != null ? String(match.away_class_id) : '',
          group_id: match.away_side_type === 'group' && match.away_group_id != null ? String(match.away_group_id) : '',
          display_name: ''
        }
      ];
    }
  }

  function addParticipant() {
    const updated = [
      ...matchForm.participants,
      { id: null, type: 'class', class_id: '', group_id: '', display_name: '' }
    ];
    matchForm = { ...matchForm, participants: updated };
  }

  function removeParticipant(index) {
    const updated = matchForm.participants.filter((_, i) => i !== index);
    matchForm = { ...matchForm, participants: updated };
  }

  function moveParticipant(index, direction) {
    const newIndex = index + direction;
    if (newIndex < 0 || newIndex >= matchForm.participants.length) {
      return;
    }
    const reordered = [...matchForm.participants];
    const [item] = reordered.splice(index, 1);
    reordered.splice(newIndex, 0, item);
    matchForm = { ...matchForm, participants: reordered };
  }

  function setParticipantType(index, type) {
    const updated = [...matchForm.participants];
    updated[index] = {
      ...updated[index],
      type,
      class_id: type === 'class' ? '' : '',
      group_id: type === 'group' ? '' : '',
    };
    matchForm = { ...matchForm, participants: updated };
  }

  function updateParticipantField(index, field, value) {
    const updated = [...matchForm.participants];
    updated[index] = {
      ...updated[index],
      [field]: value
    };
    matchForm = { ...matchForm, participants: updated };
  }

  async function submitMatch() {
    if (!session) {
      alert('先に昼競技セッションを作成してください。');
      return;
    }
    savingMatch = true;
    errorMessage = '';
    try {
      const method = matchForm.id ? 'PUT' : 'POST';
      const endpoint = matchForm.id
        ? `/api/root/noon-game/sessions/${session.id}/matches/${matchForm.id}`
        : `/api/root/noon-game/sessions/${session.id}/matches`;
      if (!matchForm.participants || matchForm.participants.length === 0) {
        throw new Error('参加者を最低1つ追加してください。');
      }

      const participantsPayload = matchForm.participants.map((participant, idx) => {
        const type = (participant.type || 'class').toLowerCase();
        const payloadParticipant = {
          id: participant.id ?? null,
          type,
          class_id: null,
          group_id: null,
          display_name: participant.display_name ? participant.display_name.trim() : null
        };

        if (type === 'class') {
          if (!participant.class_id) {
            throw new Error(`参加者${idx + 1}のクラスを選択してください。`);
          }
          payloadParticipant.class_id = Number(participant.class_id);
          if (Number.isNaN(payloadParticipant.class_id)) {
            throw new Error(`参加者${idx + 1}のクラス指定が不正です。`);
          }
          payloadParticipant.group_id = null;
        } else if (type === 'group') {
          if (!participant.group_id) {
            throw new Error(`参加者${idx + 1}のグループを選択してください。`);
          }
          payloadParticipant.group_id = Number(participant.group_id);
          if (Number.isNaN(payloadParticipant.group_id)) {
            throw new Error(`参加者${idx + 1}のグループ指定が不正です。`);
          }
          payloadParticipant.class_id = null;
        } else {
          throw new Error(`参加者${idx + 1}の種別が不正です。`);
        }

        return payloadParticipant;
      });

      const homeSideSource = participantsPayload[0] ?? { type: 'class', class_id: null, group_id: null };
      const awaySideSource =
        participantsPayload.length > 1 ? participantsPayload[1] : homeSideSource;

      const payload = {
        title: matchForm.title || null,
        scheduled_at: matchForm.scheduled_at ? new Date(matchForm.scheduled_at).toISOString() : null,
        location: matchForm.location || null,
        format: matchForm.format || null,
        memo: matchForm.memo || null,
        status: matchForm.status || 'scheduled',
        allow_draw: !!matchForm.allow_draw,
        home_side: {
          type: homeSideSource.type,
          class_id: homeSideSource.type === 'class' ? homeSideSource.class_id : null,
          group_id: homeSideSource.type === 'group' ? homeSideSource.group_id : null
        },
        away_side: {
          type: awaySideSource.type,
          class_id: awaySideSource.type === 'class' ? awaySideSource.class_id : null,
          group_id: awaySideSource.type === 'group' ? awaySideSource.group_id : null
        },
        participants: participantsPayload
      };
      const res = await fetch(endpoint, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '試合の保存に失敗しました');
      }
      const data = await res.json();
      if (data?.match) {
        updateMatchList(data.match);
      } else {
        await refetchCurrentSession();
      }
      resetMatchForm();
      alert('試合を保存しました。');
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
      alert(err.message);
    } finally {
      savingMatch = false;
    }
  }

  async function deleteMatch(matchId) {
    if (!session) return;
    if (!confirm('この試合を削除しますか？')) return;
    try {
      const res = await fetch(`/api/root/noon-game/sessions/${session.id}/matches/${matchId}`, {
        method: 'DELETE'
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '試合の削除に失敗しました');
      }
      matches = matches.filter(m => m.id !== matchId);
      alert('試合を削除しました。');
    } catch (err) {
      console.error(err);
      alert(err.message);
    }
  }

  function updateMatchList(match) {
    const exists = matches.findIndex(m => m.id === match.id);
    if (exists >= 0) {
      matches = [
        ...matches.slice(0, exists),
        match,
        ...matches.slice(exists + 1)
      ];
    } else {
      matches = [...matches, match];
    }
  }

  async function submitManualPoint() {
    if (!session) {
      alert('先に昼競技セッションを作成してください。');
      return;
    }
    if (!manualPointForm.class_id) {
      alert('クラスを選択してください。');
      return;
    }
    savingManualPoint = true;
    errorMessage = '';
    try {
      const res = await fetch(`/api/root/noon-game/sessions/${session.id}/manual-points`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          class_id: Number(manualPointForm.class_id),
          points: Number(manualPointForm.points),
          reason: manualPointForm.reason || null
        })
      });
      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || '手動加点の登録に失敗しました');
      }
      const data = await res.json();
      session = data.session;
      groups = data.groups || groups;
      matches = data.matches || matches;
      pointsSummary = data.points_summary || [];
      templateRuns = data.template_runs || templateRuns;
      manualPointForm = { class_id: null, points: 0, reason: '' };
      alert('手動加点を登録しました。');
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
      alert(err.message);
    } finally {
      savingManualPoint = false;
    }
  }

  async function refetchCurrentSession() {
    const current = get(activeEvent);
    if (current) {
      await fetchSession(current.id);
    }
  }

  function toLocalDateTime(value) {
    if (!value) return '';
    const date = new Date(value);
    const tzOffset = date.getTimezoneOffset() * 60000;
    const localISO = new Date(date.getTime() - tzOffset).toISOString().slice(0, 16);
    return localISO;
  }

  async function safeJson(response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }

  function openTemplateConfig(templateType) {
    const current = get(activeEvent);
    if (!current) {
      alert('イベントが選択されていません。');
      return;
    }

    const templateNames = {
      'year-relay': '学年対抗リレー',
      'course-relay': 'コース対抗リレー',
      'tug-of-war': '綱引き'
    };

    // テンプレートキーのマッピング（フロントエンドのキー -> バックエンドのキー）
    const templateKeyMap = {
      'year-relay': 'year_relay',
      'course-relay': 'course_relay',
      'tug-of-war': 'tug_of_war'
    };

    selectedTemplateType = templateType;

    // 既存のテンプレートランを探す
    const existingRun = templateRuns.find(run => run.template_key === templateKeyMap[templateType]);

    // デフォルトの点数設定
    let pointsByRank = {
      1: 40,
      2: 30,
      3: 20,
      4: 10
    };
    let yearRelayPoints = {
      block_a: {1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5},
      block_b: {1: 30, 2: 25, 3: 20, 4: 15, 5: 10, 6: 5},
      overall: {1: 30, 2: 20, 3: 10, 4: 0, 5: 0, 6: 0}
    };

    // 既存のテンプレートランがあれば、その点数設定を使用
    if (existingRun && existingRun.points_by_rank) {
      if (templateType === 'year-relay') {
        // 学年対抗リレーの場合、3つの点数設定を読み込む
        const pb = existingRun.points_by_rank;
        if (pb.block_a) {
          yearRelayPoints.block_a = { ...pb.block_a };
          for (let rank = 1; rank <= 6; rank++) {
            if (yearRelayPoints.block_a[rank] === undefined || yearRelayPoints.block_a[rank] === null) {
              yearRelayPoints.block_a[rank] = [30, 25, 20, 15, 10, 5][rank - 1];
            }
          }
        }
        if (pb.block_b) {
          yearRelayPoints.block_b = { ...pb.block_b };
          for (let rank = 1; rank <= 6; rank++) {
            if (yearRelayPoints.block_b[rank] === undefined || yearRelayPoints.block_b[rank] === null) {
              yearRelayPoints.block_b[rank] = [30, 25, 20, 15, 10, 5][rank - 1];
            }
          }
        }
        if (pb.overall) {
          yearRelayPoints.overall = { ...pb.overall };
          for (let rank = 1; rank <= 6; rank++) {
            if (yearRelayPoints.overall[rank] === undefined || yearRelayPoints.overall[rank] === null) {
              yearRelayPoints.overall[rank] = [30, 20, 10, 0, 0, 0][rank - 1];
            }
          }
        }
      } else {
        // コース対抗リレーと綱引きの場合
        pointsByRank = { ...existingRun.points_by_rank };
        for (let rank = 1; rank <= 4; rank++) {
          if (pointsByRank[rank] === undefined || pointsByRank[rank] === null) {
            pointsByRank[rank] = [40, 30, 20, 10][rank - 1];
          }
        }
      }
    }

    // 既存のセッションがあれば、その設定を使用
    if (session) {
      templateConfigForm = {
        name: session.name || `${templateNames[templateType]}_${current.id}`,
        description: session.description || '',
        mode: session.mode || 'group',
        win_points: session.win_points || 0,
        loss_points: session.loss_points || 0,
        draw_points: session.draw_points || 0,
        participation_points: session.participation_points || 0,
        allow_manual_points: session.allow_manual_points || false,
        points_by_rank: pointsByRank,
        year_relay_points: yearRelayPoints
      };
    } else {
      // セッションがない場合はデフォルト値を設定
      templateConfigForm = {
        name: `${templateNames[templateType]}_${current.id}`,
        description: '',
        mode: 'group',
        win_points: 0,
        loss_points: 0,
        draw_points: 0,
        participation_points: 0,
        allow_manual_points: false,
        points_by_rank: pointsByRank,
        year_relay_points: yearRelayPoints
      };
    }
  }

  function closeTemplateConfig() {
    selectedTemplateType = null;
  }

  async function createTemplate() {
    if (!selectedTemplateType) return;

    const current = get(activeEvent);
    if (!current) {
      alert('イベントが選択されていません。');
      return;
    }

    creatingTemplate = { ...creatingTemplate, [selectedTemplateType]: true };
    errorMessage = '';

    try {
      const templateNames = {
        'year-relay': '学年対抗リレー',
        'course-relay': 'コース対抗リレー',
        'tug-of-war': '綱引き'
      };

      const payload = {
        session: {
          name: templateConfigForm.name,
          description: templateConfigForm.description || null,
          mode: templateConfigForm.mode,
          win_points: Number(templateConfigForm.win_points),
          loss_points: Number(templateConfigForm.loss_points),
          draw_points: Number(templateConfigForm.draw_points),
          participation_points: Number(templateConfigForm.participation_points),
          allow_manual_points: templateConfigForm.allow_manual_points
        }
      };

      // 点数設定を追加
      if (selectedTemplateType === 'year-relay') {
        // 学年対抗リレーの場合、3つの点数設定を保存
        payload.session.points_by_rank = {
          block_a: {},
          block_b: {},
          overall: {}
        };
        for (const [rank, points] of Object.entries(templateConfigForm.year_relay_points.block_a)) {
          const pointsNum = Number(points);
          if (!Number.isNaN(pointsNum)) {
            payload.session.points_by_rank.block_a[Number(rank)] = pointsNum;
          }
        }
        for (const [rank, points] of Object.entries(templateConfigForm.year_relay_points.block_b)) {
          const pointsNum = Number(points);
          if (!Number.isNaN(pointsNum)) {
            payload.session.points_by_rank.block_b[Number(rank)] = pointsNum;
          }
        }
        for (const [rank, points] of Object.entries(templateConfigForm.year_relay_points.overall)) {
          const pointsNum = Number(points);
          if (!Number.isNaN(pointsNum)) {
            payload.session.points_by_rank.overall[Number(rank)] = pointsNum;
          }
        }
      } else if (selectedTemplateType === 'course-relay' || selectedTemplateType === 'tug-of-war') {
        // コース対抗リレーと綱引きの場合
        payload.session.points_by_rank = {};
        for (const [rank, points] of Object.entries(templateConfigForm.points_by_rank)) {
          const pointsNum = Number(points);
          if (!Number.isNaN(pointsNum) && pointsNum > 0) {
            payload.session.points_by_rank[Number(rank)] = pointsNum;
          }
        }
      }

      const endpoint = (selectedTemplateType === 'course-relay' || selectedTemplateType === 'tug-of-war')
        ? `/api/root/events/${current.id}/noon-game/templates/${selectedTemplateType}/run`
        : `/api/admin/events/${current.id}/noon-game/templates/${selectedTemplateType}/run`;
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const detail = await safeJson(res);
        throw new Error(detail?.error || `${templateNames[selectedTemplateType]}テンプレートの作成に失敗しました`);
      }

      const data = await res.json();
      if (templateRuns.length > 0) {
        alert(`${templateNames[selectedTemplateType]}テンプレートに更新しました。既存のテンプレートと関連データは削除されました。`);
      } else {
        alert(`${templateNames[selectedTemplateType]}テンプレートを作成しました。`);
      }
      closeTemplateConfig();
      await refetchCurrentSession();
    } catch (err) {
      console.error(err);
      errorMessage = err.message;
      alert(err.message);
    } finally {
      creatingTemplate = { ...creatingTemplate, [selectedTemplateType]: false };
    }
  }
</script>

<div class="space-y-8 p-4 md:p-8">
  <h1 class="text-3xl font-bold text-gray-800 border-b pb-2">昼競技管理</h1>
  {#if errorMessage}
    <div class="bg-red-100 border-l-4 border-red-400 text-red-700 p-4">
      <p class="font-semibold">エラー</p>
      <p>{errorMessage}</p>
    </div>
  {/if}

  {#if loading}
    <div class="text-gray-600">読み込み中...</div>
  {:else}
    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">テンプレート選択</h2>
      <p class="text-sm text-gray-600">
        テンプレートを選択すると、セッションが自動で作成され、必要な試合が設定されます。
        {#if templateRuns.length > 0}
          <span class="text-orange-600 font-semibold">（既にテンプレートが作成されています。別のテンプレートを選択すると、既存のテンプレートと関連データが削除され、新しいテンプレートが作成されます。）</span>
        {/if}
      </p>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="border rounded-lg p-4 space-y-3">
          <h3 class="text-lg font-semibold text-gray-800">学年対抗リレー</h3>
          <p class="text-sm text-gray-600">
            各学年から8名の走者と補欠2名を選出し、A・Bブロックで試合を行います。
            A・Bブロックそれぞれの順位点と総合順位による追加点を付与します。
          </p>
          <button
            class="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
            on:click={() => openTemplateConfig('year-relay')}
            disabled={creatingTemplate['year-relay']}>
            {templateRuns.length > 0 ? 'テンプレートを更新' : 'テンプレートを設定'}
          </button>
        </div>

        <div class="border rounded-lg p-4 space-y-3">
          <h3 class="text-lg font-semibold text-gray-800">コース対抗リレー</h3>
          <p class="text-sm text-gray-600">
            1~5年の各クラスから2名の走者と補欠1名を選出し、コースごとの3チームを作成します。
            専攻科・教員チームを加えて4チームで競技を行います。
          </p>
          <button
            class="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
            on:click={() => openTemplateConfig('course-relay')}
            disabled={creatingTemplate['course-relay']}>
            {templateRuns.length > 0 ? 'テンプレートを更新' : 'テンプレートを設定'}
          </button>
        </div>

        <div class="border rounded-lg p-4 space-y-3">
          <h3 class="text-lg font-semibold text-gray-800">綱引き</h3>
          <p class="text-sm text-gray-600">
            1~5年の各クラスから3名の選手と補欠6名を選出し、コースごとの3チームを作成します。
            専攻科・教員チームを加えて4チームで競技を行います。
          </p>
          <button
            class="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
            on:click={() => openTemplateConfig('tug-of-war')}
            disabled={creatingTemplate['tug-of-war']}>
            {templateRuns.length > 0 ? 'テンプレートを更新' : 'テンプレートを設定'}
          </button>
        </div>
      </div>
    </section>

    <!-- テンプレート設定モーダル -->
    {#if selectedTemplateType}
      <div class="fixed top-0 left-0 right-0 bottom-0 bg-black bg-opacity-50 flex items-center justify-center z-[110] min-h-screen overflow-y-auto" on:click={closeTemplateConfig}>
        <div class="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 my-4 max-h-[90vh] overflow-y-auto" on:click|stopPropagation>
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-2xl font-semibold text-gray-800">
              {#if selectedTemplateType === 'year-relay'}学年対抗リレー
              {:else if selectedTemplateType === 'course-relay'}コース対抗リレー
              {:else if selectedTemplateType === 'tug-of-war'}綱引き
              {/if} テンプレート設定
            </h2>
            <button class="text-gray-500 hover:text-gray-700" on:click={closeTemplateConfig}>×</button>
          </div>

          <div class="space-y-4">
            <div class="border rounded-lg p-4 space-y-4">
              <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">セッション設定</h3>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-2">
                  セッション名
                  <input class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.name} placeholder="例: 学年対抗リレー_1" />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-2">
                  説明
                  <textarea class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.description} rows="3" placeholder="概要やメモを入力"></textarea>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  モード
                  <select class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.mode}>
                    <option value="mixed">クラス＆グループ混在</option>
                    <option value="class">クラス対抗のみ</option>
                    <option value="group">グループ対抗のみ</option>
                  </select>
                </label>
                <label class="flex items-center space-x-2 text-sm font-medium text-gray-700">
                  <input type="checkbox" bind:checked={templateConfigForm.allow_manual_points} />
                  <span>手動加点を許可</span>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  勝利ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.win_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  敗北ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.loss_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  引き分けポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.draw_points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  参加ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={templateConfigForm.participation_points} />
                </label>
              </div>
              {#if selectedTemplateType === 'year-relay'}
                <div class="border rounded-lg p-4 space-y-4 bg-blue-50">
                  <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">点数設定</h3>
                  <p class="text-sm text-gray-600">Aブロック、Bブロック、総合順位の点数を設定します。</p>
                  
                  <div class="space-y-4">
                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">Aブロック</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.block_a[rank]} on:input={(e) => {
                              templateConfigForm.year_relay_points.block_a = {...templateConfigForm.year_relay_points.block_a, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>
                    
                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">Bブロック</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.block_b[rank]} on:input={(e) => {
                              templateConfigForm.year_relay_points.block_b = {...templateConfigForm.year_relay_points.block_b, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>
                    
                    <div>
                      <h4 class="text-md font-semibold text-gray-700 mb-2">総合順位</h4>
                      <div class="grid grid-cols-6 gap-2">
                        {#each [1, 2, 3, 4, 5, 6] as rank}
                          <label class="flex flex-col text-xs font-medium text-gray-700">
                            {rank}位
                            <input type="number" class="mt-1 border rounded px-2 py-1 text-sm" value={templateConfigForm.year_relay_points.overall[rank]} on:input={(e) => {
                              templateConfigForm.year_relay_points.overall = {...templateConfigForm.year_relay_points.overall, [rank]: Number(e.target.value) || 0};
                            }} />
                          </label>
                        {/each}
                      </div>
                    </div>
                  </div>
                </div>
              {:else if selectedTemplateType === 'course-relay' || selectedTemplateType === 'tug-of-war'}
                <div class="border rounded-lg p-4 space-y-4 bg-blue-50">
                  <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">点数設定</h3>
                  <p class="text-sm text-gray-600">順位ごとの点数を設定します。</p>
                  <div class="grid grid-cols-4 gap-4">
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      1位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[1]} on:input={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 1: Number(e.target.value) || 0};
                      }} />
                    </label>
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      2位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[2]} on:input={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 2: Number(e.target.value) || 0};
                      }} />
                    </label>
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      3位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[3]} on:input={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 3: Number(e.target.value) || 0};
                      }} />
                    </label>
                    <label class="flex flex-col text-sm font-medium text-gray-700">
                      4位の点数
                      <input type="number" class="mt-1 border rounded px-3 py-2" value={templateConfigForm.points_by_rank[4]} on:input={(e) => {
                        templateConfigForm.points_by_rank = {...templateConfigForm.points_by_rank, 4: Number(e.target.value) || 0};
                      }} />
                    </label>
                  </div>
                </div>
              {/if}
            </div>

            <div class="flex justify-end space-x-3">
              <button
                class="px-4 py-2 border rounded text-gray-700 hover:bg-gray-50"
                on:click={closeTemplateConfig}>
                キャンセル
              </button>
              <button
                class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                on:click={createTemplate}
                disabled={creatingTemplate[selectedTemplateType]}>
                {creatingTemplate[selectedTemplateType] ? '作成中...' : 'テンプレートを作成'}
              </button>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <!-- テンプレートを使用しない場合の設定 -->
    <section class="bg-white shadow rounded-lg p-6 space-y-6">
      <h2 class="text-2xl font-semibold text-gray-800 border-b pb-2">テンプレートを使用しない場合の設定</h2>
        
        <div class="space-y-6">
          <!-- 基本設定 -->
          <div class="border rounded-lg p-4 space-y-4">
            <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">基本設定</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <label class="flex flex-col text-sm font-medium text-gray-700">
                セッション名
                <input class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.name} placeholder="例: 昼休み競技 2025" />
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700">
                モード
                <select class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.mode}>
                  <option value="mixed">クラス＆グループ混在</option>
                  <option value="class">クラス対抗のみ</option>
                  <option value="group">グループ対抗のみ</option>
                </select>
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-2">
                説明
                <textarea class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.description} rows="3" placeholder="概要やメモを入力"></textarea>
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700">
                勝利ポイント
                <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.win_points} />
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700">
                敗北ポイント
                <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.loss_points} />
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700">
                引き分けポイント
                <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.draw_points} />
              </label>
              <label class="flex flex-col text-sm font-medium text-gray-700">
                参加ポイント
                <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={sessionForm.participation_points} />
              </label>
              <label class="flex items-center space-x-2 text-sm font-medium text-gray-700">
                <input type="checkbox" bind:checked={sessionForm.allow_manual_points} />
                <span>手動加点を許可</span>
              </label>
            </div>
            <button class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
              on:click={saveSession}
              disabled={savingSession}>
              {savingSession ? '保存中...' : (session ? 'セッションを更新' : 'セッションを作成')}
            </button>
          </div>

          <!-- グループ管理 -->
          {#if session}
          <div class="border rounded-lg p-4 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold text-gray-800 border-b pb-2 flex-1">グループ管理</h3>
              <button class="px-3 py-1 border rounded text-sm text-gray-600 hover:bg-gray-100" on:click={resetGroupForm}>
                フォームをリセット
              </button>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="space-y-3">
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  グループ名
                  <input class="mt-1 border rounded px-3 py-2" bind:value={groupForm.name} placeholder="例: 1年Aコース" />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  説明
                  <textarea class="mt-1 border rounded px-3 py-2" rows="3" bind:value={groupForm.description}></textarea>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  所属クラス（複数選択可）
                  <select multiple size="6" class="mt-1 border rounded px-3 py-2" bind:value={groupForm.class_ids}>
                    {#each classes as cls}
                      <option value={cls.id}>{cls.name}</option>
                    {/each}
                  </select>
                </label>
                <button class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                  on:click={submitGroup}
                  disabled={savingGroup}>
                  {groupForm.id ? (savingGroup ? '更新中...' : 'グループを更新') : (savingGroup ? '登録中...' : 'グループを登録')}
                </button>
              </div>
              <div class="space-y-4">
                {#if groups.length === 0}
                  <p class="text-gray-500">登録済みグループはありません。</p>
                {:else}
                  <ul class="space-y-3">
                    {#each groups as group}
                      <li class="border rounded px-3 py-2">
                        <div class="flex justify-between items-center">
                          <div>
                            <p class="font-semibold text-gray-800">{group.name}</p>
                            <p class="text-xs text-gray-500">{group.description}</p>
                          </div>
                          <div class="space-x-2">
                            <button class="px-3 py-1 text-sm border rounded hover:bg-gray-100" on:click={() => startEditGroup(group)}>編集</button>
                            <button class="px-3 py-1 text-sm border rounded text-red-600 hover:bg-red-50" on:click={() => deleteGroup(group.id)}>削除</button>
                          </div>
                        </div>
                        <p class="text-sm text-gray-600 mt-2">
                          メンバー: {group.members?.map(m => m.class?.name ?? `クラスID ${m.class_id}`).join('、') || '未設定'}
                        </p>
                      </li>
                    {/each}
                  </ul>
                {/if}
              </div>
            </div>
          </div>
          {:else}
          <div class="border rounded-lg p-4 space-y-4">
            <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">グループ管理</h3>
            <p class="text-gray-500">セッションを作成すると、グループ管理が利用できます。</p>
          </div>
          {/if}

          <!-- 試合設定 -->
          {#if session}
          <div class="border rounded-lg p-4 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold text-gray-800 border-b pb-2 flex-1">試合設定</h3>
              <button class="px-3 py-1 border rounded text-sm text-gray-600 hover:bg-gray-100" on:click={resetMatchForm}>
                フォームをリセット
              </button>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="space-y-3">
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  試合タイトル
                  <input class="mt-1 border rounded px-3 py-2" bind:value={matchForm.title} placeholder="例: 1年コース対抗リレー" />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  試合日時
                  <input type="datetime-local" class="mt-1 border rounded px-3 py-2" bind:value={matchForm.scheduled_at} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  場所
                  <input class="mt-1 border rounded px-3 py-2" bind:value={matchForm.location} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  形式
                  <input class="mt-1 border rounded px-3 py-2" bind:value={matchForm.format} placeholder="例: 総当たり、トーナメントなど" />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  メモ
                  <textarea class="mt-1 border rounded px-3 py-2" rows="3" bind:value={matchForm.memo}></textarea>
                </label>
                <div class="space-y-3 border rounded p-3">
                  <div class="flex items-center justify-between">
                    <p class="text-sm font-semibold text-gray-700">参加者一覧</p>
                    <button class="px-3 py-1 border rounded text-sm text-gray-600 hover:bg-gray-100" on:click={addParticipant}>
                      参加者を追加
                    </button>
                  </div>
                  {#if matchForm.participants.length === 0}
                    <p class="text-sm text-gray-500">参加者が登録されていません。</p>
                  {:else}
                    <div class="space-y-3">
                      {#each matchForm.participants as participant, index}
                        <div class="border rounded px-3 py-3 space-y-3 bg-white">
                          <div class="flex items-center justify-between">
                            <span class="text-sm font-semibold text-gray-700">参加者 {index + 1}</span>
                            <div class="space-x-2">
                              <button class="px-2 py-1 text-xs border rounded hover:bg-gray-100 disabled:opacity-40"
                                on:click={() => moveParticipant(index, -1)}
                                disabled={index === 0}>
                                上へ
                              </button>
                              <button class="px-2 py-1 text-xs border rounded hover:bg-gray-100 disabled:opacity-40"
                                on:click={() => moveParticipant(index, 1)}
                                disabled={index === matchForm.participants.length - 1}>
                                下へ
                              </button>
                              <button class="px-2 py-1 text-xs border rounded text-red-600 hover:bg-red-50"
                                on:click={() => removeParticipant(index)}>
                                削除
                              </button>
                            </div>
                          </div>
                          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <div class="space-y-1">
                              <label for={`match-form-participant-${index}-type`} class="text-xs font-semibold text-gray-600">種別</label>
                              <select
                                id={`match-form-participant-${index}-type`}
                                class="border rounded px-2 py-1 w-full"
                                value={participant.type}
                                on:change={(e) => setParticipantType(index, e.target.value)}
                              >
                                <option value="class">クラス</option>
                                <option value="group">グループ</option>
                              </select>
                            </div>
                            <div class="space-y-1">
                              <label for={`match-form-participant-${index}-${participant.type === 'group' ? 'group' : 'class'}`} class="text-xs font-semibold text-gray-600">
                                {participant.type === 'group' ? 'グループ' : 'クラス'}
                              </label>
                              {#if participant.type === 'group'}
                                <select
                                  id={`match-form-participant-${index}-group`}
                                  class="border rounded px-2 py-1 w-full"
                                  value={participant.group_id}
                                  on:change={(e) => updateParticipantField(index, 'group_id', e.target.value)}
                                >
                                  <option value="">選択</option>
                                  {#each groups as group}
                                    <option value={group.id}>{group.name}</option>
                                  {/each}
                                </select>
                              {:else}
                                <select
                                  id={`match-form-participant-${index}-class`}
                                  class="border rounded px-2 py-1 w-full"
                                  value={participant.class_id}
                                  on:change={(e) => updateParticipantField(index, 'class_id', e.target.value)}
                                >
                                  <option value="">選択</option>
                                  {#each classes as cls}
                                    <option value={cls.id}>{cls.name}</option>
                                  {/each}
                                </select>
                              {/if}
                            </div>
                            <div class="space-y-1">
                              <label for={`match-form-participant-${index}-display-name`} class="text-xs font-semibold text-gray-600">表示名（任意）</label>
                              <input
                                id={`match-form-participant-${index}-display-name`}
                                class="border rounded px-2 py-1 w-full"
                                value={participant.display_name}
                                on:input={(e) => updateParticipantField(index, 'display_name', e.target.value)}
                                placeholder="例: 1年Aチーム"
                              />
                            </div>
                          </div>
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  ステータス
                  <select class="mt-1 border rounded px-3 py-2" bind:value={matchForm.status}>
                    <option value="scheduled">予定</option>
                    <option value="in_progress">進行中</option>
                    <option value="completed">完了</option>
                    <option value="cancelled">中止</option>
                  </select>
                </label>
                <label class="flex items-center space-x-2 text-sm font-medium text-gray-700">
                  <input type="checkbox" bind:checked={matchForm.allow_draw} />
                  <span>引き分けを許可</span>
                </label>
                <button class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                  on:click={submitMatch}
                  disabled={savingMatch}>
                  {matchForm.id ? (savingMatch ? '更新中...' : '試合を更新') : (savingMatch ? '登録中...' : '試合を登録')}
                </button>
              </div>
              <div class="space-y-4">
                {#if matches.length === 0}
                  <p class="text-gray-500">登録済みの試合はありません。</p>
                {:else}
                  <div class="space-y-3 max-h-[32rem] overflow-y-auto pr-2">
                    {#each matches as match}
                      <div class="border rounded px-3 py-3 space-y-2 bg-gray-50">
                        <div class="flex justify-between items-start">
                          <div>
                            <p class="font-semibold text-gray-800">{match.title ?? `試合 #${match.id}`}</p>
                            <p class="text-xs text-gray-500">ステータス: {match.status}</p>
                            {#if match.scheduled_at}
                              <p class="text-xs text-gray-500">日時: {new Date(match.scheduled_at).toLocaleString()}</p>
                            {/if}
                          </div>
                          <div class="space-x-2">
                            <button class="px-3 py-1 text-sm border rounded hover:bg-white" on:click={() => startEditMatch(match)}>編集</button>
                            <button class="px-3 py-1 text-sm border rounded text-red-600 hover:bg-red-100" on:click={() => deleteMatch(match.id)}>削除</button>
                          </div>
                        </div>
                        {#if match.entries && match.entries.length > 0}
                          <ul class="text-sm text-gray-700 list-disc list-inside space-y-1">
                            {#each match.entries as entry}
                              <li>{entry.resolved_name}</li>
                            {/each}
                          </ul>
                        {:else}
                          <p class="text-sm text-gray-700">{match.home_display_name} vs {match.away_display_name}</p>
                        {/if}
                        {#if match.result}
                          <div class="text-xs text-green-600">
                            結果: {match.winner_display ?? '---'}
                          </div>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            </div>
          </div>
          {:else}
          <div class="border rounded-lg p-4 space-y-4">
            <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">試合設定</h3>
            <p class="text-gray-500">セッションを作成すると、試合設定が利用できます。</p>
          </div>
          {/if}

          <!-- ポイントサマリー -->
          {#if session}
          <div class="border rounded-lg p-4 space-y-4">
            <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">ポイントサマリー</h3>
            {#if pointsSummary.length === 0}
              <p class="text-gray-500">ポイントデータがまだありません。</p>
            {:else}
              <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200 text-sm">
                  <thead class="bg-gray-50">
                    <tr>
                      <th class="px-4 py-2 text-left font-semibold text-gray-600">クラス</th>
                      <th class="px-4 py-2 text-right font-semibold text-gray-600">昼競技ポイント</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white divide-y divide-gray-200">
                    {#each pointsSummary as item}
                      <tr>
                        <td class="px-4 py-2">{item.class_name}</td>
                        <td class="px-4 py-2 text-right">{item.points}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
          {:else}
          <div class="border rounded-lg p-4 space-y-4">
            <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">ポイントサマリー</h3>
            <p class="text-gray-500">セッションを作成すると、ポイントサマリーが表示されます。</p>
          </div>
          {/if}

          <!-- 手動加点 -->
          {#if session?.allow_manual_points}
            <div class="border rounded-lg p-4 space-y-4">
              <h3 class="text-lg font-semibold text-gray-800 border-b pb-2">手動加点</h3>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  クラス
                  <select class="mt-1 border rounded px-3 py-2" bind:value={manualPointForm.class_id}>
                    <option value="">クラスを選択</option>
                    {#each classes as cls}
                      <option value={cls.id}>{cls.name}</option>
                    {/each}
                  </select>
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700">
                  ポイント
                  <input type="number" class="mt-1 border rounded px-3 py-2" bind:value={manualPointForm.points} />
                </label>
                <label class="flex flex-col text-sm font-medium text-gray-700 md:col-span-1">
                  理由（任意）
                  <input class="mt-1 border rounded px-3 py-2" bind:value={manualPointForm.reason} />
                </label>
              </div>
              <button class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                on:click={submitManualPoint}
                disabled={savingManualPoint}>
                {savingManualPoint ? '登録中...' : '手動加点を登録'}
              </button>
            </div>
          {/if}
        </div>
      </section>
  {/if}
</div>

