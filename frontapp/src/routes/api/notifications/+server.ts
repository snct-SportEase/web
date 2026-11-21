import * as env from '$app/env/private';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const BACKEND_URL = env.BACKEND_URL;

export const GET: RequestHandler = async ({ cookies, url }) => {
  try {
    const sessionToken = cookies.get('session_token');
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    };

    if (sessionToken) {
      headers.cookie = `session_token=${sessionToken}`;
    }

    const response = await fetch(`${BACKEND_URL}/api/notifications${url.search}`, {
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
    console.error('[api] Failed to get notifications:', error);
    return json({ error: '通知の取得に失敗しました' }, { status: 500 });
  }
};
