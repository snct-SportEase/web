import { fail, redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, fetch }) {
  // Fetch classes for the profile setup modal
  // This assumes a /api/classes endpoint exists on the backend
  try {
    const response = await fetch(`${BACKEND_URL}/api/classes`);
    if (response.ok) {
      const classes = await response.json();
      return { user: locals.user, classes };
    }
  } catch (e) {
    // Return empty array if classes can't be fetched
    return { user: locals.user, classes: [] };
  }

  return { user: locals.user, classes: [] };
}

/** @type {import('./$types').Actions} */
export const actions = {
  logout: async ({ fetch, locals }) => {
    const sessionCookie = locals.request.headers.get('cookie');
    await fetch(`${BACKEND_URL}/api/auth/logout`, {
        method: 'POST',
        headers: {
            'cookie': sessionCookie,
        },
    });

    // Clear the user from locals and redirect
    locals.user = null;
    throw redirect(302, '/');
  },
};