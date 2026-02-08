import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;
import { json } from '@sveltejs/kit';

/** @type {import('./$types').RequestHandler} */
export async function GET({ request, cookies }) {
  try {
    const sessionToken = cookies.get('session_token');
    const headers = {
      'Content-Type': 'application/json'
    };

    if (sessionToken) {
      headers['cookie'] = `session_token=${sessionToken}`;
    }

    const response = await fetch(`${BACKEND_URL}/api/notifications/debug`, {
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
    console.error('[api] Failed to get debug info:', error);
    return json({ error: '診断情報の取得に失敗しました' }, { status: 500 });
  }
}

