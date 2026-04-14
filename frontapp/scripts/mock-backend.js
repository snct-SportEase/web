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

let events = defaultEvents();

function sendJson(res, status, body) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(body));
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

  if (url.pathname === '/api/root/events/1/notify-survey' && req.method === 'POST') {
    sendJson(res, 200, { message: 'Notification sent' });
    return;
  }

  if (url.pathname === '/api/events/active' && req.method === 'GET') {
    const activeEvent = events.find((event) => event.status === 'active') ?? events[0] ?? null;
    sendJson(res, 200, activeEvent ? { id: activeEvent.id, name: activeEvent.name } : null);
    return;
  }

  sendJson(res, 404, {
    error: `Mock backend route not found: ${req.method} ${url.pathname}`
  });
}).listen(port, '127.0.0.1', () => {
  console.log(`Mock backend listening on http://127.0.0.1:${port}`);
});
