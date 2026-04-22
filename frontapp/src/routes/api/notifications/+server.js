import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';

const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').RequestHandler} */
export async function GET({ cookies, url }) {
  try {
    const sessionToken = cookies.get('session_token');
    const headers = {
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
}
