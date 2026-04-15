#!/usr/bin/env node

import { createServer } from 'node:http';

const port = Number(process.env.MOCK_BACKEND_PORT ?? 8081);

const rootUser = {
  id: 'test-root-id',
  email: 'root@example.com',
  display_name: '管理者ユーザー',
  is_profile_complete: true,
  roles: [{ name: 'root' }]
};

const defaultEvents = () => ([
  {
    id: 1,
    name: '2025春季スポーツ大会',
    year: 2025,
    season: 'spring',
    start_date: '2025-04-01T00:00:00Z',
    end_date: '2025-04-02T00:00:00Z',
    status: 'upcoming',
    survey_url: 'https://example.com/survey',
    hide_scores: false
  }
]);

const defaultSports = () => ([
  { id: 1, name: 'バスケットボール' },
  { id: 2, name: 'バレーボール' }
]);

const defaultWhitelist = () => ([
  { id: 1, email: 'student1@sendai-nct.jp', role: 'student' },
  { id: 2, email: 'admin1@sendai-nct.jp', role: 'admin' }
]);

const defaultNotificationRequests = () => ([
  {
    id: 1,
    title: 'お知らせ配信依頼',
    body: '明日の集合時刻変更を通知したいです。',
    status: 'pending',
    target_text: '全学生',
    requester: {
      id: 'student-user-1',
      email: 'student1@sendai-nct.jp',
      display_name: '1A 代表'
    },
    messages: [
      {
        id: 1,
        message: '内容を確認お願いします。',
        created_at: '2025-04-01T09:30:00Z',
        sender: {
          id: 'student-user-1',
          email: 'student1@sendai-nct.jp',
          display_name: '1A 代表'
        }
      }
    ]
  }
]);

const defaultUsers = () => ([
  {
    id: 'user-1',
    email: 'student1@sendai-nct.jp',
    display_name: '山田太郎',
    class_id: 1,
    roles: [
      { id: 1, name: 'student' },
      { id: 2, name: '1A_rep' }
    ]
  },
  {
    id: 'user-2',
    email: 'admin1@sendai-nct.jp',
    display_name: '運営花子',
    class_id: 2,
    roles: [
      { id: 3, name: 'admin' }
    ]
  }
]);

const defaultDefaultGroups = () => ({
  year_relay: [
    { group_name: 'Aブロック', class_names: ['1A', '1B'] },
    { group_name: 'Bブロック', class_names: ['2A', '2B'] }
  ],
  course_relay: [
    { group_name: '機械系', class_names: ['1A'] },
    { group_name: '電気系', class_names: ['1B'] }
  ],
  tug_of_war: [
    { group_name: '赤組', class_names: ['1A'] },
    { group_name: '白組', class_names: ['1B'] }
  ]
});

function defaultTournaments() {
  return [
    {
      id: 't1',
      name: 'バスケットボール',
      data: {
        rounds: [{}],
        matches: [
          {
            id: 'm1',
            roundIndex: 0,
            order: 0,
            sides: [
              { contestantId: 'c1', teamId: 1 },
              { contestantId: 'c2', teamId: 2 }
            ]
          }
        ],
        contestants: {
          c1: { players: [{ title: '1A' }] },
          c2: { players: [{ title: '1B' }] }
        }
      }
    }
  ];
}

let events = defaultEvents();
let sports = defaultSports();
let eventSports = [
  {
    id: 1,
    event_id: 1,
    sport_id: 1,
    description: 'バスケットボールの説明',
    rules: 'ルール内容',
    location: '体育館',
    rules_type: 'markdown',
    rules_pdf_url: null,
    min_capacity: null,
    max_capacity: null
  },
  {
    id: 2,
    event_id: 1,
    sport_id: 2,
    description: 'バレーボールの説明',
    rules: 'ルール内容',
    location: '体育館',
    rules_type: 'markdown',
    rules_pdf_url: null,
    min_capacity: null,
    max_capacity: null
  }
];
let notifications = [
  {
    id: 1,
    title: '大会開催のお知らせ',
    body: '春季スポーツ大会を開催します。',
    type: 'general',
    target_roles: ['student'],
    created_at: '2025-04-01T09:00:00Z'
  }
];
let classes = [
  { id: 1, name: '1A', student_count: 40 },
  { id: 2, name: '1B', student_count: 38 }
];
let whitelist = defaultWhitelist();
let notificationRequests = defaultNotificationRequests();
let users = defaultUsers();
let defaultGroups = defaultDefaultGroups();
let tournamentsList = defaultTournaments();
let noonSession = null;
let noonGroups = [];
let noonMatches = [];
let noonPointsSummary = [];
let noonTemplateRuns = [];
let rainyModeSettings = [];
let classTeamMembers = [
  { id: 'class-user-1', email: 'student1@sendai-nct.jp', display_name: '山田太郎', class_id: 1 },
  { id: 'class-user-2', email: 'student2@sendai-nct.jp', display_name: '鈴木次郎', class_id: 1 },
  { id: 'class-user-3', email: 'student3@sendai-nct.jp', display_name: '佐藤花子', class_id: 2 }
];
let teamAssignments = [];

