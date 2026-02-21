import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;
import { json } from '@sveltejs/kit';

/** @type {import('./$types').RequestHandler} */
export async function GET({ cookies }) {
  try {
    const sessionToken = cookies.get('session_token');
    const headers = {
      'Content-Type': 'application/json'
    };

    if (sessionToken) {
      headers['cookie'] = `session_token=${sessionToken}`;
    }

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
}

/** @type {import('./$types').RequestHandler} */
export async function POST({ request, cookies }) {
  try {
    const sessionToken = cookies.get('session_token');
    const body = await request.json();

    const headers = {
      'Content-Type': 'application/json'
    };

    if (sessionToken) {
      headers['cookie'] = `session_token=${sessionToken}`;
    }

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
}

/** @type {import('./$types').RequestHandler} */
export async function DELETE({ request, cookies }) {
  try {
    const sessionToken = cookies.get('session_token');
    const body = await request.json();

    const headers = {
      'Content-Type': 'application/json'
    };

    if (sessionToken) {
      headers['cookie'] = `session_token=${sessionToken}`;
    }

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
}

