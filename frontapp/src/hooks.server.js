import { BACKEND_URL } from '$env/static/private';
import { redirect } from '@sveltejs/kit';

/** @type {import('@sveltejs/kit').Handle} */
export async function handle({ event, resolve }) {
  const sessionToken = event.cookies.get('session_token');

  if (sessionToken) {
    try {
      const url = `${BACKEND_URL}/api/auth/user`;
      const response = await fetch(url, {
        headers: {
          'cookie': `session_token=${sessionToken}`, 
        },
      });


      const responseText = await response.text();

      if (response.ok) {
        try {
          event.locals.user = JSON.parse(responseText);
        } catch {
          event.locals.user = null;
        }
      } else {
        event.locals.user = null;
      }
    } catch {
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
