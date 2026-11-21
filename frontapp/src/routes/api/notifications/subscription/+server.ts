import * as env from '$app/env/private';
const BACKEND_URL = env.BACKEND_URL;
import { json } from '@sveltejs/kit';
import { createBackendSessionHeaders } from '$lib/server/backendSessionHeaders.js';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ cookies }) => {
  try {
    const headers = createBackendSessionHeaders(cookies, {
      'Content-Type': 'application/json'
    });

    const response = await fetch(`${BACKEND_URL}/api/notifications/subscription`, {
      method: 'GET',
      headers
    });

    if (!response.ok) {
      const errorText = await response.text();
      return json({ error: errorText }, { status: response.status });
    }

    const data = await response.json();
    return json(data);
  } catch (error) {
    console.error('[api] Failed to get subscription:', error);
    return json({ error: '購読情報の取得に失敗しました' }, { status: 500 });
  }
};

export const POST: RequestHandler = async ({ request, cookies }) => {
  try {
    const body = await request.json();

    const headers = createBackendSessionHeaders(cookies, {
      'Content-Type': 'application/json'
    });

    const response = await fetch(`${BACKEND_URL}/api/notifications/subscription`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body)
    });

    if (!response.ok) {
      const errorText = await response.text();
      return json({ error: errorText }, { status: response.status });
    }

    const data = await response.json();
    return json(data, { status: response.status });
  } catch (error) {
    console.error('[api] Failed to save subscription:', error);
    return json({ error: '購読情報の保存に失敗しました' }, { status: 500 });
  }
};

export const DELETE: RequestHandler = async ({ request, cookies }) => {
  try {
    const body = await request.json();

    const headers = createBackendSessionHeaders(cookies, {
      'Content-Type': 'application/json'
    });

    const response = await fetch(`${BACKEND_URL}/api/notifications/subscription`, {
      method: 'DELETE',
      headers,
      body: JSON.stringify(body)
    });

    if (!response.ok) {
      const errorText = await response.text();
      return json({ error: errorText }, { status: response.status });
    }

    const data = await response.json();
    return json(data);
  } catch (error) {
    console.error('[api] Failed to delete subscription:', error);
    return json({ error: '購読情報の削除に失敗しました' }, { status: 500 });
  }
};
