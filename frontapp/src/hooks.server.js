import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';

/** @type {import('@sveltejs/kit').Handle} */
export async function handle({ event, resolve }) {
  const sessionToken = event.cookies.get('session_token');

  if (sessionToken) {
    try {
      const url = `${BACKEND_URL}/api/auth/user`;
      console.log(`[HOOKS] Fetching user from: ${url}`);
      const response = await fetch(url, {
        headers: {
          'cookie': `session_token=${sessionToken}`, 
        },
      });

      console.log(`[HOOKS] Response status: ${response.status}`);
      console.log('[HOOKS] Response headers:', Object.fromEntries(response.headers.entries()));

      const responseText = await response.text();
      console.log(`[HOOKS] Response text: ${responseText}`);

      if (response.ok) {
        try {
          event.locals.user = JSON.parse(responseText);
        } catch (e) {
          console.error('[HOOKS] Failed to parse JSON:', e);
          event.locals.user = null;
        }
      } else {
        event.locals.user = null;
      }
    } catch (error) {
      console.error('Failed to fetch user:', error);
      event.locals.user = null;
    }
  } else {
    event.locals.user = null;
  }

  // Protect dashboard route
  if (event.url.pathname.startsWith('/dashboard')) {
    if (!event.locals.user) {
      throw redirect(302, '/');
    }
  }

  // Redirect from login page if already logged in
  if (event.url.pathname === '/') {
    if (event.locals.user) {
      throw redirect(302, '/dashboard');
    }
  }

  return resolve(event);
}