function sendJson(res, status, body) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(body));
}

function sendResponse(res, status, body, headers = {}) {
  res.writeHead(status, headers);
  res.end(body);
}

function readJson(req) {
  return new Promise((resolve, reject) => {
    let body = '';

    req.on('data', (chunk) => {
      body += chunk;
    });

    req.on('end', () => {
      if (!body) {
        resolve({});
        return;
      }

      try {
        resolve(JSON.parse(body));
      } catch (error) {
        reject(error);
      }
    });

    req.on('error', reject);
  });
}

function getSessionToken(req) {
  const cookieHeader = req.headers.cookie ?? '';
  const cookie = cookieHeader
    .split(';')
    .map((value) => value.trim())
    .find((value) => value.startsWith('session_token='));

  return cookie?.split('=')[1] ?? null;
}

createServer(async (req, res) => {
  const url = new URL(req.url ?? '/', `http://${req.headers.host}`);

  if (url.pathname === '/health') {
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/__reset' && req.method === 'POST') {
    events = defaultEvents();
    sports = defaultSports();
    eventSports = [];
    notifications = [
      {
        id: 1,
        title: '大会開催のお知らせ',
        body: '春季スポーツ大会を開催します。',
        type: 'general',
        target_roles: ['student'],
        created_at: '2025-04-01T09:00:00Z'
      }
    ];
    classes = [
      { id: 1, name: '1A', student_count: 40 },
      { id: 2, name: '1B', student_count: 38 }
    ];
    whitelist = defaultWhitelist();
    notificationRequests = defaultNotificationRequests();
    users = defaultUsers();
    defaultGroups = defaultDefaultGroups();
    tournamentsList = defaultTournaments();
    noonSession = null;
    noonGroups = [];
    noonMatches = [];
    noonPointsSummary = [];
    noonTemplateRuns = [];
    rainyModeSettings = [];
    classTeamMembers = [
      { id: 'class-user-1', email: 'student1@sendai-nct.jp', display_name: '山田太郎', class_id: 1 },
      { id: 'class-user-2', email: 'student2@sendai-nct.jp', display_name: '鈴木次郎', class_id: 1 },
      { id: 'class-user-3', email: 'student3@sendai-nct.jp', display_name: '佐藤花子', class_id: 2 }
    ];
    teamAssignments = [];
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/auth/user' && req.method === 'GET') {
    if (getSessionToken(req) === 'test-session-token') {
      sendJson(res, 200, rootUser);
      return;
    }

    sendJson(res, 401, { error: 'Unauthorized' });
    return;
  }

  if (url.pathname === '/api/auth/logout' && req.method === 'POST') {
    sendJson(res, 200, { message: 'Logged out' });
    return;
  }

  if (url.pathname === '/api/root/events' && req.method === 'GET') {
    sendJson(res, 200, events);
    return;
  }

  if (url.pathname === '/api/root/events' && req.method === 'POST') {
    const body = await readJson(req);
    const nextEvent = {
      ...body,
      id: 2,
      start_date: `${body.start_date}T00:00:00Z`,
      end_date: `${body.end_date}T00:00:00Z`
    };

    events = [...events, nextEvent];
    sendJson(res, 201, { message: 'Event created', event: nextEvent });
    return;
  }

  if (url.pathname === '/api/root/events/1' && req.method === 'PUT') {
    const body = await readJson(req);

    events = events.map((event) => {
      if (event.id !== 1) return event;

      return {
        ...event,
        ...body,
        start_date: `${body.start_date}T00:00:00Z`,
        end_date: `${body.end_date}T00:00:00Z`
      };
    });

    sendJson(res, 200, { message: 'Event updated' });
    return;
  }

  if (url.pathname === '/api/root/events/1/rainy-mode' && req.method === 'PUT') {
    const body = await readJson(req);
    events = events.map((event) => event.id === 1 ? { ...event, is_rainy_mode: !!body.is_rainy_mode } : event);
    sendJson(res, 200, { is_rainy_mode: !!body.is_rainy_mode });
    return;
  }

  if (url.pathname === '/api/root/events/1/notify-survey' && req.method === 'POST') {
    sendJson(res, 200, { message: 'Notification sent' });
    return;
  }

  if (url.pathname === '/api/root/events/1/import-survey-scores' && req.method === 'POST') {
    sendJson(res, 200, { imported_classes_count: classes.length });
    return;
  }

  if (url.pathname === '/api/root/events/1/export/csv' && req.method === 'GET') {
    sendResponse(res, 200, 'class,score\n1A,100\n1B,90\n', {
      'Content-Type': 'text/csv'
    });
    return;
  }

  if (url.pathname === '/api/root/db/export' && req.method === 'GET') {
    sendResponse(res, 200, '-- mock dump', {
      'Content-Type': 'application/sql',
      'Content-Disposition': 'attachment; filename="mock_dump.sql"'
    });
    return;
  }

  if (url.pathname === '/api/scores/class' && req.method === 'GET') {
    sendJson(res, 200, [
      {
        class_name: '1A',
        rank_overall: 1,
        total_points_overall: 120,
        total_points_current_event: 60
      },
      {
        class_name: '1B',
        rank_overall: 2,
        total_points_overall: 100,
        total_points_current_event: 50
      }
    ]);
    return;
  }

  if (url.pathname === '/api/events/active' && req.method === 'GET') {
    const activeEvent = events.find((event) => event.status === 'active') ?? events[0] ?? null;
    sendJson(
      res,
      200,
      activeEvent
        ? {
            event_id: activeEvent.id,
            event_name: activeEvent.name,
            id: activeEvent.id,
            name: activeEvent.name
          }
        : null
    );
    return;
  }

  if (url.pathname === '/api/root/sports' && req.method === 'GET') {
    sendJson(res, 200, sports);
    return;
  }

  if (url.pathname === '/api/admin/allsports' && req.method === 'GET') {
    sendJson(res, 200, sports);
    return;
  }

  if (url.pathname === '/api/root/sports' && req.method === 'POST') {
    const body = await readJson(req);
    const nextSport = {
      id: sports.length + 1,
      name: body.name
    };
    sports = [...sports, nextSport];
    sendJson(res, 201, nextSport);
    return;
  }

  if (url.pathname === '/api/events/1/sports' && req.method === 'GET') {
    sendJson(res, 200, eventSports);
    return;
  }

  if (url.pathname === '/api/admin/events/1/sports' && req.method === 'POST') {
    const body = await readJson(req);
    const nextEventSport = {
      id: eventSports.length + 1,
      event_id: 1,
      sport_id: body.sport_id,
      description: body.description ?? '',
      rules: body.rules ?? '',
      location: body.location ?? 'other',
      rules_type: body.rules_type ?? 'markdown',
      rules_pdf_url: null,
      min_capacity: null,
      max_capacity: null
    };

    eventSports = [...eventSports, nextEventSport];
    sendJson(res, 201, nextEventSport);
    return;
  }

  if (url.pathname === '/api/admin/class-team/managed-class' && req.method === 'GET') {
    sendJson(res, 200, classes);
    return;
  }

  const classMembersMatch = url.pathname.match(/^\/api\/admin\/class-team\/classes\/(\d+)\/members$/);
  if (classMembersMatch && req.method === 'GET') {
    const classId = Number(classMembersMatch[1]);
    sendJson(res, 200, classTeamMembers.filter((member) => member.class_id === classId));
    return;
  }

  const teamMembersMatch = url.pathname.match(/^\/api\/admin\/class-team\/sports\/(\d+)\/members$/);
  if (teamMembersMatch && req.method === 'GET') {
    const sportId = Number(teamMembersMatch[1]);
    const classId = Number(url.searchParams.get('class_id') ?? '1');
    const assignedIds = teamAssignments
      .filter((item) => item.sport_id === sportId && item.class_id === classId)
      .map((item) => item.user_id);

    sendJson(
      res,
      200,
      classTeamMembers.filter((member) => member.class_id === classId && assignedIds.includes(member.id))
    );
    return;
  }

  if (url.pathname === '/api/admin/class-team/assign-members' && req.method === 'POST') {
    const body = await readJson(req);
    const classId = Number(body.class_id ?? 1);
    const sportId = Number(body.sport_id);
    const userIds = Array.isArray(body.user_ids) ? body.user_ids : [];

    for (const userId of userIds) {
      const exists = teamAssignments.some(
        (item) => item.class_id === classId && item.sport_id === sportId && item.user_id === userId
      );
      if (!exists) {
        teamAssignments.push({ class_id: classId, sport_id: sportId, user_id: userId });
      }
    }

    sendJson(res, 200, { message: 'メンバーの割り当てが完了しました' });
    return;
  }

  if (url.pathname === '/api/admin/class-team/remove-member' && req.method === 'DELETE') {
    const body = await readJson(req);
    const classId = Number(body.class_id ?? 1);
    const sportId = Number(body.sport_id);
    const userId = body.user_id;

    teamAssignments = teamAssignments.filter(
      (item) => !(item.class_id === classId && item.sport_id === sportId && item.user_id === userId)
    );

    sendJson(res, 200, { message: 'メンバーを削除しました' });
    return;
  }

  if (url.pathname === '/api/admin/events/1/tournaments' && req.method === 'GET') {
    sendJson(res, 200, tournamentsList);
    return;
  }

  if (url.pathname === '/api/admin/statistics/attendance' && req.method === 'GET') {
    sendJson(res, 200, { attendance_rate: 84.2 });
    return;
  }

  if (url.pathname === '/api/admin/statistics/participation' && req.method === 'GET') {
    sendJson(res, 200, {
      バスケットボール: 91,
      バレーボール: 87
    });
    return;
  }

  if (url.pathname === '/api/admin/statistics/scores' && req.method === 'GET') {
    sendJson(res, 200, {
      '2025春季': [
        { class_name: '1A', total_points_current_event: 12 },
        { class_name: '1B', total_points_current_event: 10 }
      ],
      '2025秋季': [
        { class_name: '1A', total_points_current_event: 15 },
        { class_name: '1B', total_points_current_event: 13 }
      ]
    });
    return;
  }

  if (url.pathname === '/api/admin/statistics/progress' && req.method === 'GET') {
    sendJson(res, 200, {
      バスケットボール: '進行中',
      バレーボール: '準備中',
      ドッジボール: '結果集計中'
    });
    return;
  }

  const attendanceClassDetailsMatch = url.pathname.match(/^\/api\/admin\/attendance\/class-details\/(\d+)$/);
  if (attendanceClassDetailsMatch && req.method === 'GET') {
    const classId = Number(attendanceClassDetailsMatch[1]);
    const cls = classes.find(c => c.id === classId);
    if (cls) {
      sendJson(res, 200, {
        id: cls.id,
        name: cls.name,
        studentCount: cls.student_count,
        attendancePoints: 10 // Mock value
      });
    } else {
      sendJson(res, 404, { error: 'Class not found' });
    }
    return;
  }

  if (url.pathname === '/api/admin/attendance/register' && req.method === 'POST') {
    const body = await readJson(req);
    sendJson(res, 200, { message: '出席を正常に登録しました。' });
    return;
  }

  const classNameMatch = url.pathname.match(/^\/api\/class\/name\/([^/]+)$/);
  if (classNameMatch && req.method === 'GET') {
    const className = decodeURIComponent(classNameMatch[1]);
    const cls = classes.find(c => c.name === className);
    if (cls) {
      sendJson(res, 200, cls);
    } else {
      sendJson(res, 404, { error: 'Class not found' });
    }
    return;
  }

  if (url.pathname === '/api/admin/events/active' && req.method === 'GET') {
    const activeEvent = events.find((event) => event.status === 'active') ?? events[0] ?? null;
    sendJson(res, 200, activeEvent ?? {});
    return;
  }

  const sportDetailsMatch = url.pathname.match(/^\/api\/admin\/events\/(\d+)\/sports\/(\d+)\/details$/);
  if (sportDetailsMatch && req.method === 'GET') {
    const eventId = Number(sportDetailsMatch[1]);
    const sportId = Number(sportDetailsMatch[2]);
    const detail = eventSports.find((item) => item.event_id === eventId && item.sport_id === sportId);

    sendJson(res, 200, detail ?? {
      description: '',
      rules: '',
      rules_type: 'markdown',
      rules_pdf_url: null,
      min_capacity: null,
      max_capacity: null
    });
    return;
  }

  if (sportDetailsMatch && req.method === 'PUT') {
    const body = await readJson(req);
    // Mock successful update
    sendJson(res, 200, { message: 'Sport details saved successfully' });
    return;
  }

  const matchResultMatch = url.pathname.match(/^\/api\/admin\/matches\/([^/]+)\/result$/);
  if (matchResultMatch && req.method === 'PUT') {
    const body = await readJson(req);
    // 更新成功をシミュレート
    sendJson(res, 200, { message: '試合結果を更新しました' });
    return;
  }

  const sportTeamsMatch = url.pathname.match(/^\/api\/root\/sports\/(\d+)\/teams$/);
  if (sportTeamsMatch && req.method === 'GET') {
    const sportId = Number(sportTeamsMatch[1]);
    const teams = classes.map((cls) => ({
      id: sportId * 100 + cls.id,
      event_id: 1,
      sport_id: sportId,
      class_id: cls.id,
      min_capacity: null,
      max_capacity: null
    }));
    sendJson(res, 200, teams);
    return;
  }

  const rainyModeSettingsMatch = url.pathname.match(/^\/api\/root\/events\/(\d+)\/rainy-mode\/settings(?:\/(\d+)\/(\d+))?$/);
  if (rainyModeSettingsMatch && req.method === 'GET') {
    const eventId = Number(rainyModeSettingsMatch[1]);
    sendJson(res, 200, rainyModeSettings.filter((item) => item.event_id === eventId));
    return;
  }

  if (rainyModeSettingsMatch && (req.method === 'POST' || req.method === 'PUT')) {
    const eventId = Number(rainyModeSettingsMatch[1]);
    const body = await readJson(req);
    const nextSetting = {
      event_id: eventId,
      sport_id: Number(body.sport_id),
      class_id: Number(body.class_id),
      min_capacity: body.min_capacity ?? null,
      max_capacity: body.max_capacity ?? null,
      match_start_time: body.match_start_time ?? ''
    };

    rainyModeSettings = [
      ...rainyModeSettings.filter(
        (item) =>
          !(
            item.event_id === nextSetting.event_id &&
            item.sport_id === nextSetting.sport_id &&
            item.class_id === nextSetting.class_id
          )
      ),
      nextSetting
    ];

    sendJson(res, 200, nextSetting);
    return;
  }

  if (url.pathname === '/api/root/notifications/roles' && req.method === 'GET') {
    sendJson(res, 200, {
      roles: [
        { id: 1, name: 'student' },
        { id: 2, name: 'admin' },
        { id: 3, name: 'root' }
      ]
    });
    return;
  }

  if (url.pathname === '/api/notifications' && req.method === 'GET') {
    sendJson(res, 200, { notifications });
    return;
  }

  if (url.pathname === '/api/root/notifications' && req.method === 'POST') {
    const body = await readJson(req);
    const nextNotification = {
      id: notifications.length + 1,
      ...body,
      created_at: '2025-04-02T10:00:00Z'
    };
    notifications = [nextNotification, ...notifications];
    sendJson(res, 201, { notification: nextNotification });
    return;
  }

  if (url.pathname === '/api/root/notification-requests' && req.method === 'GET') {
    sendJson(res, 200, {
      requests: notificationRequests.map(({ messages, ...request }) => request)
    });
    return;
  }

  const notificationRequestMatch = url.pathname.match(/^\/api\/root\/notification-requests\/(\d+)$/);
  if (notificationRequestMatch && req.method === 'GET') {
    const id = Number(notificationRequestMatch[1]);
    const request = notificationRequests.find((item) => item.id === id) ?? null;
    sendJson(res, request ? 200 : 404, request ? { request } : { error: 'Request not found' });
    return;
  }

  const notificationMessageMatch = url.pathname.match(/^\/api\/root\/notification-requests\/(\d+)\/messages$/);
  if (notificationMessageMatch && req.method === 'POST') {
    const id = Number(notificationMessageMatch[1]);
    const body = await readJson(req);
    notificationRequests = notificationRequests.map((item) => {
      if (item.id !== id) return item;
      const nextMessage = {
        id: item.messages.length + 1,
        message: body.message,
        created_at: '2025-04-02T10:00:00Z',
        sender: rootUser
      };
      return { ...item, messages: [...item.messages, nextMessage] };
    });
    sendJson(res, 201, { ok: true });
    return;
  }

  const notificationDecisionMatch = url.pathname.match(/^\/api\/root\/notification-requests\/(\d+)\/decision$/);
  if (notificationDecisionMatch && req.method === 'POST') {
    const id = Number(notificationDecisionMatch[1]);
    const body = await readJson(req);
    notificationRequests = notificationRequests.map((item) => item.id === id ? { ...item, status: body.status } : item);
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/classes' && req.method === 'GET') {
    sendJson(res, 200, classes);
    return;
  }

  if (url.pathname === '/api/root/whitelist' && req.method === 'GET') {
    sendJson(res, 200, whitelist);
    return;
  }

  if (url.pathname === '/api/root/whitelist' && req.method === 'POST') {
    const body = await readJson(req);
    const nextEntry = {
      id: whitelist.length + 1,
      email: body.email,
      role: body.role
    };
    whitelist = [...whitelist, nextEntry];
    sendJson(res, 201, nextEntry);
    return;
  }

  if (url.pathname === '/api/root/whitelist' && req.method === 'DELETE') {
    const body = await readJson(req);
    whitelist = whitelist.filter((entry) => entry.email !== body.email);
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/root/whitelist/bulk' && req.method === 'DELETE') {
    const body = await readJson(req);
    whitelist = whitelist.filter((entry) => !body.emails.includes(entry.email));
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/root/whitelist/csv' && req.method === 'POST') {
    whitelist = [
      ...whitelist,
      { id: whitelist.length + 1, email: 'csv-imported@sendai-nct.jp', role: 'student' }
    ];
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/root/classes/student-counts' && req.method === 'PUT') {
    const body = await readJson(req);
    classes = classes.map((cls) => {
      const updated = body.find((item) => item.class_id === cls.id);
      return updated ? { ...cls, student_count: updated.student_count } : cls;
    });
    sendJson(res, 200, { message: 'updated' });
    return;
  }

  if (url.pathname === '/api/root/classes/student-counts/csv' && req.method === 'POST') {
    classes = classes.map((cls, index) => ({ ...cls, student_count: 45 - index }));
    sendJson(res, 200, { message: 'csv updated' });
    return;
  }

  if (url.pathname === '/api/root/mic/class' && req.method === 'GET') {
    sendJson(res, 200, {
      class_name: '1A',
      total_points: 120,
      season: 'spring'
    });
    return;
  }

  if (url.pathname === '/api/admin/pdfs' && req.method === 'POST') {
    sendJson(res, 200, { url: 'https://example.com/guidelines.pdf' });
    return;
  }

  if (url.pathname === '/api/root/users' && req.method === 'GET') {
    const query = (url.searchParams.get('query') ?? '').toLowerCase();
    const searchType = url.searchParams.get('searchType') ?? '';
    let filtered = users;
    if (query) {
      filtered = users.filter((user) => {
        if (searchType === 'display_name') {
          return (user.display_name ?? '').toLowerCase().includes(query);
        }
        return user.email.toLowerCase().includes(query);
      });
    }
    sendJson(res, 200, filtered);
    return;
  }

  if (url.pathname === '/api/root/users/display-name' && req.method === 'PUT') {
    const body = await readJson(req);
    users = users.map((user) => user.id === body.user_id ? { ...user, display_name: body.display_name } : user);
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/admin/users/role' && req.method === 'PUT') {
    const body = await readJson(req);
    users = users.map((user) => {
      if (user.id !== body.user_id) return user;
      if (user.roles.some((role) => role.name === body.role)) return user;
      return {
        ...user,
        roles: [...user.roles, { id: Date.now(), name: body.role }]
      };
    });
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/admin/users/role' && req.method === 'DELETE') {
    const body = await readJson(req);
    users = users.map((user) => user.id === body.user_id ? {
      ...user,
      roles: user.roles.filter((role) => role.name !== body.role)
    } : user);
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/root/events/1/tournaments' && req.method === 'GET') {
    sendJson(res, 200, tournamentsList);
    return;
  }

  if (url.pathname === '/api/root/events/1/tournaments/generate-preview' && req.method === 'POST') {
    sendJson(res, 200, []);
    return;
  }

  if (url.pathname === '/api/root/events/1/tournaments/bulk-create' && req.method === 'POST') {
    sendJson(res, 200, { message: 'saved' });
    return;
  }

  if (url.pathname === '/api/root/events/1/noon-game/session' && req.method === 'GET') {
    sendJson(res, 200, {
      session: noonSession,
      classes,
      groups: noonGroups,
      matches: noonMatches,
      points_summary: noonPointsSummary,
      template_runs: noonTemplateRuns
    });
    return;
  }

  if (url.pathname === '/api/root/events/1/noon-game/session' && req.method === 'POST') {
    const body = await readJson(req);
    noonSession = {
      id: 1,
      name: body.name,
      description: body.description,
      mode: body.mode,
      win_points: body.win_points,
      loss_points: body.loss_points,
      draw_points: body.draw_points,
      participation_points: body.participation_points,
      allow_manual_points: body.allow_manual_points
    };
    sendJson(res, 200, {
      session: noonSession,
      classes,
      groups: noonGroups,
      matches: noonMatches,
      points_summary: noonPointsSummary,
      template_runs: noonTemplateRuns
    });
    return;
  }

  const defaultGroupsMatch = url.pathname.match(/^\/api\/root\/noon-game\/templates\/([^/]+)\/default-groups$/);
  if (defaultGroupsMatch && req.method === 'GET') {
    sendJson(res, 200, {
      groups: defaultGroups[defaultGroupsMatch[1]] ?? []
    });
    return;
  }

  if (defaultGroupsMatch && req.method === 'PUT') {
    const body = await readJson(req);
    defaultGroups = {
      ...defaultGroups,
      [defaultGroupsMatch[1]]: (body.groups ?? []).map((group) => ({
        group_name: group.group_name,
        class_names: group.class_names ?? []
      }))
    };
    sendJson(res, 200, { ok: true });
    return;
  }

  const rootTemplateRunMatch = url.pathname.match(/^\/api\/root\/events\/1\/noon-game\/templates\/([^/]+)\/run$/);
  if (rootTemplateRunMatch && req.method === 'POST') {
    const body = await readJson(req);
    noonSession = {
      id: 1,
      ...body.session
    };
    noonTemplateRuns = [{ id: 1, template_key: rootTemplateRunMatch[1].replace(/-/g, '_') }];
    sendJson(res, 200, { ok: true });
    return;
  }

  const adminTemplateRunMatch = url.pathname.match(/^\/api\/admin\/events\/1\/noon-game\/templates\/([^/]+)\/run$/);
  if (adminTemplateRunMatch && req.method === 'POST') {
    const body = await readJson(req);
    noonSession = {
      id: 1,
      ...body.session
    };
    noonTemplateRuns = [{ id: 1, template_key: adminTemplateRunMatch[1].replace(/-/g, '_') }];
    sendJson(res, 200, { ok: true });
    return;
  }

  if (url.pathname === '/api/root/events/1/competition-guidelines' && req.method === 'PUT') {
    const body = await readJson(req);
    events = events.map((event) => event.id === 1 ? { ...event, competition_guidelines_pdf_url: body.pdf_url } : event);
    sendJson(res, 200, { message: 'updated' });
    return;
  }

  sendJson(res, 404, {
    error: `Mock backend route not found: ${req.method} ${url.pathname}`
  });
}).listen(port, '127.0.0.1', () => {
  console.log(`Mock backend listening on http://127.0.0.1:${port}`);
});
