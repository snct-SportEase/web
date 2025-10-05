import { fail, redirect } from '@sveltejs/kit';
import { BACKEND_URL } from '$env/static/private';

/** @type {import('./$types').PageServerLoad} */
export async function load({ locals, fetch }) {
  const returnData = { user: locals.user, classes: [], events: [] };

  try {
    // Fetch classes
    const classesResponse = await fetch(`${BACKEND_URL}/api/classes`);
    if (classesResponse.ok) {
      returnData.classes = await classesResponse.json();
    }
  } catch (e) {
    console.error('Failed to fetch classes:', e);
  }

  // Fetch events if user is root
  const isRoot = locals.user?.roles?.some(role => role.name === 'root');
  if (isRoot && locals.user?.is_profile_complete) {
      try {
        const eventResponse = await fetch(`${BACKEND_URL}/api/root/events`, {
            headers: {
                'cookie': locals.request.headers.get('cookie'),
            }
        });
        if (eventResponse.ok) {
            returnData.events = await eventResponse.json();
        }
      } catch (e) {
        console.error('Failed to fetch events:', e);
      }
  }

  return returnData;
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