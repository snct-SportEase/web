import { env } from '$env/dynamic/private';
const BACKEND_URL = env.BACKEND_URL;

/** @type {import('./$types').PageServerLoad} */
export async function load({ fetch, request }) {
  try {
    const headers = {
      cookie: request.headers.get('cookie'),
    };
    const authHeader = request.headers.get('Authorization');
    if (authHeader) {
      headers.Authorization = authHeader;
    }

    const response = await fetch(`${BACKEND_URL}/api/root/whitelist`, {
      headers,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to fetch whitelist: ${response.status} ${errorText}`);
    }
    const whitelist = await response.json();
    return { whitelist };
  } catch (error) {
    console.error('Error loading whitelist:', error);
    return { whitelist: [], error: error.message };
  }
}
